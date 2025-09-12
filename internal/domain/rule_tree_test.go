package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRuleTree(t *testing.T) {
	t.Parallel()
	t.Run("creates empty tree with no rules", func(t *testing.T) {
		tree := NewRuleTree([]string{})

		assert.Empty(t, tree.Name)
		assert.Equal(t, RuleNodeTypeFolder, tree.Type)
		assert.Empty(t, tree.Path)
		assert.Empty(t, tree.Children)
	})

	t.Run("creates tree with single rule", func(t *testing.T) {
		tree := NewRuleTree([]string{"test-rule"})

		assert.Equal(t, RuleNodeTypeFolder, tree.Type)
		assert.Len(t, tree.Children, 1)

		rule := tree.Children["test-rule"]
		require.NotNil(t, rule)
		assert.Equal(t, "test-rule", rule.Name)
		assert.Equal(t, RuleNodeTypeRule, rule.Type)
		assert.Equal(t, "test-rule", rule.Path)
		assert.Equal(t, "test-rule", rule.RuleID)
	})

	t.Run("creates tree with nested rules", func(t *testing.T) {
		tree := NewRuleTree([]string{
			"languages/go/testing",
			"languages/go/concurrency",
			"languages/javascript/async",
			"frameworks/react/hooks",
		})

		// Check root structure
		assert.Equal(t, RuleNodeTypeFolder, tree.Type)
		assert.Len(t, tree.Children, 2) // languages and frameworks

		// Check languages folder
		languages := tree.Children["languages"]
		require.NotNil(t, languages)
		assert.Equal(t, "languages", languages.Name)
		assert.Equal(t, RuleNodeTypeFolder, languages.Type)
		assert.Equal(t, "languages", languages.Path)
		assert.Len(t, languages.Children, 2) // go and javascript

		// Check go subfolder
		goFolder := languages.Children["go"]
		require.NotNil(t, goFolder)
		assert.Equal(t, "go", goFolder.Name)
		assert.Equal(t, RuleNodeTypeFolder, goFolder.Type)
		assert.Equal(t, "languages/go", goFolder.Path)
		assert.Len(t, goFolder.Children, 2) // testing and concurrency

		// Check specific rule
		testingRule := goFolder.Children["testing"]
		require.NotNil(t, testingRule)
		assert.Equal(t, "testing", testingRule.Name)
		assert.Equal(t, RuleNodeTypeRule, testingRule.Type)
		assert.Equal(t, "languages/go/testing", testingRule.Path)
		assert.Equal(t, "languages/go/testing", testingRule.RuleID)
	})
}

func TestAddRuleToTree(t *testing.T) {
	t.Parallel()
	t.Run("handles empty path", func(t *testing.T) {
		root := &RuleNode{
			Type:     RuleNodeTypeFolder,
			Children: make(map[string]*RuleNode),
		}

		addRuleToTree(root, "")
		assert.Empty(t, root.Children)
	})

	t.Run("handles path with only separators", func(t *testing.T) {
		root := &RuleNode{
			Type:     RuleNodeTypeFolder,
			Children: make(map[string]*RuleNode),
		}

		addRuleToTree(root, "/")
		assert.Empty(t, root.Children)
	})

	t.Run("normalizes path separators", func(t *testing.T) {
		root := &RuleNode{
			Type:     RuleNodeTypeFolder,
			Children: make(map[string]*RuleNode),
		}

		addRuleToTree(root, "languages\\go\\testing")

		languages := root.Children["languages"]
		require.NotNil(t, languages)

		goFolder := languages.Children["go"]
		require.NotNil(t, goFolder)

		testing := goFolder.Children["testing"]
		require.NotNil(t, testing)
		assert.Equal(t, "languages/go/testing", testing.Path)
	})

	t.Run("handles single component path", func(t *testing.T) {
		root := &RuleNode{
			Type:     RuleNodeTypeFolder,
			Children: make(map[string]*RuleNode),
		}

		addRuleToTree(root, "single-rule")

		rule := root.Children["single-rule"]
		require.NotNil(t, rule)
		assert.Equal(t, "single-rule", rule.Name)
		assert.Equal(t, RuleNodeTypeRule, rule.Type)
		assert.Equal(t, "single-rule", rule.Path)
		assert.Equal(t, "single-rule", rule.RuleID)
	})
}

