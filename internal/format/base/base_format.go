// Package base provides the base format implementation for all output formats.
package base

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/log"
	"github.com/contextureai/contexture/internal/domain"
	contextureerrors "github.com/contextureai/contexture/internal/errors"
	"github.com/contextureai/contexture/internal/rule"
	"github.com/contextureai/contexture/internal/template"
	"github.com/spf13/afero"
)

// Base provides common functionality for all format implementations
type Base struct {
	fs             afero.Fs
	templateEngine template.Engine
	formatType     domain.FormatType
}

// NewBaseFormat creates a new base format
func NewBaseFormat(fs afero.Fs, formatType domain.FormatType) *Base {
	return &Base{
		fs:             fs,
		templateEngine: template.NewEngine(),
		formatType:     formatType,
	}
}

// ValidateRule performs common validation for all formats
func (bf *Base) ValidateRule(rule *domain.Rule) *domain.ValidationResult {
	result := &domain.ValidationResult{
		Valid:    true,
		Errors:   []error{},
		Warnings: []domain.ValidationWarning{},
		Metadata: map[string]any{
			"format": string(bf.formatType),
		},
	}

	// Check required fields
	if rule.ID == "" {
		result.AddError("id", "rule ID is required", "REQUIRED_FIELD")
	}

	if rule.Title == "" {
		result.AddError("title", "rule title is required", "REQUIRED_FIELD")
	}

	if rule.Content == "" {
		result.AddError("content", "rule content is required", "REQUIRED_FIELD")
	}

	// Check for warnings
	if rule.Description == "" {
		result.Warnings = append(result.Warnings, domain.ValidationWarning{
			Field:   "description",
			Message: "rule description is recommended for better understanding",
			Code:    "RECOMMENDED_FIELD",
		})
	}

	if len(rule.Tags) == 0 {
		result.Warnings = append(result.Warnings, domain.ValidationWarning{
			Field:   "tags",
			Message: "tags are recommended for better organization",
			Code:    "RECOMMENDED_FIELD",
		})
	}

	// Set validity based on errors
	result.Valid = len(result.Errors) == 0

	return result
}

// ProcessTemplate processes template content with common variables and optional additional variables
// This is the unified template processing method that consolidates all template rendering logic.
func (bf *Base) ProcessTemplate(
	rule *domain.Rule,
	templateContent string,
	additionalVars ...map[string]any,
) (string, error) {
	// Create base template variables
	variables := map[string]any{
		"rule":        rule,
		"id":          rule.ID,
		"title":       rule.Title,
		"description": rule.Description,
		"tags":        rule.Tags,
		"content":     rule.Content,
		"source":      rule.Source,
		"ref":         rule.Ref,
		"languages":   rule.Languages,
		"frameworks":  rule.Frameworks,
	}

	// Add trigger if it exists (convert to basic types for template compatibility)
	if rule.Trigger != nil {
		variables["trigger"] = map[string]any{
			"type":  string(rule.Trigger.Type),
			"globs": rule.Trigger.Globs,
		}
	}

	// Add variables from the parsed rule ID if they exist
	if rule.Variables != nil {
		for key, value := range rule.Variables {
			// Convert string booleans to actual booleans for proper template logic
			if strVal, ok := value.(string); ok {
				switch strings.ToLower(strVal) {
				case "true", "1", "yes":
					variables[key] = true
				case "false", "0", "no":
					variables[key] = false
				default:
					variables[key] = value
				}
			} else {
				variables[key] = value
			}
		}
	}

	// Add additional variables if provided (they can override base ones if needed)
	if len(additionalVars) > 0 && additionalVars[0] != nil {
		for key, value := range additionalVars[0] {
			variables[key] = value
		}
	}

	// Process the template (all rendering is text-based, no HTML escaping)
	content, err := bf.templateEngine.Render(templateContent, variables)
	if err != nil {
		return "", contextureerrors.Wrap(err, "base.ProcessTemplate")
	}

	return content, nil
}

// CreateTransformedRule creates a transformed rule with common metadata
func (bf *Base) CreateTransformedRule(
	rule *domain.Rule,
	content, filename, relativePath string,
	metadata map[string]any,
) *domain.TransformedRule {
	if metadata == nil {
		metadata = make(map[string]any)
	}

	metadata["format"] = string(bf.formatType)
	metadata["filename"] = filename
	metadata["relativePath"] = relativePath

	return &domain.TransformedRule{
		Rule:          rule,
		Content:       content,
		Filename:      filename,
		RelativePath:  relativePath,
		TransformedAt: time.Now(),
		Metadata:      metadata,
	}
}

