package windsurf

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	contextureerrors "github.com/contextureai/contexture/internal/errors"

	"github.com/contextureai/contexture/internal/domain"
	"github.com/contextureai/contexture/internal/format/base"
	"github.com/spf13/afero"
)

// OutputMode defines the Windsurf output mode
type OutputMode string

const (
	// ModeSingleFile outputs all rules to a single file
	ModeSingleFile OutputMode = "single"
	// ModeMultiFile outputs each rule to its own file
	ModeMultiFile OutputMode = "multi"

	singleFileFilename = "rules.md"
)

// Strategy implements the FormatStrategy interface for Windsurf format
type Strategy struct {
	fs   afero.Fs
	bf   *base.Base
	mode OutputMode
}

// NewStrategy creates a new Windsurf strategy
func NewStrategy(fs afero.Fs, bf *base.Base) *Strategy {
	return &Strategy{
		fs:   fs,
		bf:   bf,
		mode: ModeMultiFile, // Default to multi-file mode
	}
}

// NewStrategyWithMode creates a new Windsurf strategy with specified mode
func NewStrategyWithMode(fs afero.Fs, bf *base.Base, mode OutputMode) *Strategy {
	return &Strategy{
		fs:   fs,
		bf:   bf,
		mode: mode,
	}
}

// SetMode sets the output mode for the strategy
func (s *Strategy) SetMode(mode OutputMode) {
	s.mode = mode
}

// GetMode returns the current output mode
func (s *Strategy) GetMode() OutputMode {
	return s.mode
}

// GetDefaultTemplate returns the default Windsurf template with YAML frontmatter matching Windsurf spec
func (s *Strategy) GetDefaultTemplate() string {
	return `---
{{if .trigger}}{{if eq .trigger.type "always"}}trigger: always_on{{else if eq .trigger.type "manual"}}trigger: manual{{else if eq .trigger.type "model_decision"}}trigger: model_decision{{else if eq .trigger.type "glob"}}trigger: glob{{else}}trigger: manual{{end}}
{{if .description}}description: "{{.description}}"
{{end}}{{if and (eq .trigger.type "glob") .trigger.globs}}globs: "{{join .trigger.globs ","}}"
{{end}}{{else}}trigger: manual
{{end}}---

# {{.title}}

{{if .description}}> {{.description}}

{{end}}{{if or .languages .frameworks .tags}}## Context
{{if .languages}}- **Languages**: {{join_and .languages}}
{{end}}{{if .frameworks}}- **Frameworks**: {{join_and .frameworks}}
{{end}}{{if .tags}}- **Categories**: {{join_and .tags}}
{{end}}

{{end}}{{.content}}`
}

// GetOutputPath returns the output directory path for Windsurf format
func (s *Strategy) GetOutputPath(config *domain.FormatConfig) string {
	if config == nil || config.BaseDir == "" {
		return domain.WindsurfOutputDir
	}
	// For user rules, output directly to BaseDir (e.g., ~/.windsurf)
	if config.IsUserRules {
		return config.BaseDir
	}
	// For project rules, output to .windsurf/rules/ subdirectory
	return filepath.Join(config.BaseDir, domain.WindsurfOutputDir)
}

// GetFileExtension returns the file extension for Windsurf format
func (s *Strategy) GetFileExtension() string {
	return ".md"
}

// IsSingleFile returns true if in single-file mode
func (s *Strategy) IsSingleFile() bool {
	return s.mode == ModeSingleFile
}

// GenerateFilename generates a filename from rule ID
func (s *Strategy) GenerateFilename(ruleID string) string {
	if s.mode == ModeSingleFile {
		return singleFileFilename
	}
	return s.bf.GenerateFilename(ruleID)
}

// GetMetadata returns metadata about Windsurf format
func (s *Strategy) GetMetadata() *domain.FormatMetadata {
	return &domain.FormatMetadata{
		Type:        domain.FormatWindsurf,
		DisplayName: "Windsurf IDE",
		Description: "Multi-file format for Windsurf IDE (.windsurf/rules/)",
		IsDirectory: true,
	}
}

