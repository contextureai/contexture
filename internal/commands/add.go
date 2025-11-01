// Package commands provides CLI command implementations
package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"github.com/contextureai/contexture/internal/dependencies"
	"github.com/contextureai/contexture/internal/domain"
	contextureerrors "github.com/contextureai/contexture/internal/errors"
	"github.com/contextureai/contexture/internal/format"
	"github.com/contextureai/contexture/internal/git"
	"github.com/contextureai/contexture/internal/output"
	"github.com/contextureai/contexture/internal/project"
	"github.com/contextureai/contexture/internal/rule"
	"github.com/contextureai/contexture/internal/ui"
	"github.com/spf13/afero"
	"github.com/urfave/cli/v3"
)

// AddCommand implements the add command
type AddCommand struct {
	projectManager *project.Manager
	ruleFetcher    rule.Fetcher
	ruleValidator  rule.Validator
	ruleGenerator  *RuleGenerator
	registry       *format.Registry
}

// NewAddCommand creates a new add command
func NewAddCommand(deps *dependencies.Dependencies) *AddCommand {
	registry := format.GetDefaultRegistry(deps.FS)

	// Create provider registry
	providerRegistry := deps.ProviderRegistry

	ruleFetcher := rule.NewFetcher(deps.FS, newOpenRepository(deps.FS), rule.FetcherConfig{}, providerRegistry)
	ruleValidator := rule.NewValidator()

	return &AddCommand{
		projectManager: project.NewManager(deps.FS),
		ruleFetcher:    ruleFetcher,
		ruleValidator:  ruleValidator,
		ruleGenerator: NewRuleGenerator(
			ruleFetcher,
			ruleValidator,
			rule.NewProcessor(),
			registry,
			deps.FS,
		),
		registry: registry,
	}
}

