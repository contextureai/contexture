package query

import (
	"testing"

	"github.com/contextureai/contexture/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEvaluator_MatchesText(t *testing.T) {
	t.Parallel()

	evaluator := NewEvaluator()

	tests := []struct {
		name     string
		rule     *domain.Rule
		query    string
		expected bool
	}{
		{
			name: "matches in title",
			rule: &domain.Rule{
				ID:    "test/rule",
				Title: "Go Testing Best Practices",
			},
			query:    "testing",
			expected: true,
		},
		{
			name: "matches in ID",
			rule: &domain.Rule{
				ID:    "languages/go/testing",
				Title: "Best Practices",
			},
			query:    "go",
			expected: true,
		},
		{
			name: "matches multiple terms",
			rule: &domain.Rule{
				ID:    "languages/go/testing",
				Title: "Go Testing Best Practices",
			},
			query:    "go testing",
			expected: true,
		},
		{
			name: "no match",
			rule: &domain.Rule{
				ID:    "languages/python/linting",
				Title: "Python Linting",
			},
			query:    "testing",
			expected: false,
		},
		{
			name: "partial term matches",
			rule: &domain.Rule{
				ID:    "languages/go/testing",
				Title: "Testing",
			},
			query:    "go test",
			expected: true, // "test" matches as substring of "testing"
		},
		{
			name: "empty query matches all",
			rule: &domain.Rule{
				ID:    "test/rule",
				Title: "Test",
			},
			query:    "",
			expected: true,
		},
		{
			name: "case insensitive",
			rule: &domain.Rule{
				ID:    "test/rule",
				Title: "Go Testing",
			},
			query:    "GO TESTING",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := evaluator.MatchesText(tt.rule, tt.query)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestEvaluator_EvaluateExpr(t *testing.T) {
	t.Parallel()

	evaluator := NewEvaluator()

	tests := []struct {
		name      string
		rule      *domain.Rule
		expr      string
		expected  bool
		shouldErr bool
	}{
		{
			name: "tag contains",
			rule: &domain.Rule{
				ID:    "test/rule",
				Title: "Test Rule",
				Tags:  []string{"testing", "go"},
			},
			expr:     "Tag contains \"testing\"",
			expected: true,
		},
		{
			name: "title equals",
			rule: &domain.Rule{
				ID:    "test/rule",
				Title: "Test Rule",
			},
			expr:     "Title == \"Test Rule\"",
			expected: true,
		},
		{
			name: "any tag in list",
			rule: &domain.Rule{
				ID:   "test/rule",
				Tags: []string{"security", "authentication"},
			},
			expr:     "any(Tags, # in [\"security\", \"auth\"])",
			expected: true,
		},
		{
			name: "has variables",
			rule: &domain.Rule{
				ID:        "test/rule",
				Variables: map[string]any{"key": "value"},
			},
			expr:     "HasVars == true",
			expected: true,
		},
		{
			name: "complex expression",
			rule: &domain.Rule{
				ID:        "test/rule",
				Tags:      []string{"testing"},
				Languages: []string{"go"},
			},
			expr:     "Tag contains \"testing\" and Language contains \"go\"",
			expected: true,
		},
		{
			name: "no match",
			rule: &domain.Rule{
				ID:   "test/rule",
				Tags: []string{"documentation"},
			},
			expr:     "Tag contains \"testing\"",
			expected: false,
		},
		{
			name: "empty expression matches all",
			rule: &domain.Rule{
				ID: "test/rule",
			},
			expr:     "",
			expected: true,
		},
		{
			name: "invalid expression",
			rule: &domain.Rule{
				ID: "test/rule",
			},
			expr:      "invalid syntax !!",
			shouldErr: true,
		},
		{
			name: "non-boolean expression",
			rule: &domain.Rule{
				ID: "test/rule",
			},
			expr:      "Title",
			shouldErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := evaluator.EvaluateExpr(tt.rule, tt.expr)

			if tt.shouldErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestEvaluator_ExprCaching(t *testing.T) {
	t.Parallel()

	evaluator := NewEvaluator().(*evaluator)

	rule := &domain.Rule{
		ID:   "test/rule",
		Tags: []string{"testing"},
	}

	expr := "Tag contains \"testing\""

	// First evaluation should compile and cache
	result1, err := evaluator.EvaluateExpr(rule, expr)
	require.NoError(t, err)
	assert.True(t, result1)
	assert.Len(t, evaluator.programCache, 1)

	// Second evaluation should use cache
	result2, err := evaluator.EvaluateExpr(rule, expr)
	require.NoError(t, err)
	assert.True(t, result2)
	assert.Len(t, evaluator.programCache, 1) // Cache size should not increase
}
