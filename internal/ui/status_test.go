// Package ui provides user interface components and styling for the Contexture CLI.
package ui

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStatusIndicator(t *testing.T) {
	t.Parallel()
	t.Run("success_status", func(t *testing.T) {
		status := NewStatusIndicator(StatusSuccess, "Operation completed")
		result := status.Render()

		assert.Contains(t, result, "✓")
		assert.Contains(t, result, "Operation completed")
	})

	t.Run("warning_status", func(t *testing.T) {
		status := NewStatusIndicator(StatusWarning, "Warning message")
		result := status.Render()

		assert.Contains(t, result, "⚠")
		assert.Contains(t, result, "Warning message")
	})

	t.Run("error_status", func(t *testing.T) {
		status := NewStatusIndicator(StatusError, "Error occurred")
		result := status.Render()

		assert.Contains(t, result, "✗")
		assert.Contains(t, result, "Error occurred")
	})

	t.Run("info_status", func(t *testing.T) {
		status := NewStatusIndicator(StatusInfo, "Information")
		result := status.Render()

		assert.Contains(t, result, "ⓘ")
		assert.Contains(t, result, "Information")
	})

	t.Run("loading_status", func(t *testing.T) {
		status := NewStatusIndicator(StatusLoading, "Processing...")
		result := status.Render()

		assert.Contains(t, result, "Processing...")
		// Should contain a spinner character
		found := false
		for _, char := range SpinnerChars {
			if strings.Contains(result, char) {
				found = true
				break
			}
		}
		assert.True(t, found, "Should contain a spinner character")
	})

	t.Run("status_with_details", func(t *testing.T) {
		status := NewStatusIndicator(StatusSuccess, "Main message").
			WithDetails("Detail 1", "Detail 2", "Detail 3")
		result := status.Render()

		assert.Contains(t, result, "Main message")
		assert.Contains(t, result, "Detail 1")
		assert.Contains(t, result, "Detail 2")
		assert.Contains(t, result, "Detail 3")

		// Details should be indented
		lines := strings.Split(result, "\n")
		assert.Greater(t, len(lines), 1)
	})
}

func TestStatusTypeValues(t *testing.T) {
	t.Parallel()
	// Test that status type constants have expected values
	assert.Equal(t, 0, int(StatusSuccess))
	assert.Equal(t, 1, int(StatusWarning))
	assert.Equal(t, 2, int(StatusError))
	assert.Equal(t, 3, int(StatusInfo))
	assert.Equal(t, 4, int(StatusLoading))
}
