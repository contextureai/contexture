// Package commands provides CLI command implementations
package commands

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"github.com/contextureai/contexture/internal/cache"
	"github.com/contextureai/contexture/internal/dependencies"
	"github.com/contextureai/contexture/internal/domain"
	"github.com/contextureai/contexture/internal/output"
	"github.com/contextureai/contexture/internal/project"
	"github.com/contextureai/contexture/internal/rule"
	"github.com/contextureai/contexture/internal/tui"
	"github.com/contextureai/contexture/internal/ui"
	"github.com/spf13/afero"
	"github.com/urfave/cli/v3"
)

// UpdateCommand implements the update command
type UpdateCommand struct {
	projectManager *project.Manager
	ruleFetcher    rule.Fetcher
	ruleValidator  rule.Validator
	cache          *cache.SimpleCache
	fs             afero.Fs
}

// GitCommitInfo represents git commit information for a rule
type GitCommitInfo struct {
	Hash string
	Date string
}

// UpdateResult represents the result of checking/updating a rule
type UpdateResult struct {
	RuleID         string
	DisplayName    string
	CurrentVersion string
	LatestVersion  string
	HasUpdate      bool
	Error          error
	Status         UpdateStatus
	CurrentCommit  GitCommitInfo
	LatestCommit   GitCommitInfo
	Source         string // Source repository for custom rules
	Ref            string // Branch/tag reference for custom rules
}

// UpdateStatus represents the status of a rule update check
type UpdateStatus int

const (
	// StatusChecking indicates a rule is being checked for updates
	StatusChecking UpdateStatus = iota
	// StatusUpToDate indicates a rule is current
	StatusUpToDate
	// StatusUpdateAvailable indicates an update is available for the rule
	StatusUpdateAvailable
	// StatusError indicates an error occurred while checking the rule
	StatusError
	// StatusApplying indicates an update is being applied to the rule
	StatusApplying
	// StatusApplied indicates an update was successfully applied to the rule
	StatusApplied
)

// NewUpdateCommand creates a new update command with default dependencies
func NewUpdateCommand(deps *dependencies.Dependencies) *UpdateCommand {
	gitRepo := newOpenRepository(deps.FS)
	return &UpdateCommand{
		projectManager: project.NewManager(deps.FS),
		ruleFetcher:    rule.NewFetcher(deps.FS, gitRepo, rule.FetcherConfig{}),
		ruleValidator:  rule.NewValidator(),
		cache:          cache.NewSimpleCache(deps.FS, gitRepo),
		fs:             deps.FS,
	}
}

// NewUpdateCommandWithDependencies creates a new update command with explicit dependencies for testing
func NewUpdateCommandWithDependencies(
	projectManager *project.Manager,
	ruleFetcher rule.Fetcher,
	ruleValidator rule.Validator,
	cache *cache.SimpleCache,
	fs afero.Fs,
) *UpdateCommand {
	return &UpdateCommand{
		projectManager: projectManager,
		ruleFetcher:    ruleFetcher,
		ruleValidator:  ruleValidator,
		cache:          cache,
		fs:             fs,
	}
}

