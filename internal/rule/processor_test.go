package rule

import (
	"testing"

	"github.com/contextureai/contexture/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestNewProcessor(t *testing.T) {
	processor := NewProcessor()
	assert.NotNil(t, processor)
	assert.IsType(t, &TemplateProcessor{}, processor)
}

func TestTemplateProcessor_ProcessRule(t *testing.T) {
	processor := NewProcessor()

	rule := &domain.Rule{
		ID:          "[contexture:test/template-rule]",
		Title:       "Template Test Rule",
		Description: "A rule for testing template processing",
		Tags:        []string{"test", "template"},
		Content:     "# {{.rule.title}}\n\nHello {{default_if_empty .name \"World\"}}!\n\nTags: {{join_and .rule.tags}}",
		Source:      "https://github.com/test/repo.git",
		Ref:         "main",
		Variables: map[string]any{
			"name": "Alice",
		},
	}

	context := &domain.RuleContext{
		Variables: map[string]any{
			"environment": "test",
		},
		Globals: map[string]any{
			"project": "contexture",
		},
	}

	t.Run("successful processing", func(t *testing.T) {
		processed, err := processor.ProcessRule(rule, context)

		require.NoError(t, err)
		assert.NotNil(t, processed)
		assert.Equal(t, rule, processed.Rule)
		assert.Equal(t, context, processed.Context)

		// Check that content is raw (not processed)
		assert.Equal(t, rule.Content, processed.Content)

		// Check that variables are properly set for format processing
		assert.NotNil(t, processed.Variables)
		assert.Equal(t, "Alice", processed.Variables["name"])

		// Check attribution
		assert.NotEmpty(t, processed.Attribution)
		assert.Contains(t, processed.Attribution, rule.ID)
		assert.Contains(t, processed.Attribution, rule.Source)
	})

	t.Run("rule with nil context", func(t *testing.T) {
		processed, err := processor.ProcessRule(rule, nil)

		require.NoError(t, err)
		assert.NotNil(t, processed)
		// Content should be raw, variables should be available for format processing
		assert.Equal(t, rule.Content, processed.Content)
		assert.Equal(t, "Alice", processed.Variables["name"])
	})

	t.Run("rule with invalid template", func(t *testing.T) {
		invalidRule := &domain.Rule{
			ID:      "[contexture:test/invalid]",
			Title:   "Invalid Rule",
			Content: "{{if .unclosed}} {{.missing end",
		}

		processed, err := processor.ProcessRule(invalidRule, context)
		// No error should occur since we don't process templates in processor anymore
		require.NoError(t, err)
		assert.NotNil(t, processed)
		assert.Equal(t, invalidRule.Content, processed.Content)
	})
}

func TestTemplateProcessor_ProcessRules(t *testing.T) {
	processor := NewProcessor()

	rules := []*domain.Rule{
		{
			ID:      "[contexture:test/rule1]",
			Title:   "Rule 1",
			Content: "# {{.rule.title}}\nContent 1",
		},
		{
			ID:      "[contexture:test/rule2]",
			Title:   "Rule 2",
			Content: "# {{.rule.title}}\nContent 2",
		},
	}

	context := &domain.RuleContext{
		Variables: map[string]any{
			"test": true,
		},
	}

	t.Run("process multiple rules", func(t *testing.T) {
		processed, err := processor.ProcessRules(rules, context)

		require.NoError(t, err)
		assert.Len(t, processed, 2)

		// Check both rules were processed
		ruleMap := make(map[string]*domain.ProcessedRule)
		for _, p := range processed {
			ruleMap[p.Rule.ID] = p
		}

		assert.Contains(t, ruleMap, "[contexture:test/rule1]")
		assert.Contains(t, ruleMap, "[contexture:test/rule2]")

		// Content should be raw, not processed
		assert.Equal(t, "# {{.rule.title}}\nContent 1", ruleMap["[contexture:test/rule1]"].Content)
		assert.Equal(t, "# {{.rule.title}}\nContent 2", ruleMap["[contexture:test/rule2]"].Content)

		// Variables should be available for format processing
		assert.NotNil(t, ruleMap["[contexture:test/rule1]"].Variables)
		assert.NotNil(t, ruleMap["[contexture:test/rule2]"].Variables)
	})

	t.Run("empty rules slice", func(t *testing.T) {
		processed, err := processor.ProcessRules([]*domain.Rule{}, context)

		require.NoError(t, err)
		assert.Empty(t, processed)
	})
}

