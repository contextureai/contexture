package template

import (
	"errors"
	"strings"
	"testing"
	"text/template"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewEngine(t *testing.T) {
	t.Parallel()
	engine := NewEngine()
	assert.NotNil(t, engine)
}

func TestTemplateEngine_Render(t *testing.T) {
	t.Parallel()
	engine := NewEngine()

	tests := []struct {
		name      string
		template  string
		variables map[string]any
		want      string
		wantErr   bool
		errMsg    string
	}{
		{
			name:     "simple variable substitution",
			template: "Hello {{.name}}!",
			variables: map[string]any{
				"name": "World",
			},
			want:    "Hello World!",
			wantErr: false,
		},
		{
			name:     "conditional rendering",
			template: "{{if .user}}Hello {{.user.name}}{{end}}",
			variables: map[string]any{
				"user": map[string]any{
					"name": "Alice",
				},
			},
			want:    "Hello Alice",
			wantErr: false,
		},
		{
			name:     "range rendering with join",
			template: "{{join .items \", \"}}",
			variables: map[string]any{
				"items": []string{"a", "b", "c"},
			},
			want:    "a, b, c",
			wantErr: false,
		},
		{
			name:     "simple range rendering",
			template: "{{range .items}}{{.}}, {{end}}",
			variables: map[string]any{
				"items": []string{"a", "b", "c"},
			},
			want:    "a, b, c, ",
			wantErr: false,
		},
		{
			name:     "template with custom function",
			template: "{{slugify .title}}",
			variables: map[string]any{
				"title": "My Great Title",
			},
			want:    "my-great-title",
			wantErr: false,
		},
		{
			name:     "template with unknown variable",
			template: "Hello {{.unknown_variable}}!",
			variables: map[string]any{
				"variable": "value",
			},
			want:    "Hello <no value>!", // Go templates show <no value> for missing fields
			wantErr: false,
		},
		{
			name:     "nested object access",
			template: "{{.user.profile.name}}",
			variables: map[string]any{
				"user": map[string]any{
					"profile": map[string]any{
						"name": "John",
					},
				},
			},
			want:    "John",
			wantErr: false,
		},
		{
			name:     "invalid template syntax",
			template: "{{.name",
			variables: map[string]any{
				"name": "test",
			},
			wantErr: true,
			errMsg:  "parse template",
		},
		{
			name:     "execution error",
			template: "{{.func}}",
			variables: map[string]any{
				"func": func() {},
			},
			wantErr: true,
			errMsg:  "execute template",
		},
		{
			name:      "empty template",
			template:  "",
			variables: map[string]any{},
			want:      "",
			wantErr:   false,
		},
		{
			name:      "nil variables",
			template:  "static content",
			variables: nil,
			want:      "static content",
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := engine.Render(tt.template, tt.variables)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, result)
			}
		})
	}
}

func TestTemplateEngine_ParseAndValidate(t *testing.T) {
	t.Parallel()
	engine := NewEngine()

	tests := []struct {
		name     string
		template string
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "valid template",
			template: "Hello {{.name}}!",
			wantErr:  false,
		},
		{
			name:     "valid with custom function",
			template: "{{slugify .title}}",
			wantErr:  false,
		},
		{
			name:     "invalid syntax",
			template: "{{.name",
			wantErr:  true,
			errMsg:   "validate template",
		},
		{
			name:     "empty template",
			template: "",
			wantErr:  false,
		},
		{
			name:     "complex valid template",
			template: "{{range .items}}{{if .}}{{.}}{{end}}{{end}}",
			wantErr:  false,
		},
		{
			name:     "undefined function",
			template: "{{unknownFunc .data}}",
			wantErr:  true,
			errMsg:   "validate template",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := engine.ParseAndValidate(tt.template)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestTemplateEngine_ExtractVariables(t *testing.T) {
	t.Parallel()
	engine := NewEngine()

	tests := []struct {
		name     string
		template string
		want     []string
		wantErr  bool
	}{
		{
			name:     "simple variable",
			template: "Hello {{.name}}!",
			want:     []string{"name"},
		},
		{
			name:     "multiple variables",
			template: "{{.first}} {{.last}}",
			want:     []string{"first", "last"},
		},
		{
			name:     "nested variable",
			template: "{{.user.name}}",
			want:     []string{"user"},
		},
		{
			name:     "variable in if",
			template: "{{if .condition}}yes{{end}}",
			want:     []string{"condition"},
		},
		{
			name:     "variable in range",
			template: "{{range .items}}{{.}}{{end}}",
			want:     []string{"items"},
		},
		{
			name:     "variable in with",
			template: "{{with .data}}{{.field}}{{end}}",
			want:     []string{"data"},
		},
		{
			name:     "variable in function",
			template: "{{slugify .title}}",
			want:     []string{"title"},
		},
		{
			name:     "duplicate variables",
			template: "{{.name}} {{.name}} {{.name}}",
			want:     []string{"name"},
		},
		{
			name:     "no variables",
			template: "static content",
			want:     []string{},
		},
		{
			name:     "variables with filters",
			template: "{{.name | upper}} {{.title | slugify}}",
			want:     []string{"name", "title"},
		},
		{
			name:     "complex nested",
			template: "{{.user.profile.settings.theme}}",
			want:     []string{"user"},
		},
		{
			name:     "multiple in one line",
			template: "{{if .a}}{{.b}}{{else}}{{.c}}{{end}}",
			want:     []string{"a", "b", "c"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := engine.ExtractVariables(tt.template)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.ElementsMatch(t, tt.want, result)
			}
		})
	}
}

// Test custom template functions

func TestSlugify(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input string
		want  string
	}{
		{"Hello World", "hello-world"},
		{"  Test  Case  ", "test-case"},
		{"Special@#$Characters", "special-characters"},
		{"Multiple---Dashes", "multiple-dashes"},
		{"123 Numbers", "123-numbers"},
		{"", ""},
		{"already-slugified", "already-slugified"},
		{"UPPERCASE", "uppercase"},
		{"Mixed123Case456", "mixed123case456"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			assert.Equal(t, tt.want, slugify(tt.input))
		})
	}
}