// Execute runs the update command
func (c *UpdateCommand) Execute(ctx context.Context, cmd *cli.Command) error {
	// Check if JSON output mode - if so, suppress all terminal output
	outputFormat := output.Format(cmd.String("output"))
	isJSONMode := outputFormat == output.FormatJSON

	if !isJSONMode {
		// Show header like add and list commands
		commandHeaderStyle := lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.AdaptiveColor{Light: "#F793FF", Dark: "#AD58B4"})
		fmt.Printf("%s\n\n", commandHeaderStyle.Render("Update Rules"))
	}
	dryRun := cmd.Bool("dry-run")
	skipConfirmation := cmd.Bool("yes")

	// Load configuration using shared utility
	configLoad, err := LoadProjectConfig(c.projectManager)
	if err != nil {
		return err
	}

	config := configLoad.Config

	const localSource = "local"

	// Filter out local rules - they cannot be updated since they are local files
	var updatableRules []domain.RuleRef
	for _, rule := range config.Rules {
		if rule.Source != localSource {
			updatableRules = append(updatableRules, rule)
		}
	}

	if len(updatableRules) == 0 {
		// Handle output format when no rules to update
		// outputFormat already declared
		outputManager, err := output.NewManager(outputFormat)
		if err != nil {
			return fmt.Errorf("failed to create output manager: %w", err)
		}

		// Write empty output
		metadata := output.UpdateMetadata{
			RulesUpdated:  []string{},
			RulesUpToDate: []string{},
			RulesFailed:   []string{},
		}

		err = outputManager.WriteRulesUpdate(metadata)
		if err != nil {
			return fmt.Errorf("failed to write update output: %w", err)
		}

		// For default format, also show the messages
		if outputFormat == output.FormatDefault {
			theme := ui.DefaultTheme()
			mutedStyle := lipgloss.NewStyle().Foreground(theme.Muted)
			fmt.Println("No rules configured to update")
			fmt.Println(mutedStyle.Render("Add rules with: contexture add <rule-id>"))
		}

		return nil
	}

	// Check for updates with real-time progress
	theme := ui.DefaultTheme()
	if !isJSONMode {
		headerStyle := lipgloss.NewStyle().Bold(true).Foreground(theme.Primary)
		fmt.Println(headerStyle.Render("Checking for updates..."))
		fmt.Println()
	}

	updateResults := c.checkForUpdatesWithProgress(ctx, updatableRules, isJSONMode)
	if !isJSONMode {
		fmt.Println()
	}

	// Count available updates and up-to-date rules
	updatesAvailable := 0
	upToDate := 0
	errors := 0
	for _, result := range updateResults {
		switch result.Status {
		case StatusUpToDate:
			upToDate++
		case StatusUpdateAvailable:
			if result.Error == nil {
				updatesAvailable++
			}
		case StatusError:
			errors++
		case StatusChecking, StatusApplying, StatusApplied:
			// These statuses shouldn't appear in final results, but handle them for completeness
		}
	}

	// Display results with integrated counts
	if !isJSONMode {
		if updatesAvailable > 0 {
			updateStyle := lipgloss.NewStyle().
				Bold(true).
				Foreground(theme.Update)
			fmt.Println(updateStyle.Render(fmt.Sprintf("↑ %d update(s) available", updatesAvailable)))
		}

		if upToDate > 0 {
			successStyle := lipgloss.NewStyle().
				Bold(true).
				Foreground(theme.Success)
			if updatesAvailable == 0 {
				fmt.Println(successStyle.Render("All rules are up to date!"))
			} else {
				fmt.Println(successStyle.Render(fmt.Sprintf("%d rule(s) up to date", upToDate)))
			}
		}

		if errors > 0 {
			errorStyle := lipgloss.NewStyle().
				Bold(true).
				Foreground(theme.Error)
			fmt.Println(errorStyle.Render(fmt.Sprintf("✗ %d error(s) occurred", errors)))
		}
	}

	if updatesAvailable == 0 {
		// Handle output format for no updates available
		// outputFormat already declared
		outputManager, err := output.NewManager(outputFormat)
		if err != nil {
			return fmt.Errorf("failed to create output manager: %w", err)
		}

		// Collect results when no updates are available
		var rulesUpdated []string // Empty since no updates
		var rulesUpToDate []string
		var rulesFailed []string

		for _, result := range updateResults {
			switch result.Status {
			case StatusUpToDate:
				rulesUpToDate = append(rulesUpToDate, result.DisplayName)
			case StatusError:
				rulesFailed = append(rulesFailed, result.DisplayName)
			case StatusChecking, StatusUpdateAvailable, StatusApplying, StatusApplied:
				// These statuses shouldn't occur in this context (no updates available)
			}
		}

		// Write output using the appropriate format
		metadata := output.UpdateMetadata{
			RulesUpdated:  rulesUpdated,
			RulesUpToDate: rulesUpToDate,
			RulesFailed:   rulesFailed,
		}

		return outputManager.WriteRulesUpdate(metadata)
	}

	if dryRun {
		// Handle output format for dry runs
		// outputFormat already declared
		outputManager, err := output.NewManager(outputFormat)
		if err != nil {
			return fmt.Errorf("failed to create output manager: %w", err)
		}

		// Collect results for dry run output
		var rulesUpdated []string // Would be updated if not dry run
		var rulesUpToDate []string
		var rulesFailed []string

		for _, result := range updateResults {
			switch result.Status {
			case StatusUpdateAvailable:
				if result.Error == nil {
					rulesUpdated = append(rulesUpdated, result.DisplayName)
				} else {
					rulesFailed = append(rulesFailed, result.DisplayName)
				}
			case StatusUpToDate:
				rulesUpToDate = append(rulesUpToDate, result.DisplayName)
			case StatusError:
				rulesFailed = append(rulesFailed, result.DisplayName)
			case StatusChecking, StatusApplying, StatusApplied:
				// These statuses shouldn't occur in dry run mode
			}
		}

		// Write output using the appropriate format
		metadata := output.UpdateMetadata{
			RulesUpdated:  rulesUpdated,
			RulesUpToDate: rulesUpToDate,
			RulesFailed:   rulesFailed,
		}

		err = outputManager.WriteRulesUpdate(metadata)
		if err != nil {
			return fmt.Errorf("failed to write update output: %w", err)
		}

		// For default format, show the dry run message
		if !isJSONMode {
			mutedStyle := lipgloss.NewStyle().Foreground(theme.Muted)
			fmt.Println(mutedStyle.Render("Run 'contexture update' without --dry-run to apply updates"))
		}

		return nil
	}

	// Confirm update
	if !skipConfirmation {
		confirmed := true // Default to yes
		confirmForm := ui.ConfigureHuhForm(huh.NewForm(
			huh.NewGroup(
				huh.NewConfirm().
					Title(fmt.Sprintf("Apply %d update(s)?", updatesAvailable)).
					Description("This will update the affected rules to their latest versions.").
					Affirmative("Yes").
					Negative("No").
					Value(&confirmed),
			),
		))

		if err := tui.HandleFormError(confirmForm.Run()); err != nil {
			return err
		}

		if !confirmed {
			if !isJSONMode {
				theme := ui.DefaultTheme()
				mutedStyle := lipgloss.NewStyle().Foreground(theme.Muted)
				fmt.Println(mutedStyle.Render("Update cancelled"))
			}
			return nil
		}
	}

	// Apply updates
	err = c.applyUpdates(ctx, updateResults, configLoad)
	if err != nil {
		return err
	}

	// Handle output format
	// outputFormat already declared
	outputManager, err := output.NewManager(outputFormat)
	if err != nil {
		return fmt.Errorf("failed to create output manager: %w", err)
	}

	// Collect results for output
	var rulesUpdated []string
	var rulesUpToDate []string
	var rulesFailed []string

	for _, result := range updateResults {
		switch result.Status {
		case StatusApplied:
			rulesUpdated = append(rulesUpdated, result.DisplayName)
		case StatusUpToDate:
			rulesUpToDate = append(rulesUpToDate, result.DisplayName)
		case StatusError:
			rulesFailed = append(rulesFailed, result.DisplayName)
		case StatusChecking, StatusUpdateAvailable, StatusApplying:
			// These statuses shouldn't occur in final result processing
		}
	}

	// Write output using the appropriate format
	metadata := output.UpdateMetadata{
		RulesUpdated:  rulesUpdated,
		RulesUpToDate: rulesUpToDate,
		RulesFailed:   rulesFailed,
	}

	return outputManager.WriteRulesUpdate(metadata)
}

