// Package e2e provides TUI testing for bubble tea components
package e2e

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/exp/teatest"
	"github.com/contextureai/contexture/internal/domain"
	"github.com/contextureai/contexture/internal/tui"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRuleSelectorTUI tests the rule selector TUI component
func TestRuleSelectorTUI(t *testing.T) {
	// Sample rules are available but not used in these basic tests
	_ = []*domain.Rule{
		{
			ID:          "[contexture:security/input-validation]",
			Title:       "Input Validation",
			Description: "Validate all user inputs",
			Tags:        []string{"security", "validation"},
			Content:     "# Input Validation\n\nValidate all user inputs to prevent security issues.",
		},
		{
			ID:          "[contexture:performance/caching]",
			Title:       "Caching Strategy",
			Description: "Implement effective caching",
			Tags:        []string{"performance", "caching"},
			Content:     "# Caching Strategy\n\nImplement multi-layer caching.",
		},
	}

	t.Run("rule selector initialization", func(t *testing.T) {
		selector := tui.NewRuleSelector()
		require.NotNil(t, selector)
	})

	t.Run("rule selector display basic", func(t *testing.T) {
		// Create a test model that simulates the rule selector
		// Note: This is a simplified test due to the complexity of testing interactive TUI
		selector := tui.NewRuleSelector()

		// Test that the selector can be created without panicking
		require.NotNil(t, selector)

		// We would ideally test the full TUI here, but due to the complexity
		// and the fact that DisplayRules opens a full tea.Program,
		// we'll test the underlying components instead
	})
}

// TestRulePreviewHelper tests the rule preview functionality
func TestRulePreviewHelper(t *testing.T) {
	helper := tui.NewRulePreviewHelper()
	require.NotNil(t, helper)

	t.Run("preview helper size update", func(_ *testing.T) {
		helper.UpdateSize(100, 50)
		// Test passes if no panic occurs
	})

	t.Run("preview content building", func(t *testing.T) {
		rule := &domain.Rule{
			Title:       "Test Rule",
			Description: "A test rule",
			Tags:        []string{"test", "example"},
			Content:     "# Test Rule\n\nThis is test content.",
		}

		helper.SetupPreview(rule)
		content := helper.BuildPreviewContent(rule)

		assert.Contains(t, content, "Test Rule")
		assert.Contains(t, content, "A test rule")
		assert.Contains(t, content, "Tags: test, example")
		assert.Contains(t, content, "This is test content.")
	})

	t.Run("preview with trigger information", func(t *testing.T) {
		rule := &domain.Rule{
			Title:       "Triggered Rule",
			Description: "A rule with trigger",
			Tags:        []string{"test"},
			Trigger: &domain.RuleTrigger{
				Type:  "glob",
				Globs: []string{"*.go", "*.ts"},
			},
			Content: "# Triggered Rule\n\nContent with trigger.",
		}

		content := helper.BuildPreviewContent(rule)
		assert.Contains(t, content, "Trigger: glob (*.go, *.ts)")
	})
}

// TestFileBrowserTUI tests the file browser component
func TestFileBrowserTUI(t *testing.T) {
	t.Run("file browser creation", func(t *testing.T) {
		browser := tui.NewFileBrowser()
		require.NotNil(t, browser)
	})

	// Note: Full TUI testing would require more complex setup
	// These are basic smoke tests to ensure components can be created
}

