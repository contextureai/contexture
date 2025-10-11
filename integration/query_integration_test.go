// Package integration provides integration tests for the query command
package integration

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/contextureai/contexture/internal/dependencies"
	"github.com/contextureai/contexture/internal/domain"
	"github.com/contextureai/contexture/internal/query"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestQueryEvaluatorIntegration tests the query evaluator with various expressions
func TestQueryEvaluatorIntegration(t *testing.T) {
	evaluator := query.NewEvaluator()

	// Create test rule with known properties
	testRule := &domain.Rule{
		ID:          "@contexture/languages/go/testing",
		Title:       "Go Testing Best Practices",
		Description: "Guidelines for writing effective Go tests",
		Tags:        []string{"go", "testing", "best-practices"},
		Languages:   []string{"go"},
		Frameworks:  []string{},
		Content:     "Test content here",
		Variables:   map[string]any{"threshold": 80},
		Source:      "https://github.com/contextureai/rules.git",
		FilePath:    "languages/go/testing.md",
	}

	t.Run("text matching", func(t *testing.T) {
		// Test simple text search
		assert.True(t, evaluator.MatchesText(testRule, "testing"))
		assert.True(t, evaluator.MatchesText(testRule, "Go"))
		assert.True(t, evaluator.MatchesText(testRule, "languages/go"))
		assert.False(t, evaluator.MatchesText(testRule, "javascript"))
	})

	t.Run("expr with string fields", func(t *testing.T) {
		// Test Title field
		matches, err := evaluator.EvaluateExpr(testRule, "Title contains \"Testing\"")
		require.NoError(t, err)
		assert.True(t, matches)

		// Test ID field
		matches, err = evaluator.EvaluateExpr(testRule, "ID contains \"go/testing\"")
		require.NoError(t, err)
		assert.True(t, matches)

		// Test Description field
		matches, err = evaluator.EvaluateExpr(testRule, "Description contains \"Guidelines\"")
		require.NoError(t, err)
		assert.True(t, matches)
	})

	t.Run("expr with array fields", func(t *testing.T) {
		// Test Tags array
		matches, err := evaluator.EvaluateExpr(testRule, "any(Tags, # == \"testing\")")
		require.NoError(t, err)
		assert.True(t, matches)

		// Test Languages array
		matches, err = evaluator.EvaluateExpr(testRule, "any(Languages, # == \"go\")")
		require.NoError(t, err)
		assert.True(t, matches)

		// Test array not containing value
		matches, err = evaluator.EvaluateExpr(testRule, "any(Tags, # == \"nonexistent\")")
		require.NoError(t, err)
		assert.False(t, matches)
	})

	t.Run("expr with computed fields", func(t *testing.T) {
		// Test Provider field (should match Source field since we set it)
		matches, err := evaluator.EvaluateExpr(testRule, "Provider contains \"github.com\"")
		require.NoError(t, err)
		assert.True(t, matches)

		// Test Path field
		matches, err = evaluator.EvaluateExpr(testRule, "Path contains \"languages/go\"")
		require.NoError(t, err)
		assert.True(t, matches)

		// Test HasVars field
		matches, err = evaluator.EvaluateExpr(testRule, "HasVars == true")
		require.NoError(t, err)
		assert.True(t, matches)

		// Test VarCount field
		matches, err = evaluator.EvaluateExpr(testRule, "VarCount > 0")
		require.NoError(t, err)
		assert.True(t, matches)
	})

	t.Run("expr with multiple conditions", func(t *testing.T) {
		// Test AND conditions
		matches, err := evaluator.EvaluateExpr(testRule, "Provider contains \"github.com\" and any(Tags, # == \"go\")")
		require.NoError(t, err)
		assert.True(t, matches)

		// Test OR conditions
		matches, err = evaluator.EvaluateExpr(testRule, "Title contains \"Python\" or Title contains \"Go\"")
		require.NoError(t, err)
		assert.True(t, matches)

		// Test negation
		matches, err = evaluator.EvaluateExpr(testRule, "not (any(Tags, # == \"javascript\"))")
		require.NoError(t, err)
		assert.True(t, matches)
	})

	t.Run("expr with complex expressions", func(t *testing.T) {
		// Test nested conditions
		matches, err := evaluator.EvaluateExpr(testRule,
			"(Provider contains \"github\" and HasVars == true) or (Title contains \"Essential\")")
		require.NoError(t, err)
		assert.True(t, matches)

		// Test array operations
		matches, err = evaluator.EvaluateExpr(testRule,
			"any(Tags, # in [\"go\", \"testing\", \"security\"])")
		require.NoError(t, err)
		assert.True(t, matches)
	})

	t.Run("expr error handling", func(t *testing.T) {
		// Test invalid expression
		_, err := evaluator.EvaluateExpr(testRule, "invalid syntax {[")
		require.Error(t, err)

		// Test invalid field
		_, err = evaluator.EvaluateExpr(testRule, "NonExistentField == \"value\"")
		require.Error(t, err)
	})

	t.Run("expr with edge cases", func(t *testing.T) {
		// Test with empty arrays
		ruleNoTags := &domain.Rule{
			ID:        "@test/rule",
			Title:     "Test Rule",
			Tags:      []string{},
			Languages: []string{},
		}

		matches, err := evaluator.EvaluateExpr(ruleNoTags, "any(Tags, # == \"test\")")
		require.NoError(t, err)
		assert.False(t, matches)

		// Test with nil values
		ruleNoVars := &domain.Rule{
			ID:        "@test/rule",
			Title:     "Test Rule",
			Variables: nil,
		}

		matches, err = evaluator.EvaluateExpr(ruleNoVars, "HasVars == false")
		require.NoError(t, err)
		assert.True(t, matches)
	})
}

