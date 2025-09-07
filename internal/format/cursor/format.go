package cursor

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/contextureai/contexture/internal/domain"
	"github.com/contextureai/contexture/internal/format/base"
	"github.com/spf13/afero"
)

// Format implements the Cursor multi-file format
type Format struct {
	*base.Base
}

// NewFormat creates a new Cursor format implementation
func NewFormat(fs afero.Fs) *Format {
	return &Format{
		Base: base.NewBaseFormat(fs, domain.FormatCursor),
	}
}

// NewFormatFromOptions creates a new Cursor format with options
func NewFormatFromOptions(fs afero.Fs, _ map[string]any) (domain.Format, error) {
	return NewFormat(fs), nil
}

// Transform converts a processed rule to Cursor format representation
func (f *Format) Transform(processedRule *domain.ProcessedRule) (*domain.TransformedRule, error) {
	rule := processedRule.Rule
	f.LogDebug("Transforming processed rule for Cursor format", "id", rule.ID)

	// Stage 1: Render the rule content template first
	renderedContent, err := f.ProcessTemplateWithVars(rule, rule.Content, processedRule.Variables)
	if err != nil {
		return nil, fmt.Errorf("failed to render rule content template: %w", err)
	}

	// Stage 2: Use default Cursor template wrapper and include rendered content
	templateContent := f.GetDefaultTemplate()

	// Copy variables and add the rendered content
	variables := make(map[string]any)
	for k, v := range processedRule.Variables {
		variables[k] = v
	}
	variables["content"] = renderedContent

	// Process the wrapper template with rendered content
	content, err := f.ProcessTemplateWithVars(rule, templateContent, variables)
	if err != nil {
		return nil, fmt.Errorf("failed to render Cursor wrapper template: %w", err)
	}

	// Generate filename from rule ID using Cursor format's .mdc extension
	filename := f.GenerateFilename(rule.ID)
	relativePath := filepath.Join(domain.CursorOutputDir, filename)

	// Create transformed rule using BaseFormat
	transformed := f.CreateTransformedRule(rule, content, filename, relativePath, map[string]any{
		"format":    "cursor",
		"outputDir": domain.CursorOutputDir,
	})

	f.LogDebug(
		"Successfully transformed processed rule for Cursor format",
		"id",
		rule.ID,
		"filename",
		filename,
	)
	return transformed, nil
}

// Validate checks if a rule is valid for Cursor format
func (f *Format) Validate(rule *domain.Rule) (*domain.ValidationResult, error) {
	// Use BaseFormat validation
	return f.ValidateRule(rule), nil
}

// Write outputs transformed rules to the Cursor format directory
func (f *Format) Write(rules []*domain.TransformedRule, config *domain.FormatConfig) error {
	if len(rules) == 0 {
		f.LogDebug("No rules to write for Cursor format")
		return nil
	}

	f.LogDebug("Writing Cursor format files", "rules", len(rules))

	outputDir := f.getOutputDir(config)

	// Ensure output directory exists using BaseFormat
	if err := f.EnsureDirectory(outputDir); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Write each rule to its own file
	var errors []error
	for _, rule := range rules {
		filePath := filepath.Join(outputDir, rule.Filename)

		// Append tracking comment at the end instead of header at beginning
		content := f.AppendTrackingComment(rule.Content, rule.Rule.ID, rule.Rule.Variables)

		if err := f.WriteFile(filePath, []byte(content)); err != nil {
			errors = append(errors, fmt.Errorf("failed to write rule %s: %w", rule.Rule.ID, err))
			continue
		}

		f.LogDebug("Wrote Cursor rule file", "ruleID", rule.Rule.ID, "path", filePath)
	}

	if len(errors) > 0 {
		return fmt.Errorf("failed to write %d rules: %v", len(errors), errors)
	}

	f.LogInfo("Successfully wrote Cursor format files", "count", len(rules), "directory", outputDir)
	return nil
}

// Remove deletes a specific rule from the Cursor format directory
func (f *Format) Remove(ruleID string, config *domain.FormatConfig) error {
	f.LogDebug("Removing rule from Cursor format", "ruleID", ruleID)

	outputDir := f.getOutputDir(config)
	filename := f.GenerateFilename(ruleID)
	filePath := filepath.Join(outputDir, filename)

	// Check if file exists using BaseFormat
	exists, err := f.FileExists(filePath)
	if err != nil {
		return fmt.Errorf("failed to check if file exists: %w", err)
	}
	if !exists {
		f.LogDebug("Cursor rule file does not exist", "path", filePath)
		return nil
	}

	// Remove the file using BaseFormat
	if err := f.RemoveFile(filePath); err != nil {
		return fmt.Errorf("failed to remove rule file: %w", err)
	}

	// Check if directory is now empty and remove it if so
	f.CleanupEmptyDirectory(outputDir)

	f.LogInfo("Successfully removed Cursor rule file", "ruleID", ruleID, "path", filePath)
	return nil
}

