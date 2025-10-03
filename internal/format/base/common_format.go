// Package base provides common format implementation shared across all formats
package base

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/contextureai/contexture/internal/domain"
)

// FormatStrategy defines format-specific behavior that varies between formats
type FormatStrategy interface {
	// GetDefaultTemplate returns the default template for this format
	GetDefaultTemplate() string

	// GetOutputPath returns the full output path for this format
	GetOutputPath(config *domain.FormatConfig) string

	// GetFileExtension returns the file extension for this format (e.g., ".md", ".mdc")
	GetFileExtension() string

	// IsSingleFile returns true if this format outputs to a single file
	IsSingleFile() bool

	// GenerateFilename generates a filename from a rule ID (for multi-file formats)
	GenerateFilename(ruleID string) string

	// GetMetadata returns metadata about this format
	GetMetadata() *domain.FormatMetadata

	// WriteFiles handles the actual file writing (can differ between single/multi-file)
	WriteFiles(rules []*domain.TransformedRule, config *domain.FormatConfig) error

	// CleanupEmptyDirectories handles cleanup of empty directories
	CleanupEmptyDirectories(config *domain.FormatConfig) error

	// CreateDirectories creates necessary directories for this format
	CreateDirectories(config *domain.FormatConfig) error
}

// CommonFormat implements shared logic for all format implementations
// It uses the strategy pattern to delegate format-specific behavior
type CommonFormat struct {
	*Base
	strategy FormatStrategy
}

// NewCommonFormat creates a new CommonFormat with a strategy
func NewCommonFormat(base *Base, strategy FormatStrategy) *CommonFormat {
	return &CommonFormat{
		Base:     base,
		strategy: strategy,
	}
}

// Transform converts a processed rule to format representation (shared logic)
// This method implements the common 2-stage template processing used by all formats:
// Stage 1: Render the rule content template with variables
// Stage 2: Wrap rendered content in format-specific template
func (cf *CommonFormat) Transform(processedRule *domain.ProcessedRule) (*domain.TransformedRule, error) {
	rule := processedRule.Rule
	cf.LogDebug("Transforming processed rule using CommonFormat", "id", rule.ID)

	// Stage 1: Render the rule content template first
	renderedContent, err := cf.ProcessTemplate(rule, rule.Content, processedRule.Variables)
	if err != nil {
		return nil, fmt.Errorf("failed to render rule content template: %w", err)
	}

	// Stage 2: Use format-specific template wrapper and include rendered content
	templateContent := cf.strategy.GetDefaultTemplate()

	// Copy variables and add the rendered content
	variables := make(map[string]any)
	if processedRule.Variables != nil {
		for k, v := range processedRule.Variables {
			variables[k] = v
		}
	}
	variables["content"] = renderedContent

	// Process the wrapper template with rendered content
	content, err := cf.ProcessTemplate(rule, templateContent, variables)
	if err != nil {
		return nil, fmt.Errorf("failed to render wrapper template: %w", err)
	}

	// Generate filename and relative path based on format strategy
	filename := cf.strategy.GenerateFilename(rule.ID)
	outputPath := cf.strategy.GetOutputPath(nil)

	// For single-file formats, GetOutputPath may return the full file path
	// Check if it already ends with the filename to avoid doubling (e.g., CLAUDE.md/CLAUDE.md)
	var relativePath string
	if filepath.Base(outputPath) == filename {
		relativePath = outputPath
	} else {
		relativePath = filepath.Join(outputPath, filename)
	}

	// Create transformed rule using BaseFormat
	metadata := map[string]any{
		"format": string(cf.formatType),
	}
	transformed := cf.CreateTransformedRule(rule, content, filename, relativePath, metadata)

	cf.LogDebug("Successfully transformed rule", "id", rule.ID, "filename", filename)
	return transformed, nil
}

// Validate checks if a rule is valid for this format
func (cf *CommonFormat) Validate(rule *domain.Rule) (*domain.ValidationResult, error) {
	// Use BaseFormat validation
	return cf.ValidateRule(rule), nil
}

// Write outputs transformed rules using format-specific write strategy
func (cf *CommonFormat) Write(rules []*domain.TransformedRule, config *domain.FormatConfig) error {
	if len(rules) == 0 {
		cf.LogDebug("No rules to write")
		return nil
	}

	cf.LogDebug("Writing rules", "count", len(rules))

	// Delegate to format-specific write implementation
	return cf.strategy.WriteFiles(rules, config)
}

// Remove deletes a specific rule from the format
// For single-file formats: rebuilds file without the rule
// For multi-file formats: deletes the individual file
func (cf *CommonFormat) Remove(ruleID string, config *domain.FormatConfig) error {
	cf.LogDebug("Removing rule", "ruleID", ruleID)

	if cf.strategy.IsSingleFile() {
		return cf.removeSingleFile(ruleID, config)
	}
	return cf.removeMultiFile(ruleID, config)
}

