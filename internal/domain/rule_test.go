package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestRule_GetDefaultTrigger(t *testing.T) {
	t.Parallel()
	t.Run("rule with trigger", func(t *testing.T) {
		trigger := &RuleTrigger{Type: TriggerAlways}
		rule := Rule{Trigger: trigger}

		result := rule.GetDefaultTrigger()
		assert.Equal(t, trigger, result)
	})

	t.Run("rule without trigger", func(t *testing.T) {
		rule := Rule{}

		result := rule.GetDefaultTrigger()
		assert.NotNil(t, result)
		assert.Equal(t, TriggerManual, result.Type)
	})
}

func TestRule_HasLanguage(t *testing.T) {
	t.Parallel()
	rule := Rule{
		Languages: []string{"go", "typescript", "python"},
	}

	t.Run("exact match", func(t *testing.T) {
		assert.True(t, rule.HasLanguage("go"))
	})

	t.Run("case insensitive", func(t *testing.T) {
		assert.True(t, rule.HasLanguage("GO"))
		assert.True(t, rule.HasLanguage("TypeScript"))
	})

	t.Run("another language", func(t *testing.T) {
		assert.True(t, rule.HasLanguage("python"))
	})

	t.Run("not found", func(t *testing.T) {
		assert.False(t, rule.HasLanguage("java"))
	})
}

func TestRule_HasFramework(t *testing.T) {
	t.Parallel()
	rule := Rule{
		Frameworks: []string{"react", "vue", "angular"},
	}

	t.Run("exact match", func(t *testing.T) {
		assert.True(t, rule.HasFramework("react"))
	})

	t.Run("case insensitive", func(t *testing.T) {
		assert.True(t, rule.HasFramework("REACT"))
		assert.True(t, rule.HasFramework("Vue"))
	})

	t.Run("not found", func(t *testing.T) {
		assert.False(t, rule.HasFramework("svelte"))
	})
}

func TestRule_HasTag(t *testing.T) {
	t.Parallel()
	rule := Rule{
		Tags: []string{"security", "performance", "testing"},
	}

	t.Run("exact match", func(t *testing.T) {
		assert.True(t, rule.HasTag("security"))
	})

	t.Run("case insensitive", func(t *testing.T) {
		assert.True(t, rule.HasTag("SECURITY"))
		assert.True(t, rule.HasTag("Performance"))
	})

	t.Run("not found", func(t *testing.T) {
		assert.False(t, rule.HasTag("documentation"))
	})
}

func TestRuleRef_GetSource(t *testing.T) {
	t.Parallel()
	t.Run("with source", func(t *testing.T) {
		ref := RuleRef{Source: "custom"}
		assert.Equal(t, "custom", ref.GetSource())
	})

	t.Run("without source", func(t *testing.T) {
		ref := RuleRef{}
		assert.Equal(t, "contexture", ref.GetSource())
	})
}

func TestRuleRef_GetRef(t *testing.T) {
	t.Parallel()
	t.Run("with ref", func(t *testing.T) {
		ref := RuleRef{Ref: "develop"}
		assert.Equal(t, "develop", ref.GetRef())
	})

	t.Run("without ref", func(t *testing.T) {
		ref := RuleRef{}
		assert.Equal(t, "main", ref.GetRef())
	})
}

func TestValidationResult_HasErrors(t *testing.T) {
	t.Parallel()
	result := &ValidationResult{}

	// No errors initially
	assert.False(t, result.HasErrors())

	// Add an error
	result.AddError("test", "Test error", "TEST_ERROR")

	assert.True(t, result.HasErrors())
}

func TestValidationResult_HasWarnings(t *testing.T) {
	t.Parallel()
	result := &ValidationResult{}

	// No warnings initially
	assert.False(t, result.HasWarnings())

	// Add a warning
	result.Warnings = append(result.Warnings, ValidationWarning{
		Field:   "test",
		Message: "Test warning",
		Code:    "TEST_WARNING",
	})

	assert.True(t, result.HasWarnings())
}

func TestRule_MatchesGlob(t *testing.T) {
	t.Parallel()
	t.Run("matches go pattern", func(t *testing.T) {
		rule := Rule{
			Trigger: &RuleTrigger{
				Type:  TriggerGlob,
				Globs: []string{"*.go", "*.ts"},
			},
		}

		assert.True(t, rule.MatchesGlob("main.go"))
	})

	t.Run("matches ts pattern", func(t *testing.T) {
		rule := Rule{
			Trigger: &RuleTrigger{
				Type:  TriggerGlob,
				Globs: []string{"*.go", "*.ts"},
			},
		}

		assert.True(t, rule.MatchesGlob("app.ts"))
	})

	t.Run("no match", func(t *testing.T) {
		rule := Rule{
			Trigger: &RuleTrigger{
				Type:  TriggerGlob,
				Globs: []string{"*.go", "*.ts"},
			},
		}

		assert.False(t, rule.MatchesGlob("README.md"))
	})

	t.Run("non-glob trigger", func(t *testing.T) {
		rule := Rule{
			Trigger: &RuleTrigger{
				Type: TriggerAlways,
			},
		}

		assert.False(t, rule.MatchesGlob("main.go"))
	})
}

func TestRuleRef_UnmarshalYAML(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		yaml           string
		expectedID     string
		expectedSource string
		expectedRef    string
	}{
		{
			name:           "@provider/path syntax",
			yaml:           `id: "@mycompany/security/auth"`,
			expectedID:     "@mycompany/security/auth",
			expectedSource: "", // Source not extracted from @provider format, resolved by parser
			expectedRef:    "",
		},
		{
			name:           "full format with @provider",
			yaml:           `id: "[contexture(@mycompany):security/auth]"`,
			expectedID:     "[contexture(@mycompany):security/auth]",
			expectedSource: "@mycompany", // Source extracted from full format
			expectedRef:    "",
		},
		{
			name:           "full format with URL",
			yaml:           `id: "[contexture(https://github.com/user/repo.git):path/to/rule]"`,
			expectedID:     "[contexture(https://github.com/user/repo.git):path/to/rule]",
			expectedSource: "https://github.com/user/repo.git",
			expectedRef:    "",
		},
		{
			name:           "simple path format",
			yaml:           `id: "typescript/naming"`,
			expectedID:     "typescript/naming",
			expectedSource: "",
			expectedRef:    "",
		},
		{
			name: "explicit source field",
			yaml: `id: "path/to/rule"
source: "explicit-source"`,
			expectedID:     "path/to/rule",
			expectedSource: "explicit-source", // Explicit source takes precedence
			expectedRef:    "",
		},
		{
			name: "with ref field",
			yaml: `id: "@contexture/go/testing"
ref: "v1.2.0"`,
			expectedID:     "@contexture/go/testing",
			expectedSource: "",
			expectedRef:    "v1.2.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var ruleRef RuleRef
			err := yaml.Unmarshal([]byte(tt.yaml), &ruleRef)
			require.NoError(t, err)

			assert.Equal(t, tt.expectedID, ruleRef.ID)
			assert.Equal(t, tt.expectedSource, ruleRef.Source)
			assert.Equal(t, tt.expectedRef, ruleRef.Ref)
		})
	}
}
