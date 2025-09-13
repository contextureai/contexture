// Package ui provides user interface components and styling for the Contexture CLI.
package ui

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNotification(t *testing.T) {
	t.Parallel()
	t.Run("success_notification", func(t *testing.T) {
		notification := NewNotification(
			NotificationSuccess,
			"Success!",
			"Operation completed successfully",
		)
		result := notification.Render()

		assert.Contains(t, result, "✓")
		assert.Contains(t, result, "Success!")
		assert.Contains(t, result, "Operation completed successfully")
	})

	t.Run("warning_notification", func(t *testing.T) {
		notification := NewNotification(
			NotificationWarning,
			"Warning",
			"Please check your configuration",
		)
		result := notification.Render()

		assert.Contains(t, result, "⚠")
		assert.Contains(t, result, "Warning")
		assert.Contains(t, result, "Please check your configuration")
	})

	t.Run("error_notification", func(t *testing.T) {
		notification := NewNotification(NotificationError, "Error", "Something went wrong")
		result := notification.Render()

		assert.Contains(t, result, "✗")
		assert.Contains(t, result, "Error")
		assert.Contains(t, result, "Something went wrong")
	})

	t.Run("info_notification", func(t *testing.T) {
		notification := NewNotification(NotificationInfo, "Info", "Here's some information")
		result := notification.Render()

		assert.Contains(t, result, "ⓘ")
		assert.Contains(t, result, "Info")
		assert.Contains(t, result, "Here's some information")
	})

	t.Run("notification_without_message", func(t *testing.T) {
		notification := NewNotification(NotificationInfo, "Title Only", "")
		result := notification.Render()

		assert.Contains(t, result, "Title Only")
	})

	t.Run("notification_with_custom_width", func(t *testing.T) {
		notification := NewNotification(NotificationInfo, "Title", "Message").WithWidth(80)
		result := notification.Render()

		assert.Contains(t, result, "Title")
		assert.Contains(t, result, "Message")
	})
}

func TestNotificationTypeValues(t *testing.T) {
	t.Parallel()
	// Test that notification type constants have expected values
	assert.Equal(t, 0, int(NotificationSuccess))
	assert.Equal(t, 1, int(NotificationWarning))
	assert.Equal(t, 2, int(NotificationError))
	assert.Equal(t, 3, int(NotificationInfo))
}
