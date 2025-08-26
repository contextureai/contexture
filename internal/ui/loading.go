package ui

import (
	"fmt"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// LoadingIndicator creates an animated loading indicator.
type LoadingIndicator struct {
	message   string
	frame     int
	startTime time.Time
	theme     Theme
}

// NewLoadingIndicator creates a new loading indicator.
func NewLoadingIndicator(message string) *LoadingIndicator {
	return &LoadingIndicator{
		message:   message,
		startTime: time.Now(),
		theme:     DefaultTheme(),
	}
}

// WithTheme sets a custom theme.
func (l *LoadingIndicator) WithTheme(theme Theme) *LoadingIndicator {
	l.theme = theme
	return l
}

// NextFrame advances to the next animation frame.
func (l *LoadingIndicator) NextFrame() {
	l.frame = (l.frame + 1) % len(SpinnerChars)
}

// Render renders the current frame of the loading indicator.
func (l *LoadingIndicator) Render() string {
	spinnerStyle := lipgloss.NewStyle().
		Foreground(l.theme.Info).
		Bold(true)

	messageStyle := lipgloss.NewStyle().
		Foreground(l.theme.Foreground)

	spinner := SpinnerChars[l.frame]
	elapsed := time.Since(l.startTime).Truncate(time.Second)

	message := l.message
	if elapsed > 0 {
		message += fmt.Sprintf(" (%s)", elapsed)
	}

	return spinnerStyle.Render(spinner) + " " + messageStyle.Render(message)
}