func TestRuleNode_GetChildren(t *testing.T) {
	t.Parallel()
	t.Run("returns empty slice for rule node", func(t *testing.T) {
		rule := &RuleNode{
			Type: RuleNodeTypeRule,
		}

		children := rule.GetChildren()
		assert.Empty(t, children)
	})

	t.Run("returns sorted children for folder node", func(t *testing.T) {
		folder := &RuleNode{
			Type: RuleNodeTypeFolder,
			Children: map[string]*RuleNode{
				"zebra": {Name: "zebra", Type: RuleNodeTypeFolder},
				"alpha": {Name: "alpha", Type: RuleNodeTypeRule},
				"beta":  {Name: "beta", Type: RuleNodeTypeFolder},
			},
		}

		children := folder.GetChildren()
		require.Len(t, children, 3)

		// Should be sorted: folders first (beta, zebra), then rules (alpha)
		assert.Equal(t, "beta", children[0].Name)
		assert.Equal(t, RuleNodeTypeFolder, children[0].Type)

		assert.Equal(t, "zebra", children[1].Name)
		assert.Equal(t, RuleNodeTypeFolder, children[1].Type)

		assert.Equal(t, "alpha", children[2].Name)
		assert.Equal(t, RuleNodeTypeRule, children[2].Type)
	})
}

func TestRuleNode_FindNodeByPath(t *testing.T) {
	t.Parallel()
	tree := NewRuleTree([]string{
		"languages/go/testing",
		"languages/javascript/async",
		"frameworks/react/hooks",
	})

	t.Run("finds root node with empty path", func(t *testing.T) {
		node := tree.FindNodeByPath("")
		assert.Equal(t, tree, node)
	})

	t.Run("finds folder node", func(t *testing.T) {
		node := tree.FindNodeByPath("languages")
		require.NotNil(t, node)
		assert.Equal(t, "languages", node.Name)
		assert.Equal(t, RuleNodeTypeFolder, node.Type)
	})

	t.Run("finds nested folder node", func(t *testing.T) {
		node := tree.FindNodeByPath("languages/go")
		require.NotNil(t, node)
		assert.Equal(t, "go", node.Name)
		assert.Equal(t, RuleNodeTypeFolder, node.Type)
		assert.Equal(t, "languages/go", node.Path)
	})

	t.Run("finds rule node", func(t *testing.T) {
		node := tree.FindNodeByPath("languages/go/testing")
		require.NotNil(t, node)
		assert.Equal(t, "testing", node.Name)
		assert.Equal(t, RuleNodeTypeRule, node.Type)
		assert.Equal(t, "languages/go/testing", node.RuleID)
	})

	t.Run("returns nil for non-existent path", func(t *testing.T) {
		node := tree.FindNodeByPath("non/existent/path")
		assert.Nil(t, node)
	})

	t.Run("returns nil for partial non-existent path", func(t *testing.T) {
		node := tree.FindNodeByPath("languages/python")
		assert.Nil(t, node)
	})
}

func TestRuleNode_GetParentPath(t *testing.T) {
	t.Parallel()
	t.Run("returns empty string for root node", func(t *testing.T) {
		root := &RuleNode{Path: ""}
		assert.Empty(t, root.GetParentPath())
	})

	t.Run("returns dot for top-level node", func(t *testing.T) {
		node := &RuleNode{Path: "languages"}
		assert.Equal(t, ".", node.GetParentPath())
	})

	t.Run("returns parent path for nested node", func(t *testing.T) {
		node := &RuleNode{Path: "languages/go/testing"}
		assert.Equal(t, "languages/go", node.GetParentPath())
	})

	t.Run("returns parent path for two-level node", func(t *testing.T) {
		node := &RuleNode{Path: "languages/go"}
		assert.Equal(t, "languages", node.GetParentPath())
	})
}

