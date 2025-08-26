// Package ui provides user interface components and styling for the Contexture CLI.
package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// MenuItem represents a menu item.
type MenuItem struct {
	Label       string
	Description string
	Shortcut    string
	Disabled    bool
}

// Menu creates an interactive-style menu display.
type Menu struct {
	title   string
	items   []MenuItem
	width   int
	compact bool
	theme   Theme
}

// NewMenu creates a new menu.
func NewMenu(title string) *Menu {
	return &Menu{
		title: title,
		width: 50,
		theme: DefaultTheme(),
	}
}

// AddItem adds an item to the menu.
func (m *Menu) AddItem(label, description string) *Menu {
	m.items = append(m.items, MenuItem{
		Label:       label,
		Description: description,
	})
	return m
}

// AddItemWithShortcut adds an item with a keyboard shortcut.
func (m *Menu) AddItemWithShortcut(label, description, shortcut string) *Menu {
	m.items = append(m.items, MenuItem{
		Label:       label,
		Description: description,
		Shortcut:    shortcut,
	})
	return m
}

// WithCompact sets compact mode.
func (m *Menu) WithCompact(compact bool) *Menu {
	m.compact = compact
	return m
}

// WithWidth sets the menu width.
func (m *Menu) WithWidth(width int) *Menu {
	if width > 0 {
		m.width = width
	}
	return m
}

// WithTheme sets a custom theme.
func (m *Menu) WithTheme(theme Theme) *Menu {
	m.theme = theme
	return m
}

// Render renders the menu as a string.
func (m *Menu) Render() string {
	titleStyle := lipgloss.NewStyle().
		Foreground(m.theme.Primary).
		Bold(true).
		Padding(0, 1).
		MarginBottom(1)

	itemStyle := lipgloss.NewStyle().
		Foreground(m.theme.Foreground).
		Padding(0, 2)

	shortcutStyle := lipgloss.NewStyle().
		Foreground(m.theme.Secondary).
		Bold(true)

	descStyle := lipgloss.NewStyle().
		Foreground(m.theme.Muted)

	disabledStyle := lipgloss.NewStyle().
		Foreground(m.theme.Muted).
		Strikethrough(true)

	var result string
	if m.title != "" {
		result = titleStyle.Render(m.title) + "\n"
	}

	for i, item := range m.items {
		if i > 0 && !m.compact {
			result += "\n"
		}

		var line string
		if item.Shortcut != "" {
			line = shortcutStyle.Render("["+item.Shortcut+"]") + " "
		}

		line += item.Label

		if item.Description != "" && !m.compact {
			line += "\n    " + descStyle.Render(item.Description)
		}

		if item.Disabled {
			line = disabledStyle.Render(line)
		} else {
			line = itemStyle.Render(line)
		}

		result += line + "\n"
	}

	menuStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(m.theme.Border).
		Padding(1).
		Width(m.width)

	return menuStyle.Render(strings.TrimSuffix(result, "\n"))
}
