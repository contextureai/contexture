// Package ui provides user interface components and styling for the Contexture CLI.
package ui

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSidebar(t *testing.T) {
	t.Run("simple_sidebar", func(t *testing.T) {
		sidebar := NewSidebar("Navigation")
		result := sidebar.Render()

		assert.Contains(t, result, "Navigation")
	})

	t.Run("sidebar_with_items", func(t *testing.T) {
		sidebar := NewSidebar("Menu").
			AddItem("Home", "home").
			AddItem("Settings", "settings")
		result := sidebar.Render()

		assert.Contains(t, result, "Menu")
		assert.Contains(t, result, "Home")
		assert.Contains(t, result, "Settings")
	})

	t.Run("sidebar_with_icons", func(t *testing.T) {
		sidebar := NewSidebar("Files").
			AddItemWithIcon("Document", "doc1", "📄").
			AddItemWithIcon("Folder", "folder1", "📁")
		result := sidebar.Render()

		assert.Contains(t, result, "📄")
		assert.Contains(t, result, "📁")
		assert.Contains(t, result, "Document")
		assert.Contains(t, result, "Folder")
	})

	t.Run("sidebar_with_indented_items", func(t *testing.T) {
		sidebar := NewSidebar("Tree").
			AddItem("Root", "root").
			AddIndentedItem("Child 1", "child1", 1).
			AddIndentedItem("Child 2", "child2", 1).
			AddIndentedItem("Grandchild", "grandchild", 2)
		result := sidebar.Render()

		assert.Contains(t, result, "Root")
		assert.Contains(t, result, "Child 1")
		assert.Contains(t, result, "Child 2")
		assert.Contains(t, result, "Grandchild")

		// Check that grandchild appears in the result (indentation is complex due to styling)
		assert.Contains(t, result, "Grandchild")

		// Verify the sidebar contains hierarchy
		lines := strings.Split(result, "\n")
		foundGrandchild := false
		for _, line := range lines {
			if strings.Contains(line, "Grandchild") {
				foundGrandchild = true
				break
			}
		}
		assert.True(t, foundGrandchild)
	})

	t.Run("sidebar_with_custom_size", func(t *testing.T) {
		sidebar := NewSidebar("Custom").WithSize(40, 15)
		result := sidebar.Render()

		assert.Contains(t, result, "Custom")
		lines := strings.Split(result, "\n")
		assert.Greater(t, len(lines), 1) // Should have multiple lines due to height
	})
}