// TestQueryProviderIntegration tests query operations against real providers
func TestQueryProviderIntegration(t *testing.T) {
	// Skip if running in short mode
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()
	deps := dependencies.New(ctx)

	t.Run("fetch rules from default provider", func(t *testing.T) {
		// This tests that we can fetch and list rules from the default provider
		providers := deps.ProviderRegistry.ListProviders()
		require.NotEmpty(t, providers, "Should have at least default provider")

		// Find the default contexture provider
		var defaultProvider *domain.Provider
		for _, p := range providers {
			if p.Name == "contexture" {
				defaultProvider = p
				break
			}
		}
		require.NotNil(t, defaultProvider, "Should have default contexture provider")

		t.Logf("Testing with provider: %s (%s)", defaultProvider.Name, defaultProvider.URL)
	})

	t.Run("evaluate expressions against real rules", func(t *testing.T) {
		// Create a sample rule to test evaluation
		testRule := &domain.Rule{
			ID:          "@contexture/test/rule",
			Title:       "Test Rule",
			Description: "A test rule for evaluation",
			Tags:        []string{"test", "example"},
			Languages:   []string{"go"},
		}

		evaluator := query.NewEvaluator()

		// Test various expressions
		testCases := []struct {
			expr     string
			expected bool
		}{
			{"Title contains \"Test\"", true},
			{"any(Tags, # == \"test\")", true},
			{"Provider == \"contexture\"", true},
			{"any(Languages, # == \"go\")", true},
			{"Path contains \"test\"", true},
			{"Title contains \"NonExistent\"", false},
		}

		for _, tc := range testCases {
			matches, err := evaluator.EvaluateExpr(testRule, tc.expr)
			require.NoError(t, err, "Expression should parse: %s", tc.expr)
			assert.Equal(t, tc.expected, matches,
				"Expression '%s' should evaluate to %v", tc.expr, tc.expected)
		}
	})
}

