package rule

import (
	"testing"

	"github.com/contextureai/contexture/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewParser(t *testing.T) {
	parser := NewParser()
	assert.NotNil(t, parser)
	assert.IsType(t, &YAMLParser{}, parser)
}

func TestYAMLParser_ParseContent(t *testing.T) {
	parser := NewParser()

	tests := []struct {
		name     string
		content  string
		wantFM   map[string]any
		wantBody string
		wantErr  bool
	}{
		{
			name: "valid frontmatter and content",
			content: `---
title: "Test Rule"
description: "A test rule for validation"
tags: ["test", "validation"]
---

# Test Rule Content

This is the rule content.`,
			wantFM: map[string]any{
				"title":       "Test Rule",
				"description": "A test rule for validation",
				"tags":        []any{"test", "validation"},
			},
			wantBody: "\n# Test Rule Content\n\nThis is the rule content.",
			wantErr:  false,
		},
		{
			name: "content without frontmatter",
			content: `# Just Content

No frontmatter here.`,
			wantFM:   nil,
			wantBody: "# Just Content\n\nNo frontmatter here.",
			wantErr:  false,
		},
		{
			name:     "empty content",
			content:  "",
			wantFM:   nil,
			wantBody: "",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fm, body, err := parser.ParseContent(tt.content)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantFM, fm)
				assert.Equal(t, tt.wantBody, body)
			}
		})
	}
}

func TestYAMLParser_ParseRule(t *testing.T) {
	parser := NewParser()

	validContent := `---
title: "Input Validation Rule"
description: "Validates user input to prevent injection attacks"
tags: ["security", "validation", "input"]
trigger:
  type: "always"
  description: "Always apply this rule"
languages: ["javascript", "typescript"]
frameworks: ["express", "react"]
scope: "global"
global: true
variables:
  severity: "high"
  category: "security"
---

# Input Validation

Always validate user input to prevent security vulnerabilities.

## Examples

- Use parameterized queries
- Sanitize input data
- Validate data types`

	metadata := Metadata{
		ID:       "[contexture:security/input-validation]",
		FilePath: "/path/to/rule.md",
		Source:   "https://github.com/test/repo.git",
		Ref:      "main",
	}

	t.Run("valid rule parsing", func(t *testing.T) {
		rule, err := parser.ParseRule(validContent, metadata)

		require.NoError(t, err)
		assert.NotNil(t, rule)

		// Check basic fields
		assert.Equal(t, metadata.ID, rule.ID)
		assert.Equal(t, "Input Validation Rule", rule.Title)
		assert.Equal(t, "Validates user input to prevent injection attacks", rule.Description)
		assert.Contains(t, rule.Content, "# Input Validation")
		assert.Equal(t, metadata.FilePath, rule.FilePath)
		assert.Equal(t, metadata.Source, rule.Source)
		assert.Equal(t, metadata.Ref, rule.Ref)

		// Check arrays
		assert.Equal(t, []string{"security", "validation", "input"}, rule.Tags)
		assert.Equal(t, []string{"javascript", "typescript"}, rule.Languages)
		assert.Equal(t, []string{"express", "react"}, rule.Frameworks)

		// Check trigger
		assert.NotNil(t, rule.Trigger)
		assert.Equal(t, domain.TriggerAlways, rule.Trigger.Type)

		// Check variables
		assert.NotNil(t, rule.Variables)
		assert.Equal(t, "high", rule.Variables["severity"])
		assert.Equal(t, "security", rule.Variables["category"])
	})

	t.Run("missing required fields", func(t *testing.T) {
		invalidContent := `---
title: "Test Rule"
# missing description and tags
---

Content here.`

		_, err := parser.ParseRule(invalidContent, metadata)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "description: is required")
	})

	t.Run("invalid trigger type", func(t *testing.T) {
		invalidContent := `---
title: "Test Rule"
description: "Test description"
tags: ["test"]
trigger:
  type: "invalid-type"
---

Content here.`

		_, err := parser.ParseRule(invalidContent, metadata)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "must be one of:")
	})
}

func TestYAMLParser_parseTrigger(t *testing.T) {
	parser := &YAMLParser{}

	tests := []struct {
		name      string
		trigger   any
		wantType  domain.TriggerType
		wantGlobs []string
		wantErr   bool
	}{
		{
			name: "always trigger",
			trigger: map[string]any{
				"type": "always",
			},
			wantType: domain.TriggerAlways,
			wantErr:  false,
		},
		{
			name: "glob trigger with patterns",
			trigger: map[string]any{
				"type":  "glob",
				"globs": []any{"*.js", "*.ts"},
			},
			wantType:  domain.TriggerGlob,
			wantGlobs: []string{"*.js", "*.ts"},
			wantErr:   false,
		},
		{
			name: "invalid trigger type",
			trigger: map[string]any{
				"type": "invalid",
			},
			wantErr: true,
		},
		{
			name:    "trigger not an object",
			trigger: "invalid",
			wantErr: true,
		},
		{
			name: "missing trigger type",
			trigger: map[string]any{
				"description": "test",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parser.parseTrigger(tt.trigger)

			if tt.wantErr {
				require.Error(t, err)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.wantType, result.Type)
				if tt.wantGlobs != nil {
					assert.Equal(t, tt.wantGlobs, result.Globs)
				}
			}
		})
	}
}

