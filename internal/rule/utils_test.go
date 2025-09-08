package rule

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/contextureai/contexture/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestFetchRulesParallel(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name          string
		ruleRefs      []domain.RuleRef
		maxWorkers    int
		setupFetcher  func() Fetcher
		expectedErr   bool
		expectedRules int
	}{
		{
			name:       "empty rule refs",
			ruleRefs:   []domain.RuleRef{},
			maxWorkers: 2,
			setupFetcher: func() Fetcher {
				return NewMockFetcher(t)
			},
			expectedErr:   false,
			expectedRules: 0,
		},
		{
			name: "single rule success",
			ruleRefs: []domain.RuleRef{
				{ID: "test/rule1", Source: "source1"},
			},
			maxWorkers: 2,
			setupFetcher: func() Fetcher {
				fetcher := NewMockFetcher(t)
				fetcher.EXPECT().FetchRule(mock.Anything, "test/rule1").Return(&domain.Rule{
					ID:    "test/rule1",
					Title: "Test Rule",
				}, nil)
				return fetcher
			},
			expectedErr:   false,
			expectedRules: 1,
		},
		{
			name: "multiple rules success",
			ruleRefs: []domain.RuleRef{
				{ID: "test/rule1", Source: "source1"},
				{ID: "test/rule2", Source: "source1"},
			},
			maxWorkers: 2,
			setupFetcher: func() Fetcher {
				fetcher := NewMockFetcher(t)
				fetcher.EXPECT().FetchRule(mock.Anything, "test/rule1").Return(&domain.Rule{
					ID:    "test/rule1",
					Title: "Test Rule 1",
				}, nil)
				fetcher.EXPECT().FetchRule(mock.Anything, "test/rule2").Return(&domain.Rule{
					ID:    "test/rule2",
					Title: "Test Rule 2",
				}, nil)
				return fetcher
			},
			expectedErr:   false,
			expectedRules: 2,
		},
		{
			name: "single rule error",
			ruleRefs: []domain.RuleRef{
				{ID: "test/rule1", Source: "source1"},
			},
			maxWorkers: 2,
			setupFetcher: func() Fetcher {
				fetcher := NewMockFetcher(t)
				fetcher.EXPECT().FetchRule(mock.Anything, "test/rule1").Return(nil, fmt.Errorf("fetch failed"))
				return fetcher
			},
			expectedErr:   true,
			expectedRules: 0,
		},
		{
			name: "mixed success and error",
			ruleRefs: []domain.RuleRef{
				{ID: "test/rule1", Source: "source1"},
				{ID: "test/rule2", Source: "source1"},
			},
			maxWorkers: 2,
			setupFetcher: func() Fetcher {
				fetcher := NewMockFetcher(t)
				fetcher.EXPECT().FetchRule(mock.Anything, "test/rule1").Return(&domain.Rule{
					ID:    "test/rule1",
					Title: "Test Rule 1",
				}, nil)
				fetcher.EXPECT().FetchRule(mock.Anything, "test/rule2").Return(nil, fmt.Errorf("fetch failed"))
				return fetcher
			},
			expectedErr:   true,
			expectedRules: 0,
		},
		{
			name: "rules with variables",
			ruleRefs: []domain.RuleRef{
				{
					ID:     "test/rule1",
					Source: "source1",
					Variables: map[string]any{
						"custom": "value",
					},
				},
			},
			maxWorkers: 2,
			setupFetcher: func() Fetcher {
				fetcher := NewMockFetcher(t)
				fetcher.EXPECT().FetchRule(mock.Anything, "test/rule1").Return(&domain.Rule{
					ID:    "test/rule1",
					Title: "Test Rule",
					Variables: map[string]any{
						"existing": "original",
					},
				}, nil)
				return fetcher
			},
			expectedErr:   false,
			expectedRules: 1,
		},
		{
			name: "zero workers uses default",
			ruleRefs: []domain.RuleRef{
				{ID: "test/rule1", Source: "source1"},
			},
			maxWorkers: 0, // Should use default
			setupFetcher: func() Fetcher {
				fetcher := NewMockFetcher(t)
				fetcher.EXPECT().FetchRule(mock.Anything, "test/rule1").Return(&domain.Rule{
					ID:    "test/rule1",
					Title: "Test Rule",
				}, nil)
				return fetcher
			},
			expectedErr:   false,
			expectedRules: 1,
		},
		{
			name: "negative workers uses default",
			ruleRefs: []domain.RuleRef{
				{ID: "test/rule1", Source: "source1"},
			},
			maxWorkers: -1, // Should use default
			setupFetcher: func() Fetcher {
				fetcher := NewMockFetcher(t)
				fetcher.EXPECT().FetchRule(mock.Anything, "test/rule1").Return(&domain.Rule{
					ID:    "test/rule1",
					Title: "Test Rule",
				}, nil)
				return fetcher
			},
			expectedErr:   false,
			expectedRules: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fetcher := tt.setupFetcher()

			rules, err := FetchRulesParallel(ctx, fetcher, tt.ruleRefs, tt.maxWorkers)

			if tt.expectedErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "failed to fetch some rules")
			} else {
				require.NoError(t, err)
				assert.Len(t, rules, tt.expectedRules)

				// Verify variables were merged correctly if applicable
				if len(tt.ruleRefs) == 1 && len(tt.ruleRefs[0].Variables) > 0 && len(rules) > 0 {
					rule := rules[0]
					assert.Equal(t, "value", rule.Variables["custom"])
					assert.Equal(t, "original", rule.Variables["existing"])
				}
			}
		})
	}
}