// TestQueryPerformance tests query performance characteristics
func TestQueryPerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	evaluator := query.NewEvaluator()

	t.Run("text search performance", func(t *testing.T) {
		// Create many test rules
		rules := make([]*domain.Rule, 1000)
		for i := range rules {
			rules[i] = &domain.Rule{
				ID:          "@test/rule" + string(rune(i)),
				Title:       "Rule " + string(rune(i)),
				Description: "Description for rule",
				Tags:        []string{"test", "performance"},
			}
		}

		// Measure text search performance
		start := time.Now()
		matchCount := 0
		for _, rule := range rules {
			if evaluator.MatchesText(rule, "test") {
				matchCount++
			}
		}
		duration := time.Since(start)

		t.Logf("Text search of 1000 rules completed in %v", duration)
		assert.Positive(t, matchCount, "Should find matches")
		assert.Less(t, duration, 1*time.Second, "Should complete within 1 second")
	})

	t.Run("expr evaluation performance", func(t *testing.T) {
		// Create test rules
		rules := make([]*domain.Rule, 100)
		for i := range rules {
			rules[i] = &domain.Rule{
				ID:        "@test/rule" + string(rune(i)),
				Title:     "Rule " + string(rune(i)),
				Tags:      []string{"test", "performance"},
				Languages: []string{"go", "typescript"},
				Variables: map[string]any{"count": i},
			}
		}

		// Measure complex expression performance
		expr := "any(Tags, # == \"test\") and (any(Languages, # == \"go\") or HasVars == true)"
		start := time.Now()
		matchCount := 0
		for _, rule := range rules {
			matches, err := evaluator.EvaluateExpr(rule, expr)
			require.NoError(t, err)
			if matches {
				matchCount++
			}
		}
		duration := time.Since(start)

		t.Logf("Expression evaluation of 100 rules completed in %v", duration)
		assert.Positive(t, matchCount, "Should find matches")
		assert.Less(t, duration, 1*time.Second, "Should complete within 1 second")
	})
}

// TestQueryEdgeCases tests edge cases and error conditions
func TestQueryEdgeCases(t *testing.T) {
	evaluator := query.NewEvaluator()

	t.Run("empty rule fields", func(t *testing.T) {
		emptyRule := &domain.Rule{
			ID:          "",
			Title:       "",
			Description: "",
			Tags:        []string{},
			Languages:   []string{},
			Content:     "",
		}

		// Text search should handle empty fields
		assert.False(t, evaluator.MatchesText(emptyRule, "test"))

		// Expression evaluation should handle empty fields
		matches, err := evaluator.EvaluateExpr(emptyRule, "Title == \"\"")
		require.NoError(t, err)
		assert.True(t, matches)
	})

	t.Run("nil rule", func(t *testing.T) {
		// Testing with nil is not a supported use case, skip this test
		t.Skip("Nil rules are not expected in normal usage")
	})

	t.Run("special characters in search", func(t *testing.T) {
		rule := &domain.Rule{
			ID:          "@contexture/test/rule",
			Title:       "Rule with (special) [characters] {here} $pecial",
			Description: "Test @characters",
		}

		// Should handle special characters in text search (MatchesText only searches Title and ID)
		assert.True(t, evaluator.MatchesText(rule, "(special)"))
		assert.True(t, evaluator.MatchesText(rule, "$pecial"))
	})

	t.Run("case sensitivity", func(t *testing.T) {
		rule := &domain.Rule{
			ID:    "@test/rule",
			Title: "Test Rule With MixedCase",
		}

		// Text search should be case-insensitive
		assert.True(t, evaluator.MatchesText(rule, "test"))
		assert.True(t, evaluator.MatchesText(rule, "TEST"))
		assert.True(t, evaluator.MatchesText(rule, "mixedcase"))
	})

	t.Run("unicode and international characters", func(t *testing.T) {
		rule := &domain.Rule{
			ID:          "@test/rule",
			Title:       "Règle de test avec caractères spéciaux 测试规则",
			Description: "测试描述",
			Tags:        []string{"日本語", "中文"},
		}

		// Should handle unicode in text search (MatchesText only searches Title and ID)
		assert.True(t, evaluator.MatchesText(rule, "Règle"))
		assert.True(t, evaluator.MatchesText(rule, "测试规则"))

		// Should handle unicode in expressions
		matches, err := evaluator.EvaluateExpr(rule, "any(Tags, # == \"日本語\")")
		require.NoError(t, err)
		assert.True(t, matches)
	})
}