// checkForUpdatesWithProgress checks all rules for available updates with real-time progress display
func (c *UpdateCommand) checkForUpdatesWithProgress(
	ctx context.Context,
	rules []domain.RuleRef,
	isJSONMode bool,
) []UpdateResult {
	results := make([]UpdateResult, len(rules))
	theme := ui.DefaultTheme()

	// Status indicators
	checkingSpinner := []string{"⢷", "⢹", "⢺", "⢼", "⢾", "⢿", "⠹", "⢸"}
	spinnerIndex := 0

	for i, ruleRef := range rules {
		// Extract simple rule ID for display using the domain package's ExtractRulePath
		displayRuleID := domain.ExtractRulePath(ruleRef.ID)
		if displayRuleID == "" {
			displayRuleID = ruleRef.ID
		}

		result := UpdateResult{
			RuleID:      ruleRef.ID,
			DisplayName: displayRuleID,
			Status:      StatusChecking,
		}

		// Extract source and ref information for custom rules
		if parsed, err := c.ruleFetcher.ParseRuleID(ruleRef.ID); err == nil {
			result.Source = parsed.Source
			result.Ref = parsed.Ref
		}

		// Skip pinned rules
		if ruleRef.Pinned {
			result.Status = StatusUpToDate
			result.CurrentVersion = ruleRef.CommitHash
			result.LatestVersion = ruleRef.CommitHash

			// Get commit info for the pinned commit to show date and hash
			currentCommitHash := ruleRef.CommitHash
			if currentCommitHash != "" {
				// Try to get commit info, but continue even if it fails
				parsed, parseErr := c.ruleFetcher.ParseRuleID(ruleRef.ID)
				if parseErr == nil {
					repoDir, repoErr := c.cache.GetRepositoryWithUpdate(ctx, parsed.Source, parsed.Ref)
					if repoErr == nil {
						gitRepo := newOpenRepository(c.fs)
						if commitInfo, commitErr := gitRepo.GetCommitInfoByHash(repoDir, currentCommitHash); commitErr == nil {
							result.CurrentCommit = GitCommitInfo{
								Hash: commitInfo.Hash,
								Date: commitInfo.Date,
							}
						}
					}
				}
			}

			// If we couldn't get commit info, use defaults
			if result.CurrentCommit.Hash == "" {
				result.CurrentCommit = GitCommitInfo{
					Hash: currentCommitHash,
					Date: "unknown",
				}
			}

			// Clear line and show pinned status
			if !isJSONMode {
				fmt.Printf("\r") // Clear the line first
				pinnedLine := c.formatRuleDisplay(result,
					lipgloss.NewStyle().Foreground(theme.Info).Render("~"),
					lipgloss.NewStyle().Foreground(theme.Muted).Render("pinned"))
				fmt.Printf("%s\n", pinnedLine)
			}
			results[i] = result
			continue
		}

		// Get current commit hash for comparison
		currentCommitHash := ruleRef.CommitHash

		// Show checking status with spinner
		spinnerStyle := lipgloss.NewStyle().Foreground(theme.Info)
		mutedStyle := lipgloss.NewStyle().Foreground(theme.Muted)
		spinner := spinnerStyle.Render(checkingSpinner[spinnerIndex%len(checkingSpinner)])

		// Create a checking line with enough padding to avoid rendering issues
		checkingLine := fmt.Sprintf(
			"  %s %s %s",
			spinner,
			displayRuleID,
			mutedStyle.Render("checking..."),
		)
		// Show checking status with simple carriage return
		if !isJSONMode {
			fmt.Printf("\r%s", checkingLine+strings.Repeat(" ", 20)) // Add padding to clear any leftover text
		}

		// Fetch latest rule content and get latest commit info
		currentCommit, latestCommit, hasUpdate, err := c.checkRuleForUpdate(
			ctx,
			ruleRef,
			currentCommitHash,
		)
		if err != nil {
			result.Error = fmt.Errorf("failed to check rule for updates: %w", err)
			result.Status = StatusError
			// Clear line and show error with proper formatting
			if !isJSONMode {
				fmt.Printf("\r") // Clear the line first
				errorLine := c.formatRuleDisplay(result,
					lipgloss.NewStyle().Foreground(theme.Error).Render("✗"),
					lipgloss.NewStyle().Foreground(theme.Error).Render("error"))
				fmt.Printf("%s\n", errorLine)
			}
		} else {
			// Set current and latest commit info (both now have real dates)
			result.CurrentCommit = GitCommitInfo{
				Hash: currentCommit.Hash,
				Date: currentCommit.Date,
			}

			result.LatestCommit = GitCommitInfo{
				Hash: latestCommit.Hash,
				Date: latestCommit.Date,
			}

			if hasUpdate {
				result.HasUpdate = true
				result.Status = StatusUpdateAvailable
				result.CurrentVersion = currentCommitHash
				result.LatestVersion = latestCommit.Hash

				// Clear line and show update available with commit info
				if !isJSONMode {
					fmt.Printf("\r") // Clear the line first
					updateLine := c.formatRuleDisplay(result,
						lipgloss.NewStyle().Foreground(theme.Update).Render("↑"),
						lipgloss.NewStyle().Foreground(theme.Update).Render("update available"))
					fmt.Printf("%s\n", updateLine)
				}
			} else {
				result.HasUpdate = false
				result.Status = StatusUpToDate
				result.CurrentVersion = currentCommitHash
				result.LatestVersion = latestCommit.Hash

				// Clear line and show up to date with commit info
				if !isJSONMode {
					fmt.Printf("\r") // Clear the line first
					upToDateLine := c.formatRuleDisplay(result,
						lipgloss.NewStyle().Foreground(theme.Success).Render("✓"),
						lipgloss.NewStyle().Foreground(theme.Muted).Render("up to date"))
					fmt.Printf("%s\n", upToDateLine)
				}
			}
		}

		results[i] = result
		spinnerIndex++

		// Add small delay to show the checking animation
		time.Sleep(150 * time.Millisecond)
	}

	return results
}