// removeSingleFile handles removal for single-file formats
func (cf *CommonFormat) removeSingleFile(ruleID string, config *domain.FormatConfig) error {
	outputPath := cf.strategy.GetOutputPath(config)

	// Check if file exists
	exists, err := cf.FileExists(outputPath)
	if err != nil {
		return fmt.Errorf("failed to check if file exists: %w", err)
	}
	if !exists {
		cf.LogDebug("File does not exist", "path", outputPath)
		return nil
	}

	// Get current rules from the file
	currentRules, err := cf.List(config)
	if err != nil {
		return fmt.Errorf("failed to list current rules: %w", err)
	}

	// Filter out the rule we want to remove
	var remainingRules []*domain.TransformedRule
	targetRulePath := cf.extractBasePath(ruleID)

	for _, installedRule := range currentRules {
		currentRulePath := cf.extractBasePath(installedRule.Rule.ID)
		if currentRulePath != targetRulePath {
			// Keep this rule - reuse the existing TransformedRule
			remainingRules = append(remainingRules, installedRule.TransformedRule)
		}
	}

	// Rebuild the file with remaining rules
	if len(remainingRules) == 0 {
		// No rules left, remove the file
		if err := cf.RemoveFile(outputPath); err != nil {
			return fmt.Errorf("failed to remove empty file: %w", err)
		}
		cf.LogInfo("Removed file (no rules remaining)", "path", outputPath)
	} else {
		// Regenerate the file with remaining rules
		if err := cf.Write(remainingRules, config); err != nil {
			return fmt.Errorf("failed to regenerate file: %w", err)
		}
		cf.LogInfo("Successfully regenerated file", "ruleID", ruleID, "remainingRules", len(remainingRules))
	}

	return nil
}

// removeMultiFile handles removal for multi-file formats
func (cf *CommonFormat) removeMultiFile(ruleID string, config *domain.FormatConfig) error {
	outputDir := cf.strategy.GetOutputPath(config)
	filename := cf.strategy.GenerateFilename(ruleID)
	filePath := filepath.Join(outputDir, filename)

	// Check if file exists
	exists, err := cf.FileExists(filePath)
	if err != nil {
		return fmt.Errorf("failed to check if file exists: %w", err)
	}
	if !exists {
		cf.LogDebug("Rule file does not exist", "path", filePath)
		return nil
	}

	// Remove the file
	if err := cf.RemoveFile(filePath); err != nil {
		return fmt.Errorf("failed to remove rule file: %w", err)
	}

	// Check if directory is now empty and remove it if so
	cf.CleanupEmptyDirectory(outputDir)

	cf.LogInfo("Successfully removed rule file", "ruleID", ruleID, "path", filePath)
	return nil
}

// List returns all currently installed rules for this format
func (cf *CommonFormat) List(config *domain.FormatConfig) ([]*domain.InstalledRule, error) {
	cf.LogDebug("Listing installed rules")

	if cf.strategy.IsSingleFile() {
		return cf.listSingleFile(config)
	}
	return cf.listMultiFile(config)
}

// listSingleFile lists rules from a single file
func (cf *CommonFormat) listSingleFile(config *domain.FormatConfig) ([]*domain.InstalledRule, error) {
	filePath := cf.strategy.GetOutputPath(config)

	// For some single-file formats (like Windsurf), GetOutputPath returns a directory
	// In that case, we need to append the filename
	// Check if path looks like a directory (ends with "rules" or similar, not with ".md")
	var filename string
	if !strings.HasSuffix(filePath, ".md") && !strings.HasSuffix(filePath, ".mdc") {
		filename = cf.strategy.GenerateFilename("")
		filePath = filepath.Join(filePath, filename)
	} else {
		filename = filepath.Base(filePath)
	}

	// Get file info (EAFP - will fail if file doesn't exist)
	fileInfo, err := cf.GetFileInfo(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			cf.LogDebug("File does not exist", "path", filePath)
			return []*domain.InstalledRule{}, nil
		}
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}

	// Read content to parse individual rules
	content, err := cf.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			cf.LogDebug("File was deleted", "path", filePath)
			return []*domain.InstalledRule{}, nil
		}
		return nil, fmt.Errorf("failed to read file content: %w", err)
	}

	// Parse individual rules from the file
	// For relative path, use the format's default output path (without baseDir)
	relativeDir := cf.strategy.GetOutputPath(nil)
	var relativeFilePath string
	// Check if relativeDir is already the full file path (e.g., for Claude: "CLAUDE.md")
	if filepath.Base(relativeDir) == filename {
		relativeFilePath = relativeDir
	} else {
		relativeFilePath = filepath.Join(relativeDir, filename)
	}
	rules := cf.parseRulesFromContent(string(content), fileInfo, relativeFilePath)

	cf.LogDebug("Found rules in file", "count", len(rules))
	return rules, nil
}

