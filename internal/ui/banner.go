package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Banner creates a styled banner with title and optional subtitle.
type Banner struct {
	title    string
	subtitle string
	width    int
	border   lipgloss.Border
	theme    Theme
}

// NewBanner creates a new banner with the given title.
func NewBanner(title string) *Banner {
	return &Banner{
		title:  title,
		width:  60,
		border: lipgloss.RoundedBorder(),
		theme:  DefaultTheme(),
	}
}

// WithSubtitle adds a subtitle to the banner.
func (b *Banner) WithSubtitle(subtitle string) *Banner {
	b.subtitle = subtitle
	return b
}

// WithWidth sets the banner width.
func (b *Banner) WithWidth(width int) *Banner {
	if width > 0 {
		b.width = width
	}
	return b
}

// WithBorder sets the border style.
func (b *Banner) WithBorder(border lipgloss.Border) *Banner {
	b.border = border
	return b
}

// WithTheme sets a custom theme.
func (b *Banner) WithTheme(theme Theme) *Banner {
	b.theme = theme
	return b
}

// Render renders the banner as a string.
func (b *Banner) Render() string {
	titleStyle := lipgloss.NewStyle().
		Foreground(b.theme.Primary).
		Bold(true).
		Align(lipgloss.Center)

	subtitleStyle := lipgloss.NewStyle().
		Foreground(b.theme.Secondary).
		Align(lipgloss.Center)

	var content string
	if b.subtitle != "" {
		content = titleStyle.Render(b.title) + "\n" + subtitleStyle.Render(b.subtitle)
	} else {
		content = titleStyle.Render(b.title)
	}

	bannerStyle := lipgloss.NewStyle().
		Border(b.border).
		BorderForeground(b.theme.Primary).
		Padding(1, 2).
		Align(lipgloss.Center).
		Width(b.width)

	return bannerStyle.Render(content)
}

// CommandHeader creates a simple styled header for CLI commands.
// This is a convenience function that doesn't require a Banner instance.
func CommandHeader(commandName string) string {
	// Map command names to descriptive text
	commandDescriptions := map[string]string{
		"init":   "Project Initialization",
		"build":  "Build Rule Files",
		"add":    "Add New Rule",
		"remove": "Remove Rule",
		"list":   "Installed Rules",
		"update": "Update Rules",
		"config": "Project Configuration",
	}

	headerText := commandDescriptions[commandName]
	if headerText == "" {
		// Fallback: capitalize words
		words := strings.Split(strings.ReplaceAll(commandName, "-", " "), " ")
		for i, word := range words {
			if len(word) > 0 {
				words[i] = strings.ToUpper(word[:1]) + strings.ToLower(word[1:])
			}
		}
		headerText = strings.Join(words, " ")
	}

	theme := DefaultTheme()
	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.Primary).
		Foreground(theme.Primary).
		Bold(true).
		Padding(0, 2).
		Align(lipgloss.Center)

	return style.Render(headerText)
}