// ExecuteWithDeps runs the add command with explicit dependencies
func (c *AddCommand) ExecuteWithDeps(ctx context.Context, cmd *cli.Command, ruleIDs []string, deps *dependencies.Dependencies) error {
	// Check if JSON output mode - if so, suppress all terminal output
	outputFormat := output.Format(cmd.String("output"))
	isJSONMode := outputFormat == output.FormatJSON

	if !isJSONMode {
		// Show header like list command
		headerStyle := lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.AdaptiveColor{Light: "#F793FF", Dark: "#AD58B4"})
		fmt.Printf("%s\n\n", headerStyle.Render("Add Rule"))
	}

	// Get provider registry from deps
	providerRegistry := deps.ProviderRegistry

	// Parse custom data if provided
	var customData map[string]any
	if dataStr := cmd.String("data"); dataStr != "" {
		if err := json.Unmarshal([]byte(dataStr), &customData); err != nil {
			return contextureerrors.Wrap(err, "parse data")
		}
		log.Debug("Parsed custom data", "data", customData)
	}

	// Parse --var flags and merge with custom data
	if customData == nil {
		customData = make(map[string]any)
	}

	varFlags := cmd.StringSlice("var")
	for _, varFlag := range varFlags {
		key, value, err := parseVarFlag(varFlag)
		if err != nil {
			return contextureerrors.Wrap(err, "parse var")
		}
		customData[key] = value
		log.Debug("Parsed var flag", "key", key, "value", value)
	}

	// Parse --source and --ref flags for constructing rule IDs
	sourceFlag := cmd.String("source")
	refFlag := cmd.String("ref")
	log.Debug("Parsed source and ref flags", "source", sourceFlag, "ref", refFlag)

	// Check if global flag is set
	isGlobal := cmd.Bool("global")

	// Load configuration
	config, configPath, err := loadConfigByScope(c.projectManager, isGlobal)
	if err != nil {
		return err
	}

	var currentDir string
	if !isGlobal {
		currentDir, err = os.Getwd()
		if err != nil {
			return contextureerrors.Wrap(err, "get current directory")
		}
	}

	// Warn if adding global rules when Cursor is enabled with UserRulesProject mode
	// Only check if we're in a project context (not just adding to global config)
	if isGlobal && !isJSONMode && currentDir == "" {
		// Try to load project config to check for Cursor
		cwd, cwdErr := os.Getwd()
		if cwdErr == nil {
			if projectConfigResult, loadErr := c.projectManager.LoadConfig(cwd); loadErr == nil {
				for _, formatConfig := range projectConfigResult.Config.Formats {
					if formatConfig.Enabled && formatConfig.Type == domain.FormatCursor {
						mode := formatConfig.GetEffectiveUserRulesMode()
						if mode == domain.UserRulesProject {
							theme := ui.DefaultTheme()
							mutedStyle := lipgloss.NewStyle().Foreground(theme.Muted)
							fmt.Printf("%s %s\n\n",
								mutedStyle.Render("⚠"),
								mutedStyle.Render("Cursor does not support native global rules. Your global rules will be merged into project files, which may cause conflicts in team environments. Consider setting Cursor's userRulesMode to 'disabled' in .contexture.yaml"))
							break
						}
					}
				}
			}
		}
	}

	// Load providers from config into registry
	if err := providerRegistry.LoadFromProject(config); err != nil {
		return contextureerrors.Wrap(err, "load providers")
	}

	// Parse and validate rule IDs with progress indicators
	type ruleRefWithOriginal struct {
		ruleRef     domain.RuleRef
		originalID  string
		defaultVars map[string]any
	}
	var validRuleRefs []ruleRefWithOriginal

	// Validate rules
	validateFunc := func() error {
		for _, ruleID := range ruleIDs {
			// Construct proper rule ID format if --source flag is provided
			processedRuleID := ruleID
			if sourceFlag != "" {
				// If this is a simple rule ID (not already in [contexture:...] format),
				// construct the proper format using the --source and optional --ref flags
				if !strings.HasPrefix(ruleID, "[contexture") {
					if refFlag != "" {
						processedRuleID = fmt.Sprintf("[contexture(%s):%s,%s]", sourceFlag, ruleID, refFlag)
					} else {
						processedRuleID = fmt.Sprintf("[contexture(%s):%s]", sourceFlag, ruleID)
					}
					log.Debug("Constructed rule ID from flags", "original", ruleID, "constructed", processedRuleID)
				}
			}

			// Parse rule ID
			parsedID, err := c.ruleFetcher.ParseRuleID(processedRuleID)
			if err != nil {
				return contextureerrors.Wrap(err, "parse rule ID")
			}

			// Convert simple format to full format for storage (without variables)
			var fullRuleID string
			switch {
			case strings.HasPrefix(processedRuleID, "[contexture"):
				// Already in full format - extract without variables
				if strings.Contains(processedRuleID, "]{") {
					// Remove variables part from the rule ID for storage
					if bracketIdx := strings.Index(processedRuleID, "]{"); bracketIdx != -1 {
						fullRuleID = processedRuleID[:bracketIdx] + "]"
					}
				} else {
					fullRuleID = processedRuleID
				}
			case strings.HasPrefix(processedRuleID, "@"):
				// Provider format @provider/path - store without variables
				if braceIdx := strings.Index(processedRuleID, "{"); braceIdx != -1 {
					fullRuleID = processedRuleID[:braceIdx]
				} else {
					fullRuleID = processedRuleID
				}
			default:
				// Simple format - convert to full format
				fullRuleID = fmt.Sprintf("[contexture:%s]", processedRuleID)
			}

			// Check if rule already exists (check both formats)
			if c.projectManager.HasRule(config, fullRuleID) ||
				c.projectManager.HasRule(config, ruleID) {
				if !cmd.Bool("force") {
					fmt.Printf("  Rule already exists, skipping: %s\n", ruleID)
					continue
				}
				fmt.Printf("  Rule already exists, updating: %s\n", ruleID)
			}

			// Fetch and validate rule using the original ID - force remote fetching for add command
			// Try to use source-aware fetching if available to ensure we fetch from remote repository
			var fetchedRule *domain.Rule

			type sourceAwareFetcher interface {
				FetchRuleWithSource(ctx context.Context, ruleID, source string) (*domain.Rule, error)
			}

			if compositeFetcher, ok := c.ruleFetcher.(sourceAwareFetcher); ok {
				// Use the source-aware method to force remote fetching (empty source = default/remote)
				fetchedRule, err = compositeFetcher.FetchRuleWithSource(ctx, processedRuleID, "")
			} else {
				// Fallback to regular fetch
				fetchedRule, err = c.ruleFetcher.FetchRule(ctx, processedRuleID)
			}
			if err != nil {
				return contextureerrors.Wrap(err, "fetch rule")
			}

			// Validate rule
			validationResult := c.ruleValidator.ValidateRule(fetchedRule)
			if !validationResult.Valid {
				var errorMessages []string
				for _, err := range validationResult.Errors {
					errorMessages = append(errorMessages, err.Error())
				}
				return contextureerrors.ValidationErrorf("rule", "validation failed: %s", strings.Join(errorMessages, "; "))
			}

			// Create rule reference with merged variables, storing the full format
			mergedVariables := make(map[string]any)

			// Start with variables from parsed rule ID
			if parsedID.Variables != nil {
				for key, value := range parsedID.Variables {
					mergedVariables[key] = value
				}
			}

			// Merge with custom data (custom data takes precedence)
			for key, value := range customData {
				mergedVariables[key] = value
			}

			// Only set Variables if we have any
			var variables map[string]any
			if len(mergedVariables) > 0 {
				variables = mergedVariables
			}

			// Fetch the latest commit hash for this rule
			commitHash, err := c.fetchLatestCommitHash(ctx, parsedID)
			if err != nil {
				log.Warn("Failed to fetch commit hash for rule", "rule", ruleID, "error", err)
				// Continue without commit hash rather than failing
			}

			ruleRef := domain.RuleRef{
				ID:         fullRuleID,
				Variables:  variables, // Include merged variables
				CommitHash: commitHash,
			}

			// Only set Source and Ref for non-provider rules
			// Provider syntax rules (@provider/path) don't need Source/Ref since the provider contains that info
			if !strings.HasPrefix(fullRuleID, "@") {
				ruleRef.Source = parsedID.Source
				ruleRef.Ref = parsedID.Ref
			}

			validRuleRefs = append(validRuleRefs, ruleRefWithOriginal{
				ruleRef:     ruleRef,
				originalID:  ruleID,
				defaultVars: fetchedRule.DefaultVariables,
			})
		}
		return nil
	}

	// Execute validation with or without progress indicator
	if isJSONMode {
		err = validateFunc()
	} else {
		err = ui.WithProgress("Validated rules", validateFunc)
	}
	if err != nil {
		return err
	}

	if len(validRuleRefs) == 0 {
		log.Info("No new rules to add")
		return nil
	}

	// Add rules to configuration
	for _, ruleRefWithOrig := range validRuleRefs {
		err := c.projectManager.AddRule(config, ruleRefWithOrig.ruleRef)
		if err != nil {
			return contextureerrors.Wrap(err, "add rule")
		}
	}

	// Save configuration to appropriate location
	if isGlobal {
		err = c.projectManager.SaveGlobalConfig(config)
		if err != nil {
			return contextureerrors.Wrap(err, "save global config")
		}
	} else {
		// Get the appropriate config location for the project
		location := c.projectManager.GetConfigLocation(currentDir, false)
		err = c.projectManager.SaveConfig(config, location, currentDir)
		if err != nil {
			return contextureerrors.Wrap(err, "save config")
		}
	}

	// Auto-generate rules after adding them (skip in JSON mode)
	if !isJSONMode {
		if isGlobal {
			// For global rules, rebuild both global locations and project if in project context
			if err := c.rebuildAfterGlobalAdd(ctx); err != nil {
				log.Warn("Failed to auto-generate rules", "error", err)
				fmt.Println("Rules added but generation failed. Run 'contexture build' manually.")
				return nil // Exit early - rules were added but generation failed
			}
		} else {
			// For project rules, use merged config (global + project)
			if err := c.generateRulesWithMergedConfig(ctx, currentDir); err != nil {
				log.Warn("Failed to auto-generate rules", "error", err)
				fmt.Println("Rules added but generation failed. Run 'contexture build' manually.")
				return nil // Exit early - rules were added but generation failed
			}
		}
	}

	// Handle output format
	outputManager, err := output.NewManager(outputFormat)
	if err != nil {
		return contextureerrors.Wrap(err, "create output manager")
	}

	// Collect added rule IDs for output
	var addedRuleIDs []string
	for _, ruleRefWithOrig := range validRuleRefs {
		// Use the stored rule ID (which preserves @provider/path format)
		displayRuleID := domain.ExtractRulePath(ruleRefWithOrig.ruleRef.ID)
		addedRuleIDs = append(addedRuleIDs, displayRuleID)
	}

	// Write output using the appropriate format
	metadata := output.AddMetadata{
		RulesAdded: addedRuleIDs,
	}

	err = outputManager.WriteRulesAdd(metadata)
	if err != nil {
		return contextureerrors.Wrap(err, "write output")
	}

	// For default format, also display the detailed information
	if outputFormat == output.FormatDefault {
		theme := ui.DefaultTheme()
		successStyle := lipgloss.NewStyle().
			Bold(true).
			Foreground(theme.Success)

		fmt.Println()
		fmt.Println(successStyle.Render("Rules added successfully!"))

		for _, ruleRefWithOrig := range validRuleRefs {
			// Use the stored rule ID (which preserves @provider/path format)
			displayRuleID := domain.ExtractRulePath(ruleRefWithOrig.ruleRef.ID)

			// Parse to get variables
			var variables map[string]any
			if parsed, err := c.ruleFetcher.ParseRuleID(ruleRefWithOrig.originalID); err == nil {
				variables = parsed.Variables
			}

			fmt.Printf("  %s\n", displayRuleID)

			// Show source information for custom source rules (but not provider syntax)
			// Provider syntax rules start with @ and shouldn't show the underlying git URL
			if !strings.HasPrefix(ruleRefWithOrig.ruleRef.ID, "@") {
				if parsed, err := c.ruleFetcher.ParseRuleID(ruleRefWithOrig.ruleRef.ID); err == nil {
					if parsed.Source != "" && domain.IsCustomGitSource(parsed.Source) {
						darkGrayStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#666666"))
						sourceDisplay := domain.FormatSourceForDisplay(parsed.Source, parsed.Ref)
						fmt.Printf("    %s\n", darkGrayStyle.Render(sourceDisplay))
					}
				}
			}

			// Show variables only if they differ from defaults
			if rule.ShouldDisplayVariables(variables, ruleRefWithOrig.defaultVars) {
				if variablesJSON, err := json.Marshal(variables); err == nil {
					fmt.Printf("    Variables: %s\n", string(variablesJSON))
				}
			}
		}
	}

	log.Debug("Rules added",
		"count", len(validRuleRefs),
		"config_path", configPath)

	return nil
}

