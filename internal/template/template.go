// Package template provides text template processing functionality for Contexture.
//
// This package wraps Go's text/template to provide markdown-safe template rendering
// with custom functions for string manipulation, formatting, and array operations.
// All template rendering is done using text/template to avoid HTML escaping.
//
// Example usage:
//
//	engine := template.NewEngine()
//	result, err := engine.Render("Hello {{.name}}!", map[string]any{"name": "World"})
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println(result) // Output: Hello World!
package template

import (
	"fmt"
	"regexp"
	"strings"
	"text/template"
	"unicode"

	"github.com/charmbracelet/log"
)

// Pre-compiled regex patterns for better performance
var (
	// Matches {{.Variable}} and {{.Variable.Field}} patterns with optional filters
	dotVarRegex = regexp.MustCompile(
		`{{\s*\.([a-zA-Z_][a-zA-Z0-9_]*(?:\.[a-zA-Z_][a-zA-Z0-9_]*)*)\s*(?:\s*\|\s*[^}]+)?\s*}}`,
	)
	// Matches {{if .Variable}}, {{range .Items}}, {{with .Variable}} etc.
	actionVarRegex = regexp.MustCompile(
		`{{\s*(?:if|with|range)\s+\.([a-zA-Z_][a-zA-Z0-9_]*(?:\.[a-zA-Z_][a-zA-Z0-9_]*)*)`,
	)
	// Matches function calls that reference variables: {{func .Variable}}
	funcVarRegex = regexp.MustCompile(
		`{{\s*[a-zA-Z_][a-zA-Z0-9_]*\s+\.([a-zA-Z_][a-zA-Z0-9_]*(?:\.[a-zA-Z_][a-zA-Z0-9_]*)*)`,
	)
	// For slugify function
	nonAlphaNumRegex = regexp.MustCompile(`[^a-z0-9]+`)
)

// Engine defines the interface for template processing
type Engine interface {
	// Render processes a template with the given variables
	Render(template string, variables map[string]any) (string, error)
	// ParseAndValidate checks if a template is syntactically valid
	ParseAndValidate(template string) error
	// ExtractVariables finds all variables referenced in a template
	ExtractVariables(template string) ([]string, error)
}

// templateEngine implements template processing using Go's text/template
type templateEngine struct {
	funcMap template.FuncMap
}

// NewEngine creates a new template engine with custom functions
func NewEngine() Engine {
	return &templateEngine{
		funcMap: createFuncMap(),
	}
}

// Render processes a template with the given variables
func (e *templateEngine) Render(templateStr string, variables map[string]any) (string, error) {
	log.Debug("Rendering template", "vars_count", len(variables))

	// Create a new template instance for thread safety
	tmpl := template.New("render").Funcs(e.funcMap)
	tmpl, err := tmpl.Parse(templateStr)
	if err != nil {
		// Add template preview for better debugging
		preview := templateStr
		if len(preview) > 100 {
			preview = preview[:100] + "..."
		}
		return "", fmt.Errorf("template parsing failed for content %q: %w", preview, err)
	}

	// Execute template
	var result strings.Builder
	if err := tmpl.Execute(&result, variables); err != nil {
		// Add template preview for better debugging
		preview := templateStr
		if len(preview) > 100 {
			preview = preview[:100] + "..."
		}
		return "", fmt.Errorf("template execution failed for content %q: %w", preview, err)
	}

	log.Debug("Successfully rendered template")
	return result.String(), nil
}

// ParseAndValidate checks if a template is syntactically valid
func (e *templateEngine) ParseAndValidate(templateStr string) error {
	tmpl := template.New("validate").Funcs(e.funcMap)
	_, err := tmpl.Parse(templateStr)
	if err != nil {
		// Add template preview for better debugging
		preview := templateStr
		if len(preview) > 100 {
			preview = preview[:100] + "..."
		}
		return fmt.Errorf("template validation failed for content %q: %w", preview, err)
	}
	return nil
}

