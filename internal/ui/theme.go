// Package ui provides user interface components and styling for the Contexture CLI.
// It uses the charmbracelet/lipgloss library for terminal styling and provides
// a consistent theming system with adaptive colors for light/dark terminals.
package ui

import (
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

// Theme defines the color scheme for the CLI.
// All colors use lipgloss.AdaptiveColor for automatic light/dark theme support.
type Theme struct {
	Primary    lipgloss.AdaptiveColor
	Secondary  lipgloss.AdaptiveColor
	Success    lipgloss.AdaptiveColor
	Warning    lipgloss.AdaptiveColor
	Error      lipgloss.AdaptiveColor
	Info       lipgloss.AdaptiveColor
	Update     lipgloss.AdaptiveColor
	Muted      lipgloss.AdaptiveColor
	Background lipgloss.AdaptiveColor
	Foreground lipgloss.AdaptiveColor
	Border     lipgloss.AdaptiveColor
}

// DefaultTheme returns the default adaptive theme.
// Colors are based on CharmTheme for consistency with the huh library.
func DefaultTheme() Theme {
	return Theme{
		Primary: lipgloss.AdaptiveColor{Light: "#5A56E0", Dark: "#7571F9"}, // CharmTheme indigo
		Secondary: lipgloss.AdaptiveColor{
			Light: "#235",
			Dark:  "#252",
		}, // CharmTheme normalFg
		Success: lipgloss.AdaptiveColor{Light: "#02BA84", Dark: "#02BF87"}, // CharmTheme green
		Warning: lipgloss.AdaptiveColor{Light: "#D97706", Dark: "#F59E0B"}, // Amber
		Error:   lipgloss.AdaptiveColor{Light: "#FF4672", Dark: "#ED567A"}, // CharmTheme red
		Info:    lipgloss.AdaptiveColor{Light: "#5A56E0", Dark: "#7571F9"}, // Same as primary
		Update:  lipgloss.AdaptiveColor{Light: "#0EA5E9", Dark: "#0284C7"}, // Blue for updates
		Muted:   lipgloss.AdaptiveColor{Light: "#4B5563", Dark: "#6B7280"}, // Dark gray
		Background: lipgloss.AdaptiveColor{
			Light: "#FFFDF5",
			Dark:  "#000000",
		}, // CharmTheme cream/black
		Foreground: lipgloss.AdaptiveColor{
			Light: "#235",
			Dark:  "#252",
		}, // CharmTheme normalFg
		Border: lipgloss.AdaptiveColor{Light: "#E5E7EB", Dark: "#374151"}, // Light/Dark gray
	}
}

// Styles provides all the styled text rendering functions for a theme.
type Styles struct {
	theme Theme
}

// NewStyles creates a new Styles instance with the given theme.
func NewStyles(theme Theme) *Styles {
	return &Styles{theme: theme}
}

// Header returns a styled header with "▶" prefix
func (s *Styles) Header(text string) string {
	style := lipgloss.NewStyle().
		Bold(true).
		Foreground(s.theme.Primary).
		Background(s.theme.Background).
		Padding(0, 1).
		MarginTop(1)
	return style.Render("▶ " + text)
}

// Success returns styled success text with "✓" prefix
func (s *Styles) Success(text string) string {
	style := lipgloss.NewStyle().
		Foreground(s.theme.Success).
		Bold(true)
	return style.Render("✓ " + text)
}

// Warning returns styled warning text with "⚠" prefix
func (s *Styles) Warning(text string) string {
	style := lipgloss.NewStyle().
		Foreground(s.theme.Warning).
		Bold(true)
	return style.Render("⚠ " + text)
}

// Error returns styled error text with "✗" prefix
func (s *Styles) Error(text string) string {
	style := lipgloss.NewStyle().
		Foreground(s.theme.Error).
		Bold(true)
	return style.Render("✗ " + text)
}

// Info returns styled info text with "ⓘ" prefix
func (s *Styles) Info(text string) string {
	style := lipgloss.NewStyle().
		Foreground(s.theme.Info)
	return style.Render("ⓘ " + text)
}

// Muted returns styled muted text
func (s *Styles) Muted(text string) string {
	style := lipgloss.NewStyle().
		Foreground(s.theme.Muted)
	return style.Render(text)
}

// PrimaryStyle returns a style configured with the primary color
func (s *Styles) PrimaryStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(s.theme.Primary)
}

// SecondaryStyle returns a style configured with the secondary color
func (s *Styles) SecondaryStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(s.theme.Secondary)
}

// SuccessStyle returns a style configured with the success color
func (s *Styles) SuccessStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(s.theme.Success)
}

// ErrorStyle returns a style configured with the error color
func (s *Styles) ErrorStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(s.theme.Error)
}

// UpdateStyle returns a style configured with the update color
func (s *Styles) UpdateStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(s.theme.Update)
}

// BorderStyle returns a style with a normal border
func (s *Styles) BorderStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(s.theme.Border).
		Padding(1, 2)
}

// ConfigureHuhForm applies our theme to a huh form.
// It uses CharmTheme which is designed to work well with our color scheme.
func ConfigureHuhForm(form *huh.Form) *huh.Form {
	keymap := huh.NewDefaultKeyMap()
	keymap.Quit.SetKeys("ctrl+c", "esc", "q")

	return form.
		WithTheme(huh.ThemeCharm()).
		WithKeyMap(keymap)
}

// SpinnerChars are the frames for animated spinners
var SpinnerChars = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