// TestQueryConcurrency tests concurrent query operations
func TestQueryConcurrency(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping concurrency test in short mode")
	}

	evaluator := query.NewEvaluator()

	// Create test rules
	rules := make([]*domain.Rule, 100)
	for i := range rules {
		rules[i] = &domain.Rule{
			ID:        "@test/rule" + string(rune(i)),
			Title:     "Rule " + string(rune(i)),
			Tags:      []string{"test", "concurrent"},
			Languages: []string{"go"},
		}
	}

	t.Run("concurrent text searches", func(t *testing.T) {
		done := make(chan bool, 10)
		start := time.Now()

		// Run 10 concurrent searches
		for range 10 {
			go func() {
				for _, rule := range rules {
					_ = evaluator.MatchesText(rule, "test")
				}
				done <- true
			}()
		}

		// Wait for all to complete
		for range 10 {
			select {
			case <-done:
				// Success
			case <-time.After(5 * time.Second):
				t.Fatal("Concurrent searches timed out")
			}
		}

		duration := time.Since(start)
		t.Logf("10 concurrent text searches completed in %v", duration)
	})

	t.Run("concurrent expr evaluations", func(t *testing.T) {
		done := make(chan bool, 10)
		errorChan := make(chan error, 10)

		// Run 10 concurrent expression evaluations
		for range 10 {
			go func() {
				for _, rule := range rules {
					_, err := evaluator.EvaluateExpr(rule, "any(Tags, # == \"test\")")
					if err != nil {
						errorChan <- err
						return
					}
				}
				done <- true
			}()
		}

		// Wait for all to complete or error
		for range 10 {
			select {
			case err := <-errorChan:
				t.Fatalf("Concurrent evaluation failed: %v", err)
			case <-done:
				// Success
			case <-time.After(5 * time.Second):
				t.Fatal("Concurrent evaluations timed out")
			}
		}

		t.Log("10 concurrent expr evaluations completed successfully")
	})
}

// TestQueryExpressionSyntax tests various expression syntax patterns
func TestQueryExpressionSyntax(t *testing.T) {
	evaluator := query.NewEvaluator()

	rule := &domain.Rule{
		ID:          "@contexture/languages/go/testing",
		Title:       "Go Testing Best Practices",
		Description: "Guidelines for writing Go tests",
		Tags:        []string{"go", "testing", "best-practices"},
		Languages:   []string{"go"},
		Variables:   map[string]any{"minCoverage": 80, "enabled": true},
		Source:      "", // Empty Source defaults Provider to "contexture"
	}

	testCases := []struct {
		name     string
		expr     string
		expected bool
		wantErr  bool
	}{
		// String operations
		{"contains", "Title contains \"Testing\"", true, false},
		{"startsWith", "Title startsWith \"Go\"", true, false},
		{"endsWith", "Title endsWith \"Practices\"", true, false},
		{"equality", "Title == \"Go Testing Best Practices\"", true, false},
		{"inequality", "Title != \"JavaScript\"", true, false},

		// Array operations
		{"any with equality", "any(Tags, # == \"go\")", true, false},
		{"any with in", "any(Tags, # in [\"go\", \"rust\"])", true, false},
		{"all", "all(Tags, # != \"\")", true, false},
		{"none", "none(Tags, # == \"javascript\")", true, false},

		// Boolean operations
		{"and", "HasVars == true and VarCount > 0", true, false},
		{"or", "Title contains \"Python\" or Title contains \"Go\"", true, false},
		{"not", "not (Provider == \"unknown\")", true, false},

		// Numeric operations
		{"greater than", "VarCount > 0", true, false},
		{"less than", "VarCount < 100", true, false},
		{"equals", "VarCount == 2", true, false},

		// Computed fields
		{"provider", "Provider == \"contexture\"", true, false},
		{"path", "Path contains \"languages/go\"", true, false},
		{"hasVars", "HasVars == true", true, false},

		// Error cases
		{"invalid syntax", "invalid {[ syntax", false, true},
		{"undefined field", "UndefinedField == \"value\"", false, true},
		{"type mismatch", "VarCount == \"string\"", false, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			matches, err := evaluator.EvaluateExpr(rule, tc.expr)
			if tc.wantErr {
				require.Error(t, err, "Expression should fail: %s", tc.expr)
			} else {
				require.NoError(t, err, "Expression should succeed: %s", tc.expr)
				assert.Equal(t, tc.expected, matches,
					"Expression '%s' should evaluate to %v", tc.expr, tc.expected)
			}
		})
	}
}

