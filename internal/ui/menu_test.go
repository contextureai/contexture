// Package ui provides user interface components and styling for the Contexture CLI.
package ui

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMenu(t *testing.T) {
	t.Parallel()
	t.Run("simple_menu", func(t *testing.T) {
		menu := NewMenu("Main Menu")
		result := menu.Render()

		assert.Contains(t, result, "Main Menu")
	})

	t.Run("menu_with_items", func(t *testing.T) {
		menu := NewMenu("Options").
			AddItem("Option 1", "First option").
			AddItem("Option 2", "Second option")
		result := menu.Render()

		assert.Contains(t, result, "Options")
		assert.Contains(t, result, "Option 1")
		assert.Contains(t, result, "Option 2")
		assert.Contains(t, result, "First option")
		assert.Contains(t, result, "Second option")
	})

	t.Run("menu_with_shortcuts", func(t *testing.T) {
		menu := NewMenu("Commands").
			AddItemWithShortcut("Initialize", "Initialize project", "i").
			AddItemWithShortcut("Generate", "Generate output", "g")
		result := menu.Render()

		assert.Contains(t, result, "[i]")
		assert.Contains(t, result, "[g]")
		assert.Contains(t, result, "Initialize")
		assert.Contains(t, result, "Generate")
	})

	t.Run("compact_menu", func(t *testing.T) {
		menu := NewMenu("Compact").
			AddItem("Item 1", "Description 1").
			AddItem("Item 2", "Description 2").
			WithCompact(true)
		result := menu.Render()

		assert.Contains(t, result, "Item 1")
		assert.Contains(t, result, "Item 2")
		// In compact mode, descriptions might not be shown or formatted differently
	})

	t.Run("menu_with_custom_width", func(t *testing.T) {
		menu := NewMenu("Wide Menu").WithWidth(80)
		result := menu.Render()

		assert.Contains(t, result, "Wide Menu")
	})
}