// checkRuleForUpdate checks if a rule has updates by comparing commit hashes from cached repository
func (c *UpdateCommand) checkRuleForUpdate(
	ctx context.Context,
	ruleRef domain.RuleRef,
	currentCommitHash string,
) (*GitCommitInfo, *GitCommitInfo, bool, error) {
	// Parse the rule ID to get the rule path and source information
	parsed, err := c.ruleFetcher.ParseRuleID(ruleRef.ID)
	if err != nil {
		return nil, nil, false, fmt.Errorf("failed to parse rule ID: %w", err)
	}

	// Get repository with updates using cache
	repoDir, err := c.cache.GetRepositoryWithUpdate(ctx, parsed.Source, parsed.Ref)
	if err != nil {
		return nil, nil, false, fmt.Errorf("failed to get repository: %w", err)
	}

	// Get the rule file path within the repository
	ruleFilePath := parsed.RulePath + ".md"

	// Create git repository instance for the cached directory
	gitRepo := newOpenRepository(c.fs)

	// Get the latest commit information for this specific file
	latestCommitInfo, err := gitRepo.GetFileCommitInfo(repoDir, ruleFilePath, parsed.Ref)
	if err != nil {
		return nil, nil, false, fmt.Errorf("failed to get file commit info: %w", err)
	}

	latestCommit := &GitCommitInfo{
		Hash: latestCommitInfo.Hash,
		Date: latestCommitInfo.Date,
	}

	// Get current commit info if we have a hash
	var currentCommit *GitCommitInfo
	if currentCommitHash != "" {
		currentCommitInfo, err := gitRepo.GetCommitInfoByHash(repoDir, currentCommitHash)
		if err != nil {
			// Print warning on new line with proper formatting
			fmt.Printf("\n")
			log.Warn("Failed to get current commit info", "hash", currentCommitHash, "error", err)
			currentCommit = &GitCommitInfo{
				Hash: currentCommitHash,
				Date: "unknown",
			}
		} else {
			currentCommit = &GitCommitInfo{
				Hash: currentCommitInfo.Hash,
				Date: currentCommitInfo.Date,
			}
		}
	} else {
		currentCommit = &GitCommitInfo{
			Hash: "none",
			Date: "unknown",
		}
	}

	// Check if there's an update by comparing commit hashes
	hasUpdate := currentCommitHash != latestCommit.Hash

	// If no current commit hash, consider it as needs initial setup
	if currentCommitHash == "" {
		hasUpdate = true
	}

	return currentCommit, latestCommit, hasUpdate, nil
}