// Execute runs the add command
func (c *AddCommand) Execute(ctx context.Context, cmd *cli.Command, ruleIDs []string) error {
	// This is a compatibility wrapper that creates dependencies
	// In practice, this won't be called as we use AddAction which has access to deps
	deps := dependencies.New(ctx)
	return c.ExecuteWithDeps(ctx, cmd, ruleIDs, deps)
}

// rebuildAfterGlobalAdd rebuilds outputs after adding a global rule
// 1. Generates to native user rules locations (e.g., ~/.claude/CLAUDE.md)
// 2. If in a project, also rebuilds project files based on UserRulesMode
func (c *AddCommand) rebuildAfterGlobalAdd(ctx context.Context) error {
	// Load global config with local rules to get all global rules
	globalConfigResult, err := c.projectManager.LoadGlobalConfigWithLocalRules()
	if err != nil {
		return contextureerrors.Wrap(err, "load global config with local rules")
	}

	globalConfig := globalConfigResult.Config

	// Get formats from global config that have native user rules support
	var userFormats []domain.FormatConfig
	for _, formatConfig := range globalConfig.Formats {
		if !formatConfig.Enabled {
			continue
		}

		caps, exists := c.registry.GetCapabilities(formatConfig.Type)
		if !exists {
			continue
		}

		// Only include formats that support native user rules
		if caps.SupportsUserRules && caps.UserRulesPath != "" {
			// Create format config for user rules generation
			userFormatConfig := formatConfig
			userDir := filepath.Dir(caps.UserRulesPath)
			userFormatConfig.BaseDir = userDir
			userFormatConfig.IsUserRules = true
			userFormats = append(userFormats, userFormatConfig)
		}
	}
	log.Debug("Found formats with native user rules support", "count", len(userFormats))

	// Generate to native user rules locations if we have formats that support it
	if len(userFormats) > 0 {
		userConfig := &domain.Project{}
		*userConfig = *globalConfig
		// Use only global rules for user locations
		userConfig.Rules = globalConfig.Rules

		log.Debug("Auto-generating global rules to user locations", "formats", len(userFormats), "rules", len(userConfig.Rules))
		if err := c.ruleGenerator.GenerateRulesWithScope(ctx, userConfig, userFormats, "global"); err != nil {
			log.Error("Failed to generate global rules to user locations", "error", err)
			return contextureerrors.Wrap(err, "generate global rules to user locations")
		}
		log.Debug("Successfully generated global rules to user locations")
	}

	// Check if we're in a project context
	currentDir, err := os.Getwd()
	if err != nil {
		// Can't determine current directory, skip project rebuild
		log.Debug("Cannot determine current directory, skipping project rebuild", "error", err)
		return nil
	}

	// Try to load project config with merged rules
	merged, projectErr := c.projectManager.LoadConfigMergedWithLocalRules(currentDir)
	if projectErr != nil {
		// Not in a project, this is normal - global rules were added successfully
		//nolint:nilerr // Intentionally returning nil - not being in a project is not an error
		return nil
	}

	// We're in a project - rebuild project files based on UserRulesMode per format
	log.Debug("Detected project context, rebuilding based on UserRulesMode")

	// Separate user (global) rules from project rules
	var projectRules, userRules []domain.RuleRef
	for _, rws := range merged.MergedRules {
		if rws.Source == domain.RuleSourceUser {
			userRules = append(userRules, rws.RuleRef)
		} else {
			projectRules = append(projectRules, rws.RuleRef)
		}
	}

	// Get enabled formats from project config
	targetFormats := merged.Project.GetEnabledFormats()
	if len(targetFormats) == 0 {
		return nil
	}

	// Group formats by which rules they need based on UserRulesMode
	var nativeOnlyFormats []domain.FormatConfig // UserRulesNative - project rules only
	var mergedFormats []domain.FormatConfig     // UserRulesProject - project + user rules
	var disabledFormats []domain.FormatConfig   // UserRulesDisabled - project rules only

	for _, formatConfig := range targetFormats {
		mode := formatConfig.GetEffectiveUserRulesMode()
		switch mode {
		case domain.UserRulesNative:
			nativeOnlyFormats = append(nativeOnlyFormats, formatConfig)
		case domain.UserRulesProject:
			mergedFormats = append(mergedFormats, formatConfig)
		case domain.UserRulesDisabled:
			disabledFormats = append(disabledFormats, formatConfig)
		}
	}

	// Generate for formats with UserRulesNative or UserRulesDisabled (project rules only)
	projectOnlyFormats := make([]domain.FormatConfig, 0, len(nativeOnlyFormats)+len(disabledFormats))
	projectOnlyFormats = append(projectOnlyFormats, nativeOnlyFormats...)
	projectOnlyFormats = append(projectOnlyFormats, disabledFormats...)
	if len(projectOnlyFormats) > 0 && len(projectRules) > 0 {
		config := &domain.Project{}
		*config = *merged.Project
		config.Rules = projectRules

		if err := c.ruleGenerator.GenerateRulesWithScope(ctx, config, projectOnlyFormats, "project"); err != nil {
			return err
		}
	}

	// Generate for formats with UserRulesProject (project + user rules)
	if len(mergedFormats) > 0 && (len(projectRules) > 0 || len(userRules) > 0) {
		config := &domain.Project{}
		*config = *merged.Project
		config.Rules = append(append([]domain.RuleRef{}, projectRules...), userRules...)

		if err := c.ruleGenerator.GenerateRulesWithScope(ctx, config, mergedFormats, "project"); err != nil {
			return err
		}
	}

	return nil
}