func TestRuleNode_GetAllRules(t *testing.T) {
	t.Parallel()
	tree := NewRuleTree([]string{
		"languages/go/testing",
		"languages/go/concurrency",
		"languages/javascript/async",
		"frameworks/react/hooks",
	})

	t.Run("returns all rules from root", func(t *testing.T) {
		rules := tree.GetAllRules()
		assert.Len(t, rules, 4)

		// Collect rule IDs for verification
		ruleIDs := make([]string, len(rules))
		for i, rule := range rules {
			ruleIDs[i] = rule.RuleID
		}

		assert.Contains(t, ruleIDs, "languages/go/testing")
		assert.Contains(t, ruleIDs, "languages/go/concurrency")
		assert.Contains(t, ruleIDs, "languages/javascript/async")
		assert.Contains(t, ruleIDs, "frameworks/react/hooks")
	})

	t.Run("returns rules from subfolder", func(t *testing.T) {
		goFolder := tree.FindNodeByPath("languages/go")
		require.NotNil(t, goFolder)

		rules := goFolder.GetAllRules()
		assert.Len(t, rules, 2)

		ruleIDs := make([]string, len(rules))
		for i, rule := range rules {
			ruleIDs[i] = rule.RuleID
		}

		assert.Contains(t, ruleIDs, "languages/go/testing")
		assert.Contains(t, ruleIDs, "languages/go/concurrency")
	})

	t.Run("returns single rule for rule node", func(t *testing.T) {
		rule := tree.FindNodeByPath("languages/go/testing")
		require.NotNil(t, rule)

		rules := rule.GetAllRules()
		assert.Len(t, rules, 1)
		assert.Equal(t, rule, rules[0])
	})
}

func TestRuleNode_GetBreadcrumb(t *testing.T) {
	t.Parallel()
	tree := NewRuleTree([]string{
		"languages/go/testing",
	})

	t.Run("returns root breadcrumb for root", func(t *testing.T) {
		breadcrumb := tree.GetBreadcrumb()
		assert.Equal(t, []string{"/"}, breadcrumb)
	})

	t.Run("returns single element for top-level folder", func(t *testing.T) {
		languages := tree.FindNodeByPath("languages")
		require.NotNil(t, languages)

		breadcrumb := languages.GetBreadcrumb()
		assert.Equal(t, []string{"/", "languages"}, breadcrumb)
	})

	t.Run("returns full path for nested folder", func(t *testing.T) {
		goFolder := tree.FindNodeByPath("languages/go")
		require.NotNil(t, goFolder)

		breadcrumb := goFolder.GetBreadcrumb()
		assert.Equal(t, []string{"/", "languages", "go"}, breadcrumb)
	})

	t.Run("returns full path for rule", func(t *testing.T) {
		testing := tree.FindNodeByPath("languages/go/testing")
		require.NotNil(t, testing)

		breadcrumb := testing.GetBreadcrumb()
		assert.Equal(t, []string{"/", "languages", "go", "testing"}, breadcrumb)
	})
}

func TestExtractRulePath(t *testing.T) {
	t.Parallel()
	t.Run("extracts path from contexture rule ID", func(t *testing.T) {
		ruleID := "[contexture:languages/go/testing]"
		path := ExtractRulePath(ruleID)
		assert.Equal(t, "languages/go/testing", path)
	})

	t.Run("extracts path from contexture rule ID with source", func(t *testing.T) {
		ruleID := "[contexture(github):languages/go/testing]"
		path := ExtractRulePath(ruleID)
		assert.Equal(t, "languages/go/testing", path)
	})

	t.Run("returns input for invalid format", func(t *testing.T) {
		ruleID := "invalid-format"
		path := ExtractRulePath(ruleID)
		assert.Equal(t, "invalid-format", path)
	})

	t.Run("returns empty string for empty input", func(t *testing.T) {
		path := ExtractRulePath("")
		assert.Empty(t, path)
	})

	t.Run("handles complex paths", func(t *testing.T) {
		ruleID := "[contexture:frameworks/react/advanced-patterns]"
		path := ExtractRulePath(ruleID)
		assert.Equal(t, "frameworks/react/advanced-patterns", path)
	})

	t.Run("extracts path from rule ID with variables", func(t *testing.T) {
		ruleID := "[contexture:languages/go/code-organization]{\"extended\": true}"
		path := ExtractRulePath(ruleID)
		assert.Equal(t, "languages/go/code-organization", path)
	})

	t.Run("extracts path from rule ID with branch and variables", func(t *testing.T) {
		ruleID := "[contexture:typescript/strict,v2.0.0]{\"target\": \"es2022\"}"
		path := ExtractRulePath(ruleID)
		assert.Equal(t, "typescript/strict", path)
	})

	t.Run("extracts path from rule ID with source and variables", func(t *testing.T) {
		ruleID := "[contexture(local):languages/go/testing]{\"strict\": false}"
		path := ExtractRulePath(ruleID)
		assert.Equal(t, "languages/go/testing", path)
	})
}