// updateRuleCommitHash updates the commit hash for a specific rule in the config
func (c *UpdateCommand) updateRuleCommitHash(config *domain.Project, ruleID, newCommitHash string) {
	for i, rule := range config.Rules {
		if rule.ID == ruleID {
			config.Rules[i].CommitHash = newCommitHash
			break
		}
	}
}

// applyUpdates applies the available updates with progress feedback
func (c *UpdateCommand) applyUpdates(
	ctx context.Context,
	results []UpdateResult,
	configLoad *ConfigLoadResult,
) error {
	config := configLoad.Config
	theme := ui.DefaultTheme()
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(theme.Primary)
	successStyle := lipgloss.NewStyle().Foreground(theme.Success)
	errorStyle := lipgloss.NewStyle().Foreground(theme.Error)
	mutedStyle := lipgloss.NewStyle().Foreground(theme.Muted)

	fmt.Println(headerStyle.Render("Applying updates..."))
	fmt.Println()

	updatedCount := 0
	var errors []string

	// Show progress for each update
	for _, result := range results {
		if !result.HasUpdate || result.Error != nil {
			continue
		}

		// Show applying status
		applyingLine := fmt.Sprintf("  %s %s %s",
			lipgloss.NewStyle().Foreground(theme.Update).Render("↑"),
			result.DisplayName,
			mutedStyle.Render("applying..."))
		fmt.Printf("\r\033[K%s", applyingLine)

		// Simulate some processing time
		time.Sleep(200 * time.Millisecond)

		// Fetch and validate the updated rule
		fetchedRule, err := c.ruleFetcher.FetchRule(ctx, result.RuleID)
		if err != nil {
			// Clear line and show error
			fmt.Printf("\r") // Clear the line first
			errorLine := fmt.Sprintf("  %s %s %s",
				errorStyle.Render("✗"),
				result.DisplayName,
				errorStyle.Render("failed"))
			fmt.Printf("%s\n", errorLine)
			errors = append(errors, fmt.Sprintf("%s: %v", result.DisplayName, err))
			continue
		}

		// Validate the updated rule
		validationResult := c.ruleValidator.ValidateRule(fetchedRule)
		if !validationResult.Valid {
			var errorMessages []string
			for _, validationErr := range validationResult.Errors {
				errorMessages = append(errorMessages, validationErr.Error())
			}
			errorMsg := fmt.Sprintf("validation failed: %s", strings.Join(errorMessages, ", "))
			// Clear line and show validation error
			fmt.Printf("\r") // Clear the line first
			validationErrorLine := fmt.Sprintf("  %s %s %s",
				errorStyle.Render("✗"),
				result.DisplayName,
				errorStyle.Render("validation failed"))
			fmt.Printf("%s\n", validationErrorLine)
			errors = append(errors, fmt.Sprintf("%s: %s", result.DisplayName, errorMsg))
			continue
		}

		// Update the commit hash in the config
		c.updateRuleCommitHash(config, result.RuleID, result.LatestCommit.Hash)

		// Update status to applied
		for i := range results {
			if results[i].RuleID == result.RuleID {
				results[i].Status = StatusApplied
				break
			}
		}

		// Clear line and show success
		fmt.Printf("\r\033[K") // Clear the line first
		successLine := fmt.Sprintf("  %s %s %s",
			successStyle.Render("✓"),
			result.DisplayName,
			successStyle.Render("updated"))
		fmt.Printf("%s\n", successLine)
		updatedCount++
	}

	// Save configuration using shared utility
	if err := configLoad.SaveConfig(c.projectManager); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	// Display final results
	fmt.Println()
	if updatedCount > 0 {
		headerStyle := lipgloss.NewStyle().Bold(true).Foreground(theme.Success)
		message := fmt.Sprintf("✓ Successfully updated %d rule(s)", updatedCount)
		fmt.Println(headerStyle.Render(message))
	}

	if len(errors) > 0 {
		headerStyle := lipgloss.NewStyle().Bold(true).Foreground(theme.Error)
		fmt.Println(headerStyle.Render(fmt.Sprintf("✗ %d error(s) occurred:", len(errors))))
		for _, err := range errors {
			fmt.Printf("  %s\n", errorStyle.Render(err))
		}
		fmt.Println()
	}

	if updatedCount > 0 {
		// Automatically regenerate files after updates
		fmt.Println()
		headerStyle := lipgloss.NewStyle().Bold(true).Foreground(theme.Primary)
		fmt.Println(headerStyle.Render("Regenerating format files..."))
		fmt.Println()

		// Create build command and execute it
		buildCmd := NewBuildCommand(&dependencies.Dependencies{
			FS:      c.fs,
			Context: ctx,
		})

		// Create a minimal CLI command for generation
		dummyCmd := &cli.Command{}
		if err := buildCmd.Execute(ctx, dummyCmd); err != nil {
			log.Warn("Failed to regenerate files after update", "error", err)
		}
	}

	return nil
}