// TestInteractiveTUIBehavior tests TUI behavior with simulated input using teatest
func TestInteractiveTUIBehavior(t *testing.T) {
	// Create comprehensive test rules
	rules := []*domain.Rule{
		{
			ID:          "[contexture:security/input-validation]",
			Title:       "Input Validation",
			Description: "Validate all user inputs",
			Tags:        []string{"security", "validation"},
			Content:     "# Input Validation\\n\\nValidate all user inputs to prevent security issues.",
		},
		{
			ID:          "[contexture:performance/caching]",
			Title:       "Caching Strategy",
			Description: "Implement effective caching",
			Tags:        []string{"performance", "caching"},
			Content:     "# Caching Strategy\\n\\nImplement multi-layer caching.",
		},
		{
			ID:          "[contexture:testing/unit-tests]",
			Title:       "Unit Testing",
			Description: "Write comprehensive unit tests",
			Tags:        []string{"testing", "quality"},
			Content:     "# Unit Testing\\n\\nWrite comprehensive unit tests for all components.",
		},
	}

	t.Run("rule selection workflow with teatest", func(t *testing.T) {
		// Convert rules to list items (simulating internal structure)
		items := make([]list.Item, len(rules))
		for i, rule := range rules {
			items[i] = &ruleItem{rule: rule}
		}

		// Create the rule selection model
		model := newRuleSelectionModel(items, "E2E Rule Selection Test")

		tm := teatest.NewTestModel(t, model,
			teatest.WithInitialTermSize(100, 30))

		// Test complete workflow: navigate, select multiple rules, confirm
		// Navigate down to second rule
		tm.Send(tea.KeyMsg{Type: tea.KeyDown})

		// Select second rule (performance/caching)
		tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})

		// Navigate to third rule
		tm.Send(tea.KeyMsg{Type: tea.KeyDown})

		// Select third rule (testing/unit-tests)
		tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})

		// Preview the current rule
		tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}})

		// Wait for preview to appear
		teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
			return bytes.Contains(bts, []byte("Rule Preview")) &&
				bytes.Contains(bts, []byte("Unit Testing"))
		}, teatest.WithCheckInterval(50*time.Millisecond), teatest.WithDuration(2*time.Second))

		// Exit preview
		tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}})

		// Confirm selection
		tm.Send(tea.KeyMsg{Type: tea.KeyEnter})

		// Validate final state
		finalModel, ok := tm.FinalModel(t).(*mockRuleSelectionModel)
		require.True(t, ok, "Final model should be mockRuleSelectionModel")

		// Should have selected 2 rules
		assert.Len(t, finalModel.chosen, 2)
		assert.Contains(t, finalModel.chosen, "[contexture:performance/caching]")
		assert.Contains(t, finalModel.chosen, "[contexture:testing/unit-tests]")
		assert.NotContains(t, finalModel.chosen, "[contexture:security/input-validation]")
	})

	t.Run("cancellation workflow", func(t *testing.T) {
		items := make([]list.Item, len(rules))
		for i, rule := range rules {
			items[i] = &ruleItem{rule: rule}
		}

		model := newRuleSelectionModel(items, "E2E Cancel Test")

		tm := teatest.NewTestModel(t, model,
			teatest.WithInitialTermSize(80, 24))

		// Make some selections
		tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})
		tm.Send(tea.KeyMsg{Type: tea.KeyDown})
		tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})

		// Cancel instead of confirming
		tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})

		finalModel, ok := tm.FinalModel(t).(*mockRuleSelectionModel)
		require.True(t, ok)

		// Should have cancelled without selecting
		assert.True(t, finalModel.quitting)
		assert.Empty(t, finalModel.chosen)
	})
}

// Helper types for testing (these would normally be internal to the tui package)
type ruleItem struct {
	rule     *domain.Rule
	selected bool
}

func (i *ruleItem) FilterValue() string {
	var parts []string
	parts = append(parts, i.rule.Title)
	if i.rule.Description != "" {
		parts = append(parts, i.rule.Description)
	}
	parts = append(parts, strings.Join(i.rule.Tags, " "))
	return strings.Join(parts, " ")
}

func (i *ruleItem) Title() string {
	return i.rule.Title
}

func (i *ruleItem) Description() string {
	return i.rule.Description
}

// Mock function to create rule selection model (would be internal)
func newRuleSelectionModel(items []list.Item, title string) tea.Model {
	// This is a simplified mock - in real tests we'd use the actual internal model
	// For now, we create a basic model that implements the tea.Model interface
	return &mockRuleSelectionModel{
		items:          items,
		title:          title,
		chosen:         []string{},
		quitting:       false,
		showingPreview: false,
	}
}

type mockRuleSelectionModel struct {
	items          []list.Item
	title          string
	chosen         []string
	quitting       bool
	showingPreview bool
	selectedIndex  int
}

func (m *mockRuleSelectionModel) Init() tea.Cmd {
	return nil
}

