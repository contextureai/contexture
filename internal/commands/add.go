// Package commands provides CLI command implementations
package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"github.com/contextureai/contexture/internal/dependencies"
	"github.com/contextureai/contexture/internal/domain"
	"github.com/contextureai/contexture/internal/format"
	"github.com/contextureai/contexture/internal/git"
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
	ruleFetcher := rule.NewFetcher(deps.FS, git.NewRepository(deps.FS), rule.FetcherConfig{})
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
		),
		registry: registry,
	}
}

// Execute runs the add command
func (c *AddCommand) Execute(ctx context.Context, cmd *cli.Command, ruleIDs []string) error {
	// Parse custom data if provided
	var customData map[string]any
	if dataStr := cmd.String("data"); dataStr != "" {
		if err := json.Unmarshal([]byte(dataStr), &customData); err != nil {
			return fmt.Errorf("invalid JSON in --data parameter: %w", err)
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
			return fmt.Errorf("invalid --var parameter '%s': %w", varFlag, err)
		}
		customData[key] = value
		log.Debug("Parsed var flag", "key", key, "value", value)
	}

	// Parse --source and --ref flags for constructing rule IDs
	sourceFlag := cmd.String("source")
	refFlag := cmd.String("ref")
	log.Debug("Parsed source and ref flags", "source", sourceFlag, "ref", refFlag)

	// Get current directory and load configuration
	currentDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	configResult, err := c.projectManager.LoadConfigWithLocalRules(currentDir)
	if err != nil {
		return fmt.Errorf("no Contexture project found in '%s' or parent directories.\n\nOriginal error: %w", currentDir, err)
	}
	config := configResult.Config
	configPath := configResult.Path

	// Parse and validate rule IDs with progress indicators
	type ruleRefWithOriginal struct {
		ruleRef     domain.RuleRef
		originalID  string
		defaultVars map[string]any
	}
	var validRuleRefs []ruleRefWithOriginal

	err = ui.WithProgressTiming("Validated rules", func() error {
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
				return fmt.Errorf("invalid rule ID '%s'.\n\nRule IDs should be in one of these formats:\n  - [contexture:path/to/rule]           (from default repository)\n  - [contexture(source):path/to/rule]   (from custom source)\n  - path/to/rule                        (shorthand for default registry)\n  - path/to/rule --source URL           (with custom source flag)\n\nExamples:\n  - [contexture:languages/go/testing]\n  - languages/go/testing\n  - test/lemon --source https://github.com/user/repo.git\n\nOriginal error: %w", ruleID, err)
			}

			// Convert simple format to full format for storage (without variables)
			var fullRuleID string
			if !strings.HasPrefix(processedRuleID, "[contexture") {
				// This is a simple format, convert to full format
				fullRuleID = fmt.Sprintf("[contexture:%s]", processedRuleID)
			} else {
				// Extract the rule ID without variables for storage
				if strings.Contains(processedRuleID, "]{") {
					// Remove variables part from the rule ID for storage
					if bracketIdx := strings.Index(processedRuleID, "]{"); bracketIdx != -1 {
						fullRuleID = processedRuleID[:bracketIdx] + "]"
					}
				} else {
					fullRuleID = processedRuleID
				}
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
				return fmt.Errorf("failed to fetch rule '%s'.\n\nOriginal error: %w", processedRuleID, err)
			}

			// Validate rule
			validationResult := c.ruleValidator.ValidateRule(fetchedRule)
			if !validationResult.Valid {
				var errorMessages []string
				for _, err := range validationResult.Errors {
					errorMessages = append(errorMessages, err.Error())
				}
				return fmt.Errorf("rule '%s' failed validation\n\nvalidation errors:\n  - %s\n\nthis usually indicates:\n  - the rule file is malformed or incomplete\n  - required fields are missing\n  - invalid YAML syntax\n\nplease check the rule source or report this issue to the rule maintainer",
					ruleID, strings.Join(errorMessages, "\n  - "))
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
				Source:     parsedID.Source,
				Ref:        parsedID.Ref,
				Variables:  variables, // Include merged variables
				CommitHash: commitHash,
			}

			validRuleRefs = append(validRuleRefs, ruleRefWithOriginal{
				ruleRef:     ruleRef,
				originalID:  ruleID,
				defaultVars: fetchedRule.DefaultVariables,
			})
		}
		return nil
	})
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
			return fmt.Errorf("failed to add rule '%s' to project configuration.\n\nOriginal error: %w", ruleRefWithOrig.ruleRef.ID, err)
		}
	}

	// Get the appropriate config location for the project
	location := c.projectManager.GetConfigLocation(currentDir, false)
	err = c.projectManager.SaveConfig(config, location, currentDir)
	if err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	// Auto-generate rules after adding them
	if err := c.generateRules(ctx, config, currentDir); err != nil {
		log.Warn("Failed to auto-generate rules", "error", err)
		fmt.Println("Rules added but generation failed. Run 'contexture build' manually.")
	}

	// Success message
	theme := ui.DefaultTheme()
	successStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(theme.Success)

	fmt.Println()
	fmt.Println(successStyle.Render("Rules added successfully!"))

	for _, ruleRefWithOrig := range validRuleRefs {
		// Extract simple rule ID for display (remove [contexture:] wrapper if present)
		displayRuleID := ruleRefWithOrig.originalID
		var variables map[string]any
		var parsed *domain.ParsedRuleID

		if strings.HasPrefix(ruleRefWithOrig.originalID, "[contexture:") {
			// Parse to extract just the path component
			var err error
			parsed, err = c.ruleFetcher.ParseRuleID(ruleRefWithOrig.originalID)
			if err == nil && parsed.RulePath != "" {
				displayRuleID = parsed.RulePath
				variables = parsed.Variables
			}
		} else {
			// Parse the full rule ID from ruleRef to get source info
			var err error
			parsed, err = c.ruleFetcher.ParseRuleID(ruleRefWithOrig.ruleRef.ID)
			if err == nil && parsed.RulePath != "" {
				displayRuleID = parsed.RulePath
				variables = parsed.Variables
			}
		}

		fmt.Printf("  %s\n", displayRuleID)

		// Show source information for custom source rules (like in remove command)
		if parsed != nil && parsed.Source != "" && domain.IsCustomGitSource(parsed.Source) {
			darkGrayStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#666666"))
			sourceDisplay := domain.FormatSourceForDisplay(parsed.Source, parsed.Ref)
			fmt.Printf("    %s\n", darkGrayStyle.Render(sourceDisplay))
		}

		// Show variables only if they differ from defaults
		if rule.ShouldDisplayVariables(variables, ruleRefWithOrig.defaultVars) {
			if variablesJSON, err := json.Marshal(variables); err == nil {
				fmt.Printf("    Variables: %s\n", string(variablesJSON))
			}
		}
	}

	log.Debug("Rules added",
		"count", len(validRuleRefs),
		"config_path", configPath)

	return nil
}

