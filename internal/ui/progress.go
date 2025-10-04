// Package ui provides user interface components and styling for the Contexture CLI.
package ui

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/lipgloss"
	contextureerrors "github.com/contextureai/contexture/internal/errors"
	"golang.org/x/term"
)

const (
	// DefaultTerminalWidth is the default terminal width for clearing lines
	DefaultTerminalWidth = 80
	// DefaultProgressBarWidth is the default width for progress bars
	DefaultProgressBarWidth = 40
)

// isTerminal checks if stdout is a terminal (TTY)
func isTerminal() bool {
	return term.IsTerminal(int(os.Stdout.Fd()))
}

// ProgressIndicator provides simple progress feedback for CLI operations.
type ProgressIndicator struct {
	spinner  spinner.Model
	progress progress.Model
	message  string
	done     bool
	mu       sync.Mutex
	theme    Theme
}

// NewProgressIndicator creates a new progress indicator.
func NewProgressIndicator(message string) *ProgressIndicator {
	theme := DefaultTheme()
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(theme.Primary)

	p := progress.New(progress.WithDefaultGradient())
	p.Width = DefaultProgressBarWidth

	return &ProgressIndicator{
		spinner:  s,
		progress: p,
		message:  message,
		theme:    theme,
	}
}

// Start begins displaying the progress indicator, showing a spinner in TTY mode or simple text in non-TTY environments.
func (pi *ProgressIndicator) Start() {
	pi.mu.Lock()
	defer pi.mu.Unlock()

	if pi.done {
		return
	}

	// Only show interactive spinner in TTY
	if !isTerminal() {
		fmt.Printf("%s\n", pi.message)
		return
	}

	fmt.Printf("%s %s", pi.spinner.View(), pi.message)
}

// Update updates the progress bar with a percentage (0.0 to 1.0) and optional message, clearing the line in TTY mode.
func (pi *ProgressIndicator) Update(percent float64, message string) {
	pi.mu.Lock()
	defer pi.mu.Unlock()

	if pi.done {
		return
	}

	if message != "" {
		pi.message = message
	}

	// Only show interactive progress bar in TTY
	if !isTerminal() {
		// Simple log output for non-TTY (CI/CD, files, etc.)
		fmt.Printf("%s (%.0f%%)\n", pi.message, percent*100)
		return
	}

	// Clear the line and show progress bar
	fmt.Printf("\r%s", strings.Repeat(" ", DefaultTerminalWidth))
	fmt.Printf("\r%s %s", pi.progress.ViewAs(percent), pi.message)
}

// UpdateSpinner updates the spinner message for indeterminate progress, showing animated spinner in TTY mode.
func (pi *ProgressIndicator) UpdateSpinner(message string) {
	pi.mu.Lock()
	defer pi.mu.Unlock()

	if pi.done {
		return
	}

	if message != "" {
		pi.message = message
	}

	// Only show interactive spinner in TTY
	if !isTerminal() {
		// Simple log output for non-TTY
		fmt.Printf("%s\n", pi.message)
		return
	}

	// Clear the line and show spinner
	fmt.Printf("\r%s", strings.Repeat(" ", DefaultTerminalWidth))
	fmt.Printf("\r%s %s", pi.spinner.View(), pi.message)
}

// Finish completes the progress indicator with a success checkmark and final message.
func (pi *ProgressIndicator) Finish(message string) {
	pi.mu.Lock()
	defer pi.mu.Unlock()

	if pi.done {
		return
	}

	pi.done = true

	successStyle := lipgloss.NewStyle().Foreground(pi.theme.Success)

	// Only clear line in TTY
	if isTerminal() {
		fmt.Printf("\r%s", strings.Repeat(" ", DefaultTerminalWidth))
		fmt.Printf("\r%s %s\n", successStyle.Render("✓"), message)
	} else {
		fmt.Printf("✓ %s\n", message)
	}
}

// FinishWithError completes the progress indicator with an error symbol and error message.
func (pi *ProgressIndicator) FinishWithError(message string) {
	pi.mu.Lock()
	defer pi.mu.Unlock()

	if pi.done {
		return
	}

	pi.done = true

	errorStyle := lipgloss.NewStyle().Foreground(pi.theme.Error)

	// Only clear line in TTY
	if isTerminal() {
		fmt.Printf("\r%s", strings.Repeat(" ", DefaultTerminalWidth))
		fmt.Printf("\r%s %s\n", errorStyle.Render("✗"), message)
	} else {
		fmt.Printf("✗ %s\n", message)
	}
}

// BubblesSpinner provides a spinner using bubbles components (no manual goroutines).
type BubblesSpinner struct {
	spinner spinner.Model
	message string
	done    bool
	mu      sync.Mutex
	theme   Theme
}

// NewBubblesSpinner creates a spinner using bubbles components.
func NewBubblesSpinner(message string) *BubblesSpinner {
	theme := DefaultTheme()
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(theme.Info)

	return &BubblesSpinner{
		spinner: s,
		message: message,
		theme:   theme,
	}
}

// View renders the current spinner state following Bubble Tea's view interface pattern.
func (s *BubblesSpinner) View() string {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.done {
		return ""
	}

	messageStyle := lipgloss.NewStyle().Foreground(s.theme.Muted)

	return s.spinner.View() + " " + messageStyle.Render(s.message)
}