// generateRulesWithMergedConfig generates rules using merged global + project + local config
// respecting UserRulesMode to determine which rules go where
func (c *AddCommand) generateRulesWithMergedConfig(ctx context.Context, currentDir string) error {
	// Load merged config with local rules
	merged, err := c.projectManager.LoadConfigMergedWithLocalRules(currentDir)
	if err != nil {
		return err
	}

	// Separate user (global) rules from project rules
	var projectRules, userRules []domain.RuleRef
	for _, rws := range merged.MergedRules {
		if rws.Source == domain.RuleSourceUser {
			userRules = append(userRules, rws.RuleRef)
		} else {
			projectRules = append(projectRules, rws.RuleRef)
		}
	}

	if len(projectRules) == 0 && len(userRules) == 0 {
		return nil
	}

	// Get enabled formats from project config
	targetFormats := merged.Project.GetEnabledFormats()
	if len(targetFormats) == 0 {
		return nil
	}

	// Group formats by which rules they need based on UserRulesMode
	var nativeOnlyFormats []domain.FormatConfig // UserRulesNative/Disabled - project rules only
	var mergedFormats []domain.FormatConfig     // UserRulesProject - project + user rules

	for _, formatConfig := range targetFormats {
		mode := formatConfig.GetEffectiveUserRulesMode()
		switch mode {
		case domain.UserRulesNative, domain.UserRulesDisabled:
			nativeOnlyFormats = append(nativeOnlyFormats, formatConfig)
		case domain.UserRulesProject:
			mergedFormats = append(mergedFormats, formatConfig)
		}
	}

	// Generate for formats with UserRulesNative or UserRulesDisabled (project rules only)
	if len(nativeOnlyFormats) > 0 && len(projectRules) > 0 {
		config := &domain.Project{}
		*config = *merged.Project
		config.Rules = projectRules

		if err := c.ruleGenerator.GenerateRulesWithScope(ctx, config, nativeOnlyFormats, "project"); err != nil {
			return err
		}
	}

	// Generate for formats with UserRulesProject (project + user rules)
	if len(mergedFormats) > 0 && (len(projectRules) > 0 || len(userRules) > 0) {
		config := &domain.Project{}
		*config = *merged.Project
		config.Rules = append(append([]domain.RuleRef{}, projectRules...), userRules...)

		if err := c.ruleGenerator.GenerateRulesWithScope(ctx, config, mergedFormats, "project"); err != nil {
			return err
		}
	}

	log.Debug("Auto-generating rules with merged config",
		"project_rules", len(projectRules),
		"user_rules", len(userRules),
		"formats", len(targetFormats))

	return nil
}