// GenerateFilename creates a safe filename from a rule ID for multi-file formats
func (bf *Base) GenerateFilename(ruleID string) string {
	// Extract the path part from [contexture:path/to/rule] or [contexture(source):path/to/rule]
	// Use a simpler extraction pattern for now
	matches := domain.RuleIDParsePatternRegex.FindStringSubmatch(ruleID)

	if len(matches) < 3 {
		// Fallback: use the entire ID but sanitize it
		filename := strings.ReplaceAll(ruleID, "[", "")
		filename = strings.ReplaceAll(filename, "]", "")
		filename = strings.ReplaceAll(filename, ":", "-")
		filename = strings.ReplaceAll(filename, "/", "-")
		return filename + ".md"
	}

	// Use the path part (matches[2] according to the regex pattern)
	path := matches[2]

	// Replace slashes with hyphens and ensure it's a valid filename
	filename := strings.ReplaceAll(path, "/", "-")
	filename = strings.ReplaceAll(filename, "\\", "-")
	filename = domain.FilenameCleanRegex.ReplaceAllString(filename, "_")

	// Ensure it ends with .md
	if !strings.HasSuffix(filename, ".md") {
		filename += ".md"
	}

	return filename
}

// ExtractRuleIDFromFilename attempts to reverse the filename generation
func (bf *Base) ExtractRuleIDFromFilename(filename string) string {
	base := strings.TrimSuffix(filename, ".md")
	path := strings.ReplaceAll(base, "-", "/")
	return fmt.Sprintf("[contexture:%s]", path)
}

// ExtractTitleFromContent extracts the title from file content
func (bf *Base) ExtractTitleFromContent(content string) string {
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "# ") {
			return strings.TrimSpace(line[2:])
		}
	}
	return ""
}

// CalculateContentHash calculates a hash of content
func (bf *Base) CalculateContentHash(content []byte) string {
	hash := sha256.Sum256(content)
	return fmt.Sprintf("%x", hash)
}

// EnsureDirectory ensures a directory exists
func (bf *Base) EnsureDirectory(dir string) error {
	return bf.fs.MkdirAll(dir, domain.DirPermission)
}

// WriteFile writes content to a file safely
func (bf *Base) WriteFile(path string, content []byte) error {
	dir := filepath.Dir(path)
	if err := bf.EnsureDirectory(dir); err != nil {
		return contextureerrors.Wrap(err, "base.WriteFile")
	}

	return afero.WriteFile(bf.fs, path, content, domain.FilePermission)
}

// ReadFile reads a file safely
func (bf *Base) ReadFile(path string) ([]byte, error) {
	return afero.ReadFile(bf.fs, path)
}

// FileExists checks if a file exists
func (bf *Base) FileExists(path string) (bool, error) {
	return afero.Exists(bf.fs, path)
}

// DirExists checks if a directory exists
func (bf *Base) DirExists(path string) (bool, error) {
	return afero.DirExists(bf.fs, path)
}

// RemoveFile removes a file
func (bf *Base) RemoveFile(path string) error {
	return bf.fs.Remove(path)
}

// GetFileInfo gets file information
func (bf *Base) GetFileInfo(path string) (os.FileInfo, error) {
	return bf.fs.Stat(path)
}

// ListDirectory lists files in a directory
func (bf *Base) ListDirectory(path string) ([]os.FileInfo, error) {
	return afero.ReadDir(bf.fs, path)
}