func TestCamelCase(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input string
		want  string
	}{
		{"hello world", "helloWorld"},
		{"Hello World", "helloWorld"},
		{"hello-world", "helloWorld"},
		{"hello_world", "helloWorld"},
		{"", ""},
		{"a", "a"},
		{"aB", "aB"},
		{"test 123 case", "test123Case"},
		{"XML parser", "xmlParser"},
		{"IOError", "ioError"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			assert.Equal(t, tt.want, camelCase(tt.input))
		})
	}
}

func TestPascalCase(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input string
		want  string
	}{
		{"hello world", "HelloWorld"},
		{"Hello World", "HelloWorld"},
		{"hello-world", "HelloWorld"},
		{"hello_world", "HelloWorld"},
		{"", ""},
		{"a", "A"},
		{"aB", "AB"},
		{"test 123 case", "Test123Case"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			assert.Equal(t, tt.want, pascalCase(tt.input))
		})
	}
}

func TestSnakeCase(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input string
		want  string
	}{
		{"hello world", "hello_world"},
		{"Hello World", "hello_world"},
		{"hello-world", "hello_world"},
		{"hello_world", "hello_world"},
		{"", ""},
		{"CamelCase", "camel_case"},
		{"HTTPRequest", "http_request"},
		{"test 123 case", "test_123_case"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			assert.Equal(t, tt.want, snakeCase(tt.input))
		})
	}
}

func TestKebabCase(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input string
		want  string
	}{
		{"hello world", "hello-world"},
		{"Hello World", "hello-world"},
		{"hello-world", "hello-world"},
		{"hello_world", "hello-world"},
		{"", ""},
		{"CamelCase", "camel-case"},
		{"test 123 case", "test-123-case"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			assert.Equal(t, tt.want, kebabCase(tt.input))
		})
	}
}

func TestTitleCase(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input string
		want  string
	}{
		{"hello world", "Hello World"},
		{"HELLO WORLD", "Hello World"},
		{"hElLo wOrLd", "Hello World"},
		{"", ""},
		{"a", "A"},
		{"test 123 case", "Test 123 Case"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			assert.Equal(t, tt.want, titleCase(tt.input))
		})
	}
}

func TestJoinAnd(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input any
		want  string
	}{
		{[]string{}, ""},
		{[]string{"one"}, "one"},
		{[]string{"one", "two"}, "one and two"},
		{[]string{"one", "two", "three"}, "one, two, and three"},
		{[]string{"a", "b", "c", "d"}, "a, b, c, and d"},
		{[]any{"x", "y", "z"}, "x, y, and z"},
		{"single", "single"},
		{[]any{1, 2, 3}, "1, 2, and 3"},
		{nil, "<nil>"},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			assert.Equal(t, tt.want, joinAnd(tt.input))
		})
	}
}

func TestUnique(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input any
		want  []string
	}{
		{[]string{"a", "b", "a", "c", "b"}, []string{"a", "b", "c"}},
		{[]string{}, []string{}},
		{[]string{"x"}, []string{"x"}},
		{[]any{"a", "b", "a"}, []string{"a", "b"}},
		{"single", []string{"single"}},
		{[]any{1, 2, 1, 3}, []string{"1", "2", "3"}},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			assert.Equal(t, tt.want, unique(tt.input))
		})
	}
}

func TestIndent(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input  string
		spaces int
		want   string
	}{
		{"line1\nline2", 2, "  line1\n  line2"},
		{"line1\n\nline3", 2, "  line1\n\n  line3"},
		{"", 4, ""},
		{"single", 4, "    single"},
		{"line1\nline2", 0, "line1\nline2"},
		{"line1\nline2", -1, "line1\nline2"},
		{"  already\n  indented", 2, "    already\n    indented"},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			assert.Equal(t, tt.want, indent(tt.input, tt.spaces))
		})
	}
}

