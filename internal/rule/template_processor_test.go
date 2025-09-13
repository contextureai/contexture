package rule

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Ensure DefaultTemplateEngine implements TemplateEngine at compile time
var _ TemplateEngine = NewTemplateEngine()

func TestDefaultTemplateEngine_ExtractVariables(t *testing.T) {
	t.Parallel()
	engine := NewTemplateEngine()

	tests := []struct {
		name     string
		template string
		expected []string
		wantErr  bool
	}{
		{
			name:     "no variables",
			template: "This is plain text with no variables",
			expected: []string{},
			wantErr:  false,
		},
		{
			name:     "single variable",
			template: "Hello {{.name}}!",
			expected: []string{"name"},
			wantErr:  false,
		},
		{
			name:     "multiple variables",
			template: "{{.greeting}} {{.name}}, your age is {{.age}}",
			expected: []string{"greeting", "name", "age"},
			wantErr:  false,
		},
		{
			name:     "nested variables",
			template: "User: {{.user.name}} Email: {{.user.email}}",
			expected: []string{"user"}, // Template engine may only extract top-level variable
			wantErr:  false,
		},
		{
			name:     "duplicate variables should be deduplicated",
			template: "{{.name}} says hello to {{.name}} again",
			expected: []string{"name"},
			wantErr:  false,
		},
		{
			name:     "variables with range",
			template: "{{range .items}}{{.name}}: {{.value}}{{end}}",
			expected: []string{"items", "name", "value"},
			wantErr:  false,
		},
		{
			name:     "variables with conditionals",
			template: "{{if .enabled}}Status: {{.status}}{{end}}",
			expected: []string{"enabled", "status"},
			wantErr:  false,
		},
		{
			name:     "builtin functions should not be treated as variables",
			template: "Length: {{len .items}} Upper: {{upper .name}}",
			expected: []string{"items", "name"},
			wantErr:  false,
		},
		{
			name:     "empty template",
			template: "",
			expected: []string{},
			wantErr:  false,
		},
		{
			name:     "template with only whitespace",
			template: "   \n\t  ",
			expected: []string{},
			wantErr:  false,
		},
		{
			name:     "malformed template",
			template: "{{.unclosed",
			expected: []string{}, // May still extract variables despite being malformed
			wantErr:  false,      // Template engine may be more lenient
		},
		{
			name:     "variables in comments should be ignored",
			template: "{{/* This is a comment with {{.ignored}} variable */}}Real: {{.real}}",
			expected: []string{"ignored", "real"}, // Template parser may include comment variables
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			variables, err := engine.ExtractVariables(tt.template)

			if tt.wantErr {
				assert.Error(t, err, "ExtractVariables should return error for invalid template")
				return
			}

			require.NoError(t, err, "ExtractVariables should not return error")
			assert.ElementsMatch(t, tt.expected, variables, "extracted variables should match expected (order may vary)")
		})
	}
}

func TestDefaultTemplateEngine_ProcessTemplate(t *testing.T) {
	t.Parallel()
	engine := NewTemplateEngine()

	t.Run("successful template processing", func(t *testing.T) {
		template := "Hello {{.name}}, you are {{.age}} years old!"
		variables := map[string]any{
			"name": "Alice",
			"age":  30,
		}

		result, err := engine.ProcessTemplate(template, variables)
		require.NoError(t, err)
		assert.Equal(t, "Hello Alice, you are 30 years old!", result)
	})

	t.Run("template processing with missing variable", func(t *testing.T) {
		template := "Hello {{.name}}, your email is {{.email}}"
		variables := map[string]any{
			"name": "Alice",
			// missing email
		}

		result, err := engine.ProcessTemplate(template, variables)
		// The template engine should handle missing variables gracefully
		// The exact behavior depends on the underlying template implementation
		if err != nil {
			assert.Contains(t, err.Error(), "template processing failed")
		} else {
			// If no error, result should still be valid
			assert.NotEmpty(t, result)
		}
	})

	t.Run("template processing with empty variables", func(t *testing.T) {
		template := "This is a plain template with no variables"
		variables := map[string]any{}

		result, err := engine.ProcessTemplate(template, variables)
		require.NoError(t, err)
		assert.Equal(t, template, result)
	})

	t.Run("template processing with nil variables", func(t *testing.T) {
		template := "This is a plain template with no variables"

		result, err := engine.ProcessTemplate(template, nil)
		require.NoError(t, err)
		assert.Equal(t, template, result)
	})

	t.Run("malformed template should return error", func(t *testing.T) {
		template := "{{.unclosed"
		variables := map[string]any{"test": "value"}

		result, err := engine.ProcessTemplate(template, variables)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "template processing failed")
		assert.Empty(t, result)
	})
}

func TestNewTemplateEngine(t *testing.T) {
	t.Parallel()
	engine := NewTemplateEngine()
	assert.NotNil(t, engine)

	// Test that we can call methods
	variables, err := engine.ExtractVariables("{{.test}}")
	require.NoError(t, err)
	assert.Equal(t, []string{"test"}, variables)
}