// fetchLatestCommitHash fetches the latest commit hash for a specific rule file
func (c *AddCommand) fetchLatestCommitHash(
	ctx context.Context,
	parsedID *domain.ParsedRuleID,
) (string, error) {
	// Clone the repository to a temporary directory
	tempDir, cleanup, err := c.cloneRepositoryToTemp(ctx, parsedID.Source, parsedID.Ref)
	if err != nil {
		return "", contextureerrors.Wrap(err, "clone repository")
	}
	defer cleanup()

	// Get the rule file path within the repository
	ruleFilePath := parsedID.RulePath + ".md"

	// Create git repository instance for the cloned directory
	gitRepo := newOpenRepository(afero.NewOsFs())

	// Get the latest commit information for this specific file
	commitInfo, err := gitRepo.GetFileCommitInfo(tempDir, ruleFilePath, parsedID.Ref)
	if err != nil {
		return "", contextureerrors.Wrap(err, "get file commit info")
	}

	return commitInfo.Hash, nil
}

// cloneRepositoryToTemp clones a repository to a temporary directory (similar to update command)
func (c *AddCommand) cloneRepositoryToTemp(
	ctx context.Context,
	repoURL, branch string,
) (string, func(), error) {
	// Create temporary directory
	tempDir, err := afero.TempDir(afero.NewOsFs(), "", "contexture-add-*")
	if err != nil {
		return "", nil, contextureerrors.Wrap(err, "create temp directory")
	}

	// Cleanup function
	cleanup := func() {
		if err := afero.NewOsFs().RemoveAll(tempDir); err != nil {
			log.Warn("Failed to cleanup temporary directory", "path", tempDir, "error", err)
		}
	}

	// Create git repository instance
	gitRepo := newOpenRepository(afero.NewOsFs())

	// Clone repository with the specified branch
	err = gitRepo.Clone(ctx, repoURL, tempDir, git.WithBranch(branch))
	if err != nil {
		cleanup() // Clean up on error
		return "", nil, contextureerrors.Wrap(err, "clone repository")
	}

	return tempDir, cleanup, nil
}