// WriteFiles handles writing rules for Windsurf format (single or multi-file based on mode)
func (s *Strategy) WriteFiles(rules []*domain.TransformedRule, config *domain.FormatConfig) error {
	outputDir := s.GetOutputPath(config)

	// When no rules, delete output files/directory
	if len(rules) == 0 {
		s.bf.LogDebug("No rules to write for Windsurf format, deleting output")
		exists, err := s.bf.DirExists(outputDir)
		if err != nil {
			s.bf.LogDebug("Failed to check if directory exists", "path", outputDir, "error", err)
			return nil
		}
		if exists {
			// Remove the entire rules directory
			if err := s.bf.RemoveDirectory(outputDir); err != nil {
				return contextureerrors.WithOpf("delete output directory", "failed to delete %s: %w", outputDir, err)
			}
			s.bf.LogInfo("Deleted Windsurf format directory", "path", outputDir)

			// Also clean up parent .windsurf directory if it's now empty
			if config != nil {
				baseDir := config.BaseDir
				if baseDir == "" {
					baseDir = "."
				}
				parentDir := filepath.Join(baseDir, ".windsurf")
				s.bf.CleanupEmptyDirectory(parentDir)
			}
		}
		return nil
	}

	s.bf.LogDebug("Writing Windsurf format files", "rules", len(rules), "mode", s.mode)

	// Check character limits for each rule individually
	for _, rule := range rules {
		if len(rule.Content) > domain.WindsurfMaxSingleRuleChars {
			return contextureerrors.ValidationErrorf(
				rule.Rule.ID,
				"rule '%s' exceeds Windsurf per-file limit of %d characters (current: %d)",
				rule.Rule.ID,
				domain.WindsurfMaxSingleRuleChars,
				len(rule.Content),
			)
		}
	}

	// Ensure output directory exists
	if err := s.bf.EnsureDirectory(outputDir); err != nil {
		return contextureerrors.Wrap(err, "windsurf.WriteFiles: create output directory")
	}

	// Force single-file mode for user rules (global_rules.md)
	useSingleFile := s.mode == ModeSingleFile || (config != nil && config.IsUserRules)
	if useSingleFile {
		return s.writeSingleFile(rules, outputDir, config)
	}
	return s.writeMultiFile(rules, outputDir)
}

// CleanupEmptyDirectories handles cleanup of empty directories for Windsurf format
func (s *Strategy) CleanupEmptyDirectories(config *domain.FormatConfig) error {
	outputDir := s.GetOutputPath(config)

	baseDir := config.BaseDir
	if baseDir == "" {
		baseDir = "."
	}
	parentDir := filepath.Join(baseDir, ".windsurf")

	// First clean up the rules directory
	s.bf.CleanupEmptyDirectory(outputDir)
	// Then clean up the parent .windsurf directory if it's also empty
	s.bf.CleanupEmptyDirectory(parentDir)

	return nil
}

// CreateDirectories creates necessary directories for Windsurf format
func (s *Strategy) CreateDirectories(config *domain.FormatConfig) error {
	outputDir := s.GetOutputPath(config)
	return s.bf.EnsureDirectory(outputDir)
}

// writeSingleFile writes all rules to a single file
func (s *Strategy) writeSingleFile(rules []*domain.TransformedRule, outputDir string, config *domain.FormatConfig) error {
	filename := singleFileFilename
	// For user rules, use global_rules.md instead of rules.md
	if config != nil && config.IsUserRules {
		filename = "global_rules.md"
	}
	filePath := filepath.Join(outputDir, filename)

	var content strings.Builder
	content.Grow(s.estimateContentSize(rules))

	// Write header
	content.WriteString(s.getSingleFileHeader(len(rules)))
	content.WriteString("\n\n")

	// Write each rule
	for i, rule := range rules {
		if i > 0 {
			content.WriteString("\n\n---\n\n")
		}

		// Write rule content with tracking comment appended, only including non-default variables
		ruleContent := s.bf.AppendTrackingCommentWithDefaults(rule.Content, rule.Rule.ID, rule.Rule.Variables, rule.Rule.DefaultVariables)
		content.WriteString(ruleContent)
	}

	// Write footer
	content.WriteString("\n\n")
	content.WriteString(s.getSingleFileFooter())

	// Write to file
	if err := s.bf.WriteFile(filePath, []byte(content.String())); err != nil {
		return contextureerrors.Wrap(err, "windsurf.writeSingleFile")
	}

	s.bf.LogInfo("Successfully wrote Windsurf single file", "path", filePath, "rules", len(rules))
	return nil
}

