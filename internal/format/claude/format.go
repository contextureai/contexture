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

	// Combine all rules into a single document
	var content strings.Builder

	// Write header
	content.WriteString(f.getFileHeader(len(rules)))
	content.WriteString("\n\n")

	// Write each rule
	for i, rule := range rules {
		if i > 0 {
			content.WriteString("\n\n---\n\n")
		}

		// Write rule content
		ruleContent := rule.Content

		// Append tracking comment using the new system
		ruleContent = f.AppendTrackingComment(ruleContent, rule.Rule.ID, rule.Rule.Variables)

		content.WriteString(ruleContent)
	}

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

// Remove deletes a specific rule from the Claude format file
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

	// Read current content
	content, err := f.ReadFile(outputPath)
	if err != nil {
		return fmt.Errorf("failed to read Claude format file: %w", err)
	}

	// Remove the rule from content by parsing sections
	contentStr := string(content)
	updatedContent := f.removeRuleFromContent(contentStr, ruleID)

	// Write back the updated content
	if err := f.WriteFile(outputPath, []byte(updatedContent)); err != nil {
		return fmt.Errorf("failed to write updated Claude format file: %w", err)
	}

	f.LogInfo("Successfully removed rule from Claude format", "ruleID", ruleID)
	return nil
}

// List returns all currently installed rules for Claude format
func (f *Format) List(config *domain.FormatConfig) ([]*domain.InstalledRule, error) {
	f.LogDebug("Listing installed rules for Claude format")

	outputPath := f.getOutputPath(config)

	// Check if file exists
	exists, err := f.FileExists(outputPath)
	if err != nil {
		return nil, fmt.Errorf("failed to check if file exists: %w", err)
	}
	if !exists {
		f.LogDebug("Claude format file does not exist", "path", outputPath)
		return []*domain.InstalledRule{}, nil
	}

	// Get file info
	fileInfo, err := f.GetFileInfo(outputPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}

	// Read content to parse individual rules
	content, err := f.ReadFile(outputPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file content: %w", err)
	}

	// Parse individual rules from the Claude format file
	rules := f.parseRulesFromContent(string(content), fileInfo)

	f.LogDebug("Found rules in Claude format file", "count", len(rules))
	return rules, nil
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
{{end}}
{{.content}}`
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

// removeRuleFromContent removes a specific rule from Claude format content
func (f *Format) removeRuleFromContent(content, ruleID string) string {
	// Split content by rule separators
	sections := strings.Split(content, "\n---\n")
	var filteredSections []string

	for _, section := range sections {
		// Check if this section contains the tracking comment for the rule we want to remove
		// The ruleID comes in as just the path (e.g., "core/typescript/strict-config")
		// but we need to format it as the full rule ID for the tracking comment
		fullRuleID := fmt.Sprintf("[contexture:%s]", ruleID)
		trackingComment := f.CreateTrackingComment(fullRuleID, nil)
		if !strings.Contains(section, trackingComment) {
			filteredSections = append(filteredSections, section)
		}
	}

	return strings.Join(filteredSections, "\n---\n")
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