// shortHash returns the first 7 characters of a commit hash for display
func shortHash(hash string) string {
	if len(hash) >= 7 {
		return hash[:7]
	}
	return hash
}

// formatDateForAlignment ensures consistent date formatting for column alignment
func formatDateForAlignment(date string) string {
	// Pad "unknown" to match typical date width for better alignment
	if date == "unknown" {
		return "unknown     " // Pad to match typical date length
	}
	return date
}

// formatRuleDisplay formats the rule display line with commit info and proper alignment
func (c *UpdateCommand) formatRuleDisplay(result UpdateResult, status, statusText string) string {
	darkGrayStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#666666"))

	// Define column widths for alignment
	const (
		statusWidth = 2  // "✓ " or "↑ "
		nameWidth   = 35 // rule name column
		textWidth   = 18 // "up to date" / "update available"
		dateWidth   = 12 // "23 Jul 2025" or "unknown"
	)

	// Truncate rule name if too long
	displayName := result.DisplayName
	if len(displayName) > nameWidth-1 {
		displayName = displayName[:nameWidth-4] + "..."
	}

	var mainLine string
	if result.HasUpdate && result.Status == StatusUpdateAvailable {
		// Show current → latest for updates
		mainLine = fmt.Sprintf(
			"%s %-*s %-*s %s → %s",
			status,
			nameWidth, displayName,
			textWidth, statusText,
			darkGrayStyle.Render(
				fmt.Sprintf(
					"%-*s %s",
					dateWidth,
					formatDateForAlignment(result.CurrentCommit.Date),
					shortHash(result.CurrentCommit.Hash),
				),
			),
			darkGrayStyle.Render(
				fmt.Sprintf(
					"%-*s %s",
					dateWidth,
					formatDateForAlignment(result.LatestCommit.Date),
					shortHash(result.LatestCommit.Hash),
				),
			),
		)
	} else {
		// Show just current for up-to-date or error
		mainLine = fmt.Sprintf(
			"%s %-*s %-*s %s",
			status,
			nameWidth, displayName,
			textWidth, statusText,
			darkGrayStyle.Render(
				fmt.Sprintf(
					"%-*s %s",
					dateWidth,
					formatDateForAlignment(result.CurrentCommit.Date),
					shortHash(result.CurrentCommit.Hash),
				),
			),
		)
	}

	// Check if this is a custom source rule (not empty and not default)
	if result.Source != "" && domain.IsCustomGitSource(result.Source) {
		// Add source line in dark gray underneath
		sourceLine := domain.FormatSourceForDisplay(result.Source, result.Ref)
		sourceDisplay := darkGrayStyle.Render(fmt.Sprintf("    %s", sourceLine))
		return fmt.Sprintf("%s\n%s", mainLine, sourceDisplay)
	}

	return mainLine
}

// UpdateAction is the CLI action handler for the update command
func UpdateAction(ctx context.Context, cmd *cli.Command, deps *dependencies.Dependencies) error {
	updateCmd := NewUpdateCommand(deps)
	return updateCmd.Execute(ctx, cmd)
}