// writeMultiFile writes each rule to its own file
func (s *Strategy) writeMultiFile(rules []*domain.TransformedRule, outputDir string) error {
	var errors []error

	// Write each rule to its own file
	for _, rule := range rules {
		filePath := filepath.Join(outputDir, rule.Filename)

		// Append tracking comment at the end instead of header at beginning, only including non-default variables
		content := s.bf.AppendTrackingCommentWithDefaults(rule.Content, rule.Rule.ID, rule.Rule.Variables, rule.Rule.DefaultVariables)

		if err := s.bf.WriteFile(filePath, []byte(content)); err != nil {
			errors = append(errors, contextureerrors.Wrap(err, "windsurf.writeMultiFile: write rule "+rule.Rule.ID))
			continue
		}

		s.bf.LogDebug("Wrote Windsurf rule file", "ruleID", rule.Rule.ID, "path", filePath)
	}

	if len(errors) > 0 {
		return contextureerrors.WithOpf("windsurf.writeMultiFile", "failed to write %d rules: %v", len(errors), errors)
	}

	s.bf.LogInfo("Successfully wrote Windsurf multi-file format", "count", len(rules), "directory", outputDir)
	return nil
}

// getSingleFileHeader returns the header for single file mode
func (s *Strategy) getSingleFileHeader(ruleCount int) string {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	return fmt.Sprintf(`# Windsurf Rules

This file contains %d contexture rules for Windsurf AI assistant.

Generated at: %s
Mode: Single File`, ruleCount, timestamp)
}

// getSingleFileFooter returns the footer for single file mode
func (s *Strategy) getSingleFileFooter() string {
	return `---

*This file was generated by Contexture CLI in single-file mode. Do not edit manually.*`
}

// estimateContentSize estimates the total size needed for the Windsurf single file
func (s *Strategy) estimateContentSize(rules []*domain.TransformedRule) int {
	// Start with header + footer overhead
	size := 1024

	// Add space for each rule's content plus separators
	for _, rule := range rules {
		size += len(rule.Content) + 200 // Content + tracking comment + separator overhead
	}

	return size
}

// Format implements the Windsurf format with support for both single and multi-file modes
type Format struct {
	*base.CommonFormat

	strategy *Strategy
}

// NewFormat creates a new Windsurf format implementation
func NewFormat(fs afero.Fs) *Format {
	bf := base.NewBaseFormat(fs, domain.FormatWindsurf)
	strategy := NewStrategy(fs, bf)
	commonFormat := base.NewCommonFormat(bf, strategy)

	return &Format{
		CommonFormat: commonFormat,
		strategy:     strategy,
	}
}

// NewFormatWithMode creates a new Windsurf format implementation with specified mode
func NewFormatWithMode(fs afero.Fs, mode OutputMode) *Format {
	bf := base.NewBaseFormat(fs, domain.FormatWindsurf)
	strategy := NewStrategyWithMode(fs, bf, mode)
	commonFormat := base.NewCommonFormat(bf, strategy)

	return &Format{
		CommonFormat: commonFormat,
		strategy:     strategy,
	}
}

// NewFormatFromOptions creates a new Windsurf format with options
func NewFormatFromOptions(fs afero.Fs, options map[string]any) (domain.Format, error) {
	format := NewFormat(fs)

	// Check for mode option
	if modeStr, ok := options["mode"].(string); ok {
		mode := OutputMode(modeStr)
		if mode == ModeSingleFile || mode == ModeMultiFile {
			format.SetMode(mode)
		}
	}

	return format, nil
}

// SetMode sets the output mode for the format
func (f *Format) SetMode(mode OutputMode) {
	f.strategy.SetMode(mode)
}

// GetMode returns the current output mode
func (f *Format) GetMode() OutputMode {
	return f.strategy.GetMode()
}

// Transform converts a processed rule to Windsurf format representation
// Overrides CommonFormat.Transform to add mode-specific metadata
func (f *Format) Transform(processedRule *domain.ProcessedRule) (*domain.TransformedRule, error) {
	// Use CommonFormat's Transform
	transformed, err := f.CommonFormat.Transform(processedRule)
	if err != nil {
		return nil, err
	}

	// Add mode to metadata
	transformed.Metadata["mode"] = string(f.strategy.GetMode())

	return transformed, nil
}