// ExtractVariables finds all variables referenced in a template
func (e *templateEngine) ExtractVariables(templateStr string) ([]string, error) {
	seen := make(map[string]bool)
	var variables []string

	// Extract variables from different template patterns
	extractMatches := func(matches [][]string, isWithinBlock bool) {
		for _, match := range matches {
			if len(match) > 1 {
				// Get root variable name (before first dot)
				varName := strings.Split(match[1], ".")[0]
				if !seen[varName] && !isWithinBlock {
					seen[varName] = true
					variables = append(variables, varName)
				}
			}
		}
	}

	// First, handle with/range blocks to avoid extracting variables from within them
	// Find and remove content within {{with .var}}...{{end}} blocks
	withBlockRegex := regexp.MustCompile(
		`{{\s*with\s+\.([a-zA-Z_][a-zA-Z0-9_]*(?:\.[a-zA-Z_][a-zA-Z0-9_]*)*)\s*}}.*?{{\s*end\s*}}`,
	)
	withMatches := withBlockRegex.FindAllStringSubmatch(templateStr, -1)

	// Extract variables from with statements themselves
	for _, match := range withMatches {
		if len(match) > 1 {
			varName := strings.Split(match[1], ".")[0]
			if !seen[varName] {
				seen[varName] = true
				variables = append(variables, varName)
			}
		}
	}

	// Remove with block contents to avoid extracting inner variables
	cleanedTemplate := withBlockRegex.ReplaceAllString(templateStr, "")

	// Extract from {{.Variable}} patterns (not within blocks)
	extractMatches(dotVarRegex.FindAllStringSubmatch(cleanedTemplate, -1), false)
	// Extract from {{if .Variable}} patterns
	extractMatches(actionVarRegex.FindAllStringSubmatch(cleanedTemplate, -1), false)
	// Extract from function calls
	extractMatches(funcVarRegex.FindAllStringSubmatch(cleanedTemplate, -1), false)

	return variables, nil
}

// createFuncMap creates all custom functions used by Contexture
func createFuncMap() template.FuncMap {
	return template.FuncMap{
		// String manipulation functions
		"slugify":    slugify,
		"camelcase":  camelCase,
		"pascalcase": pascalCase,
		"snakecase":  snakeCase,
		"kebabcase":  kebabCase,
		"titlecase":  titleCase,

		// Array functions
		"join_and": joinAnd,
		"unique":   unique,

		// Formatting functions
		"indent": indent,

		// Conditional functions
		"default_if_empty": defaultIfEmpty,

		// Standard string functions
		"join":    strings.Join,
		"lower":   strings.ToLower,
		"upper":   strings.ToUpper,
		"trim":    strings.TrimSpace,
		"replace": strings.ReplaceAll,
		"len":     length,
	}
}

// length returns the length of various types
func length(v any) int {
	switch val := v.(type) {
	case []string:
		return len(val)
	case []any:
		return len(val)
	case string:
		return len(val)
	case map[string]any:
		return len(val)
	default:
		return 0
	}
}

// slugify converts a string to a URL-friendly slug
func slugify(input string) string {
	// Convert to lowercase
	result := strings.ToLower(input)
	// Replace non-alphanumeric with hyphens
	result = nonAlphaNumRegex.ReplaceAllString(result, "-")
	// Remove leading/trailing hyphens
	return strings.Trim(result, "-")
}

// splitWords splits a string into words, handling various separators
func splitWords(input string) []string {
	if input == "" {
		return []string{}
	}

	words := []string{}
	runes := []rune(input)
	start := 0

	for i := range runes {
		// Check for word boundaries
		if i > 0 {
			curr := runes[i]
			prev := runes[i-1]

			// Split on non-letter/digit
			if !unicode.IsLetter(curr) && !unicode.IsDigit(curr) {
				if i > start {
					words = append(words, string(runes[start:i]))
				}
				start = i + 1
				continue
			}

			// Split when transitioning from letter to digit or vice versa
			if unicode.IsLetter(prev) && unicode.IsDigit(curr) ||
				unicode.IsDigit(prev) && unicode.IsLetter(curr) {
				if i > start {
					words = append(words, string(runes[start:i]))
				}
				start = i
				continue
			}

			// Split on case changes
			if unicode.IsLetter(prev) && unicode.IsLetter(curr) {
				// Lower to upper
				if unicode.IsLower(prev) && unicode.IsUpper(curr) {
					if i > start {
						words = append(words, string(runes[start:i]))
					}
					start = i
					continue
				}

				// Upper to upper followed by lower (e.g., "HTTPSConnection" -> "HTTPS", "Connection")
				if unicode.IsUpper(prev) && unicode.IsUpper(curr) && i+1 < len(runes) &&
					unicode.IsLower(runes[i+1]) {
					if i > start {
						words = append(words, string(runes[start:i]))
					}
					start = i
					continue
				}
			}
		}
	}

	// Add the last word
	if start < len(runes) {
		words = append(words, string(runes[start:]))
	}

	return words
}

