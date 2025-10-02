package claude

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

// Format implements the Claude single-file format
type Format struct {
	*base.Base
}

// NewFormat creates a new Claude format implementation
func NewFormat(fs afero.Fs) *Format {
	return &Format{
		Base: base.NewBaseFormat(fs, domain.FormatClaude),
	}
}

// NewFormatFromOptions creates a new Claude format with options
func NewFormatFromOptions(fs afero.Fs, _ map[string]any) (domain.Format, error) {
	return NewFormat(fs), nil
}

// Transform converts a processed rule to Claude format representation
func (f *Format) Transform(processedRule *domain.ProcessedRule) (*domain.TransformedRule, error) {
	rule := processedRule.Rule
	f.LogDebug("Transforming processed rule for Claude format", "id", rule.ID)

	// Stage 1: Render the rule content template first
	renderedContent, err := f.ProcessTemplateWithVars(rule, rule.Content, processedRule.Variables)
	if err != nil {
		return nil, fmt.Errorf("failed to render rule content template: %w", err)
	}

	// Stage 2: Use default Claude template wrapper and include rendered content
	templateContent := f.getDefaultTemplate()

	// Copy variables and add the rendered content
	variables := make(map[string]any)
	for k, v := range processedRule.Variables {
		variables[k] = v
	}
	variables["content"] = renderedContent

	// Process the wrapper template with rendered content
	content, err := f.ProcessTemplateWithVars(rule, templateContent, variables)
	if err != nil {
		return nil, fmt.Errorf("failed to render Claude wrapper template: %w", err)
	}

	// Create transformed rule using BaseFormat
	transformed := f.CreateTransformedRule(
		rule,
		content,
		f.getOutputFilename(),
		f.getOutputFilename(),
		map[string]any{
			"format": "claude",
		},
	)

	f.LogDebug("Successfully transformed processed rule for Claude format", "id", rule.ID)
	return transformed, nil
}

// Validate checks if a rule is valid for Claude format
func (f *Format) Validate(rule *domain.Rule) (*domain.ValidationResult, error) {
	// Use BaseFormat validation
	return f.ValidateRule(rule), nil
}

// Write outputs transformed rules to the Claude format file
func (f *Format) Write(rules []*domain.TransformedRule, config *domain.FormatConfig) error {
	if len(rules) == 0 {
		f.LogDebug("No rules to write for Claude format")
		return nil
	}

	f.LogDebug("Writing Claude format file", "rules", len(rules))

	// Get output path
	outputPath := f.getOutputPath(config)

	// Check if a custom template is specified
	if config != nil && config.Template != "" {
		return f.writeWithTemplate(rules, config, outputPath)
	}

	// Default behavior: use extracted method for consistency
	return f.writeWithoutTemplate(rules, config, outputPath)
}

// Remove deletes a specific rule from the Claude format file by rebuilding it from scratch
func (f *Format) Remove(ruleID string, config *domain.FormatConfig) error {
	f.LogDebug("Removing rule from Claude format", "ruleID", ruleID)

	outputPath := f.getOutputPath(config)

	// Check if file exists
	exists, err := f.FileExists(outputPath)
	if err != nil {
		return fmt.Errorf("failed to check if file exists: %w", err)
	}
	if !exists {
		f.LogDebug("Claude format file does not exist", "path", outputPath)
		return nil
	}

	// Get current rules from the file
	currentRules, err := f.List(config)
	if err != nil {
		return fmt.Errorf("failed to list current rules: %w", err)
	}

	// Filter out the rule we want to remove
	var remainingRules []*domain.TransformedRule
	targetRulePath := f.extractBasePath(ruleID)

	for _, installedRule := range currentRules {
		currentRulePath := f.extractBasePath(installedRule.Rule.ID)
		if currentRulePath != targetRulePath {
			// Keep this rule - reuse the existing TransformedRule
			remainingRules = append(remainingRules, installedRule.TransformedRule)
		}
	}

	// Rebuild the file with remaining rules
	if len(remainingRules) == 0 {
		// No rules left, remove the file
		if err := f.RemoveFile(outputPath); err != nil {
			return fmt.Errorf("failed to remove empty Claude format file: %w", err)
		}
		f.LogInfo("Removed Claude format file (no rules remaining)", "path", outputPath)
	} else {
		// Regenerate the file with remaining rules
		if err := f.Write(remainingRules, config); err != nil {
			return fmt.Errorf("failed to regenerate Claude format file: %w", err)
		}
		f.LogInfo("Successfully regenerated Claude format file", "ruleID", ruleID, "remainingRules", len(remainingRules))
	}

	return nil
}

