// Package ui provides user interface components and styling for the Contexture CLI.
package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// DividerStyle represents different divider styles.
type DividerStyle int

const (
	// DividerPlain represents a plain divider style.
	DividerPlain DividerStyle = iota
	// DividerDashed represents a dashed divider style.
	DividerDashed
	// DividerDotted represents a dotted divider style.
	DividerDotted
	// DividerDouble represents a double-line divider style.
	DividerDouble
)

// Divider creates a styled divider line.
type Divider struct {
	text      string
	width     int
	character string
	style     DividerStyle
	theme     Theme
}

// NewDivider creates a new divider.
func NewDivider() *Divider {
	return &Divider{
		width:     60,
		character: "─",
		style:     DividerPlain,
		theme:     DefaultTheme(),
	}
}

// WithText adds text to the divider.
func (d *Divider) WithText(text string) *Divider {
	d.text = text
	return d
}

// WithWidth sets the divider width.
func (d *Divider) WithWidth(width int) *Divider {
	if width > 0 {
		d.width = width
	}
	return d
}

// WithStyle sets the divider style.
func (d *Divider) WithStyle(style DividerStyle) *Divider {
	d.style = style
	switch style {
	case DividerPlain:
		d.character = "─"
	case DividerDashed:
		d.character = "┄"
	case DividerDotted:
		d.character = "┈"
	case DividerDouble:
		d.character = "═"
	default:
		d.character = "─"
	}
	return d
}

// WithTheme sets a custom theme.
func (d *Divider) WithTheme(theme Theme) *Divider {
	d.theme = theme
	return d
}

// Render renders the divider as a string.
func (d *Divider) Render() string {
	dividerStyle := lipgloss.NewStyle().
		Foreground(d.theme.Border)

	if d.text == "" {
		return dividerStyle.Render(strings.Repeat(d.character, d.width))
	}

	textStyle := lipgloss.NewStyle().
		Foreground(d.theme.Secondary).
		Bold(true)

	textWidth := len(d.text) + 2 // Add padding spaces
	if textWidth >= d.width {
		return textStyle.Render(" " + d.text + " ")
	}

	sideWidth := (d.width - textWidth) / 2
	leftSide := strings.Repeat(d.character, sideWidth)
	rightSide := strings.Repeat(d.character, d.width-sideWidth-textWidth)

	left := dividerStyle.Render(leftSide)
	text := textStyle.Render(" " + d.text + " ")
	right := dividerStyle.Render(rightSide)

	return left + text + right
}