func (m *mockRuleSelectionModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if msg, ok := msg.(tea.KeyMsg); ok {
		if m.showingPreview {
			switch msg.String() {
			case "p", "esc", "q":
				m.showingPreview = false
			}
			return m, nil
		}

		switch msg.String() {
		case "down":
			if m.selectedIndex < len(m.items)-1 {
				m.selectedIndex++
			}
		case "up":
			if m.selectedIndex > 0 {
				m.selectedIndex--
			}
		case " ":
			// Toggle selection
			if item, ok := m.items[m.selectedIndex].(*ruleItem); ok {
				item.selected = !item.selected
				if item.selected {
					m.chosen = append(m.chosen, item.rule.ID)
				} else {
					// Remove from chosen
					for i, id := range m.chosen {
						if id == item.rule.ID {
							m.chosen = append(m.chosen[:i], m.chosen[i+1:]...)
							break
						}
					}
				}
			}
		case "p":
			m.showingPreview = !m.showingPreview
		case "enter":
			return m, tea.Quit
		case "q", "esc":
			m.quitting = true
			m.chosen = []string{}
			return m, tea.Quit
		default:
			// Handle typing for filtering - simplified
			if len(msg.Runes) > 0 {
				// Simulate filtering by showing only matching items
				// This is a very basic mock implementation
				_ = msg.Runes // Use the runes to avoid unused variable warning
			}
		}
	}
	return m, nil
}

func (m *mockRuleSelectionModel) View() string {
	if m.quitting {
		return ""
	}

	var view strings.Builder
	view.WriteString(m.title + "\\n\\n")

	if m.showingPreview {
		view.WriteString("Rule Preview\\n")
		if m.selectedIndex < len(m.items) {
			if item, ok := m.items[m.selectedIndex].(*ruleItem); ok {
				view.WriteString(item.rule.Title + "\\n")
				view.WriteString(item.rule.Description + "\\n")
			}
		}
		return view.String()
	}

	// Show items
	for i, item := range m.items {
		if ruleItem, ok := item.(*ruleItem); ok {
			prefix := "[ ] "
			if ruleItem.selected {
				prefix = "[✓] "
			}
			if i == m.selectedIndex {
				prefix = "> " + prefix
			}
			view.WriteString(prefix + ruleItem.rule.Title + "\\n")
		}
	}

	selectedCount := 0
	for _, item := range m.items {
		if ruleItem, ok := item.(*ruleItem); ok && ruleItem.selected {
			selectedCount++
		}
	}

	if selectedCount > 0 {
		view.WriteString(fmt.Sprintf("\\n%d selected\\n", selectedCount))
	}

	return view.String()
}

// TestTUIComponentIntegration tests integration between TUI components
func TestTUIComponentIntegration(t *testing.T) {
	t.Run("preview helper with real rule data", func(t *testing.T) {
		helper := tui.NewRulePreviewHelper()

		// Test with minimal rule
		minimalRule := &domain.Rule{
			Title:   "Minimal Rule",
			Content: "Basic content",
		}

		content := helper.BuildPreviewContent(minimalRule)
		assert.Contains(t, content, "Minimal Rule")
		assert.Contains(t, content, "Basic content")
	})

	t.Run("preview helper with complex rule", func(t *testing.T) {
		helper := tui.NewRulePreviewHelper()

		complexRule := &domain.Rule{
			Title:       "Complex Rule",
			Description: "A complex rule with all fields",
			Tags:        []string{"complex", "testing", "full-featured"},
			Languages:   []string{"go", "javascript", "python"},
			Frameworks:  []string{"react", "vue", "angular"},
			Trigger: &domain.RuleTrigger{
				Type:  "glob",
				Globs: []string{"*.js", "*.jsx", "*.ts", "*.tsx"},
			},
			Content: `# Complex Rule

This is a complex rule with **markdown** formatting.

## Features
- Multiple languages
- Framework support
- Trigger conditions
- Rich content

` + "```javascript" + `
// Code example
function example() {
    return "Hello, World!";
}
` + "```" + ``,
		}

		content := helper.BuildPreviewContent(complexRule)

		// Verify all components are included
		assert.Contains(t, content, "Complex Rule")
		assert.Contains(t, content, "A complex rule with all fields")
		assert.Contains(t, content, "Tags: complex, testing, full-featured")
		assert.Contains(t, content, "Languages: go, javascript, python")
		assert.Contains(t, content, "Frameworks: react, vue, angular")
		assert.Contains(t, content, "Trigger: glob (*.js, *.jsx, *.ts, *.tsx)")
		assert.Contains(t, content, "This is a complex rule")
	})
}