// camelCase converts a string to camelCase
func camelCase(input string) string {
	words := splitWords(input)
	if len(words) == 0 {
		return ""
	}

	var result strings.Builder
	result.WriteString(strings.ToLower(words[0]))
	for i := 1; i < len(words); i++ {
		if len(words[i]) > 0 {
			result.WriteString(strings.ToUpper(words[i][:1]))
			result.WriteString(strings.ToLower(words[i][1:]))
		}
	}
	return result.String()
}

// pascalCase converts a string to PascalCase
func pascalCase(input string) string {
	words := splitWords(input)
	var result strings.Builder
	for _, word := range words {
		if len(word) > 0 {
			result.WriteString(strings.ToUpper(word[:1]))
			result.WriteString(strings.ToLower(word[1:]))
		}
	}
	return result.String()
}

// snakeCase converts a string to snake_case
func snakeCase(input string) string {
	words := splitWords(input)
	for i := range words {
		words[i] = strings.ToLower(words[i])
	}
	return strings.Join(words, "_")
}

// kebabCase converts a string to kebab-case
func kebabCase(input string) string {
	words := splitWords(input)
	for i := range words {
		words[i] = strings.ToLower(words[i])
	}
	return strings.Join(words, "-")
}

// titleCase converts a string to Title Case
func titleCase(input string) string {
	words := strings.Fields(input)
	for i, word := range words {
		if len(word) > 0 {
			words[i] = strings.ToUpper(word[:1]) + strings.ToLower(word[1:])
		}
	}
	return strings.Join(words, " ")
}

// joinAnd joins array elements with commas and "and" for the last element
func joinAnd(input any) string {
	items := toStringSlice(input)

	switch len(items) {
	case 0:
		return ""
	case 1:
		return items[0]
	case 2:
		return items[0] + " and " + items[1]
	default:
		return strings.Join(items[:len(items)-1], ", ") + ", and " + items[len(items)-1]
	}
}

// unique removes duplicate elements from an array
func unique(input any) []string {
	items := toStringSlice(input)
	seen := make(map[string]bool)
	result := []string{} // Initialize as empty slice, not nil

	for _, item := range items {
		if !seen[item] {
			seen[item] = true
			result = append(result, item)
		}
	}

	return result
}

// toStringSlice converts various types to []string
func toStringSlice(input any) []string {
	switch v := input.(type) {
	case []string:
		return v
	case []any:
		result := make([]string, len(v))
		for i, item := range v {
			result[i] = fmt.Sprintf("%v", item)
		}
		return result
	default:
		return []string{fmt.Sprintf("%v", input)}
	}
}

// indent indents each line of text by the specified number of spaces
func indent(input string, spaces int) string {
	if spaces <= 0 {
		return input
	}

	indent := strings.Repeat(" ", spaces)
	lines := strings.Split(input, "\n")

	for i, line := range lines {
		if strings.TrimSpace(line) != "" { // Don't indent empty lines
			lines[i] = indent + line
		}
	}

	return strings.Join(lines, "\n")
}

// defaultIfEmpty returns a default value if the input is empty
func defaultIfEmpty(input any, defaultValue any) any {
	switch v := input.(type) {
	case string:
		if strings.TrimSpace(v) == "" {
			return defaultValue
		}
	case []any:
		if len(v) == 0 {
			return defaultValue
		}
	case []string:
		if len(v) == 0 {
			return defaultValue
		}
	case nil:
		return defaultValue
	}
	return input
}
