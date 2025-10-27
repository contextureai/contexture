package claude

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/contextureai/contexture/internal/domain"
	contextureerrors "github.com/contextureai/contexture/internal/errors"
	"github.com/contextureai/contexture/internal/format/base"
	"github.com/spf13/afero"
)

const defaultClaudeFilename = "CLAUDE.md"

// Strategy implements the FormatStrategy interface for Claude format
type Strategy struct {
	fs afero.Fs
	bf *base.Base
}

// NewStrategy creates a new Claude strategy
func NewStrategy(fs afero.Fs, bf *base.Base) *Strategy {
	return &Strategy{
		fs: fs,
		bf: bf,
	}
}

// GetDefaultTemplate returns the default Claude template
func (s *Strategy) GetDefaultTemplate() string {
	return `# {{.title}}

{{if .description}}{{.description}}

{{end}}{{if .trigger}}{{if eq .trigger.type "always"}}**Applies:** Always active
{{else if eq .trigger.type "glob"}}**Applies:** When working with {{join_and .trigger.globs}} files
{{else if eq .trigger.type "model_decision"}}**Applies:** When {{.description}}
{{else}}**Applies:** When explicitly requested
{{end}}

{{end}}{{if .tags}}**Tags:** {{join_and .tags}}
{{end}}{{if .frameworks}}**Frameworks:** {{join_and .frameworks}}
{{end}}{{.content}}`
}

// GetOutputPath returns the full output path for the Claude format file
func (s *Strategy) GetOutputPath(config *domain.FormatConfig) string {
	filename := defaultClaudeFilename

	if config == nil {
		return filename
	}

	baseDir := config.BaseDir
	if baseDir == "" {
		baseDir = "."
	}

	// For user rules, BaseDir is already set to ~/.claude/
	// so we just join with the filename
	return filepath.Join(baseDir, filename)
}

// GetFileExtension returns the file extension for Claude format
func (s *Strategy) GetFileExtension() string {
	return ".md"
}

// IsSingleFile returns true since Claude format outputs to a single file
func (s *Strategy) IsSingleFile() bool {
	return true
}

// GenerateFilename generates a filename from a rule ID (not used for single-file format)
func (s *Strategy) GenerateFilename(_ string) string {
	return defaultClaudeFilename
}

// GetMetadata returns metadata about Claude format
func (s *Strategy) GetMetadata() *domain.FormatMetadata {
	return &domain.FormatMetadata{
		Type:        domain.FormatClaude,
		DisplayName: "Claude AI Assistant",
		Description: "Single-file format for Claude AI assistant (CLAUDE.md)",
		IsDirectory: false,
	}
}

// WriteFiles handles writing rules for Claude format (single file or custom template)
func (s *Strategy) WriteFiles(rules []*domain.TransformedRule, config *domain.FormatConfig) error {
	if len(rules) == 0 {
		s.bf.LogDebug("No rules to write for Claude format")
		return nil
	}

	s.bf.LogDebug("Writing Claude format file", "rules", len(rules))
	outputPath := s.GetOutputPath(config)

	// Check if a custom template is specified
	if config != nil && config.Template != "" {
		return s.writeWithTemplate(rules, config, outputPath)
	}

	// Default behavior: write without custom template
	return s.writeWithoutTemplate(rules, outputPath)
}

// CleanupEmptyDirectories handles cleanup for Claude format (no-op since it's file-based)
func (s *Strategy) CleanupEmptyDirectories(_ *domain.FormatConfig) error {
	s.bf.LogDebug("Claude format doesn't need directory cleanup (file-based)")
	return nil
}

// CreateDirectories creates necessary directories for Claude format (no-op since it's file-based)
func (s *Strategy) CreateDirectories(_ *domain.FormatConfig) error {
	s.bf.LogDebug("Claude format doesn't need directory creation (file-based)")
	return nil
}

