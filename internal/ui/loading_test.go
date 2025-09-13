// Package ui provides user interface components and styling for the Contexture CLI.
package ui

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLoadingIndicator(t *testing.T) {
	t.Parallel()
	t.Run("initial_loading", func(t *testing.T) {
		loading := NewLoadingIndicator("Loading data...")
		result := loading.Render()

		assert.Contains(t, result, "Loading data...")
		// Should contain the first spinner character
		assert.Contains(t, result, SpinnerChars[0])
	})

	t.Run("next_frame", func(t *testing.T) {
		loading := NewLoadingIndicator("Processing...")

		// Initial frame
		result1 := loading.Render()
		assert.Contains(t, result1, SpinnerChars[0])

		// Next frame
		loading.NextFrame()
		result2 := loading.Render()
		assert.Contains(t, result2, SpinnerChars[1])

		// Frames should be different
		assert.NotEqual(t, result1, result2)
	})

	t.Run("frame_cycling", func(t *testing.T) {
		loading := NewLoadingIndicator("Testing...")

		// Advance through all frames and one more
		for i := 0; i <= len(SpinnerChars); i++ {
			loading.NextFrame()
		}

		// Should cycle back to first frame
		result := loading.Render()
		assert.Contains(t, result, SpinnerChars[1]) // Frame should have cycled
	})

	t.Run("elapsed_time", func(t *testing.T) {
		// Create a loading indicator with a past start time
		loading := NewLoadingIndicator("Waiting...")
		loading.startTime = time.Now().Add(-2 * time.Second)

		result := loading.Render()
		assert.Contains(t, result, "Waiting...")
		assert.Contains(t, result, "2s") // Should show elapsed time
	})
}