func TestYAMLParser_parseStringSlice(t *testing.T) {
	parser := &YAMLParser{}

	tests := []struct {
		name      string
		value     any
		fieldName string
		want      []string
		wantErr   bool
	}{
		{
			name:      "string slice",
			value:     []string{"a", "b", "c"},
			fieldName: "test",
			want:      []string{"a", "b", "c"},
			wantErr:   false,
		},
		{
			name:      "interface slice with strings",
			value:     []any{"x", "y", "z"},
			fieldName: "test",
			want:      []string{"x", "y", "z"},
			wantErr:   false,
		},
		{
			name:      "single string",
			value:     "single",
			fieldName: "test",
			want:      []string{"single"},
			wantErr:   false,
		},
		{
			name:      "interface slice with non-string",
			value:     []any{"a", 123, "c"},
			fieldName: "test",
			wantErr:   true,
		},
		{
			name:      "invalid type",
			value:     123,
			fieldName: "test",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parser.parseStringSlice(tt.value, tt.fieldName)

			if tt.wantErr {
				require.Error(t, err)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, result)
			}
		})
	}
}

func TestYAMLParser_ValidateRule(t *testing.T) {
	parser := NewParser()

	t.Run("valid rule", func(t *testing.T) {
		rule := &domain.Rule{
			ID:          "[contexture:test/rule]",
			Title:       "Test Rule",
			Description: "A test rule",
			Tags:        []string{"test"},
			Content:     "# Test content",
		}

		err := parser.ValidateRule(rule)
		assert.NoError(t, err)
	})

	t.Run("missing required fields", func(t *testing.T) {
		rule := &domain.Rule{
			ID: "[contexture:test/rule]",
			// Missing title, description, tags, content
		}

		err := parser.ValidateRule(rule)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "validation errors")
	})

	t.Run("glob trigger without globs", func(t *testing.T) {
		rule := &domain.Rule{
			ID:          "[contexture:test/rule]",
			Title:       "Test Rule",
			Description: "A test rule",
			Tags:        []string{"test"},
			Content:     "# Test content",
			Trigger: &domain.RuleTrigger{
				Type: domain.TriggerGlob,
				// Missing globs
			},
		}

		err := parser.ValidateRule(rule)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed required_if validation")
	})

	t.Run("empty content", func(t *testing.T) {
		rule := &domain.Rule{
			ID:          "[contexture:test/rule]",
			Title:       "Test Rule",
			Description: "A test rule",
			Tags:        []string{"test"},
			Content:     "   \n\t  ", // Only whitespace
		}

		err := parser.ValidateRule(rule)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "rule content cannot be empty")
	})

	t.Run("duplicate tags", func(t *testing.T) {
		rule := &domain.Rule{
			ID:          "[contexture:test/rule]",
			Title:       "Test Rule",
			Description: "A test rule",
			Tags:        []string{"test", "validation", "test"}, // Duplicate
			Content:     "# Test content",
		}

		err := parser.ValidateRule(rule)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "duplicate tag")
	})
}

func TestParseRules(t *testing.T) {
	parser := NewParser()

	validContent := `---
title: "Valid Rule"
description: "A valid test rule"
tags: ["test"]
---

# Valid Content`

	invalidContent := `---
title: "Invalid Rule"
# Missing description and tags
---

# Invalid Content`

	rules := []Content{
		{
			Content: validContent,
			Metadata: Metadata{
				ID:       "[contexture:test/valid]",
				FilePath: "/valid.md",
				Source:   "test",
				Ref:      "main",
			},
		},
		{
			Content: invalidContent,
			Metadata: Metadata{
				ID:       "[contexture:test/invalid]",
				FilePath: "/invalid.md",
				Source:   "test",
				Ref:      "main",
			},
		},
	}

	result := ParseRules(parser, rules)

	assert.Len(t, result.Rules, 1)
	assert.Len(t, result.Errors, 1)
	assert.Len(t, result.Skipped, 1)

	assert.Equal(t, "Valid Rule", result.Rules[0].Title)
	assert.Equal(t, "[contexture:test/invalid]", result.Skipped[0])
	assert.Contains(t, result.Errors[0].Error(), "failed to parse rule")
}