// writeWithTemplate processes rules using a custom template file
func (s *Strategy) writeWithTemplate(rules []*domain.TransformedRule, config *domain.FormatConfig, outputPath string) error {
	s.bf.LogDebug("Using custom template for Claude format", "template", config.Template)

	// Get template path - relative to project directory with validation
	var templatePath string
	if config.BaseDir != "" {
		templatePath = filepath.Join(config.BaseDir, config.Template)

		// Validate path is within base directory to prevent path traversal
		cleanPath, err := filepath.Abs(templatePath)
		if err != nil {
			return contextureerrors.Wrap(err, "invalid template path")
		}

		cleanBase, err := filepath.Abs(config.BaseDir)
		if err != nil {
			return contextureerrors.Wrap(err, "invalid base directory")
		}

		// Ensure template path is within base directory
		if !strings.HasPrefix(cleanPath, cleanBase+string(filepath.Separator)) && cleanPath != cleanBase {
			return contextureerrors.WithOpf("validate template path", "template path %q is outside base directory %q", config.Template, config.BaseDir)
		}
	} else {
		templatePath = config.Template
	}

	// Check if template file exists
	exists, err := s.bf.FileExists(templatePath)
	if err != nil {
		return contextureerrors.Wrap(err, "failed to check template file")
	}
	if !exists {
		s.bf.LogWarn("Template file not found, falling back to default format", "template", templatePath)
		return s.writeWithoutTemplate(rules, outputPath)
	}

	// Read template content
	templateBytes, err := s.bf.ReadFile(templatePath)
	if err != nil {
		return contextureerrors.WithOpf("read template file", "failed to read template file %s: %w", templatePath, err)
	}
	templateContent := string(templateBytes)

	// Generate rules content (same as default format but without header/footer)
	rulesContent := s.generateRulesContent(rules)

	// Process template with rules content
	variables := map[string]any{
		"Rules": rulesContent,
	}

	// Create a dummy rule for template processing (we only need the template engine functionality)
	dummyRule := &domain.Rule{ID: "template", Title: "Template Processing"}
	processedContent, err := s.bf.ProcessTemplate(dummyRule, templateContent, variables)
	if err != nil {
		return contextureerrors.Wrap(err, "failed to process template")
	}

	// Write to file
	if err := s.bf.WriteFile(outputPath, []byte(processedContent)); err != nil {
		return contextureerrors.Wrap(err, "failed to write Claude format file with template")
	}

	s.bf.LogInfo("Successfully wrote Claude format file using template", "path", outputPath, "template", config.Template, "rules", len(rules))
	return nil
}

// writeWithoutTemplate is the default write behavior
func (s *Strategy) writeWithoutTemplate(rules []*domain.TransformedRule, outputPath string) error {
	// Combine all rules into a single document
	var content strings.Builder
	content.Grow(s.estimateContentSize(rules))

	// Write header
	content.WriteString(s.getFileHeader(len(rules)))
	content.WriteString("\n\n")

	// Write rules content
	content.WriteString(s.generateRulesContent(rules))

	// Write footer
	content.WriteString("\n\n")
	content.WriteString(s.getFileFooter())

	// Write to file
	if err := s.bf.WriteFile(outputPath, []byte(content.String())); err != nil {
		return contextureerrors.Wrap(err, "failed to write Claude format file")
	}

	s.bf.LogInfo("Successfully wrote Claude format file", "path", outputPath, "rules", len(rules))
	return nil
}

// generateRulesContent creates the formatted rules content without header/footer
func (s *Strategy) generateRulesContent(rules []*domain.TransformedRule) string {
	var content strings.Builder

	for i, rule := range rules {
		if i > 0 {
			content.WriteString("\n\n---\n\n")
		}

		// Write rule content
		ruleContent := rule.Content

		// Append tracking comment using the new system, only including non-default variables
		ruleContent = s.bf.AppendTrackingCommentWithDefaults(ruleContent, rule.Rule.ID, rule.Rule.Variables, rule.Rule.DefaultVariables)

		content.WriteString(ruleContent)
	}

	return content.String()
}

// getFileHeader returns the header for the Claude format file
func (s *Strategy) getFileHeader(_ int) string {
	return "# claude.md"
}

// getFileFooter returns the footer for the Claude format file
func (s *Strategy) getFileFooter() string {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	return fmt.Sprintf("---\n\n<!-- Generated by Contexture CLI at %s -->", timestamp)
}

// estimateContentSize estimates the total size needed for the Claude format file
func (s *Strategy) estimateContentSize(rules []*domain.TransformedRule) int {
	// Start with header + footer overhead
	size := 1024

	// Add space for each rule's content plus separators
	for _, rule := range rules {
		size += len(rule.Content) + 200 // Content + tracking comment + separator overhead
	}

	return size
}

// Format implements the Claude single-file format using CommonFormat
type Format struct {
	*base.CommonFormat

	strategy *Strategy
}

// NewFormat creates a new Claude format implementation
func NewFormat(fs afero.Fs) *Format {
	bf := base.NewBaseFormat(fs, domain.FormatClaude)
	strategy := NewStrategy(fs, bf)
	commonFormat := base.NewCommonFormat(bf, strategy)

	return &Format{
		CommonFormat: commonFormat,
		strategy:     strategy,
	}
}

// NewFormatFromOptions creates a new Claude format with options
func NewFormatFromOptions(fs afero.Fs, _ map[string]any) (domain.Format, error) {
	return NewFormat(fs), nil
}

// Test helper methods to expose strategy methods
// These are used by tests to verify private implementation details

func (f *Format) getOutputPath(config *domain.FormatConfig) string {
	return f.strategy.GetOutputPath(config)
}

func (f *Format) getDefaultTemplate() string {
	return f.strategy.GetDefaultTemplate()
}

func (f *Format) getFileHeader(ruleCount int) string {
	return f.strategy.getFileHeader(ruleCount)
}

func (f *Format) getFileFooter() string {
	return f.strategy.getFileFooter()
}

func (f *Format) getOutputFilename() string {
	return defaultClaudeFilename
}

func (f *Format) generateRulesContent(rules []*domain.TransformedRule) string {
	return f.strategy.generateRulesContent(rules)
}