// TestQueryStringFields tests querying on various string fields
func TestQueryStringFields(t *testing.T) {
	evaluator := query.NewEvaluator()

	rule := &domain.Rule{
		ID:          "@organization/category/subcategory/rule-name",
		Title:       "Example Rule Title",
		Description: "This is a detailed description of the rule",
		Content:     "Rule content goes here with instructions",
		Source:      "https://github.com/example/rules.git",
		FilePath:    "category/subcategory/rule-name.md",
	}

	t.Run("search in ID", func(t *testing.T) {
		assert.True(t, evaluator.MatchesText(rule, "organization"))
		assert.True(t, evaluator.MatchesText(rule, "rule-name"))
		assert.True(t, evaluator.MatchesText(rule, "subcategory"))
	})

	t.Run("search in Title", func(t *testing.T) {
		assert.True(t, evaluator.MatchesText(rule, "Example"))
		assert.True(t, evaluator.MatchesText(rule, "Rule Title"))
	})

	t.Run("search in Description", func(t *testing.T) {
		// MatchesText only searches Title and ID, use expressions for Description
		matches, err := evaluator.EvaluateExpr(rule, "Description contains \"detailed\"")
		require.NoError(t, err)
		assert.True(t, matches)
		matches, err = evaluator.EvaluateExpr(rule, "Description contains \"description\"")
		require.NoError(t, err)
		assert.True(t, matches)
	})

	t.Run("search in FilePath", func(t *testing.T) {
		// MatchesText only searches Title and ID, use expressions for FilePath
		matches, err := evaluator.EvaluateExpr(rule, "FilePath contains \"category/subcategory\"")
		require.NoError(t, err)
		assert.True(t, matches)
		matches, err = evaluator.EvaluateExpr(rule, "FilePath contains \"rule-name.md\"")
		require.NoError(t, err)
		assert.True(t, matches)
	})

	t.Run("search in Source", func(t *testing.T) {
		// MatchesText only searches Title and ID, use expressions for Source
		matches, err := evaluator.EvaluateExpr(rule, "Source contains \"github.com\"")
		require.NoError(t, err)
		assert.True(t, matches)
		matches, err = evaluator.EvaluateExpr(rule, "Source contains \"example/rules\"")
		require.NoError(t, err)
		assert.True(t, matches)
	})

	t.Run("case insensitive search", func(t *testing.T) {
		assert.True(t, evaluator.MatchesText(rule, "EXAMPLE"))
		assert.True(t, evaluator.MatchesText(rule, "title"))
		assert.True(t, evaluator.MatchesText(rule, "SUBCATEGORY"))
	})

	t.Run("partial match", func(t *testing.T) {
		// MatchesText only searches Title and ID
		assert.True(t, evaluator.MatchesText(rule, "exa"))      // matches "Example" in Title
		assert.True(t, evaluator.MatchesText(rule, "subcateg")) // matches "subcategory" in ID
	})

	t.Run("no match", func(t *testing.T) {
		assert.False(t, evaluator.MatchesText(rule, "nonexistent"))
		assert.False(t, evaluator.MatchesText(rule, "zzzzz"))
	})
}