// listMultiFile lists rules from multiple files in a directory
func (cf *CommonFormat) listMultiFile(config *domain.FormatConfig) ([]*domain.InstalledRule, error) {
	outputDir := cf.strategy.GetOutputPath(config)

	// Read directory contents (EAFP - will fail if directory doesn't exist)
	files, err := cf.ListDirectory(outputDir)
	if err != nil {
		if os.IsNotExist(err) {
			cf.LogDebug("Directory does not exist", "path", outputDir)
			return []*domain.InstalledRule{}, nil
		}
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	var installedRules []*domain.InstalledRule
	fileExt := cf.strategy.GetFileExtension()

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		// Skip index file and other non-rule files
		if file.Name() == "index.md" || file.Name() == ".gitkeep" {
			continue
		}

		// Only process files with the correct extension
		if !strings.HasSuffix(file.Name(), fileExt) {
			continue
		}

		filePath := filepath.Join(outputDir, file.Name())

		// Read file content
		content, err := cf.ReadFile(filePath)
		if err != nil {
			cf.LogWarn("Failed to read rule file", "path", filePath, "error", err)
			continue
		}

		// Calculate content hash
		contentHash := cf.CalculateContentHash(content)

		// Extract rule ID and title from content (handles both old and new tracking comment formats)
		ruleID, title := cf.ParseRuleFromContent(string(content))

		// Fallback to filename extraction if no tracking comment found
		if ruleID == "" {
			ruleID = cf.ExtractRuleIDFromFilename(file.Name())
		}

		// Fallback to filename if no title found
		if title == "" {
			title = strings.TrimSuffix(file.Name(), fileExt)
		}

		// Create a mock rule for the transformed rule
		mockRule := &domain.Rule{
			ID:     ruleID,
			Title:  title,
			Source: "unknown",
			Ref:    "",
		}

		// Create transformed rule
		// For relative path, use the format's default output path (without baseDir)
		relativeDir := cf.strategy.GetOutputPath(nil)
		transformed := &domain.TransformedRule{
			Rule:          mockRule,
			Content:       string(content),
			Filename:      file.Name(),
			RelativePath:  filepath.Join(relativeDir, file.Name()),
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

	cf.LogDebug("Found rules", "count", len(installedRules))
	return installedRules, nil
}

// parseRulesFromContent parses individual rules from single-file content
func (cf *CommonFormat) parseRulesFromContent(
	content string,
	fileInfo os.FileInfo,
	filePath string,
) []*domain.InstalledRule {
	var rules []*domain.InstalledRule

	// Calculate content hash
	contentHash := cf.CalculateContentHash([]byte(content))

	// Split content by rule separators
	sections := strings.Split(content, "\n---\n")

	// Skip header section (first section usually contains file header)
	for i, section := range sections {
		if i == 0 {
			// Skip header section
			continue
		}

		// Parse rule ID and title from section
		ruleID, title := cf.extractRuleFromSection(section)
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
			Filename:      filepath.Base(filePath),
			RelativePath:  filePath,
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

// extractRuleFromSection extracts rule ID and title from a format section
func (cf *CommonFormat) extractRuleFromSection(section string) (string, string) {
	lines := strings.Split(section, "\n")
	var ruleID, title string

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Look for tracking comment in new format: <!-- id: [contexture:...] -->
		if strings.Contains(line, domain.RuleIDCommentPrefix) {
			if extractedRuleID, err := cf.ParseTrackingComment(line); err == nil {
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

// extractBasePath extracts the base rule path (without variables or brackets) for matching
func (cf *CommonFormat) extractBasePath(ruleID string) string {
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

// GetOutputPath returns the output path for this format
func (cf *CommonFormat) GetOutputPath(config *domain.FormatConfig) string {
	return cf.strategy.GetOutputPath(config)
}

// CleanupEmptyDirectories handles cleanup of empty directories
func (cf *CommonFormat) CleanupEmptyDirectories(config *domain.FormatConfig) error {
	return cf.strategy.CleanupEmptyDirectories(config)
}

// CreateDirectories creates necessary directories for this format
func (cf *CommonFormat) CreateDirectories(config *domain.FormatConfig) error {
	return cf.strategy.CreateDirectories(config)
}

// GetMetadata returns metadata about this format
func (cf *CommonFormat) GetMetadata() *domain.FormatMetadata {
	return cf.strategy.GetMetadata()
}
