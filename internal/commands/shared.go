// Package commands provides CLI command implementations
package commands

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"github.com/contextureai/contexture/internal/domain"
	contextureerrors "github.com/contextureai/contexture/internal/errors"
	"github.com/contextureai/contexture/internal/format"
	"github.com/contextureai/contexture/internal/rule"
	"github.com/contextureai/contexture/internal/ui"
	"github.com/spf13/afero"
)

// RuleGenerator provides shared rule generation functionality
type RuleGenerator struct {
	ruleFetcher   rule.Fetcher
	ruleValidator rule.Validator
	ruleProcessor rule.Processor
	registry      *format.Registry
	fs            afero.Fs
}

// NewRuleGenerator creates a new rule generator
func NewRuleGenerator(
	fetcher rule.Fetcher,
	validator rule.Validator,
	processor rule.Processor,
	registry *format.Registry,
	fs afero.Fs,
) *RuleGenerator {
	if fs == nil {
		fs = afero.NewOsFs()
	}
	return &RuleGenerator{
		ruleFetcher:   fetcher,
		ruleValidator: validator,
		ruleProcessor: processor,
		registry:      registry,
		fs:            fs,
	}
}

// GenerateRules handles the complete rule generation process with consistent UI
func (g *RuleGenerator) GenerateRules(
	ctx context.Context,
	config *domain.Project,
	targetFormats []domain.FormatConfig,
) error {
	return g.GenerateRulesWithScope(ctx, config, targetFormats, "")
}

// GenerateRulesWithScopeAndWarning handles the complete rule generation process with scope tags and optional warnings
func (g *RuleGenerator) GenerateRulesWithScopeAndWarning(
	ctx context.Context,
	config *domain.Project,
	targetFormats []domain.FormatConfig,
	scope string, // "project", "global", or "" for no scope
	hasGlobalRules bool, // whether global rules are being merged (for warnings)
) error {
	return g.generateRulesWithScopeInternal(ctx, config, targetFormats, scope, hasGlobalRules)
}

// GenerateRulesWithScope handles the complete rule generation process with scope tags
func (g *RuleGenerator) GenerateRulesWithScope(
	ctx context.Context,
	config *domain.Project,
	targetFormats []domain.FormatConfig,
	scope string, // "project", "global", or "" for no scope
) error {
	return g.generateRulesWithScopeInternal(ctx, config, targetFormats, scope, false)
}

// generateRulesWithScopeInternal is the internal implementation
func (g *RuleGenerator) generateRulesWithScopeInternal(
	ctx context.Context,
	config *domain.Project,
	targetFormats []domain.FormatConfig,
	scope string, // "project", "global", or "" for no scope
	hasGlobalRules bool, // whether global rules are being merged (for warnings)
) error {
	if len(targetFormats) == 0 {
		return contextureerrors.ValidationErrorf("formats", "no target formats available")
	}

	// If no rules, we still need to generate (which will trigger cleanup/deletion in format handlers)
	var processedRules []*domain.ProcessedRule
	if len(config.Rules) > 0 {
		// Fetch all rules in parallel with progress indicator and timing
		var rules []*domain.Rule
		scopeLabel := ""
		if scope != "" {
			theme := ui.DefaultTheme()
			mutedStyle := lipgloss.NewStyle().Foreground(theme.Muted)
			scopeLabel = " " + mutedStyle.Render(fmt.Sprintf("[%s]", scope))
		}

		err := ui.WithProgress("Fetched rules"+scopeLabel, func() error {
			var fetchErr error
			rules, fetchErr = rule.FetchRulesParallel(
				ctx,
				g.ruleFetcher,
				config.Rules,
				config.GetGeneration().ParallelFetches,
			)
			return fetchErr
		})
		if err != nil {
			return contextureerrors.Wrap(err, "fetch rules")
		}

		// Sort rules deterministically for consistent output
		parser := rule.NewRuleIDParser("", nil)
		rules = rule.SortRulesDeterministically(rules, parser)

		// Process rules (templates, validation) with progress indicator and timing
		err = ui.WithProgress("Generated rules"+scopeLabel, func() error {
			var processErr error
			processedRules, processErr = g.processRules(ctx, rules)
			return processErr
		})
		if err != nil {
			return contextureerrors.Wrap(err, "process rules")
		}
	} else {
		log.Debug("No rules configured, will trigger cleanup in format handlers")
	}

	// Generate output for each format (even with 0 rules to trigger cleanup)
	for _, formatConfig := range targetFormats {
		if err := g.generateFormat(ctx, processedRules, formatConfig); err != nil {
			log.Warn("Failed to generate format", "format", formatConfig.Type, "error", err)
			continue
		}

		// Show format completion with scope tag (only if we had rules to process)
		if len(processedRules) > 0 {
			if handler, exists := g.registry.GetHandler(formatConfig.Type); exists {
				theme := ui.DefaultTheme()
				successStyle := lipgloss.NewStyle().Foreground(theme.Success)
				mutedStyle := lipgloss.NewStyle().Foreground(theme.Muted)

				displayName := handler.GetDisplayName()
				if scope != "" {
					displayName += " " + mutedStyle.Render(fmt.Sprintf("[%s]", scope))
				}
				fmt.Printf("  %s %s\n", successStyle.Render("✓"), displayName)

				// Show warning for Cursor when global rules are being merged
				if hasGlobalRules && formatConfig.Type == domain.FormatCursor && scope == "project" {
					fmt.Printf("     %s %s\n",
						mutedStyle.Render("⚠"),
						mutedStyle.Render("Cursor does not support native global rules. Your global rules will be merged into project files, which may cause conflicts in team environments. Consider setting Cursor's userRulesMode to 'disabled' in .contexture.yaml"))
				}
			}
		}
	}

	log.Debug("Rule generation completed",
		"rules", len(processedRules),
		"formats", len(targetFormats),
		"scope", scope)
	return nil
}