// List returns all currently installed rules for Cursor format
func (f *Format) List(config *domain.FormatConfig) ([]*domain.InstalledRule, error) {
	f.LogDebug("Listing installed rules for Cursor format")

	outputDir := f.getOutputDir(config)

	// Check if directory exists using BaseFormat
	exists, err := f.DirExists(outputDir)
	if err != nil {
		return nil, fmt.Errorf("failed to check if directory exists: %w", err)
	}
	if !exists {
		f.LogDebug("Cursor format directory does not exist", "path", outputDir)
		return []*domain.InstalledRule{}, nil
	}

	// Read directory contents using BaseFormat
	files, err := f.ListDirectory(outputDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	var installedRules []*domain.InstalledRule
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		// Skip index file and other non-rule files
		if file.Name() == "index.md" || file.Name() == ".gitkeep" {
			continue
		}

		// Only process .mdc files (Cursor format)
		if !strings.HasSuffix(file.Name(), ".mdc") {
			continue
		}

		filePath := filepath.Join(outputDir, file.Name())

		// Read file content using BaseFormat
		content, err := f.ReadFile(filePath)
		if err != nil {
			f.LogWarn("Failed to read rule file", "path", filePath, "error", err)
			continue
		}

		// Calculate content hash using BaseFormat
		contentHash := f.CalculateContentHash(content)

		// Extract rule ID from tracking comment in content
		trackingComments := f.ExtractTrackingComments(string(content))
		var ruleID string
		if len(trackingComments) > 0 {
			ruleID = trackingComments[0] // Use the first tracking comment found
		} else {
			// Fallback to filename extraction for backward compatibility
			ruleID = f.ExtractRuleIDFromFilename(file.Name())
		}

		// Try to extract title from content using BaseFormat
		title := f.ExtractTitleFromContent(string(content))
		if title == "" {
			title = strings.TrimSuffix(file.Name(), ".mdc")
		}

		// Create a mock rule for the transformed rule
		mockRule := &domain.Rule{
			ID:     ruleID,
			Title:  title,
			Source: "unknown", // Could be extracted from file content metadata
			Ref:    "",
		}

		// Create transformed rule
		transformed := &domain.TransformedRule{
			Rule:          mockRule,
			Content:       string(content),
			Filename:      file.Name(),
			RelativePath:  filepath.Join(domain.CursorOutputDir, file.Name()),
			TransformedAt: file.ModTime(),
			Size:          file.Size(),
			ContentHash:   contentHash,
		}

		// Convert to installed rule
		installedRule := &domain.InstalledRule{
			TransformedRule: transformed,
			InstalledAt:     file.ModTime(),
		}

		installedRules = append(installedRules, installedRule)
	}

	f.LogDebug("Found Cursor format rules", "count", len(installedRules))
	return installedRules, nil
}

// GetDefaultTemplate returns the default Cursor template with YAML frontmatter matching Cursor spec
func (f *Format) GetDefaultTemplate() string {
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

// GenerateFilename generates a .mdc filename from rule ID (overrides BaseFormat method)
func (f *Format) GenerateFilename(ruleID string) string {
	filename := f.Base.GenerateFilename(ruleID)
	return strings.TrimSuffix(filename, ".md") + ".mdc"
}

// ExtractRuleIDFromFilename extracts rule ID from .mdc filename (overrides BaseFormat method)
func (f *Format) ExtractRuleIDFromFilename(filename string) string {
	base := strings.TrimSuffix(filename, ".mdc")
	path := strings.ReplaceAll(base, "-", "/")
	return fmt.Sprintf("[contexture:%s]", path)
}

// getOutputDir returns the output directory for Cursor format
func (f *Format) getOutputDir(config *domain.FormatConfig) string {
	if config == nil || config.BaseDir == "" {
		return domain.CursorOutputDir
	}
	return filepath.Join(config.BaseDir, domain.CursorOutputDir)
}

// GetOutputPath returns the output directory path for Cursor format
func (f *Format) GetOutputPath(config *domain.FormatConfig) string {
	return f.getOutputDir(config)
}

// CleanupEmptyDirectories handles cleanup of empty directories for Cursor format
func (f *Format) CleanupEmptyDirectories(config *domain.FormatConfig) error {
	outputDir := f.getOutputDir(config)
	
	baseDir := config.BaseDir
	if baseDir == "" {
		baseDir = "."
	}
	parentDir := filepath.Join(baseDir, ".cursor")
	
	// First clean up the rules directory
	f.CleanupEmptyDirectory(outputDir)
	// Then clean up the parent .cursor directory if it's also empty
	f.CleanupEmptyDirectory(parentDir)
	
	return nil
}

// CreateDirectories creates necessary directories for Cursor format
func (f *Format) CreateDirectories(config *domain.FormatConfig) error {
	outputDir := f.getOutputDir(config)
	return f.EnsureDirectory(outputDir)
}

// GetMetadata returns metadata about Cursor format
func (f *Format) GetMetadata() *domain.FormatMetadata {
	return &domain.FormatMetadata{
		Type:        domain.FormatCursor,
		DisplayName: "Cursor IDE",
		Description: "Multi-file format for Cursor IDE (.cursor/rules/)",
		IsDirectory: true,
	}
}