func TestDefaultIfEmpty(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input        interface{}
		defaultValue any
		want         any
	}{
		{"", "default", "default"},
		{"value", "default", "value"},
		{"   ", "default", "default"},
		{[]string{}, "default", "default"},
		{[]string{"a"}, "default", []string{"a"}},
		{[]any{}, "default", "default"},
		{[]any{"x"}, "default", []any{"x"}},
		{nil, "default", "default"},
		{123, "default", 123},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			assert.Equal(t, tt.want, defaultIfEmpty(tt.input, tt.defaultValue))
		})
	}
}

func TestLength(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input any
		want  int
	}{
		{[]string{"a", "b", "c"}, 3},
		{[]any{1, 2}, 2},
		{"hello", 5},
		{map[string]any{"a": 1, "b": 2}, 2},
		{nil, 0},
		{123, 0},
		{[]string{}, 0},
		{"", 0},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			assert.Equal(t, tt.want, length(tt.input))
		})
	}
}

func TestSplitWords(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input string
		want  []string
	}{
		{"helloWorld", []string{"hello", "World"}},
		{"hello-world", []string{"hello", "world"}},
		{"hello_world", []string{"hello", "world"}},
		{"hello world", []string{"hello", "world"}},
		{"HTTPSConnection", []string{"HTTPS", "Connection"}},
		{"test123case", []string{"test", "123", "case"}},
		{"", []string{}},
		{"a", []string{"a"}},
		{"ABC", []string{"ABC"}},
		{"AbC", []string{"Ab", "C"}},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			assert.Equal(t, tt.want, splitWords(tt.input))
		})
	}
}

func TestToStringSlice(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input any
		want  []string
	}{
		{[]string{"a", "b"}, []string{"a", "b"}},
		{[]any{"x", "y"}, []string{"x", "y"}},
		{[]any{1, 2}, []string{"1", "2"}},
		{"single", []string{"single"}},
		{123, []string{"123"}},
		{nil, []string{"<nil>"}},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			assert.Equal(t, tt.want, toStringSlice(tt.input))
		})
	}
}

func TestConcurrentRenderSafety(t *testing.T) {
	t.Parallel()
	engine := NewEngine()
	template := "Hello {{.name}}!"

	// Run multiple renders concurrently
	done := make(chan bool, 10)
	for i := range 10 {
		go func(n int) {
			vars := map[string]any{"name": strings.Repeat("x", n)}
			_, err := engine.Render(template, vars)
			assert.NoError(t, err)
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for range 10 {
		<-done
	}
}

func TestComplexTemplateScenarios(t *testing.T) {
	t.Parallel()
	engine := NewEngine()

	tests := []struct {
		name     string
		template string
		vars     map[string]any
		want     string
	}{
		{
			name:     "combined functions",
			template: `{{.title | slugify | upper}}`,
			vars:     map[string]any{"title": "Hello World"},
			want:     "HELLO-WORLD",
		},
		{
			name:     "nested conditionals",
			template: `{{if .a}}{{if .b}}both{{else}}only a{{end}}{{else}}neither{{end}}`,
			vars:     map[string]any{"a": true, "b": false},
			want:     "only a",
		},
		{
			name:     "range with index",
			template: `{{range $i, $v := .items}}{{$i}}:{{$v}} {{end}}`,
			vars:     map[string]any{"items": []string{"a", "b"}},
			want:     "0:a 1:b ",
		},
		{
			name:     "with statement",
			template: `{{with .user}}Name: {{.name}}, Age: {{.age}}{{end}}`,
			vars: map[string]any{
				"user": map[string]any{"name": "John", "age": 30},
			},
			want: "Name: John, Age: 30",
		},
		{
			name:     "join_and with mixed types",
			template: `{{join_and .items}}`,
			vars: map[string]any{
				"items": []any{"apple", 2, "orange"},
			},
			want: "apple, 2, and orange",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := engine.Render(tt.template, tt.vars)
			require.NoError(t, err)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestErrorHandling(t *testing.T) {
	t.Parallel()
	engine := NewEngine()

	// Test template execution error
	t.Run("execution error with nil pointer", func(t *testing.T) {
		template := "{{.user.name.first}}"
		vars := map[string]any{"user": nil}

		_, err := engine.Render(template, vars)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "execute template")
	})

	// Test with function that returns error
	t.Run("function error", func(t *testing.T) {
		tmpl := template.New("test").Funcs(template.FuncMap{
			"errorFunc": func() (string, error) {
				return "", errors.New("function error")
			},
		})

		_, err := tmpl.Parse("{{errorFunc}}")
		require.NoError(t, err)

		var buf strings.Builder
		err = tmpl.Execute(&buf, nil)
		require.Error(t, err)
	})
}