// Validate checks if a rule is valid for Windsurf format (adds character limit validation)
func (f *Format) Validate(rule *domain.Rule) (*domain.ValidationResult, error) {
	// Use CommonFormat validation and add mode metadata
	result, err := f.CommonFormat.Validate(rule)
	if err != nil {
		return nil, err
	}

	result.Metadata["mode"] = string(f.strategy.GetMode())

	// Add Windsurf-specific character limit validation
	contentLength := len(rule.Content)
	if contentLength > domain.WindsurfMaxSingleRuleChars {
		result.Errors = append(result.Errors, contextureerrors.ValidationErrorf(
			"content",
			"rule content exceeds Windsurf limit of %d characters (current: %d)",
			domain.WindsurfMaxSingleRuleChars,
			contentLength,
		))
		result.Valid = false
	}

	return result, nil
}

// List returns all currently installed rules for Windsurf format
// Note: We override this to handle the single-file mode's special parsing
func (f *Format) List(config *domain.FormatConfig) ([]*domain.InstalledRule, error) {
	f.strategy.bf.LogDebug("Listing installed rules for Windsurf format", "mode", f.strategy.mode)

	outputDir := f.strategy.GetOutputPath(config)

	// Check if directory exists
	exists, err := f.strategy.bf.DirExists(outputDir)
	if err != nil {
		return nil, contextureerrors.Wrap(err, "windsurf.List: check directory exists")
	}
	if !exists {
		f.strategy.bf.LogDebug("Windsurf format directory does not exist", "path", outputDir)
		return []*domain.InstalledRule{}, nil
	}

	// Use CommonFormat's list implementation
	return f.CommonFormat.List(config)
}

// Remove deletes a specific rule from the Windsurf format
// Note: We override this to handle single-file mode's content rebuilding
func (f *Format) Remove(ruleID string, config *domain.FormatConfig) error {
	f.strategy.bf.LogDebug("Removing rule from Windsurf format", "ruleID", ruleID, "mode", f.strategy.mode)

	if f.strategy.mode == ModeSingleFile {
		return f.removeSingleFile(ruleID, config)
	}
	// Use CommonFormat's removeMultiFile logic
	return f.CommonFormat.Remove(ruleID, config)
}

// GetSingleFileFilename returns the filename for single-file mode.
func (f *Format) GetSingleFileFilename() string {
	return singleFileFilename
}

// GetDefaultTemplate returns the default template for the format.
func (f *Format) GetDefaultTemplate() string {
	return f.strategy.GetDefaultTemplate()
}

// removeSingleFile removes a rule from the single file by parsing and rebuilding content
func (f *Format) removeSingleFile(ruleID string, config *domain.FormatConfig) error {
	outputDir := f.strategy.GetOutputPath(config)
	filename := singleFileFilename
	filePath := filepath.Join(outputDir, filename)

	// Read current content (EAFP - will fail if file doesn't exist)
	content, err := f.strategy.bf.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			f.strategy.bf.LogDebug("Windsurf single file does not exist", "path", filePath)
			return nil
		}
		return contextureerrors.Wrap(err, "windsurf.removeSingleFile: read file")
	}

	// Remove the rule from content by parsing sections
	contentStr := string(content)
	updatedContent := f.removeRuleFromContent(contentStr, ruleID)

	// Write back the updated content
	if err := f.strategy.bf.WriteFile(filePath, []byte(updatedContent)); err != nil {
		return contextureerrors.Wrap(err, "windsurf.removeSingleFile: write file")
	}

	f.strategy.bf.LogInfo("Successfully removed rule from Windsurf single file", "ruleID", ruleID)
	return nil
}

// removeRuleFromContent removes a specific rule from Windsurf format content
func (f *Format) removeRuleFromContent(content, ruleID string) string {
	// Split content by rule separators
	sections := strings.Split(content, "\n---\n")
	var filteredSections []string

	for _, section := range sections {
		// Check if this section contains the tracking comment for the rule we want to remove
		trackingComment := f.strategy.bf.CreateTrackingComment(ruleID, nil)
		if !strings.Contains(section, trackingComment) {
			filteredSections = append(filteredSections, section)
		}
	}

	return strings.Join(filteredSections, "\n---\n")
}

// Test helper methods to expose strategy methods
// These are used by tests to verify private implementation details

func (f *Format) getOutputDir(config *domain.FormatConfig) string {
	return f.strategy.GetOutputPath(config)
}

func (f *Format) getSingleFileHeader(ruleCount int) string {
	return f.strategy.getSingleFileHeader(ruleCount)
}

func (f *Format) getSingleFileFooter() string {
	return f.strategy.getSingleFileFooter()
}
