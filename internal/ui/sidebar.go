// Package ui provides user interface components and styling for the Contexture CLI.
package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// SidebarItem represents an item in the sidebar.
type SidebarItem struct {
	Label    string
	Value    string
	Icon     string
	Selected bool
	Indent   int
}

// Sidebar creates a sidebar layout component.
type Sidebar struct {
	title  string
	items  []SidebarItem
	width  int
	height int
	theme  Theme
}

// NewSidebar creates a new sidebar.
func NewSidebar(title string) *Sidebar {
	return &Sidebar{
		title:  title,
		width:  30,
		height: 20,
		theme:  DefaultTheme(),
	}
}

// AddItem adds an item to the sidebar.
func (s *Sidebar) AddItem(label, value string) *Sidebar {
	s.items = append(s.items, SidebarItem{
		Label: label,
		Value: value,
	})
	return s
}

// AddItemWithIcon adds an item with an icon.
func (s *Sidebar) AddItemWithIcon(label, value, icon string) *Sidebar {
	s.items = append(s.items, SidebarItem{
		Label: label,
		Value: value,
		Icon:  icon,
	})
	return s
}

// AddIndentedItem adds an indented item (for hierarchy).
func (s *Sidebar) AddIndentedItem(label, value string, indent int) *Sidebar {
	s.items = append(s.items, SidebarItem{
		Label:  label,
		Value:  value,
		Indent: indent,
	})
	return s
}

// WithSize sets the sidebar dimensions.
func (s *Sidebar) WithSize(width, height int) *Sidebar {
	if width > 0 {
		s.width = width
	}
	if height > 0 {
		s.height = height
	}
	return s
}

// WithTheme sets a custom theme.
func (s *Sidebar) WithTheme(theme Theme) *Sidebar {
	s.theme = theme
	return s
}

// Render renders the sidebar as a string.
func (s *Sidebar) Render() string {
	titleStyle := lipgloss.NewStyle().
		Foreground(s.theme.Primary).
		Bold(true).
		Padding(0, 1).
		Background(s.theme.Background).
		Width(s.width - 2).
		Align(lipgloss.Center)

	selectedStyle := lipgloss.NewStyle().
		Foreground(s.theme.Primary).
		Background(s.theme.Background).
		Bold(true).
		Padding(0, 1)

	normalStyle := lipgloss.NewStyle().
		Foreground(s.theme.Foreground).
		Padding(0, 1)

	var content string
	if s.title != "" {
		content = titleStyle.Render(s.title) + "\n"
	}

	for _, item := range s.items {
		indent := strings.Repeat("  ", item.Indent)

		var line string
		if item.Icon != "" {
			line = item.Icon + " "
		}
		line += item.Label

		if item.Selected {
			line = selectedStyle.Render(indent + line)
		} else {
			line = normalStyle.Render(indent + line)
		}

		content += line + "\n"
	}

	sidebarStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(s.theme.Border).
		Width(s.width).
		Height(s.height)

	return sidebarStyle.Render(strings.TrimSuffix(content, "\n"))
}