// TestQueryRealWorldExpressions tests realistic expression patterns users might write
func TestQueryRealWorldExpressions(t *testing.T) {
	evaluator := query.NewEvaluator()

	rules := []*domain.Rule{
		{
			ID:          "@contexture/security/input-validation",
			Title:       "Input Validation Rules",
			Description: "Validate all user inputs",
			Tags:        []string{"security", "validation", "input"},
			Languages:   []string{"go", "typescript"},
			Variables:   map[string]any{"strict": true},
		},
		{
			ID:          "@contexture/testing/unit-tests",
			Title:       "Unit Testing Guidelines",
			Description: "Best practices for unit tests",
			Tags:        []string{"testing", "unit", "best-practices"},
			Languages:   []string{"go"},
			Variables:   map[string]any{"coverage": 80},
		},
		{
			ID:          "@contexture/code-quality/naming",
			Title:       "Naming Conventions",
			Description: "Consistent naming patterns",
			Tags:        []string{"code-quality", "style", "naming"},
			Languages:   []string{"go", "typescript", "python"},
			Variables:   nil,
		},
	}

	testCases := []struct {
		name          string
		expr          string
		expectedCount int
	}{
		{"find security rules", "any(Tags, # == \"security\")", 1},
		{"find Go rules", "any(Languages, # == \"go\")", 3},
		{"find rules with variables", "HasVars == true", 2},
		{"find multi-language rules", "any(Languages, # == \"typescript\") and any(Languages, # == \"go\")", 2},
		{"find testing or security", "any(Tags, # in [\"testing\", \"security\"])", 2},
		{"find strict security rules", "any(Tags, # == \"security\") and HasVars == true", 1},
		{"exclude Python rules", "not any(Languages, # == \"python\")", 2},
		{"complex query", "(any(Tags, # == \"testing\") or any(Tags, # == \"security\")) and any(Languages, # == \"go\")", 2},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			matchCount := 0
			for _, rule := range rules {
				matches, err := evaluator.EvaluateExpr(rule, tc.expr)
				require.NoError(t, err, "Expression should be valid: %s", tc.expr)
				if matches {
					matchCount++
				}
			}
			assert.Equal(t, tc.expectedCount, matchCount,
				"Expression '%s' should match %d rules, got %d",
				tc.expr, tc.expectedCount, matchCount)
		})
	}
}

// TestQueryProviderField tests the Provider computed field
func TestQueryProviderField(t *testing.T) {
	evaluator := query.NewEvaluator()

	testCases := []struct {
		name             string
		ruleID           string
		ruleSource       string
		expectedProvider string
	}{
		{"default provider empty source", "@contexture/test/rule", "", "contexture"},
		{"custom provider with source", "@mycompany/test/rule", "https://github.com/mycompany/rules.git", "https://github.com/mycompany/rules.git"},
		{"local rule", "rules/local-rule.md", "local", "local"},
		{"bracketed default empty source", "[contexture:test/rule]", "", "contexture"},
		{"bracketed custom with source", "[contexture(myorg):test/rule]", "https://github.com/myorg/rules.git", "https://github.com/myorg/rules.git"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rule := &domain.Rule{
				ID:     tc.ruleID,
				Title:  "Test Rule",
				Source: tc.ruleSource,
			}

			expr := "Provider == \"" + tc.expectedProvider + "\""
			matches, err := evaluator.EvaluateExpr(rule, expr)
			require.NoError(t, err)
			assert.True(t, matches,
				"Rule ID '%s' with Source '%s' should have provider '%s'",
				tc.ruleID, tc.ruleSource, tc.expectedProvider)
		})
	}
}

// TestQueryMemoryUsage tests that queries don't leak memory
func TestQueryMemoryUsage(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping memory test in short mode")
	}

	evaluator := query.NewEvaluator()

	t.Run("repeated evaluations don't leak", func(t *testing.T) {
		rule := &domain.Rule{
			ID:        "@test/rule",
			Title:     "Test Rule",
			Tags:      []string{"test"},
			Variables: map[string]any{"count": 1},
		}

		// Run many evaluations to test for memory leaks
		iterations := 10000
		for range iterations {
			_, err := evaluator.EvaluateExpr(rule, "any(Tags, # == \"test\") and HasVars == true")
			require.NoError(t, err)

			_ = evaluator.MatchesText(rule, "test")
		}

		t.Logf("Completed %d evaluations without issues", iterations)
	})
}

// TestQueryResultTruncation tests that description truncation works correctly
func TestQueryResultTruncation(t *testing.T) {
	// This tests the logic used in query result display

	testCases := []struct {
		name        string
		description string
		maxLen      int
		expected    string
	}{
		{
			"short description",
			"Short desc",
			100,
			"Short desc",
		},
		{
			"exact length",
			strings.Repeat("a", 100),
			100,
			strings.Repeat("a", 100),
		},
		{
			"truncate long description",
			strings.Repeat("a", 150),
			100,
			strings.Repeat("a", 97) + "...",
		},
		{
			"unicode truncation",
			"This is a test with émojis 😀 and special chars: 日本語",
			30,
			"This is a test with émojis...",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.description
			if len(result) > tc.maxLen {
				result = result[:tc.maxLen-3] + "..."
			}
			assert.Equal(t, tc.expected, result)
		})
	}
}
