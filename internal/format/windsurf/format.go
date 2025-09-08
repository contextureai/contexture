package windsurf

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

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
)

// Format implements the Windsurf format with support for both single and multi-file modes
type Format struct {
	*base.Base

	mode OutputMode
}

// NewFormat creates a new Windsurf format implementation
func NewFormat(fs afero.Fs) *Format {
	return &Format{
		Base: base.NewBaseFormat(fs, domain.FormatWindsurf),
		mode: ModeMultiFile, // Default to multi-file mode
	}
}

// NewFormatWithMode creates a new Windsurf format implementation with specified mode
func NewFormatWithMode(fs afero.Fs, mode OutputMode) *Format {
	return &Format{
		Base: base.NewBaseFormat(fs, domain.FormatWindsurf),
		mode: mode,
	}
}

// SetMode sets the output mode for the format
func (f *Format) SetMode(mode OutputMode) {
	f.mode = mode
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

// GetMode returns the current output mode
func (f *Format) GetMode() OutputMode {
	return f.mode
}

// Transform converts a processed rule to Windsurf format representation
func (f *Format) Transform(processedRule *domain.ProcessedRule) (*domain.TransformedRule, error) {
	rule := processedRule.Rule
	f.LogDebug("Transforming processed rule for Windsurf format", "id", rule.ID, "mode", f.mode)

	// Stage 1: Render the rule content template first
	renderedContent, err := f.ProcessTextTemplateWithVars(rule, rule.Content, processedRule.Variables)
	if err != nil {
		return nil, fmt.Errorf("failed to render rule content template: %w", err)
	}

	// Stage 2: Use default Windsurf template wrapper and include rendered content
	templateContent := f.GetDefaultTemplate()

	// Copy variables and add the rendered content and mode
	variables := make(map[string]any)
	if processedRule.Variables != nil {
		for k, v := range processedRule.Variables {
			variables[k] = v
		}
	}
	variables["content"] = renderedContent
	variables["mode"] = string(f.mode)

	// Process the wrapper template with rendered content (using text template to avoid HTML escaping)
	content, err := f.ProcessTextTemplateWithVars(rule, templateContent, variables)
	if err != nil {
		return nil, fmt.Errorf("failed to render Windsurf wrapper template: %w", err)
	}

	// Generate filename and path based on mode
	var filename, relativePath string
	if f.mode == ModeSingleFile {
		filename = f.GetSingleFileFilename()
		relativePath = filepath.Join(domain.WindsurfOutputDir, filename)
	} else {
		filename = f.GenerateFilename(rule.ID)
		relativePath = filepath.Join(domain.WindsurfOutputDir, filename)
	}

	// Create transformed rule using BaseFormat
	transformed := f.CreateTransformedRule(rule, content, filename, relativePath, map[string]any{
		"format":    "windsurf",
		"mode":      string(f.mode),
		"outputDir": domain.WindsurfOutputDir,
	})

	f.LogDebug(
		"Successfully transformed rule for Windsurf format",
		"id",
		rule.ID,
		"filename",
		filename,
		"mode",
		f.mode,
	)
	return transformed, nil
}

// Validate checks if a rule is valid for Windsurf format
func (f *Format) Validate(rule *domain.Rule) (*domain.ValidationResult, error) {
	// Use BaseFormat validation and add mode metadata
	result := f.ValidateRule(rule)
	result.Metadata["mode"] = string(f.mode)

	// Add Windsurf-specific character limit validation
	contentLength := len(rule.Content)
	if contentLength > domain.WindsurfMaxSingleRuleChars {
		result.Errors = append(result.Errors, fmt.Errorf(
			"rule content exceeds Windsurf limit of %d characters (current: %d)",
			domain.WindsurfMaxSingleRuleChars,
			contentLength,
		))
		result.Valid = false
	}

	return result, nil
}

// Write outputs transformed rules to the Windsurf format
func (f *Format) Write(rules []*domain.TransformedRule, config *domain.FormatConfig) error {
	if len(rules) == 0 {
		f.LogDebug("No rules to write for Windsurf format")
		return nil
	}

	f.LogDebug("Writing Windsurf format files", "rules", len(rules), "mode", f.mode)

	// Check character limits for each rule individually
	for _, rule := range rules {
		if len(rule.Content) > domain.WindsurfMaxSingleRuleChars {
			return fmt.Errorf(
				"rule '%s' exceeds Windsurf per-file limit of %d characters (current: %d)",
				rule.Rule.ID,
				domain.WindsurfMaxSingleRuleChars,
				len(rule.Content),
			)
		}
	}

	outputDir := f.getOutputDir(config)

	// Ensure output directory exists using BaseFormat
	if err := f.EnsureDirectory(outputDir); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	if f.mode == ModeSingleFile {
		return f.writeSingleFile(rules, outputDir)
	}
	return f.writeMultiFile(rules, outputDir)
}

// GetDefaultTemplate returns the default Windsurf template with YAML frontmatter matching Windsurf spec
func (f *Format) GetDefaultTemplate() string {
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

// GetSingleFileFilename returns the filename for single file mode
func (f *Format) GetSingleFileFilename() string {
	return "rules.md"
}

// List returns all currently installed rules for Windsurf format
func (f *Format) List(config *domain.FormatConfig) ([]*domain.InstalledRule, error) {
	f.LogDebug("Listing installed rules for Windsurf format", "mode", f.mode)

	outputDir := f.getOutputDir(config)

	// Check if directory exists using BaseFormat
	exists, err := f.DirExists(outputDir)
	if err != nil {
		return nil, fmt.Errorf("failed to check if directory exists: %w", err)
	}
	if !exists {
		f.LogDebug("Windsurf format directory does not exist", "path", outputDir)
		return []*domain.InstalledRule{}, nil
	}

	if f.mode == ModeSingleFile {
		return f.listSingleFile(outputDir)
	}
	return f.listMultiFile(outputDir)
}

// Remove deletes a specific rule from the Windsurf format
func (f *Format) Remove(ruleID string, config *domain.FormatConfig) error {
	f.LogDebug("Removing rule from Windsurf format", "ruleID", ruleID, "mode", f.mode)

	outputDir := f.getOutputDir(config)

	if f.mode == ModeSingleFile {
		return f.removeSingleFile(ruleID, outputDir)
	}
	return f.removeMultiFile(ruleID, outputDir)
}

// GetOutputPath returns the output directory path for Windsurf format
func (f *Format) GetOutputPath(config *domain.FormatConfig) string {
	return f.getOutputDir(config)
}

// CleanupEmptyDirectories handles cleanup of empty directories for Windsurf format
func (f *Format) CleanupEmptyDirectories(config *domain.FormatConfig) error {
	outputDir := f.getOutputDir(config)

	baseDir := config.BaseDir
	if baseDir == "" {
		baseDir = "."
	}
	parentDir := filepath.Join(baseDir, ".windsurf")

	// First clean up the rules directory
	f.CleanupEmptyDirectory(outputDir)
	// Then clean up the parent .windsurf directory if it's also empty
	f.CleanupEmptyDirectory(parentDir)

	return nil
}

// CreateDirectories creates necessary directories for Windsurf format
func (f *Format) CreateDirectories(config *domain.FormatConfig) error {
	outputDir := f.getOutputDir(config)
	return f.EnsureDirectory(outputDir)
}

// GetMetadata returns metadata about Windsurf format
func (f *Format) GetMetadata() *domain.FormatMetadata {
	return &domain.FormatMetadata{
		Type:        domain.FormatWindsurf,
		DisplayName: "Windsurf IDE",
		Description: "Multi-file format for Windsurf IDE (.windsurf/rules/)",
		IsDirectory: true,
	}
}

// writeSingleFile writes all rules to a single file
func (f *Format) writeSingleFile(rules []*domain.TransformedRule, outputDir string) error {
	filename := f.GetSingleFileFilename()
	filePath := filepath.Join(outputDir, filename)

	var content strings.Builder

	// Write header
	content.WriteString(f.getSingleFileHeader(len(rules)))
	content.WriteString("\n\n")

	// Write each rule
	for i, rule := range rules {
		if i > 0 {
			content.WriteString("\n\n---\n\n")
		}

		// Write rule content with tracking comment appended
		ruleContent := f.AppendTrackingComment(rule.Content, rule.Rule.ID, rule.Rule.Variables)
		content.WriteString(ruleContent)
	}

	// Write footer
	content.WriteString("\n\n")
	content.WriteString(f.getSingleFileFooter())

	// Write to file using BaseFormat
	if err := f.WriteFile(filePath, []byte(content.String())); err != nil {
		return fmt.Errorf("failed to write Windsurf single file: %w", err)
	}

	f.LogInfo("Successfully wrote Windsurf single file", "path", filePath, "rules", len(rules))
	return nil
}

// writeMultiFile writes each rule to its own file
func (f *Format) writeMultiFile(rules []*domain.TransformedRule, outputDir string) error {
	var errors []error

	// Write each rule to its own file
	for _, rule := range rules {
		filePath := filepath.Join(outputDir, rule.Filename)

		// Append tracking comment at the end instead of header at beginning
		content := f.AppendTrackingComment(rule.Content, rule.Rule.ID, rule.Rule.Variables)

		if err := f.WriteFile(filePath, []byte(content)); err != nil {
			errors = append(errors, fmt.Errorf("failed to write rule %s: %w", rule.Rule.ID, err))
			continue
		}

		f.LogDebug("Wrote Windsurf rule file", "ruleID", rule.Rule.ID, "path", filePath)
	}

	if len(errors) > 0 {
		return fmt.Errorf("failed to write %d rules: %v", len(errors), errors)
	}

	f.LogInfo(
		"Successfully wrote Windsurf multi-file format",
		"count",
		len(rules),
		"directory",
		outputDir,
	)
	return nil
}

// removeSingleFile removes a rule from the single file
func (f *Format) removeSingleFile(ruleID string, outputDir string) error {
	filename := f.GetSingleFileFilename()
	filePath := filepath.Join(outputDir, filename)

	// Check if file exists using BaseFormat
	exists, err := f.FileExists(filePath)
	if err != nil {
		return fmt.Errorf("failed to check if file exists: %w", err)
	}
	if !exists {
		f.LogDebug("Windsurf single file does not exist", "path", filePath)
		return nil
	}

	// Read current content
	content, err := f.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read Windsurf single file: %w", err)
	}

	// Remove the rule from content by parsing sections
	contentStr := string(content)
	updatedContent := f.removeRuleFromContent(contentStr, ruleID)

	// Write back the updated content
	if err := f.WriteFile(filePath, []byte(updatedContent)); err != nil {
		return fmt.Errorf("failed to write updated Windsurf single file: %w", err)
	}

	f.LogInfo("Successfully removed rule from Windsurf single file", "ruleID", ruleID)
	return nil
}

// removeMultiFile removes a rule file from multi-file mode
func (f *Format) removeMultiFile(ruleID string, outputDir string) error {
	filename := f.GenerateFilename(ruleID)
	filePath := filepath.Join(outputDir, filename)

	// Check if file exists using BaseFormat
	exists, err := f.FileExists(filePath)
	if err != nil {
		return fmt.Errorf("failed to check if file exists: %w", err)
	}
	if !exists {
		f.LogDebug("Windsurf rule file does not exist", "path", filePath)
		return nil
	}

	// Remove the file using BaseFormat
	if err := f.RemoveFile(filePath); err != nil {
		return fmt.Errorf("failed to remove rule file: %w", err)
	}

	// Check if directory is now empty and remove it if so
	f.CleanupEmptyDirectory(outputDir)

	f.LogInfo("Successfully removed Windsurf rule file", "ruleID", ruleID, "path", filePath)
	return nil
}

// listSingleFile lists rules from single file mode
func (f *Format) listSingleFile(outputDir string) ([]*domain.InstalledRule, error) {
	filename := f.GetSingleFileFilename()
	filePath := filepath.Join(outputDir, filename)

	// Check if file exists using BaseFormat
	exists, err := f.FileExists(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to check if file exists: %w", err)
	}
	if !exists {
		return []*domain.InstalledRule{}, nil
	}

	// Get file info using BaseFormat
	fileInfo, err := f.GetFileInfo(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}

	// Read content using BaseFormat
	content, err := f.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file content: %w", err)
	}

	// Parse individual rules from the Windsurf format file
	rules := f.parseRulesFromContent(string(content), fileInfo)

	return rules, nil
}