// Update updates the spinner state following Bubble Tea's update interface pattern.
func (s *BubblesSpinner) Update(msg any) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.done {
		return
	}

	s.spinner, _ = s.spinner.Update(msg)
}

// Stop stops the spinner and displays a success checkmark with the final message.
func (s *BubblesSpinner) Stop(finalMessage string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.done {
		return
	}

	s.done = true

	// Clear line and show final message
	fmt.Printf("\r%s", strings.Repeat(" ", DefaultTerminalWidth))
	if finalMessage != "" {
		successStyle := lipgloss.NewStyle().Foreground(s.theme.Success)
		fmt.Printf("\r%s %s\n", successStyle.Render("✓"), finalMessage)
	} else {
		fmt.Print("\r")
	}
}

// StopWithError stops the spinner and displays an error symbol with the error message.
func (s *BubblesSpinner) StopWithError(errorMessage string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.done {
		return
	}

	s.done = true

	// Clear line and show error message
	fmt.Printf("\r%s", strings.Repeat(" ", DefaultTerminalWidth))
	if errorMessage != "" {
		errorStyle := lipgloss.NewStyle().Foreground(s.theme.Error)
		fmt.Printf("\r%s %s\n", errorStyle.Render("✗"), errorMessage)
	} else {
		fmt.Print("\r")
	}
}

// ProgressBar creates a simple text-based progress bar for operations with known total steps.
func ProgressBar(current, total int, message string) {
	if total == 0 {
		return
	}

	percent := float64(current) / float64(total)
	width := DefaultProgressBarWidth
	filled := int(percent * float64(width))

	// Ensure filled doesn't exceed width to avoid negative repeat counts
	if filled > width {
		filled = width
	}

	bar := strings.Repeat("█", filled) + strings.Repeat("░", width-filled)
	percentage := int(percent * 100)

	fmt.Printf("\r[%s] %d%% (%d/%d) %s", bar, percentage, current, total, message)

	if current >= total {
		fmt.Println()
	}
}

// formatDuration formats duration in a user-friendly way
func formatDuration(d time.Duration) string {
	if d < time.Microsecond {
		return fmt.Sprintf("%dns", d.Nanoseconds())
	}
	if d < time.Millisecond {
		return fmt.Sprintf("%dµs", d.Microseconds())
	}
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	return d.Round(time.Millisecond).String()
}

// getTerminalWidth returns the terminal width, fallback to 80
func getTerminalWidth() int {
	width, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		return 80 // fallback
	}
	return width
}

// WithProgress wraps a function execution with a spinner, showing success or error on completion.
func WithProgress(message string, fn func() error) error {
	if fn == nil {
		return contextureerrors.ValidationErrorf("fn", "progress function cannot be nil")
	}

	spinner := NewBubblesSpinner(message)

	// Show initial state
	fmt.Print(spinner.View())

	err := fn()
	if err != nil {
		spinner.StopWithError(fmt.Sprintf("%s failed", message))
		return err
	}

	spinner.Stop(message)
	return nil
}

// WithProgressTiming wraps a function execution with a spinner and displays elapsed time on completion.
func WithProgressTiming(message string, fn func() error) error {
	if fn == nil {
		return contextureerrors.ValidationErrorf("fn", "progress function cannot be nil")
	}

	spinner := NewBubblesSpinner(message)
	start := time.Now()

	// Show initial state
	fmt.Print(spinner.View())

	err := fn()
	duration := time.Since(start)

	if err != nil {
		spinner.StopWithError(fmt.Sprintf("%s failed", message))
		return err
	}

	// Show completion with right-aligned timing
	showTimedCompletion("✓", message, duration, 0)
	return nil
}

// showTimedCompletion shows a completion message with right-aligned timing
func showTimedCompletion(icon, message string, duration time.Duration, indent int) {
	termWidth := getTerminalWidth()
	durationText := fmt.Sprintf("[%s]", formatDuration(duration))
	// Use RuneCountInString to count visual characters, not bytes
	visualTextLength := utf8.RuneCountInString(durationText)

	// Apply styling
	indentStr := strings.Repeat(" ", indent)
	theme := DefaultTheme()
	successStyle := lipgloss.NewStyle().Foreground(theme.Success)
	mutedStyle := lipgloss.NewStyle().Foreground(theme.Muted)

	// Print left side (message)
	fmt.Printf("\r%s", strings.Repeat(" ", termWidth)) // Clear entire line
	fmt.Printf("\r%s%s %s", indentStr, successStyle.Render(icon), message)

	// Calculate exact start position so timing ends at column termWidth
	// Use visual character count for proper alignment with Unicode characters
	timingStartColumn := termWidth - visualTextLength + 1
	if timingStartColumn > 0 {
		// Use ANSI positioning to place cursor at exact column
		fmt.Printf("\033[%dG%s", timingStartColumn, mutedStyle.Render(durationText))
	}
	fmt.Println() // Move to next line
}

// ShowFormatCompletion displays format completion status with right-aligned timing information.
func ShowFormatCompletion(formatName string, duration time.Duration) {
	showTimedCompletion("✓", formatName, duration, 2)
}