// AddAction is the CLI action handler for the add command
func AddAction(ctx context.Context, cmd *cli.Command, deps *dependencies.Dependencies) error {
	ruleIDs := cmd.Args().Slice()
	addCmd := NewAddCommand(deps)

	// If no rule IDs provided, show helpful error message
	if len(ruleIDs) == 0 {
		return contextureerrors.ValidationErrorf("rule-id", "no rule IDs provided")
	}

	return addCmd.ExecuteWithDeps(ctx, cmd, ruleIDs, deps)
}

// parseVarFlag parses a single --var flag in the format "key=value"
// The value can be a simple string or JSON for complex values
func parseVarFlag(varFlag string) (string, any, error) {
	parts := strings.SplitN(varFlag, "=", 2)
	if len(parts) != 2 {
		return "", nil, contextureerrors.ValidationErrorf("var", "format should be 'key=value'")
	}

	key := parts[0]
	valueStr := parts[1]

	if key == "" {
		return "", nil, contextureerrors.ValidationErrorf("var", "key cannot be empty")
	}

	// Try to parse as JSON first (for complex values)
	var value any
	if err := json.Unmarshal([]byte(valueStr), &value); err != nil {
		// If JSON parsing fails, treat it as a simple string value
		value = valueStr
	}

	return key, value, nil
}