func TestExtractRuleDisplayPath(t *testing.T) {
	t.Parallel()
	t.Run("standard contexture rule returns just path", func(t *testing.T) {
		ruleID := "[contexture:languages/go/basics]"
		result := ExtractRuleDisplayPath(ruleID)
		assert.Equal(t, "languages/go/basics", result)
	})

	t.Run("custom SSH source includes source in path", func(t *testing.T) {
		ruleID := "[contexture(git@github.com:ryanskidmore/secretrules.git):test/lemon]"
		result := ExtractRuleDisplayPath(ruleID)
		assert.Equal(t, "git@github.com:ryanskidmore/secretrules/test/lemon", result)
	})

	t.Run("custom HTTPS source converts to friendly format", func(t *testing.T) {
		ruleID := "[contexture(https://github.com/user/repo.git):path/to/rule]"
		result := ExtractRuleDisplayPath(ruleID)
		assert.Equal(t, "github.com/user/repo/path/to/rule", result)
	})

	t.Run("custom source with branch shows branch in parentheses", func(t *testing.T) {
		ruleID := "[contexture(git@github.com:user/repo.git):test/rule,develop]"
		result := ExtractRuleDisplayPath(ruleID)
		assert.Equal(t, "git@github.com:user/repo/test/rule (develop)", result)
	})

	t.Run("custom source with variables strips variables", func(t *testing.T) {
		ruleID := "[contexture(git@example.com:org/rules.git):security/auth]{\"level\": \"strict\"}"
		result := ExtractRuleDisplayPath(ruleID)
		assert.Equal(t, "git@example.com:org/rules/security/auth", result)
	})

	t.Run("HTTPS without .git suffix", func(t *testing.T) {
		ruleID := "[contexture(https://gitlab.com/company/rules):policies/security]"
		result := ExtractRuleDisplayPath(ruleID)
		assert.Equal(t, "gitlab.com/company/rules/policies/security", result)
	})

	t.Run("default repo reference shows only path", func(t *testing.T) {
		ruleID := "[contexture(github):typescript/strict-config]"
		result := ExtractRuleDisplayPath(ruleID)
		assert.Equal(t, "typescript/strict-config", result)
	})

	t.Run("local repo reference shows only path", func(t *testing.T) {
		ruleID := "[contexture(local):user/custom-rule]"
		result := ExtractRuleDisplayPath(ruleID)
		assert.Equal(t, "user/custom-rule", result)
	})

	t.Run("empty rule ID returns empty", func(t *testing.T) {
		result := ExtractRuleDisplayPath("")
		assert.Empty(t, result)
	})

	t.Run("fallback for malformed custom source", func(t *testing.T) {
		ruleID := "[contexture(malformed:test/rule]"
		result := ExtractRuleDisplayPath(ruleID)
		// Should fallback to standard extraction which returns the malformed input
		assert.Equal(t, "[contexture(malformed:test/rule", result)
	})

	t.Run("custom source with non-default ref shows ref in parentheses", func(t *testing.T) {
		ruleID := "[contexture(git@github.com:user/repo.git):test/rule,develop]"
		result := ExtractRuleDisplayPath(ruleID)
		assert.Equal(t, "git@github.com:user/repo/test/rule (develop)", result)
	})

	t.Run("custom source with main ref does not show ref", func(t *testing.T) {
		ruleID := "[contexture(git@github.com:user/repo.git):test/rule,main]"
		result := ExtractRuleDisplayPath(ruleID)
		assert.Equal(t, "git@github.com:user/repo/test/rule", result)
	})

	t.Run("default repository does not show source even with custom ref", func(t *testing.T) {
		ruleID := "[contexture(https://github.com/contextureai/rules.git):some/rule,feature]"
		result := ExtractRuleDisplayPath(ruleID)
		assert.Equal(t, "some/rule", result) // Should not show source for default repo
	})
}

