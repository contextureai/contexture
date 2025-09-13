// Package ui provides user interface components and styling for the Contexture CLI.
package ui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/stretchr/testify/assert"
)

func TestCard(t *testing.T) {
	t.Parallel()
	t.Run("simple_card", func(t *testing.T) {
		card := NewCard("Card Title")
		result := card.Render()

		assert.Contains(t, result, "Card Title")
	})

	t.Run("card_with_content", func(t *testing.T) {
		card := NewCard("Title").WithContent("This is the card content")
		result := card.Render()

		assert.Contains(t, result, "Title")
		assert.Contains(t, result, "This is the card content")
	})

	t.Run("card_with_custom_size", func(t *testing.T) {
		card := NewCard("Title").WithSize(60, 10)
		result := card.Render()

		assert.Contains(t, result, "Title")
		lines := strings.Split(result, "\n")
		assert.GreaterOrEqual(t, len(lines), 3) // At least title + borders
	})

	t.Run("card_with_custom_border", func(t *testing.T) {
		card := NewCard("Title").WithBorder(lipgloss.DoubleBorder())
		result := card.Render()

		assert.Contains(t, result, "Title")
	})
}
