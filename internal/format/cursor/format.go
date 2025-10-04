package cursor

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/contextureai/contexture/internal/domain"
	contextureerrors "github.com/contextureai/contexture/internal/errors"
	"github.com/contextureai/contexture/internal/format/base"
	"github.com/spf13/afero"
)

// Strategy implements the FormatStrategy interface for Cursor format
type Strategy struct {
	fs afero.Fs
	bf *base.Base
}

// NewStrategy creates a new Cursor strategy
func NewStrategy(fs afero.Fs, bf *base.Base) *Strategy {
	return &Strategy{
		fs: fs,
		bf: bf,
	}
}

// GetDefaultTemplate returns the default Cursor template with YAML frontmatter matching Cursor spec
func (s *Strategy) GetDefaultTemplate() string {
	return `---
{{if .trigger}}{{if eq .trigger.type "always"}}alwaysApply: true{{else}}alwaysApply: false{{end}}
{{if .description}}description: "{{.description}}"
{{end}}{{if and (eq .trigger.type "glob") .trigger.globs}}globs: "{{join .trigger.globs ","}}"
{{end}}{{else}}alwaysApply: false
{{end}}---

# {{.title}}

{{if .description}}{{.description}}

{{end}}{{.content}}`
}

// GetOutputPath returns the output directory path for Cursor format
func (s *Strategy) GetOutputPath(config *domain.FormatConfig) string {
	if config == nil || config.BaseDir == "" {
		return domain.CursorOutputDir
	}
	return filepath.Join(config.BaseDir, domain.CursorOutputDir)
}

// GetFileExtension returns the file extension for Cursor format (.mdc)
func (s *Strategy) GetFileExtension() string {
	return ".mdc"
}

// IsSingleFile returns false since Cursor format outputs multiple files
func (s *Strategy) IsSingleFile() bool {
	return false
}

// GenerateFilename generates a .mdc filename from rule ID
func (s *Strategy) GenerateFilename(ruleID string) string {
	filename := s.bf.GenerateFilename(ruleID)
	return strings.TrimSuffix(filename, ".md") + ".mdc"
}

// GetMetadata returns metadata about Cursor format
func (s *Strategy) GetMetadata() *domain.FormatMetadata {
	return &domain.FormatMetadata{
		Type:        domain.FormatCursor,
		DisplayName: "Cursor IDE",
		Description: "Multi-file format for Cursor IDE (.cursor/rules/)",
		IsDirectory: true,
	}
}

// WriteFiles handles writing rules for Cursor format (multi-file)
func (s *Strategy) WriteFiles(rules []*domain.TransformedRule, config *domain.FormatConfig) error {
	if len(rules) == 0 {
		s.bf.LogDebug("No rules to write for Cursor format")
		return nil
	}

	s.bf.LogDebug("Writing Cursor format files", "rules", len(rules))

	outputDir := s.GetOutputPath(config)

	// Ensure output directory exists
	if err := s.bf.EnsureDirectory(outputDir); err != nil {
		return contextureerrors.Wrap(err, "failed to create output directory")
	}

	// Write each rule to its own file
	var errors []error
	for _, rule := range rules {
		filePath := filepath.Join(outputDir, rule.Filename)

		// Append tracking comment at the end instead of header at beginning, only including non-default variables
		content := s.bf.AppendTrackingCommentWithDefaults(rule.Content, rule.Rule.ID, rule.Rule.Variables, rule.Rule.DefaultVariables)

		if err := s.bf.WriteFile(filePath, []byte(content)); err != nil {
			errors = append(errors, contextureerrors.WithOpf("failed to write rule", "%s: %w", rule.Rule.ID, err))
			continue
		}

		s.bf.LogDebug("Wrote Cursor rule file", "ruleID", rule.Rule.ID, "path", filePath)
	}

	if len(errors) > 0 {
		return contextureerrors.WithOpf("WriteFiles", "failed to write %d rules: %v", len(errors), errors)
	}

	s.bf.LogInfo("Successfully wrote Cursor format files", "count", len(rules), "directory", outputDir)
	return nil
}

// CleanupEmptyDirectories handles cleanup of empty directories for Cursor format
func (s *Strategy) CleanupEmptyDirectories(config *domain.FormatConfig) error {
	outputDir := s.GetOutputPath(config)

	baseDir := config.BaseDir
	if baseDir == "" {
		baseDir = "."
	}
	parentDir := filepath.Join(baseDir, ".cursor")

	// First clean up the rules directory
	s.bf.CleanupEmptyDirectory(outputDir)
	// Then clean up the parent .cursor directory if it's also empty
	s.bf.CleanupEmptyDirectory(parentDir)

	return nil
}

// CreateDirectories creates necessary directories for Cursor format
func (s *Strategy) CreateDirectories(config *domain.FormatConfig) error {
	outputDir := s.GetOutputPath(config)
	return s.bf.EnsureDirectory(outputDir)
}

// Format implements the Cursor multi-file format using CommonFormat
type Format struct {
	*base.CommonFormat

	strategy *Strategy
}

// NewFormat creates a new Cursor format implementation
func NewFormat(fs afero.Fs) *Format {
	bf := base.NewBaseFormat(fs, domain.FormatCursor)
	strategy := NewStrategy(fs, bf)
	commonFormat := base.NewCommonFormat(bf, strategy)

	return &Format{
		CommonFormat: commonFormat,
		strategy:     strategy,
	}
}

// NewFormatFromOptions creates a new Cursor format with options
func NewFormatFromOptions(fs afero.Fs, _ map[string]any) (domain.Format, error) {
	return NewFormat(fs), nil
}

// GenerateFilename generates a .mdc filename from rule ID (overrides BaseFormat method)
func (f *Format) GenerateFilename(ruleID string) string {
	return f.strategy.GenerateFilename(ruleID)
}

// ExtractRuleIDFromFilename extracts rule ID from .mdc filename
func (f *Format) ExtractRuleIDFromFilename(filename string) string {
	base := strings.TrimSuffix(filename, ".mdc")
	path := strings.ReplaceAll(base, "-", "/")
	return fmt.Sprintf("[contexture:%s]", path)
}

// GetDefaultTemplate returns the default template for the format.
func (f *Format) GetDefaultTemplate() string {
	return f.strategy.GetDefaultTemplate()
}

// Test helper methods to expose strategy methods
// These are used by tests to verify private implementation details

func (f *Format) getOutputDir(config *domain.FormatConfig) string {
	return f.strategy.GetOutputPath(config)
}