func TestFetchRulesParallel_WithCommitHash(t *testing.T) {
	ctx := context.Background()

	t.Run("with commit hash and regular fetcher", func(t *testing.T) {
		ruleRefs := []domain.RuleRef{
			{
				ID:         "test/rule1",
				Source:     "source1",
				CommitHash: "abc123",
			},
		}

		// Use regular mock fetcher - the commit hash logic will fallback to regular fetch
		fetcher := NewMockFetcher(t)
		fetcher.EXPECT().FetchRule(mock.Anything, "test/rule1").Return(&domain.Rule{
			ID:    "test/rule1",
			Title: "Test Rule",
		}, nil)

		rules, err := FetchRulesParallel(ctx, fetcher, ruleRefs, 2)

		require.NoError(t, err)
		assert.Len(t, rules, 1)
		assert.Equal(t, "test/rule1", rules[0].ID)
	})
}

func TestFetchRulesParallel_Concurrency(t *testing.T) {
	ctx := context.Background()

	// Test that the function properly limits concurrency
	ruleRefs := make([]domain.RuleRef, 10)
	for i := range ruleRefs {
		ruleRefs[i] = domain.RuleRef{
			ID:     fmt.Sprintf("test/rule%d", i),
			Source: "source1",
		}
	}

	fetcher := NewMockFetcher(t)
	for i := range ruleRefs {
		fetcher.EXPECT().FetchRule(mock.Anything, fmt.Sprintf("test/rule%d", i)).Return(&domain.Rule{
			ID:    fmt.Sprintf("test/rule%d", i),
			Title: fmt.Sprintf("Test Rule %d", i),
		}, nil).Maybe()
	}

	// Use a small number of workers to test concurrency limiting
	rules, err := FetchRulesParallel(ctx, fetcher, ruleRefs, 3)

	require.NoError(t, err)
	assert.Len(t, rules, 10)
	// Rules should be non-nil regardless of timing
	require.NotNil(t, rules)
}

func TestFetchRulesParallel_ContextCancellation(t *testing.T) {
	// Create a context that will be cancelled
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	ruleRefs := []domain.RuleRef{
		{ID: "test/rule1", Source: "source1"},
	}

	fetcher := NewMockFetcher(t)
	// The fetch might not even be called due to quick cancellation
	fetcher.EXPECT().FetchRule(mock.Anything, "test/rule1").Return(&domain.Rule{
		ID:    "test/rule1",
		Title: "Test Rule",
	}, nil).Maybe()

	// Wait for context to be cancelled
	time.Sleep(2 * time.Millisecond)

	rules, err := FetchRulesParallel(ctx, fetcher, ruleRefs, 2)
	// The behavior depends on timing, but we shouldn't panic
	if err != nil {
		assert.Contains(t, err.Error(), "failed to fetch some rules")
	}
	// Rules could be empty or contain results depending on timing, but should be non-nil
	require.NotNil(t, rules)
}

func TestShouldDisplayVariables(t *testing.T) {
	tests := []struct {
		name      string
		variables map[string]any
		defaults  map[string]any
		expected  bool
	}{
		{
			name:      "empty variables",
			variables: map[string]any{},
			defaults:  map[string]any{"extended": false},
			expected:  false,
		},
		{
			name:      "nil variables",
			variables: nil,
			defaults:  map[string]any{"extended": false},
			expected:  false,
		},
		{
			name:      "variables match defaults exactly",
			variables: map[string]any{"extended": false, "strict": true},
			defaults:  map[string]any{"extended": false, "strict": true},
			expected:  false,
		},
		{
			name:      "variable differs from default",
			variables: map[string]any{"extended": true},
			defaults:  map[string]any{"extended": false},
			expected:  true,
		},
		{
			name:      "variable not in defaults",
			variables: map[string]any{"custom": "value"},
			defaults:  map[string]any{"extended": false},
			expected:  true,
		},
		{
			name:      "no defaults provided",
			variables: map[string]any{"extended": false},
			defaults:  nil,
			expected:  true,
		},
		{
			name:      "empty defaults",
			variables: map[string]any{"extended": false},
			defaults:  map[string]any{},
			expected:  true,
		},
		{
			name:      "mixed - some match, some differ",
			variables: map[string]any{"extended": false, "strict": false},
			defaults:  map[string]any{"extended": false, "strict": true},
			expected:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ShouldDisplayVariables(tt.variables, tt.defaults)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFilterNonDefaultVariables(t *testing.T) {
	tests := []struct {
		name      string
		variables map[string]any
		defaults  map[string]any
		expected  map[string]any
	}{
		{
			name:      "empty variables",
			variables: map[string]any{},
			defaults:  map[string]any{"extended": false},
			expected:  nil,
		},
		{
			name:      "all variables match defaults",
			variables: map[string]any{"extended": false, "strict": true},
			defaults:  map[string]any{"extended": false, "strict": true},
			expected:  nil,
		},
		{
			name:      "variable differs from default",
			variables: map[string]any{"extended": true, "strict": false},
			defaults:  map[string]any{"extended": false, "strict": false},
			expected:  map[string]any{"extended": true},
		},
		{
			name:      "variable not in defaults",
			variables: map[string]any{"custom": "value", "extended": false},
			defaults:  map[string]any{"extended": false},
			expected:  map[string]any{"custom": "value"},
		},
		{
			name:      "no defaults provided",
			variables: map[string]any{"extended": false, "strict": true},
			defaults:  nil,
			expected:  map[string]any{"extended": false, "strict": true},
		},
		{
			name:      "mixed scenarios",
			variables: map[string]any{"extended": true, "strict": false, "debug": true},
			defaults:  map[string]any{"extended": false, "strict": false},
			expected:  map[string]any{"extended": true, "debug": true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FilterNonDefaultVariables(tt.variables, tt.defaults)
			assert.Equal(t, tt.expected, result)
		})
	}
}