func TestTemplateProcessor_ProcessTemplate(t *testing.T) {
	processor := NewProcessor()

	tests := []struct {
		name      string
		template  string
		variables map[string]any
		wantErr   bool
		contains  []string
	}{
		{
			name:     "simple variable substitution",
			template: "Hello {{.name}}!",
			variables: map[string]any{
				"name": "World",
			},
			wantErr:  false,
			contains: []string{"Hello World!"},
		},
		{
			name:      "built-in variables",
			template:  "Generated on {{.date}} by {{.contexture.engine}}",
			variables: map[string]any{},
			wantErr:   false,
			contains:  []string{"Generated on", "by go"},
		},
		{
			name:     "custom filters",
			template: "{{slugify .title}} and {{join_and .items}}",
			variables: map[string]any{
				"title": "My Great Title",
				"items": []string{"a", "b", "c"},
			},
			wantErr:  false,
			contains: []string{"my-great-title", "a, b, and c"},
		},
		{
			name:      "invalid template",
			template:  "{{.unclosed",
			variables: map[string]any{},
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := processor.ProcessTemplate(tt.template, tt.variables)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				for _, contains := range tt.contains {
					assert.Contains(t, result, contains)
				}
			}
		})
	}
}

func TestTemplateProcessor_GenerateAttribution(t *testing.T) {
	processor := NewProcessor()

	rule := &domain.Rule{
		ID:     "[contexture:test/rule]",
		Source: "https://github.com/test/repo.git",
		Ref:    "main",
	}

	attribution := processor.GenerateAttribution(rule)

	assert.NotEmpty(t, attribution)
	assert.Contains(t, attribution, rule.ID)
	assert.Contains(t, attribution, rule.Source)
	assert.Contains(t, attribution, "Generated by Contexture")
}

func TestTemplateProcessor_buildVariableMap(t *testing.T) {
	variableManager := NewVariableManager()

	rule := &domain.Rule{
		ID:          "[contexture:test/rule]",
		Title:       "Test Rule",
		Description: "A test rule",
		Tags:        []string{"test"},
		Variables: map[string]any{
			"ruleVar": "rule-value",
			"shared":  "rule-wins", // Should override context and global
		},
	}

	context := &domain.RuleContext{
		Variables: map[string]any{
			"contextVar": "context-value",
			"shared":     "context-value", // Should be overridden by rule
		},
		Globals: map[string]any{
			"globalVar": "global-value",
			"shared":    "global-value", // Should be overridden by rule and context
		},
	}

	variables := variableManager.BuildVariableMap(rule, context)

	// Check all variables are present
	assert.Equal(t, "rule-value", variables["ruleVar"])
	assert.Equal(t, "context-value", variables["contextVar"])
	assert.Equal(t, "global-value", variables["globalVar"])

	// Check precedence (rule > context > global)
	assert.Equal(t, "rule-wins", variables["shared"])

	// Check rule metadata is available
	ruleData, ok := variables["rule"].(map[string]any)
	assert.True(t, ok, "rule should be a map")
	assert.Equal(t, rule.ID, ruleData["id"])
	assert.Equal(t, rule.Title, ruleData["title"])
	assert.Equal(t, rule.Tags, ruleData["tags"])
}

func TestTemplateProcessor_addBuiltinVariables(t *testing.T) {
	variableManager := NewVariableManager()

	variables := map[string]any{
		"existing": "value",
	}

	enriched := variableManager.EnrichWithBuiltins(variables)

	// Check existing variables are preserved
	assert.Equal(t, "value", enriched["existing"])

	// Check built-in variables are added
	assert.NotNil(t, enriched["now"])
	assert.NotNil(t, enriched["date"])
	assert.NotNil(t, enriched["year"])
	assert.NotNil(t, enriched["timestamp"])

	// Check contexture-specific variables
	contexture := enriched["contexture"].(map[string]any)
	assert.Equal(t, "go", contexture["engine"])
	assert.NotEmpty(t, contexture["version"])
}

func TestTemplateProcessor_ValidateTemplate(t *testing.T) {
	processor := &TemplateProcessor{}
	mockEngine := NewMockTemplateEngine(t)
	mockEngine.EXPECT().
		ExtractVariables(mock.AnythingOfType("string")).
		Return([]string{"name", "site"}, nil).
		Maybe()
	mockEngine.EXPECT().
		ProcessTemplate(mock.AnythingOfType("string"), mock.Anything).
		Return("processed template", nil).
		Maybe()
	processor.templateEngine = mockEngine

	tests := []struct {
		name          string
		template      string
		requiredVars  []string
		wantErr       bool
		errorContains string
	}{
		{
			name:         "valid template with all required vars",
			template:     "Hello {{.name}}, welcome to {{.site}}!",
			requiredVars: []string{"name", "site"},
			wantErr:      false,
		},
		{
			name:          "missing required variable",
			template:      "Hello {{.name}}!",
			requiredVars:  []string{"name", "missing"},
			wantErr:       true,
			errorContains: "missing required variables",
		},
		{
			name:         "no required variables",
			template:     "Static content",
			requiredVars: []string{},
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := processor.ValidateTemplate(tt.template, tt.requiredVars)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