// generateRules automatically generates output after adding rules
func (c *AddCommand) generateRules(
	ctx context.Context,
	config *domain.Project,
	_ string,
) error {
	if len(config.Rules) == 0 {
		return nil // No rules to generate
	}

	// Get target formats (all enabled formats)
	targetFormats := config.GetEnabledFormats()
	if len(targetFormats) == 0 {
		return fmt.Errorf("no enabled formats found for rule generation.\n\nTo fix this:\n  1. Run 'contexture init' to configure formats\n  2. Or manually add formats to your .contexture.yaml:\n     formats:\n       - name: claude\n         enabled: true\n\nAvailable formats: claude, cursor, windsurf")
	}

	log.Debug("Auto-generating rules", "rules", len(config.Rules), "formats", len(targetFormats))

	// Use shared rule generator with consistent UI styling
	return c.ruleGenerator.GenerateRules(ctx, config, targetFormats)
}

// fetchLatestCommitHash fetches the latest commit hash for a specific rule file
func (c *AddCommand) fetchLatestCommitHash(
	ctx context.Context,
	parsedID *domain.ParsedRuleID,
) (string, error) {
	// Clone the repository to a temporary directory
	tempDir, cleanup, err := c.cloneRepositoryToTemp(ctx, parsedID.Source, parsedID.Ref)
	if err != nil {
		return "", fmt.Errorf("failed to clone repository: %w", err)
	}
	defer cleanup()

	// Get the rule file path within the repository
	ruleFilePath := parsedID.RulePath + ".md"

	// Create git repository instance for the cloned directory
	gitRepo := git.NewRepository(afero.NewOsFs())

	// Get the latest commit information for this specific file
	commitInfo, err := gitRepo.GetFileCommitInfo(tempDir, ruleFilePath, parsedID.Ref)
	if err != nil {
		return "", fmt.Errorf("failed to get file commit info: %w", err)
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
		return "", nil, fmt.Errorf("failed to create temp directory: %w", err)
	}

	// Cleanup function
	cleanup := func() {
		if err := afero.NewOsFs().RemoveAll(tempDir); err != nil {
			log.Warn("Failed to cleanup temporary directory", "path", tempDir, "error", err)
		}
	}

	// Create git repository instance
	gitRepo := git.NewRepository(afero.NewOsFs())

	// Clone repository with the specified branch
	err = gitRepo.Clone(ctx, repoURL, tempDir, git.WithBranch(branch))
	if err != nil {
		cleanup() // Clean up on error
		return "", nil, fmt.Errorf("failed to clone repository: %w", err)
	}

	return tempDir, cleanup, nil
}

// AddAction is the CLI action handler for the add command
func AddAction(ctx context.Context, cmd *cli.Command, deps *dependencies.Dependencies) error {
	ruleIDs := cmd.Args().Slice()
	addCmd := NewAddCommand(deps)

	// If no rule IDs provided, show helpful error message
	if len(ruleIDs) == 0 {
		return fmt.Errorf("no rule IDs provided\n\nUsage:\n  contexture rules add [rule-id...]\n\nExamples:\n  # Add specific rules (simple format)\n  contexture rules add languages/go/code-organization testing/unit-tests\n  \n  # Add rules (full format)\n  contexture rules add \"[contexture:languages/go/advanced-patterns]\" \"[contexture:security/input-validation]\"\n  \n  # Add rule with variables\n  contexture rules add languages/go/testing --var threshold=90\n  \n  # Add from custom source\n  contexture rules add my/custom-rule --source \"https://github.com/my-org/rules.git\"\n\nTo browse available rules:\n  1. Check the repository at https://github.com/contextureai/rules\n  2. Use 'contexture rules list' to see currently installed rules\n  \nRun 'contexture rules add --help' for more options")
	}

	return addCmd.Execute(ctx, cmd, ruleIDs)
}

// parseVarFlag parses a single --var flag in the format "key=value"
// The value can be a simple string or JSON for complex values
func parseVarFlag(varFlag string) (string, any, error) {
	parts := strings.SplitN(varFlag, "=", 2)
	if len(parts) != 2 {
		return "", nil, fmt.Errorf("format should be 'key=value', got: %s", varFlag)
	}

	key := parts[0]
	valueStr := parts[1]

	if key == "" {
		return "", nil, fmt.Errorf("key cannot be empty")
	}

	// Try to parse as JSON first (for complex values)
	var value any
	if err := json.Unmarshal([]byte(valueStr), &value); err != nil {
		// If JSON parsing fails, treat it as a simple string value
		value = valueStr
	}

	return key, value, nil
}
