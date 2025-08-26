// Package ui provides user interface components and styling for the Contexture CLI.
package ui

import (
	"github.com/charmbracelet/lipgloss"
)

// NotificationType represents notification types.
type NotificationType int

const (
	// NotificationSuccess represents a success notification.
	NotificationSuccess NotificationType = iota
	// NotificationWarning represents a warning notification.
	NotificationWarning
	// NotificationError represents an error notification.
	NotificationError
	// NotificationInfo represents an informational notification.
	NotificationInfo
)

// Notification creates a styled notification.
type Notification struct {
	notifType NotificationType
	title     string
	message   string
	width     int
	theme     Theme
}

// NewNotification creates a new notification.
func NewNotification(notifType NotificationType, title, message string) *Notification {
	return &Notification{
		notifType: notifType,
		title:     title,
		message:   message,
		width:     60,
		theme:     DefaultTheme(),
	}
}

// WithWidth sets the notification width.
func (n *Notification) WithWidth(width int) *Notification {
	if width > 0 {
		n.width = width
	}
	return n
}

// WithTheme sets a custom theme.
func (n *Notification) WithTheme(theme Theme) *Notification {
	n.theme = theme
	return n
}

// Render renders the notification as a string.
func (n *Notification) Render() string {
	var icon string
	var borderColor lipgloss.AdaptiveColor
	var titleColor lipgloss.AdaptiveColor

	switch n.notifType {
	case NotificationSuccess:
		icon = "✓"
		borderColor = n.theme.Success
		titleColor = n.theme.Success
	case NotificationWarning:
		icon = "⚠"
		borderColor = n.theme.Warning
		titleColor = n.theme.Warning
	case NotificationError:
		icon = "✗"
		borderColor = n.theme.Error
		titleColor = n.theme.Error
	case NotificationInfo:
		icon = "ⓘ"
		borderColor = n.theme.Info
		titleColor = n.theme.Info
	}

	titleStyle := lipgloss.NewStyle().
		Foreground(titleColor).
		Bold(true)

	messageStyle := lipgloss.NewStyle().
		Foreground(n.theme.Foreground).
		MarginTop(1)

	notificationStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Padding(1, 2).
		Width(n.width)

	content := titleStyle.Render(icon + " " + n.title)
	if n.message != "" {
		content += "\n" + messageStyle.Render(n.message)
	}

	return notificationStyle.Render(content)
}