// List returns all currently installed rules for Claude format
func (f *Format) List(config *domain.FormatConfig) ([]*domain.InstalledRule, error) {
	f.LogDebug("Listing installed rules for Claude format")

	outputPath := f.getOutputPath(config)

	// Get file info (EAFP - will fail if file doesn't exist)
	fileInfo, err := f.GetFileInfo(outputPath)
	if err != nil {
		if os.IsNotExist(err) {
			f.LogDebug("Claude format file does not exist", "path", outputPath)
			return []*domain.InstalledRule{}, nil
		}
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}

	// Read content to parse individual rules
	content, err := f.ReadFile(outputPath)
	if err != nil {
		if os.IsNotExist(err) {
			f.LogDebug("Claude format file was deleted", "path", outputPath)
			return []*domain.InstalledRule{}, nil
		}
		return nil, fmt.Errorf("failed to read file content: %w", err)
	}

	// Parse individual rules from the Claude format file
	rules := f.parseRulesFromContent(string(content), fileInfo)

	f.LogDebug("Found rules in Claude format file", "count", len(rules))
	return rules, nil
}

// GetOutputPath returns the full output path for the Claude format file
func (f *Format) GetOutputPath(config *domain.FormatConfig) string {
	return f.getOutputPath(config)
}

// CleanupEmptyDirectories handles cleanup for Claude format (no-op since it's file-based)
func (f *Format) CleanupEmptyDirectories(_ *domain.FormatConfig) error {
	// Claude format creates a single file, not directories, so no cleanup needed
	f.LogDebug("Claude format doesn't need directory cleanup (file-based)")
	return nil
}

// CreateDirectories creates necessary directories for Claude format (no-op since it's file-based)
func (f *Format) CreateDirectories(_ *domain.FormatConfig) error {
	// Claude format creates a single file, not directories, so no directory creation needed
	f.LogDebug("Claude format doesn't need directory creation (file-based)")
	return nil
}

// GetMetadata returns metadata about Claude format
func (f *Format) GetMetadata() *domain.FormatMetadata {
	return &domain.FormatMetadata{
		Type:        domain.FormatClaude,
		DisplayName: "Claude AI Assistant",
		Description: "Single-file format for Claude AI assistant (CLAUDE.md)",
		IsDirectory: false,
	}
}

