package ui

import (
	"github.com/charmbracelet/lipgloss"
)

// Card creates a styled card layout.
type Card struct {
	title   string
	content string
	width   int
	height  int
	border  lipgloss.Border
	theme   Theme
}

// NewCard creates a new card with the given title.
func NewCard(title string) *Card {
	return &Card{
		title:  title,
		width:  40,
		border: lipgloss.RoundedBorder(),
		theme:  DefaultTheme(),
	}
}

// WithContent sets the card content.
func (c *Card) WithContent(content string) *Card {
	c.content = content
	return c
}

// WithSize sets the card dimensions.
func (c *Card) WithSize(width, height int) *Card {
	if width > 0 {
		c.width = width
	}
	if height > 0 {
		c.height = height
	}
	return c
}

// WithBorder sets the card border.
func (c *Card) WithBorder(border lipgloss.Border) *Card {
	c.border = border
	return c
}

// WithTheme sets a custom theme.
func (c *Card) WithTheme(theme Theme) *Card {
	c.theme = theme
	return c
}

// Render renders the card as a string.
func (c *Card) Render() string {
	titleStyle := lipgloss.NewStyle().
		Foreground(c.theme.Primary).
		Bold(true).
		Padding(0, 1).
		Background(c.theme.Background)

	contentStyle := lipgloss.NewStyle().
		Foreground(c.theme.Foreground).
		Padding(1)

	cardStyle := lipgloss.NewStyle().
		Border(c.border).
		BorderForeground(c.theme.Border).
		Width(c.width)

	if c.height > 0 {
		cardStyle = cardStyle.Height(c.height)
	}

	content := titleStyle.Render(c.title)
	if c.content != "" {
		content += "\n" + contentStyle.Render(c.content)
	}

	return cardStyle.Render(content)
}
