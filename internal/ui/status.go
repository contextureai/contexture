package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// StatusType represents the type of status.
type StatusType int

const (
	// StatusSuccess represents a successful status.
	StatusSuccess StatusType = iota
	// StatusWarning represents a warning status.
	StatusWarning
	// StatusError represents an error status.
	StatusError
	// StatusInfo represents an informational status.
	StatusInfo
	// StatusLoading represents a loading status.
	StatusLoading
)

// StatusIndicator represents different status types.
type StatusIndicator struct {
	status  StatusType
	message string
	details []string
	theme   Theme
}

// NewStatusIndicator creates a new status indicator.
func NewStatusIndicator(status StatusType, message string) *StatusIndicator {
	return &StatusIndicator{
		status:  status,
		message: message,
		details: make([]string, 0),
		theme:   DefaultTheme(),
	}
}

// WithDetails adds details to the status indicator.
func (s *StatusIndicator) WithDetails(details ...string) *StatusIndicator {
	// Create a new slice to avoid mutation issues
	newDetails := make([]string, len(s.details)+len(details))
	copy(newDetails, s.details)
	copy(newDetails[len(s.details):], details)
	s.details = newDetails
	return s
}

// WithTheme sets a custom theme.
func (s *StatusIndicator) WithTheme(theme Theme) *StatusIndicator {
	s.theme = theme
	return s
}

// Render renders the status indicator as a string.
func (s *StatusIndicator) Render() string {
	var icon string
	var color lipgloss.AdaptiveColor

	switch s.status {
	case StatusSuccess:
		icon = "✓"
		color = s.theme.Success
	case StatusWarning:
		icon = "⚠"
		color = s.theme.Warning
	case StatusError:
		icon = "✗"
		color = s.theme.Error
	case StatusInfo:
		icon = "ⓘ"
		color = s.theme.Info
	case StatusLoading:
		icon = SpinnerChars[0] // Use first spinner char for static rendering
		color = s.theme.Info
	default:
		icon = "•"
		color = s.theme.Foreground
	}

	mainStyle := lipgloss.NewStyle().
		Foreground(color).
		Bold(true)

	var sb strings.Builder
	sb.WriteString(mainStyle.Render(icon + " " + s.message))

	if len(s.details) > 0 {
		detailStyle := lipgloss.NewStyle().
			Foreground(s.theme.Muted).
			MarginLeft(4)

		for _, detail := range s.details {
			sb.WriteString("\n")
			sb.WriteString(detailStyle.Render("• " + detail))
		}
	}

	return sb.String()
}