// getDefaultTemplate returns the default Claude template
func (f *Format) getDefaultTemplate() string {
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

// getFileHeader returns the header for the Claude format file
func (f *Format) getFileHeader(_ int) string {
	return "# claude.md"
}

// getFileFooter returns the footer for the Claude format file
func (f *Format) getFileFooter() string {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	return fmt.Sprintf("---\n\n<!-- Generated by Contexture CLI at %s -->", timestamp)
}

// getOutputFilename returns the default output filename
func (f *Format) getOutputFilename() string {
	return "CLAUDE.md"
}

// getOutputPath returns the full output path for the Claude format file
func (f *Format) getOutputPath(config *domain.FormatConfig) string {
	if config == nil {
		return f.getOutputFilename()
	}

	baseDir := config.BaseDir
	if baseDir == "" {
		baseDir = "."
	}

	filename := f.getOutputFilename()

	return filepath.Join(baseDir, filename)
}

// extractBasePath extracts the base rule path (without variables or brackets) for matching
func (f *Format) extractBasePath(ruleID string) string {
	rulePath := ruleID

	// Remove [contexture: prefix and ] suffix if present
	if strings.HasPrefix(rulePath, "[contexture:") {
		rulePath = strings.TrimPrefix(rulePath, "[contexture:")
		rulePath = strings.TrimSuffix(rulePath, "]")
	}

	// Remove variables part if present (path]{variables})
	if bracketIdx := strings.Index(rulePath, "]{"); bracketIdx != -1 {
		rulePath = rulePath[:bracketIdx]
	}

	return rulePath
}

// parseRulesFromContent parses individual rules from Claude format content
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
			Content:       content,
			Filename:      filepath.Base(f.getOutputFilename()),
			RelativePath:  f.getOutputFilename(),
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

// extractRuleFromSection extracts rule ID and title from a Claude format section
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

// writeWithTemplate processes rules using a custom template file
func (f *Format) writeWithTemplate(rules []*domain.TransformedRule, config *domain.FormatConfig, outputPath string) error {
	f.LogDebug("Using custom template for Claude format", "template", config.Template)

	// Get template path - relative to project directory with validation
	var templatePath string
	if config.BaseDir != "" {
		templatePath = filepath.Join(config.BaseDir, config.Template)

		// Validate path is within base directory to prevent path traversal
		cleanPath, err := filepath.Abs(templatePath)
		if err != nil {
			return fmt.Errorf("invalid template path: %w", err)
		}

		cleanBase, err := filepath.Abs(config.BaseDir)
		if err != nil {
			return fmt.Errorf("invalid base directory: %w", err)
		}

		// Ensure template path is within base directory
		if !strings.HasPrefix(cleanPath, cleanBase+string(filepath.Separator)) && cleanPath != cleanBase {
			return fmt.Errorf("template path %q is outside base directory %q", config.Template, config.BaseDir)
		}
	} else {
		templatePath = config.Template
	}

	// Check if template file exists
	exists, err := f.FileExists(templatePath)
	if err != nil {
		return fmt.Errorf("failed to check template file: %w", err)
	}
	if !exists {
		f.LogWarn("Template file not found, falling back to default format", "template", templatePath)
		return f.writeWithoutTemplate(rules, config, outputPath)
	}

	// Read template content
	templateBytes, err := f.ReadFile(templatePath)
	if err != nil {
		return fmt.Errorf("failed to read template file %s: %w", templatePath, err)
	}
	templateContent := string(templateBytes)

	// Generate rules content (same as default format but without header/footer)
	rulesContent := f.generateRulesContent(rules)

	// Process template with rules content
	variables := map[string]any{
		"Rules": rulesContent,
	}

	// Create a dummy rule for template processing (we only need the template engine functionality)
	dummyRule := &domain.Rule{ID: "template", Title: "Template Processing"}
	processedContent, err := f.ProcessTemplateWithVars(dummyRule, templateContent, variables)
	if err != nil {
		return fmt.Errorf("failed to process template: %w", err)
	}

	// Write to file using BaseFormat
	if err := f.WriteFile(outputPath, []byte(processedContent)); err != nil {
		return fmt.Errorf("failed to write Claude format file with template: %w", err)
	}

	f.LogInfo("Successfully wrote Claude format file using template", "path", outputPath, "template", config.Template, "rules", len(rules))
	return nil
}

// estimateClaudeContentSize estimates the total size needed for the Claude format file
func (f *Format) estimateClaudeContentSize(rules []*domain.TransformedRule) int {
	// Start with header + footer overhead
	size := 1024

	// Add space for each rule's content plus separators
	for _, rule := range rules {
		size += len(rule.Content) + 200 // Content + tracking comment + separator overhead
	}

	return size
}

// writeWithoutTemplate is the default write behavior (extracted for reuse)
func (f *Format) writeWithoutTemplate(rules []*domain.TransformedRule, _ *domain.FormatConfig, outputPath string) error {
	// Combine all rules into a single document
	var content strings.Builder
	content.Grow(f.estimateClaudeContentSize(rules))

	// Write header
	content.WriteString(f.getFileHeader(len(rules)))
	content.WriteString("\n\n")

	// Write rules content
	content.WriteString(f.generateRulesContent(rules))

	// Write footer
	content.WriteString("\n\n")
	content.WriteString(f.getFileFooter())

	// Write to file using BaseFormat
	if err := f.WriteFile(outputPath, []byte(content.String())); err != nil {
		return fmt.Errorf("failed to write Claude format file: %w", err)
	}

	f.LogInfo("Successfully wrote Claude format file", "path", outputPath, "rules", len(rules))
	return nil
}

// generateRulesContent creates the formatted rules content without header/footer
func (f *Format) generateRulesContent(rules []*domain.TransformedRule) string {
	var content strings.Builder

	for i, rule := range rules {
		if i > 0 {
			content.WriteString("\n\n---\n\n")
		}

		// Write rule content
		ruleContent := rule.Content

		// Append tracking comment using the new system, only including non-default variables
		ruleContent = f.AppendTrackingCommentWithDefaults(ruleContent, rule.Rule.ID, rule.Rule.Variables, rule.Rule.DefaultVariables)

		content.WriteString(ruleContent)
	}

	return content.String()
}