// listMultiFile lists rules from multi-file mode
func (f *Format) listMultiFile(outputDir string) ([]*domain.InstalledRule, error) {
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

		// Only process .md files
		if !strings.HasSuffix(file.Name(), ".md") {
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
			title = strings.TrimSuffix(file.Name(), ".md")
		}

		// Create a mock rule for the transformed rule
		mockRule := &domain.Rule{
			ID:     ruleID,
			Title:  title,
			Source: "unknown",
			Ref:    "",
		}

		// Create transformed rule
		transformed := &domain.TransformedRule{
			Rule:          mockRule,
			Content:       string(content),
			Filename:      file.Name(),
			RelativePath:  filepath.Join(domain.WindsurfOutputDir, file.Name()),
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

	return installedRules, nil
}

// getSingleFileHeader returns the header for single file mode
func (f *Format) getSingleFileHeader(ruleCount int) string {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	return fmt.Sprintf(`# Windsurf Rules

This file contains %d contexture rules for Windsurf AI assistant.

Generated at: %s
Mode: Single File`, ruleCount, timestamp)
}

// getSingleFileFooter returns the footer for single file mode
func (f *Format) getSingleFileFooter() string {
	return `---

*This file was generated by Contexture CLI in single-file mode. Do not edit manually.*`
}

// getOutputDir returns the output directory for Windsurf format
func (f *Format) getOutputDir(config *domain.FormatConfig) string {
	if config == nil || config.BaseDir == "" {
		return domain.WindsurfOutputDir
	}
	return filepath.Join(config.BaseDir, domain.WindsurfOutputDir)
}

// removeRuleFromContent removes a specific rule from Windsurf format content
func (f *Format) removeRuleFromContent(content, ruleID string) string {
	// Split content by rule separators
	sections := strings.Split(content, "\n---\n")
	var filteredSections []string

	for _, section := range sections {
		// Check if this section contains the tracking comment for the rule we want to remove
		trackingComment := f.CreateTrackingComment(ruleID, nil)
		if !strings.Contains(section, trackingComment) {
			filteredSections = append(filteredSections, section)
		}
	}

	return strings.Join(filteredSections, "\n---\n")
}

// parseRulesFromContent parses individual rules from Windsurf format content
func (f *Format) parseRulesFromContent(
	content string,
	fileInfo os.FileInfo,
) []*domain.InstalledRule {
	var rules []*domain.InstalledRule

	// Calculate content hash
	contentHash := f.CalculateContentHash([]byte(content))

	// Split content by rule separators
	sections := strings.Split(content, "\n---\n")

	// Skip header section (first section usually contains file header)
	for i, section := range sections {
		if i == 0 {
			// Skip header section
			continue
		}

		// Parse rule ID and title from section
		ruleID, title := f.extractRuleFromSection(section)
		if ruleID == "" {
			continue
		}

		// Create a mock rule for the transformed rule
		mockRule := &domain.Rule{
			ID:     ruleID,
			Title:  title,
			Source: "local",
			Ref:    "",
		}

		// Create transformed rule
		transformed := &domain.TransformedRule{
			Rule:          mockRule,
			Content:       section,
			Filename:      filepath.Base(f.GetSingleFileFilename()),
			RelativePath:  filepath.Join(domain.WindsurfOutputDir, f.GetSingleFileFilename()),
			TransformedAt: fileInfo.ModTime(),
			Size:          fileInfo.Size(),
			ContentHash:   contentHash,
		}

		// Convert to installed rule
		rule := &domain.InstalledRule{
			TransformedRule: transformed,
			InstalledAt:     fileInfo.ModTime(),
		}

		rules = append(rules, rule)
	}

	return rules
}

// extractRuleFromSection extracts rule ID and title from a Windsurf format section
func (f *Format) extractRuleFromSection(section string) (string, string) {
	lines := strings.Split(section, "\n")
	var ruleID, title string

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Look for tracking comment in new format: <!-- id: [contexture:...] -->
		if strings.Contains(line, domain.RuleIDCommentPrefix) {
			if extractedRuleID, err := f.ParseTrackingComment(line); err == nil {
				ruleID = extractedRuleID
			}
		}

		// Look for title in markdown header
		if strings.HasPrefix(line, "# ") && title == "" {
			title = strings.TrimSpace(line[2:])
		}
	}

	return ruleID, title
}