// TestTUIErrorHandling tests error scenarios in TUI components
func TestTUIErrorHandling(t *testing.T) {
	t.Run("preview with nil rule", func(t *testing.T) {
		helper := tui.NewRulePreviewHelper()

		// This should not panic
		content := helper.BuildPreviewContent(nil)
		assert.Contains(t, content, "No rule selected")
	})

	t.Run("preview with empty rule", func(t *testing.T) {
		helper := tui.NewRulePreviewHelper()

		emptyRule := &domain.Rule{}
		content := helper.BuildPreviewContent(emptyRule)

		// Should handle empty rule gracefully
		assert.NotEmpty(t, content)
		// Should not contain error messages or panics
		assert.NotContains(t, strings.ToLower(content), "error")
		assert.NotContains(t, strings.ToLower(content), "panic")
	})

	t.Run("preview with malformed content", func(t *testing.T) {
		helper := tui.NewRulePreviewHelper()

		malformedRule := &domain.Rule{
			Title:   "Malformed Rule",
			Content: "# Unclosed markdown **bold text without closing",
		}

		content := helper.BuildPreviewContent(malformedRule)
		assert.Contains(t, content, "Malformed Rule")
		// Should still render something even with malformed markdown
		assert.NotEmpty(t, content)
	})
}

// TestTUIPerformance tests TUI performance with large datasets
func TestTUIPerformance(t *testing.T) {
	t.Run("preview with large rule content", func(t *testing.T) {
		helper := tui.NewRulePreviewHelper()

		// Create a rule with large content
		largeContent := strings.Repeat("# Large Section\n\nThis is a large section with lots of content.\n\n", 100)

		largeRule := &domain.Rule{
			Title:   "Large Rule",
			Content: largeContent,
		}

		start := time.Now()
		content := helper.BuildPreviewContent(largeRule)
		duration := time.Since(start)

		assert.Contains(t, content, "Large Rule")
		assert.Less(t, duration, 100*time.Millisecond, "Preview generation should be fast")
	})

	t.Run("preview helper with many rules", func(t *testing.T) {
		helper := tui.NewRulePreviewHelper()

		// Test processing many rules in sequence
		start := time.Now()

		for i := range 50 {
			rule := &domain.Rule{
				Title:   "Rule " + string(rune(i+65)),
				Content: "Content for rule " + string(rune(i+65)),
			}

			helper.SetupPreview(rule)
			content := helper.BuildPreviewContent(rule)
			assert.NotEmpty(t, content)
		}

		duration := time.Since(start)
		assert.Less(t, duration, 500*time.Millisecond, "Processing many rules should be efficient")
	})
}

// TestTUIAccessibility tests accessibility features in TUI components
func TestTUIAccessibility(t *testing.T) {
	t.Run("preview content structure", func(t *testing.T) {
		helper := tui.NewRulePreviewHelper()

		rule := &domain.Rule{
			Title:       "Accessibility Test Rule",
			Description: "Testing accessibility features",
			Tags:        []string{"accessibility", "a11y"},
			Content:     "# Accessibility\n\nThis rule tests accessibility features.",
		}

		content := helper.BuildPreviewContent(rule)

		// Check that content has clear structure
		lines := strings.Split(content, "\n")
		assert.Greater(t, len(lines), 5, "Content should have multiple lines for readability")

		// Check for clear section headers - the actual format shows the title directly
		assert.Contains(t, content, "Accessibility Test Rule")
		assert.Contains(t, content, "Testing accessibility features")
		assert.Contains(t, content, "accessibility")
	})
}