func TestFormatSourceForDisplay(t *testing.T) {
	t.Parallel()
	t.Run("SSH URL with develop ref shows ref in parentheses", func(t *testing.T) {
		result := FormatSourceForDisplay("git@github.com:user/repo.git", "develop")
		assert.Equal(t, "git@github.com:user/repo (develop)", result)
	})

	t.Run("SSH URL with main ref does not show ref", func(t *testing.T) {
		result := FormatSourceForDisplay("git@github.com:user/repo.git", "main")
		assert.Equal(t, "git@github.com:user/repo", result)
	})

	t.Run("HTTPS URL with feature ref shows ref in parentheses", func(t *testing.T) {
		result := FormatSourceForDisplay("https://github.com/user/repo.git", "feature")
		assert.Equal(t, "github.com/user/repo (feature)", result)
	})
}

func TestIsCustomGitSource(t *testing.T) {
	t.Parallel()
	t.Run("default repository returns false", func(t *testing.T) {
		result := IsCustomGitSource(DefaultRepository)
		assert.False(t, result)
	})

	t.Run("default repository without .git returns false", func(t *testing.T) {
		result := IsCustomGitSource("https://github.com/contextureai/rules")
		assert.False(t, result)
	})

	t.Run("custom SSH URL returns true", func(t *testing.T) {
		result := IsCustomGitSource("git@github.com:user/repo.git")
		assert.True(t, result)
	})

	t.Run("custom HTTPS URL returns true", func(t *testing.T) {
		result := IsCustomGitSource("https://github.com/user/repo.git")
		assert.True(t, result)
	})
}

func TestRuleTreeEdgeCases(t *testing.T) {
	t.Parallel()
	t.Run("handles duplicate rule paths", func(t *testing.T) {
		tree := NewRuleTree([]string{
			"languages/go/testing",
			"languages/go/testing", // duplicate
		})

		goFolder := tree.FindNodeByPath("languages/go")
		require.NotNil(t, goFolder)

		// Should only have one child
		assert.Len(t, goFolder.Children, 1)

		testing := goFolder.Children["testing"]
		require.NotNil(t, testing)
		assert.Equal(t, "testing", testing.Name)
	})

	t.Run("handles mixed case in paths", func(t *testing.T) {
		tree := NewRuleTree([]string{
			"Languages/Go/Testing",
		})

		// Should preserve case
		languages := tree.Children["Languages"]
		require.NotNil(t, languages)

		goFolder := languages.Children["Go"]
		require.NotNil(t, goFolder)

		testing := goFolder.Children["Testing"]
		require.NotNil(t, testing)
		assert.Equal(t, "Languages/Go/Testing", testing.Path)
	})

	t.Run("handles very deep nesting", func(t *testing.T) {
		tree := NewRuleTree([]string{
			"a/b/c/d/e/f/g/deep-rule",
		})

		deepRule := tree.FindNodeByPath("a/b/c/d/e/f/g/deep-rule")
		require.NotNil(t, deepRule)
		assert.Equal(t, "deep-rule", deepRule.Name)
		assert.Equal(t, RuleNodeTypeRule, deepRule.Type)

		breadcrumb := deepRule.GetBreadcrumb()
		expected := []string{"/", "a", "b", "c", "d", "e", "f", "g", "deep-rule"}
		assert.Equal(t, expected, breadcrumb)
	})
}