// ParseRuleFromContent parses rule metadata from file content
func (bf *Base) ParseRuleFromContent(content string) (string, string) {
	lines := strings.Split(content, "\n")
	var ruleID, title string

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Look for rule ID in old format comment
		if strings.HasPrefix(line, "<!-- Contexture Rule:") {
			start := strings.Index(line, ": ") + 2
			end := strings.Index(line, " -->")
			if start > 1 && end > start {
				ruleID = strings.TrimSpace(line[start:end])
			}
		}

		// Look for rule ID in new tracking comment format
		if strings.Contains(line, domain.RuleIDCommentPrefix) {
			if extractedRuleID, err := bf.ParseTrackingComment(line); err == nil {
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

// LogDebug logs a debug message with format context
func (bf *Base) LogDebug(message string, args ...any) {
	allArgs := append([]any{"format", bf.formatType}, args...)
	log.Debug(message, allArgs...)
}

// LogInfo logs an info message with format context
func (bf *Base) LogInfo(message string, args ...any) {
	allArgs := append([]any{"format", bf.formatType}, args...)
	log.Debug(message, allArgs...)
}

// LogWarn logs a warning message with format context
func (bf *Base) LogWarn(message string, args ...any) {
	allArgs := append([]any{"format", bf.formatType}, args...)
	log.Warn(message, allArgs...)
}

// CreateTrackingComment creates a spec-compliant tracking comment
// Format: <!-- id: [contexture:path/to/rule]{variables} -->
func (bf *Base) CreateTrackingComment(ruleID string, variables map[string]any) string {
	comment := ruleID

	// Add variables if they exist
	if len(variables) > 0 {
		if variablesJSON, err := json.Marshal(variables); err == nil {
			comment += string(variablesJSON)
		}
	}

	return fmt.Sprintf("%s%s%s", domain.RuleIDCommentPrefix, comment, domain.RuleIDCommentSuffix)
}

// CreateTrackingCommentWithDefaults creates a tracking comment that only includes variables differing from defaults
// Format: <!-- id: [contexture:path/to/rule]{variables} -->
func (bf *Base) CreateTrackingCommentWithDefaults(ruleID string, variables, defaultVariables map[string]any) string {
	comment := ruleID

	// Import rule package for the utility function
	filteredVars := rule.FilterNonDefaultVariables(variables, defaultVariables)

	// Add variables only if they differ from defaults
	if len(filteredVars) > 0 {
		if variablesJSON, err := json.Marshal(filteredVars); err == nil {
			comment += string(variablesJSON)
		}
	}

	return fmt.Sprintf("%s%s%s", domain.RuleIDCommentPrefix, comment, domain.RuleIDCommentSuffix)
}

// CreateTrackingCommentFromParsed creates a tracking comment from a ParsedRuleID
func (bf *Base) CreateTrackingCommentFromParsed(parsed *domain.ParsedRuleID) string {
	// Reconstruct the rule ID string (already includes variables)
	ruleID := bf.FormatRuleID(parsed)
	// Don't pass variables again since they're already included in the formatted rule ID
	return bf.CreateTrackingComment(ruleID, nil)
}

// FormatRuleID converts a ParsedRuleID back to its string representation
func (bf *Base) FormatRuleID(parsed *domain.ParsedRuleID) string {
	var parts []string

	// Build the basic rule ID format
	switch parsed.Source {
	case "", domain.DefaultRepository:
		// Default source - no source identifier needed
		parts = append(parts, fmt.Sprintf("[contexture:%s", parsed.RulePath))
	case "local":
		// Local rules - use (local) identifier
		parts = append(parts, fmt.Sprintf("[contexture(local):%s", parsed.RulePath))
	default:
		// Custom source - use source identifier
		parts = append(parts, fmt.Sprintf("[contexture(%s):%s", parsed.Source, parsed.RulePath))
	}

	// Add ref if specified and not default
	if parsed.Ref != "" && parsed.Ref != "main" {
		parts[0] += fmt.Sprintf(",%s", parsed.Ref)
	}

	// Close the bracket
	parts[0] += "]"

	// Add variables if they exist
	if len(parsed.Variables) > 0 {
		if variablesJSON, err := json.Marshal(parsed.Variables); err == nil {
			parts[0] += string(variablesJSON)
		}
	}

	return parts[0]
}

// ParseTrackingComment extracts rule ID from a tracking comment
// Format: <!-- id: [contexture:path/to/rule]{variables} -->
func (bf *Base) ParseTrackingComment(content string) (string, error) {
	// Look for tracking comment pattern
	matches := domain.TrackingCommentRegex.FindStringSubmatch(content)

	if len(matches) != 2 {
		return "", &contextureerrors.Error{
			Op:      "base.ParseTrackingComment",
			Kind:    contextureerrors.KindFormat,
			Message: "invalid tracking comment format",
		}
	}

	return strings.TrimSpace(matches[1]), nil
}

// ExtractTrackingComments finds all tracking comments in content
func (bf *Base) ExtractTrackingComments(content string) []string {
	matches := domain.TrackingCommentRegex.FindAllStringSubmatch(content, -1)

	var ruleIDs []string
	for _, match := range matches {
		if len(match) == 2 {
			ruleIDs = append(ruleIDs, strings.TrimSpace(match[1]))
		}
	}

	return ruleIDs
}

// HasTrackingComment checks if content contains a specific tracking comment
func (bf *Base) HasTrackingComment(content, ruleID string) bool {
	trackingComment := bf.CreateTrackingComment(ruleID, nil)
	return strings.Contains(content, trackingComment)
}

// RemoveTrackingComment removes a specific tracking comment from content
func (bf *Base) RemoveTrackingComment(content, ruleID string) string {
	trackingComment := bf.CreateTrackingComment(ruleID, nil)
	return strings.ReplaceAll(content, trackingComment, "")
}

// AppendTrackingComment adds a tracking comment to the end of content
func (bf *Base) AppendTrackingComment(
	content string, ruleID string, variables map[string]any,
) string {
	// If the ruleID already contains variables (has }), don't add them again
	var passVariables map[string]any
	if !strings.Contains(ruleID, "]{") {
		passVariables = variables
	}
	trackingComment := bf.CreateTrackingComment(ruleID, passVariables)

	// Ensure there's a newline before the comment if content doesn't end with one
	if !strings.HasSuffix(content, "\n") {
		content += "\n"
	}

	// Add an extra newline for separation, then the tracking comment
	return content + "\n" + trackingComment
}

// AppendTrackingCommentWithDefaults adds a tracking comment to the end of content, only including non-default variables
func (bf *Base) AppendTrackingCommentWithDefaults(
	content string, ruleID string, variables, defaultVariables map[string]any,
) string {
	// If the ruleID already contains variables (has }), don't add them again
	var passVariables, passDefaults map[string]any
	if !strings.Contains(ruleID, "]{") {
		passVariables = variables
		passDefaults = defaultVariables
	}
	trackingComment := bf.CreateTrackingCommentWithDefaults(ruleID, passVariables, passDefaults)

	// Ensure there's a newline before the comment if content doesn't end with one
	if !strings.HasSuffix(content, "\n") {
		content += "\n"
	}

	// Add an extra newline for separation, then the tracking comment
	return content + "\n" + trackingComment
}

// RemoveDirectory removes a directory using the filesystem
func (bf *Base) RemoveDirectory(dir string) error {
	return bf.fs.RemoveAll(dir)
}

// CleanupEmptyDirectory removes the directory if it's empty
func (bf *Base) CleanupEmptyDirectory(dir string) {
	// Check if directory exists
	exists, err := bf.DirExists(dir)
	if err != nil || !exists {
		return
	}

	// List directory contents
	files, err := bf.ListDirectory(dir)
	if err != nil {
		bf.LogDebug("Could not list directory for cleanup", "dir", dir, "error", err)
		return
	}

	// Remove directory if empty
	if len(files) == 0 {
		if err := bf.RemoveDirectory(dir); err != nil {
			bf.LogDebug("Could not remove empty directory", "dir", dir, "error", err)
		} else {
			bf.LogDebug("Removed empty directory", "dir", dir)
		}
	}
}

// GetOutputPath returns the default output path for the format (to be overridden by specific formats)
func (bf *Base) GetOutputPath(_ *domain.FormatConfig) string {
	// Default implementation returns empty string, should be overridden by specific formats
	bf.LogDebug("Using default GetOutputPath implementation, should be overridden", "formatType", bf.formatType)
	return ""
}

// CleanupEmptyDirectories handles cleanup of empty directories (default implementation does nothing)
func (bf *Base) CleanupEmptyDirectories(_ *domain.FormatConfig) error {
	// Default implementation does nothing, can be overridden by specific formats
	bf.LogDebug("Using default CleanupEmptyDirectories implementation", "formatType", bf.formatType)
	return nil
}

// CreateDirectories creates necessary directories for this format (default implementation does nothing)
func (bf *Base) CreateDirectories(_ *domain.FormatConfig) error {
	// Default implementation does nothing, can be overridden by specific formats
	bf.LogDebug("Using default CreateDirectories implementation", "formatType", bf.formatType)
	return nil
}

// GetMetadata returns metadata about this format (default implementation with basic info)
func (bf *Base) GetMetadata() *domain.FormatMetadata {
	return &domain.FormatMetadata{
		Type:        bf.formatType,
		DisplayName: string(bf.formatType),
		Description: "Base format implementation",
		IsDirectory: false, // Default to file-based, override in directory-based formats
	}
}