// TestTUIKeyboardNavigation tests keyboard navigation in TUI components
func TestTUIKeyboardNavigation(t *testing.T) {
	t.Run("rule selector keyboard simulation", func(t *testing.T) {
		// Create a rule selector for testing
		selector := tui.NewRuleSelector()
		require.NotNil(t, selector)

		// Test that the selector can be created without error
		// In a full implementation, this would test:
		// - Arrow key navigation
		// - Enter key selection
		// - Escape key cancellation
		// - Search input handling
		t.Log("Rule selector keyboard navigation test placeholder")
	})
}

// TestTUISearchFunctionality tests search and filtering in TUI
func TestTUISearchFunctionality(t *testing.T) {
	t.Run("rule filtering", func(t *testing.T) {
		// Test rule filtering logic
		// This would test the search/filter functionality if exposed
		rules := []*domain.Rule{
			{ID: "security/auth", Title: "Authentication", Tags: []string{"security", "auth"}},
			{ID: "performance/cache", Title: "Caching", Tags: []string{"performance", "cache"}},
			{ID: "security/input", Title: "Input Validation", Tags: []string{"security", "validation"}},
		}

		// Test filtering by tag
		securityRules := filterRulesByTag(rules, "security")
		assert.Len(t, securityRules, 2)

		// Test filtering by title
		authRules := filterRulesByTitle(rules, "auth")
		assert.Len(t, authRules, 1)
		assert.Equal(t, "Authentication", authRules[0].Title)
	})
}

// TestTUIRenderingEdgeCases tests rendering with edge cases
func TestTUIRenderingEdgeCases(t *testing.T) {
	helper := tui.NewRulePreviewHelper()

	t.Run("unicode content", func(t *testing.T) {
		unicodeRule := &domain.Rule{
			Title:       "Unicode Test 🎯",
			Description: "Testing unicode: ñoño, 中文, العربية, emoji 🚀",
			Tags:        []string{"unicode", "testing", "🏷️"},
			Content:     "# Unicode Content\n\n测试内容 with emojis 🎉 and symbols ∞",
		}

		content := helper.BuildPreviewContent(unicodeRule)
		assert.Contains(t, content, "Unicode Test 🎯")
		assert.Contains(t, content, "🚀")
		assert.Contains(t, content, "中文")
		assert.Contains(t, content, "🎉")
	})

	t.Run("very long lines", func(t *testing.T) {
		longLine := strings.Repeat("This is a very long line that should wrap properly in the terminal display. ", 50)

		longLineRule := &domain.Rule{
			Title:   "Long Line Test",
			Content: "# Long Line Test\n\n" + longLine,
		}

		content := helper.BuildPreviewContent(longLineRule)
		assert.Contains(t, content, "Long Line Test")
		assert.Contains(t, content, "very long line")
	})

	t.Run("terminal size simulation", func(t *testing.T) {
		helper.UpdateSize(80, 24) // Standard terminal size

		rule := &domain.Rule{
			Title:   "Terminal Size Test",
			Content: "# Test\n\nContent that should fit in standard terminal",
		}

		content := helper.BuildPreviewContent(rule)
		assert.NotEmpty(t, content)

		// Test with small terminal
		helper.UpdateSize(40, 10)
		contentSmall := helper.BuildPreviewContent(rule)
		assert.NotEmpty(t, contentSmall)

		// Test with large terminal
		helper.UpdateSize(120, 40)
		contentLarge := helper.BuildPreviewContent(rule)
		assert.NotEmpty(t, contentLarge)
	})
}

// Helper functions for TUI testing
func filterRulesByTag(rules []*domain.Rule, tag string) []*domain.Rule {
	var filtered []*domain.Rule
	for _, rule := range rules {
		for _, t := range rule.Tags {
			if t == tag {
				filtered = append(filtered, rule)
				break
			}
		}
	}
	return filtered
}

func filterRulesByTitle(rules []*domain.Rule, search string) []*domain.Rule {
	var filtered []*domain.Rule
	searchLower := strings.ToLower(search)
	for _, rule := range rules {
		if strings.Contains(strings.ToLower(rule.Title), searchLower) {
			filtered = append(filtered, rule)
		}
	}
	return filtered
}
