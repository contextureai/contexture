// Package commands provides CLI command implementations
package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"github.com/contextureai/contexture/internal/dependencies"
	"github.com/contextureai/contexture/internal/domain"
	"github.com/contextureai/contexture/internal/format"
	"github.com/contextureai/contexture/internal/git"
	"github.com/contextureai/contexture/internal/project"
	"github.com/contextureai/contexture/internal/rule"
	"github.com/contextureai/contexture/internal/tui"
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
	// Show command header (this is non-interactive mode with rule IDs provided)
	fmt.Println(ui.CommandHeader("add"))
	fmt.Println()

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
		ruleRef    domain.RuleRef
		originalID string
	}
	var validRuleRefs []ruleRefWithOriginal

	err = ui.WithProgressTiming("Fetched rules", func() error {
		for _, ruleID := range ruleIDs {
			// Parse rule ID
			parsedID, err := c.ruleFetcher.ParseRuleID(ruleID)
			if err != nil {
				return fmt.Errorf("invalid rule ID '%s'.\n\nRule IDs should be in one of these formats:\n  - [contexture:path/to/rule]           (from default repository)\n  - [contexture(source):path/to/rule]   (from custom source)\n  - path/to/rule                        (shorthand for default registry)\n\nExamples:\n  - [contexture:languages/go/testing]\n  - languages/go/testing\n\nOriginal error: %w", ruleID, err)
			}

			// Convert simple format to full format for storage
			fullRuleID := ruleID
			if !strings.HasPrefix(ruleID, "[contexture") {
				// This is a simple format, convert to full format
				fullRuleID = fmt.Sprintf("[contexture:%s]", ruleID)
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
				fetchedRule, err = compositeFetcher.FetchRuleWithSource(ctx, ruleID, "")
			} else {
				// Fallback to regular fetch
				fetchedRule, err = c.ruleFetcher.FetchRule(ctx, ruleID)
			}
			if err != nil {
				return fmt.Errorf("failed to fetch rule '%s'.\n\nOriginal error: %w", ruleID, err)
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
				ruleRef:    ruleRef,
				originalID: ruleID,
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
		fmt.Printf("  %s\n", ruleRefWithOrig.originalID)
	}

	log.Debug("Rules added",
		"count", len(validRuleRefs),
		"config_path", configPath)

	return nil
}

// ShowAvailableRules lists available rules from remote repositories for adding
func (c *AddCommand) ShowAvailableRules(ctx context.Context, cmd *cli.Command) error {
	// Get source repository and branch
	defaultRepo := domain.DefaultRepository
	defaultBranch := domain.DefaultBranch

	// Fetch available rules with spinner (no completion message)
	spinner := ui.NewBubblesSpinner("Fetching available rules")
	fmt.Print(spinner.View())

	rules, err := c.ruleFetcher.ListAvailableRules(ctx, defaultRepo, defaultBranch)
	spinner.Stop("") // Stop without message
	if err != nil {
		log.Error("Failed to fetch available rules", "error", err)
		fmt.Printf("Failed to fetch available rules: %v\n", err)
		return nil
	}

	if len(rules) == 0 {
		fmt.Println("\nNo rules found in the repository.")
		return nil
	}

	// No filters available since --search flag was removed

	// Sort rules for consistent output
	sort.Strings(rules)

	// Always show interactive mode since no flags are available for non-interactive filtering
	return c.showInteractiveRulesForAdding(ctx, cmd, rules, defaultRepo, defaultBranch)
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

// showInteractiveRulesForAdding shows an interactive searchable list of rules for adding
func (c *AddCommand) showInteractiveRulesForAdding(
	ctx context.Context,
	cmd *cli.Command,
	rules []string,
	_, _ string,
) error {
	// Check if we're in a project
	currentDir, _ := os.Getwd()

	var config *domain.Project
	var hasValidConfig bool
	configResult, err := c.projectManager.LoadConfigWithLocalRules(currentDir)
	if err == nil && configResult != nil {
		hasValidConfig = true
		config = configResult.Config
	}

	if !hasValidConfig {
		fmt.Println("No project configuration found. Please run 'contexture init' first.")
		return nil
	}

	// Filter out already added rules
	var availableRules []string
	for _, ruleID := range rules {
		// Convert simple format to full format for comparison
		fullRuleID := fmt.Sprintf("[contexture:%s]", ruleID)

		// Check if rule already exists (check both formats)
		if !c.projectManager.HasRule(config, fullRuleID) &&
			!c.projectManager.HasRule(config, ruleID) {
			availableRules = append(availableRules, ruleID)
		}
	}

	if len(availableRules) == 0 {
		fmt.Println("All available rules have already been added to your project.")
		return nil
	}

	rules = availableRules

	// Fetch detailed rule information with spinner
	detailSpinner := ui.NewBubblesSpinner("Loading rule details")
	fmt.Print(detailSpinner.View())

	var detailedRules []*domain.Rule
	for _, ruleID := range rules {
		// For rules from ListAvailableRules, we know they're remote rules from the default repository
		// Force them to be fetched via the GitFetcher
		var rule *domain.Rule
		var fetchErr error

		// Try to use source-aware fetching if available
		type sourceAwareFetcher interface {
			FetchRuleWithSource(ctx context.Context, ruleID, source string) (*domain.Rule, error)
		}

		if compositeFetcher, ok := c.ruleFetcher.(sourceAwareFetcher); ok {
			// Use the source-aware method to force remote fetching (empty source = default/remote)
			rule, fetchErr = compositeFetcher.FetchRuleWithSource(ctx, ruleID, "")
		} else {
			// Fallback to regular fetch
			rule, fetchErr = c.ruleFetcher.FetchRule(ctx, ruleID)
		}

		if fetchErr != nil {
			log.Warn("Failed to fetch rule details", "rule", ruleID, "error", fetchErr)
			// Continue with other rules even if one fails
			continue
		}
		detailedRules = append(detailedRules, rule)
	}
	detailSpinner.Stop("") // Stop without message

	if len(detailedRules) == 0 {
		fmt.Println("No rule details could be loaded.")
		return nil
	}

	// Show interactive list for rule selection
	filteredRules, err := c.showRuleListSelection(detailedRules)
	if err != nil {
		return err
	}

	if len(filteredRules) == 0 {
		log.Info("No rules selected")
		return nil
	}

	// Process selected rules using existing add logic
	return c.Execute(ctx, cmd, filteredRules)
}

// showRuleListSelection shows an interactive bubbles list for rule selection using file browser
func (c *AddCommand) showRuleListSelection(rules []*domain.Rule) ([]string, error) {
	return showInteractiveRuleBrowser(rules, "Select Rules to Add")
}

// showInteractiveRuleBrowser shows an interactive file browser for rule selection
func showInteractiveRuleBrowser(rules []*domain.Rule, title string) ([]string, error) {
	// Extract rule paths from rules for building the tree
	var rulePaths []string
	for _, rule := range rules {
		rulePath := domain.ExtractRulePath(rule.ID)
		if rulePath == "" {
			rulePath = rule.FilePath
		}
		if rulePath != "" {
			rulePaths = append(rulePaths, rulePath)
		}
	}

	if len(rulePaths) == 0 {
		return nil, fmt.Errorf("no valid rules found for selection")
	}

	// Build the rule tree
	ruleTree := domain.NewRuleTree(rulePaths)

	// Use the file browser
	browser := tui.NewFileBrowser()
	return browser.BrowseRules(ruleTree, rules, title)
}

// showInteractiveRuleSelection shows an interactive searchable list of rules for selection (shared function)
func showInteractiveRuleSelection(rules []*domain.Rule, title string) ([]string, error) {
	selector := tui.NewRuleSelector()
	return selector.SelectRules(rules, title)
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

	// If no rule IDs provided, show available rules to add
	if len(ruleIDs) == 0 {
		return addCmd.ShowAvailableRules(ctx, cmd)
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