// processRules validates and processes rules through templates
func (g *RuleGenerator) processRules(
	_ context.Context,
	rules []*domain.Rule,
) ([]*domain.ProcessedRule, error) {
	var processedRules []*domain.ProcessedRule
	var errors []string

	for _, rule := range rules {
		// Validate rule
		validationResult := g.ruleValidator.ValidateRule(rule)
		if !validationResult.Valid {
			var errorMessages []string
			for _, err := range validationResult.Errors {
				errorMessages = append(errorMessages, err.Error())
			}
			errors = append(errors, fmt.Sprintf("rule %s validation failed: %s",
				rule.ID, strings.Join(errorMessages, ", ")))
			continue
		}

		// Process rule templates
		processedRule, err := g.ruleProcessor.ProcessRule(rule, &domain.RuleContext{})
		if err != nil {
			errors = append(errors, fmt.Sprintf("rule %s processing failed: %v", rule.ID, err))
			continue
		}

		processedRules = append(processedRules, processedRule)
	}

	if len(errors) > 0 {
		return nil, contextureerrors.ValidationErrorf("rules", "processing errors: %v", errors)
	}

	return processedRules, nil
}

// generateFormat generates output for a single format
func (g *RuleGenerator) generateFormat(
	_ context.Context,
	rules []*domain.ProcessedRule,
	formatConfig domain.FormatConfig,
) error {
	// Create format instance
	format, err := g.registry.CreateFormat(formatConfig.Type, g.fs, nil)
	if err != nil {
		return contextureerrors.Wrap(err, "create format")
	}

	// Transform rules for this format
	var transformedRules []*domain.TransformedRule
	for _, processedRule := range rules {
		transformed, err := format.Transform(processedRule)
		if err != nil {
			return contextureerrors.Wrap(err, "transform rule")
		}
		transformedRules = append(transformedRules, transformed)
	}

	// Write format output
	err = format.Write(transformedRules, &formatConfig)
	if err != nil {
		return contextureerrors.Wrap(err, "write format output")
	}

	// Clean up empty directories if no rules were written
	if len(transformedRules) == 0 {
		g.cleanupEmptyFormatDirectory(format, &formatConfig)
	}

	log.Debug("Format generated", "type", formatConfig.Type, "rules", len(transformedRules))
	return nil
}

// cleanupEmptyFormatDirectory removes empty output directories for formats that support it
func (g *RuleGenerator) cleanupEmptyFormatDirectory(format domain.Format, config *domain.FormatConfig) {
	// Check if the format has a method to get the output directory and access to BaseFormat
	if f, ok := format.(interface {
		getOutputDir(*domain.FormatConfig) string
		CleanupEmptyDirectory(string)
	}); ok {
		outputDir := f.getOutputDir(config)
		if outputDir != "" {
			// Use the centralized cleanup method from BaseFormat
			f.CleanupEmptyDirectory(outputDir)
		}
	}
}
