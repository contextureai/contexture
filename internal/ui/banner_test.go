// Package ui provides user interface components and styling for the Contexture CLI.
package ui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/stretchr/testify/assert"
)

func TestBanner(t *testing.T) {
	t.Run("simple_banner", func(t *testing.T) {
		banner := NewBanner("Test Banner")
		result := banner.Render()

		assert.Contains(t, result, "Test Banner")
		// Should contain border characters
		assert.Greater(t, len(result), len("Test Banner"))
	})

	t.Run("banner_with_subtitle", func(t *testing.T) {
		banner := NewBanner("Main Title").WithSubtitle("Subtitle Text")
		result := banner.Render()

		assert.Contains(t, result, "Main Title")
		assert.Contains(t, result, "Subtitle Text")
	})

	t.Run("banner_with_custom_width", func(t *testing.T) {
		banner := NewBanner("Test").WithWidth(80)
		result := banner.Render()

		assert.Contains(t, result, "Test")
		// Width affects the output size
		lines := strings.Split(result, "\n")
		assert.Greater(t, len(lines), 1) // Should have multiple lines due to border
	})

	t.Run("banner_with_custom_border", func(t *testing.T) {
		banner := NewBanner("Test").WithBorder(lipgloss.NormalBorder())
		result := banner.Render()

		assert.Contains(t, result, "Test")
	})
}

// Benchmark tests
func BenchmarkBannerRender(b *testing.B) {
	banner := NewBanner("Benchmark Test").WithSubtitle("Performance Testing")
	b.ResetTimer()
	for range b.N {
		banner.Render()
	}
}
