// Package ui provides user interface components and styling for the Contexture CLI.
package ui

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDivider(t *testing.T) {
	t.Parallel()
	t.Run("plain_divider", func(t *testing.T) {
		divider := NewDivider()
		result := divider.Render()

		assert.Contains(t, result, "─")
		// Don't check exact length due to ANSI color codes
		assert.GreaterOrEqual(t, len(result), 60)
	})

	t.Run("divider_with_text", func(t *testing.T) {
		divider := NewDivider().WithText("Section")
		result := divider.Render()

		assert.Contains(t, result, "Section")
		assert.Contains(t, result, "─")
	})

	t.Run("divider_with_custom_width", func(t *testing.T) {
		divider := NewDivider().WithWidth(40)
		result := divider.Render()

		// Don't check exact length due to ANSI color codes
		assert.GreaterOrEqual(t, len(result), 40)
	})

	t.Run("dashed_divider", func(t *testing.T) {
		divider := NewDivider().WithStyle(DividerDashed)
		result := divider.Render()

		assert.Contains(t, result, "┄")
	})

	t.Run("dotted_divider", func(t *testing.T) {
		divider := NewDivider().WithStyle(DividerDotted)
		result := divider.Render()

		assert.Contains(t, result, "┈")
	})

	t.Run("double_divider", func(t *testing.T) {
		divider := NewDivider().WithStyle(DividerDouble)
		result := divider.Render()

		assert.Contains(t, result, "═")
	})

	t.Run("divider_text_too_long", func(t *testing.T) {
		divider := NewDivider().
			WithWidth(10).
			WithText("This is a very long text that exceeds width")
		result := divider.Render()

		assert.Contains(t, result, "This is a very long text that exceeds width")
	})
}

func TestDividerStyleValues(t *testing.T) {
	t.Parallel()
	// Test that divider style constants have expected values
	assert.Equal(t, 0, int(DividerPlain))
	assert.Equal(t, 1, int(DividerDashed))
	assert.Equal(t, 2, int(DividerDotted))
	assert.Equal(t, 3, int(DividerDouble))
}
