// Package tui provides terminal user interface components for Contexture.
//
// This package contains reusable TUI components built with Bubble Tea and Charm libraries,
// providing consistent and interactive user experiences across the application.
package tui

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/contextureai/contexture/internal/domain"
	"github.com/contextureai/contexture/internal/ui"
)

const (
	escKey     = "esc"
	quitKey    = "q"
	previewKey = "p"
)

// RulePreviewHelper provides shared preview functionality for all models
type RulePreviewHelper struct {
	viewport       viewport.Model
	contentBuilder *PreviewContentBuilder
}

// NewRulePreviewHelper creates a new preview helper
func NewRulePreviewHelper() *RulePreviewHelper {
	previewViewport := viewport.New(0, 0)
	previewViewport.SetContent("")

	contentBuilder := NewPreviewContentBuilder(
		previewTitleStyle,
		previewLabelStyle,
		previewMetadataStyle,
		previewErrorStyle,
		previewEmptyStyle,
	)

	return &RulePreviewHelper{
		viewport:       previewViewport,
		contentBuilder: contentBuilder,
	}
}

// SetupPreview initializes the preview for a rule
func (h *RulePreviewHelper) SetupPreview(rule *domain.Rule) {
	// Initialize viewport size if not set
	if h.viewport.Width == 0 || h.viewport.Height == 0 {
		// Set a reasonable default size
		h.viewport.Width = 80
		h.viewport.Height = 20
	}

	// Build preview content
	content := h.BuildPreviewContent(rule)

	// Set content and ensure viewport is at the top
	h.viewport.SetContent(content)
	h.viewport.GotoTop()
}

// UpdateSize updates the preview viewport size
func (h *RulePreviewHelper) UpdateSize(width, height int) {
	previewWidth, previewHeight := CalculatePreviewDimensions(width, height)
	h.viewport.Width = previewWidth
	h.viewport.Height = previewHeight
}

// BuildPreviewContent creates the content for the preview
func (h *RulePreviewHelper) BuildPreviewContent(rule *domain.Rule) string {
	return h.contentBuilder.BuildPreviewContentWithMarkdown(rule, h.RenderMarkdown)
}

// RenderMarkdown renders markdown content using glamour with user-friendly error handling
func (h *RulePreviewHelper) RenderMarkdown(markdown string) (string, error) {
	// Determine appropriate width for word wrapping
	width := h.viewport.Width - 4 // Account for padding
	if width <= 0 {
		width = 76 // Reasonable default width
	}

	// Create a glamour renderer with a dark theme that fits our color scheme
	renderer, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(width),
	)
	if err != nil {
		return "", &MarkdownRenderError{
			Type:    "renderer_init",
			Message: "Failed to initialize markdown renderer",
			Cause:   err,
		}
	}

	// Render the markdown
	renderedContent, err := renderer.Render(markdown)
	if err != nil {
		return "", &MarkdownRenderError{
			Type:    "render_failed",
			Message: "Failed to render markdown content",
			Cause:   err,
		}
	}

	return renderedContent, nil
}

// MarkdownRenderError provides detailed error information for markdown rendering failures
type MarkdownRenderError struct {
	Type    string // "renderer_init" or "render_failed"
	Message string
	Cause   error
}

func (e *MarkdownRenderError) Error() string {
	return e.Message
}

// GetUserFriendlyMessage returns a user-friendly error message with suggested actions
func (e *MarkdownRenderError) GetUserFriendlyMessage() string {
	switch e.Type {
	case "renderer_init":
		return "Preview unavailable: Markdown renderer could not be initialized. Rule content will be shown as plain text."
	case "render_failed":
		return "Preview formatting failed: The markdown content contains syntax that could not be rendered. Showing content as plain text."
	default:
		return "Preview error: Unable to format the rule content. Showing content as plain text."
	}
}

// RenderOverlay renders the preview dialog overlay
func (h *RulePreviewHelper) RenderOverlay(background string) string {
	// Create the preview box style
	previewStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#C084FC")).
		Padding(1, 2).
		Foreground(lipgloss.Color("#DDDDDD"))

	// Header with just the title
	headerText := "Rule Preview"
	headerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#EE6FF8")).
		Bold(true).
		Align(lipgloss.Center).
		Width(h.viewport.Width)

	header := headerStyle.Render(headerText)
	viewportContent := h.viewport.View()

	// Footer with help text
	helpText := "Press 'p', 'q', or 'esc' to close"
	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#6B7280")). // Muted gray color
		Align(lipgloss.Center).
		Width(h.viewport.Width)

	footer := helpStyle.Render(helpText)

	// Combine header, content, and footer
	previewContent := header + "\n\n" + viewportContent + "\n\n" + footer

	// Apply the preview box style
	styledPreview := previewStyle.Render(previewContent)

	// Center the preview on the screen
	return lipgloss.Place(
		lipgloss.Width(background), lipgloss.Height(background),
		lipgloss.Center, lipgloss.Center,
		styledPreview,
	)
}

// Update handles viewport updates and returns the command
func (h *RulePreviewHelper) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	h.viewport, cmd = h.viewport.Update(msg)
	return cmd
}

// Additional rule selector specific color constants
var (
	errorRed     = lipgloss.Color("#FF6B6B") // Error messages
	successGreen = lipgloss.AdaptiveColor{Light: "#02BA84", Dark: "#02BF87"}
	neutralWhite = lipgloss.Color("#FFFFFF") // White text
)

// Shared style constants
var (
	appStyle = lipgloss.NewStyle().Padding(0, 2)

	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFDF5")).
			Background(lipgloss.Color("#7571F9")).
			Padding(0, 1)

	statusMessageStyle = lipgloss.NewStyle().
				Foreground(successGreen).
				Render

	// Preview content styles
	previewTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(primaryPink)

	previewLabelStyle = lipgloss.NewStyle().
				Foreground(secondaryPurple)

	previewMetadataStyle = lipgloss.NewStyle().
				Foreground(mutedGray)

	previewErrorStyle = lipgloss.NewStyle().
				Foreground(errorRed)

	previewEmptyStyle = lipgloss.NewStyle().
				Foreground(darkGray)

	// Filter indicator styles
	filterLabelStyle = lipgloss.NewStyle().
				Foreground(secondaryPurple).
				Bold(true)

	filterValueStyle = lipgloss.NewStyle().
				Foreground(neutralWhite)
)

// buildFilterIndicator creates a standalone filter indicator when active
func buildFilterIndicator(listModel *list.Model) string {
	if listModel.IsFiltered() && listModel.FilterValue() != "" {
		filterLabel := filterLabelStyle.Render("Filter: ")
		filterValue := filterValueStyle.Render(listModel.FilterValue())
		return filterLabel + filterValue
	}
	return ""
}

// RuleSelector provides an interactive rule selection interface
type RuleSelector struct{}

// NewRuleSelector creates a new rule selector
func NewRuleSelector() *RuleSelector {
	return &RuleSelector{}
}

// SelectRules shows an interactive rule selection interface
func (rs *RuleSelector) SelectRules(rules []*domain.Rule, title string) ([]string, error) {
	if len(rules) == 0 {
		return nil, fmt.Errorf("no rules available for selection")
	}

	// Sort rules by title
	sort.Slice(rules, func(i, j int) bool {
		return rules[i].Title < rules[j].Title
	})

	// Convert to list items
	items := make([]list.Item, len(rules))
	for i, rule := range rules {
		items[i] = &ruleItem{rule: rule}
	}

	// Create model
	model := newRuleSelectionModel(items, title)

	// Run the program
	p := tea.NewProgram(model, tea.WithAltScreen())
	finalModel, err := p.Run()
	if err != nil {
		return nil, fmt.Errorf("failed to run rule selector: %w", err)
	}

	// Extract selected rules
	m, ok := finalModel.(*ruleSelectionModel)
	if !ok || m.quitting {
		return nil, nil // User cancelled
	}

	return m.chosen, nil
}

// DisplayRules shows rules in a read-only interface for browsing
func (rs *RuleSelector) DisplayRules(rules []*domain.Rule, title string) error {
	if len(rules) == 0 {
		fmt.Println("No rules found.")
		return nil
	}

	// Sort rules by title
	sort.Slice(rules, func(i, j int) bool {
		return rules[i].Title < rules[j].Title
	})

	// Convert to list items
	items := make([]list.Item, len(rules))
	for i, rule := range rules {
		items[i] = &ruleItem{rule: rule}
	}

	// Create model with read-only mode
	model := newRuleDisplayModel(items, title)

	// Run the program
	p := tea.NewProgram(model, tea.WithAltScreen())
	_, err := p.Run()
	if err != nil {
		return fmt.Errorf("failed to run rule display: %w", err)
	}

	return nil
}

// ruleItem implements list.Item for rules
type ruleItem struct {
	rule     *domain.Rule
	selected bool
}

func (i *ruleItem) FilterValue() string {
	// Include multiple fields for comprehensive filtering
	var parts []string

	// Extract rule path from ID
	rulePath := extractRulePath(i.rule.ID)

	// Add rule path for search
	if rulePath != "" {
		parts = append(parts, rulePath)
	}

	// Add title (highest priority)
	parts = append(parts, i.rule.Title)

	// Add description
	if i.rule.Description != "" {
		parts = append(parts, i.rule.Description)
	}

	// Add tags, frameworks, and languages (metadata)
	parts = append(parts, strings.Join(i.rule.Tags, " "))
	parts = append(parts, strings.Join(i.rule.Frameworks, " "))
	parts = append(parts, strings.Join(i.rule.Languages, " "))

	return strings.Join(parts, " ")
}

// MatchScore calculates how many matches this item has for a given filter
func (i *ruleItem) MatchScore(filter string) int {
	if filter == "" {
		return 0
	}

	score := 0
	filterLower := strings.ToLower(filter)

	// Extract rule path from ID
	rulePath := extractRulePath(i.rule.ID)

	// Count matches in different fields (with different weights)
	// Title matches are worth more
	score += countMatches(strings.ToLower(i.rule.Title), filterLower) * 3

	// Path matches
	if rulePath != "" {
		score += countMatches(strings.ToLower(rulePath), filterLower) * 2
	}

	// Description matches
	if i.rule.Description != "" {
		score += countMatches(strings.ToLower(i.rule.Description), filterLower)
	}

	// Metadata matches (tags, frameworks, languages)
	for _, tag := range i.rule.Tags {
		score += countMatches(strings.ToLower(tag), filterLower)
	}
	for _, framework := range i.rule.Frameworks {
		score += countMatches(strings.ToLower(framework), filterLower)
	}
	for _, language := range i.rule.Languages {
		score += countMatches(strings.ToLower(language), filterLower)
	}

	return score
}

func (i *ruleItem) Title() string {
	return i.rule.Title
}

func (i *ruleItem) Description() string {
	return i.rule.Description
}

// ruleSelectionModel is the Bubble Tea model for rule selection
type ruleSelectionModel struct {
	list           list.Model
	keys           *ruleSelectionKeyMap
	delegateKeys   *ruleDelegateKeyMap
	quitting       bool
	chosen         []string
	showingPreview bool
	previewRule    *domain.Rule
	previewHelper  *RulePreviewHelper
	baseTitle      string
}

func newRuleSelectionModel(items []list.Item, title string) *ruleSelectionModel {
	var (
		delegateKeys = newRuleDelegateKeyMap()
		listKeys     = newRuleSelectionKeyMap()
	)

	theme := ui.DefaultTheme()

	// Setup list with custom delegate
	delegate := newRuleItemDelegate(delegateKeys)
	ruleList := list.New(items, delegate, 0, 0)
	ruleList.Title = title
	ruleList.Styles.Title = titleStyle
	ruleList.SetShowStatusBar(false)  // Disable built-in status bar
	ruleList.SetShowPagination(false) // Hide pagination numbers

	// Override any default styles that might use inappropriate colors
	purpleColor := lipgloss.Color("#C084FC")

	ruleList.Styles.StatusBar = lipgloss.NewStyle().
		Foreground(theme.Muted).       // Visible for our selected count
		Background(lipgloss.Color("")) // No background
	ruleList.Styles.StatusEmpty = lipgloss.NewStyle().
		Foreground(lipgloss.Color("")). // Invisible
		Background(lipgloss.Color("")). // No background
		Width(0).                       // Zero width
		Height(0)                       // Zero height

	// Try all possible filter-related styles to override the yellow
	ruleList.Styles.StatusBarActiveFilter = lipgloss.NewStyle().
		Foreground(purpleColor).
		Bold(true).
		Background(lipgloss.Color(""))
	ruleList.Styles.StatusBarFilterCount = lipgloss.NewStyle().
		Foreground(lipgloss.Color("")). // Invisible - hide the count
		Background(lipgloss.Color("")). // No background
		Width(0).                       // Zero width to completely hide
		Height(0)                       // Zero height to completely hide

	// Set filter input and prompt styles
	ruleList.Styles.FilterPrompt = lipgloss.NewStyle().
		Foreground(purpleColor).
		Bold(true).
		Background(lipgloss.Color(""))
	ruleList.Styles.FilterCursor = lipgloss.NewStyle().
		Foreground(purpleColor)
	ruleList.FilterInput.PromptStyle = lipgloss.NewStyle().
		Foreground(purpleColor).
		Bold(true).
		Background(lipgloss.Color(""))

	// Try to set any additional styles that might affect filter display
	ruleList.Styles.NoItems = lipgloss.NewStyle().
		Foreground(theme.Muted).
		Background(lipgloss.Color("")) // No background
	ruleList.Styles.HelpStyle = lipgloss.NewStyle().
		Foreground(theme.Muted).
		Background(lipgloss.Color("")).
		Margin(0).
		Padding(0) // No background

	// Set to show help and enable filtering
	ruleList.SetShowHelp(true)
	ruleList.SetFilteringEnabled(true)

	// Get the default key map and customize it
	keyMap := list.DefaultKeyMap()
	// Disable unwanted keys by setting them to empty bindings
	keyMap.ClearFilter = key.NewBinding()
	// Keep AcceptWhileFiltering so users can apply filters with enter
	// Keep CancelWhileFiltering so users can escape from filter mode with esc
	keyMap.Quit = key.NewBinding()
	keyMap.ForceQuit = key.NewBinding()
	ruleList.KeyMap = keyMap

	// Add our custom keys to the short help by using AdditionalShortHelpKeys
	ruleList.AdditionalShortHelpKeys = func() []key.Binding {
		return []key.Binding{
			listKeys.preview,
			listKeys.quit,
		}
	}

	// Also add to full help for completeness
	ruleList.AdditionalFullHelpKeys = func() []key.Binding {
		return []key.Binding{
			listKeys.toggle,
			listKeys.choose,
			listKeys.preview,
			listKeys.quit,
		}
	}

	// Initialize preview helper
	previewHelper := NewRulePreviewHelper()

	return &ruleSelectionModel{
		list:          ruleList,
		keys:          listKeys,
		delegateKeys:  delegateKeys,
		previewHelper: previewHelper,
		baseTitle:     title,
	}
}

func newRuleDisplayModel(items []list.Item, title string) *ruleDisplayModel {
	theme := ui.DefaultTheme()

	// Setup list with custom delegate for read-only display
	delegate := newRuleDisplayDelegate()
	ruleList := list.New(items, delegate, 0, 0)
	ruleList.Title = title
	ruleList.Styles.Title = titleStyle
	ruleList.SetShowStatusBar(false)  // Disable built-in status bar
	ruleList.SetShowPagination(false) // Hide pagination numbers

	// Override any default styles that might use inappropriate colors
	purpleColor := lipgloss.Color("#C084FC")

	ruleList.Styles.StatusBar = lipgloss.NewStyle().
		Foreground(theme.Muted).       // Visible for our selected count
		Background(lipgloss.Color("")) // No background
	ruleList.Styles.StatusEmpty = lipgloss.NewStyle().
		Foreground(lipgloss.Color("")). // Invisible
		Background(lipgloss.Color("")). // No background
		Width(0).                       // Zero width
		Height(0)                       // Zero height

	// Try all possible filter-related styles to override the yellow
	ruleList.Styles.StatusBarActiveFilter = lipgloss.NewStyle().
		Foreground(purpleColor).
		Bold(true).
		Background(lipgloss.Color(""))
	ruleList.Styles.StatusBarFilterCount = lipgloss.NewStyle().
		Foreground(lipgloss.Color("")). // Invisible - hide the count
		Background(lipgloss.Color("")). // No background
		Width(0).                       // Zero width to completely hide
		Height(0)                       // Zero height to completely hide

	// Set filter input and prompt styles
	ruleList.Styles.FilterPrompt = lipgloss.NewStyle().
		Foreground(purpleColor).
		Bold(true).
		Background(lipgloss.Color(""))
	ruleList.Styles.FilterCursor = lipgloss.NewStyle().
		Foreground(purpleColor)
	ruleList.FilterInput.PromptStyle = lipgloss.NewStyle().
		Foreground(purpleColor).
		Bold(true).
		Background(lipgloss.Color(""))

	// Try to set any additional styles that might affect filter display
	ruleList.Styles.NoItems = lipgloss.NewStyle().
		Foreground(theme.Muted).
		Background(lipgloss.Color("")) // No background
	ruleList.Styles.HelpStyle = lipgloss.NewStyle().
		Foreground(theme.Muted).
		Background(lipgloss.Color("")).
		Margin(0).
		Padding(0) // No background

	// Set to show help and enable filtering
	ruleList.SetShowHelp(true)
	ruleList.SetFilteringEnabled(true)

	// Get the default key map and customize it
	keyMap := list.DefaultKeyMap()
	// Disable unwanted keys by setting them to empty bindings
	keyMap.ClearFilter = key.NewBinding()
	// Keep AcceptWhileFiltering so users can apply filters with enter
	// Keep CancelWhileFiltering so users can escape from filter mode with esc
	keyMap.Quit = key.NewBinding()
	keyMap.ForceQuit = key.NewBinding()
	ruleList.KeyMap = keyMap

	// Add preview and quit keys to short help so they're always visible
	previewKey := key.NewBinding(
		key.WithKeys("p"),
		key.WithHelp("p", "preview rule"),
	)
	quitKey := key.NewBinding(
		key.WithKeys("q", "esc"),
		key.WithHelp("q/esc", "quit"),
	)
	ruleList.AdditionalShortHelpKeys = func() []key.Binding {
		return []key.Binding{previewKey, quitKey}
	}

	// Also add to full help for completeness
	ruleList.AdditionalFullHelpKeys = func() []key.Binding {
		return []key.Binding{previewKey, quitKey}
	}

	// Initialize preview helper
	previewHelper := NewRulePreviewHelper()

	return &ruleDisplayModel{
		list:          ruleList,
		quitting:      false,
		previewHelper: previewHelper,
		baseTitle:     title,
	}
}

// ruleDisplayModel is the Bubble Tea model for read-only rule display
type ruleDisplayModel struct {
	list           list.Model
	quitting       bool
	showingPreview bool
	previewRule    *domain.Rule
	previewHelper  *RulePreviewHelper
	baseTitle      string
}

// ruleSelectionKeyMap defines keyboard shortcuts for the list
type ruleSelectionKeyMap struct {
	toggle  key.Binding
	choose  key.Binding
	preview key.Binding
	quit    key.Binding
}

func newRuleSelectionKeyMap() *ruleSelectionKeyMap {
	return &ruleSelectionKeyMap{
		toggle: key.NewBinding(
			key.WithKeys(" "),
			key.WithHelp("space", "toggle"),
		),
		choose: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "confirm selection"),
		),
		preview: key.NewBinding(
			key.WithKeys("p"),
			key.WithHelp("p", "preview rule"),
		),
		quit: key.NewBinding(
			key.WithKeys("q", "esc", "ctrl+c"),
			key.WithHelp("q/esc", "quit"),
		),
	}
}

// ruleDelegateKeyMap defines keyboard shortcuts for individual items
type ruleDelegateKeyMap struct {
	toggle key.Binding
}

func newRuleDelegateKeyMap() *ruleDelegateKeyMap {
	return &ruleDelegateKeyMap{
		toggle: key.NewBinding(
			key.WithKeys(" "),
			key.WithHelp("space", "toggle"),
		),
	}
}

func (d ruleDelegateKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{
		d.toggle,
	}
}

func (d ruleDelegateKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{
			d.toggle,
		},
	}
}

// Init initializes the model
func (m *ruleSelectionModel) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (m *ruleSelectionModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		h, v := appStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)
		// Update preview viewport size if showing preview
		if m.showingPreview {
			m.previewHelper.UpdateSize(msg.Width, msg.Height)
			// Regenerate content with new viewport width for proper word wrapping
			if m.previewRule != nil {
				m.previewHelper.SetupPreview(m.previewRule)
			}
		}

	case tea.KeyMsg:
		// Handle preview mode first
		if m.showingPreview {
			switch msg.String() {
			case escKey, quitKey, previewKey:
				m.showingPreview = false
				return m, nil
			default:
				// Pass scroll events to viewport
				cmd := m.previewHelper.Update(msg)
				return m, cmd
			}
		}

		// When actively filtering (typing), let the list handle the key first
		if m.list.FilterState() == list.Filtering {
			// Let list handle filtering first
			newListModel, cmd := m.list.Update(msg)
			m.list = newListModel
			cmds = append(cmds, cmd)
			return m, tea.Batch(cmds...)
		}

		switch {
		case key.Matches(msg, m.keys.preview):
			// Show preview for current item
			selectedItem := m.list.SelectedItem()
			if selectedItem != nil {
				if item, ok := selectedItem.(*ruleItem); ok {
					m.showingPreview = true
					m.previewRule = item.rule
					m.previewHelper.SetupPreview(item.rule)
					return m, nil
				}
			}
			return m, nil

		case key.Matches(msg, m.keys.toggle):
			// Toggle selection on current item
			// Get the actual selected item from the filtered view
			selectedItem := m.list.SelectedItem()
			if selectedItem != nil {
				if item, ok := selectedItem.(*ruleItem); ok {
					item.selected = !item.selected
					// Count will be updated in View function
					return m, nil
				}
			}
			return m, nil

		case key.Matches(msg, m.keys.choose):
			// Collect selected rules
			m.chosen = []string{}
			items := m.list.Items()
			for _, item := range items {
				if ruleItem, ok := item.(*ruleItem); ok && ruleItem.selected {
					m.chosen = append(m.chosen, ruleItem.rule.ID)
				}
			}
			return m, tea.Quit

		case key.Matches(msg, m.keys.quit):
			// Cancel/quit without selecting anything
			m.quitting = true
			return m, tea.Quit
		}
	}

	// This will also call our delegate's update function.
	newListModel, cmd := m.list.Update(msg)
	m.list = newListModel
	cmds = append(cmds, cmd)

	// Keep original title - filter will be rendered separately
	m.list.Title = m.baseTitle

	return m, tea.Batch(cmds...)
}

// View renders the model
func (m *ruleSelectionModel) View() string {
	if m.quitting {
		return ""
	}

	// Count selected items
	selectedCount := 0
	for _, item := range m.list.Items() {
		if ruleItem, ok := item.(*ruleItem); ok && ruleItem.selected {
			selectedCount++
		}
	}

	// Render list with custom status message above help menu
	listView := m.list.View()

	// Add filter indicator if active
	filterIndicator := buildFilterIndicator(&m.list)
	if filterIndicator != "" {
		listView = listView + "\n" + filterIndicator
	}

	// Add our selected count above the help menu if any items are selected
	if selectedCount > 0 {
		countText := fmt.Sprintf("%d selected", selectedCount)
		statusMsg := statusMessageStyle(countText)
		// Insert the status message before the help menu by finding and replacing the help section
		// The help appears at the bottom, so we'll add our message just before it
		listView = statusMsg + "\n" + listView
	}

	if m.showingPreview {
		return m.previewHelper.RenderOverlay(listView)
	}

	return appStyle.Render(listView)
}

// Init initializes the display model
func (m *ruleDisplayModel) Init() tea.Cmd {
	return nil
}

// Update handles messages for read-only display
func (m *ruleDisplayModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		h, v := appStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)
		// Update preview viewport size if showing preview
		if m.showingPreview {
			m.previewHelper.UpdateSize(msg.Width, msg.Height)
			// Regenerate content with new viewport width for proper word wrapping
			if m.previewRule != nil {
				m.previewHelper.SetupPreview(m.previewRule)
			}
		}

	case tea.KeyMsg:
		// Handle preview mode first
		if m.showingPreview {
			switch msg.String() {
			case previewKey, quitKey, escKey:
				m.showingPreview = false
				m.previewRule = nil
				return m, nil
			}
			// In preview mode, allow viewport scrolling
			cmd := m.previewHelper.Update(msg)
			return m, cmd
		}

		// When actively filtering (typing), let the list handle the key first
		if m.list.FilterState() == list.Filtering {
			// Let list handle filtering first
			newListModel, cmd := m.list.Update(msg)
			m.list = newListModel
			cmds = append(cmds, cmd)
			return m, tea.Batch(cmds...)
		}

		// Allow 'q' or 'esc' to quit the display
		switch msg.String() {
		case quitKey, escKey, "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		case previewKey:
			// Show preview for current item
			selectedItem := m.list.SelectedItem()
			if selectedItem != nil {
				if item, ok := selectedItem.(*ruleItem); ok {
					m.showingPreview = true
					m.previewRule = item.rule
					m.previewHelper.SetupPreview(item.rule)
					return m, nil
				}
			}
			return m, nil
		}
	}

	// Update the list (allows filtering and navigation)
	newListModel, cmd := m.list.Update(msg)
	m.list = newListModel
	cmds = append(cmds, cmd)

	// Keep original title - filter will be rendered separately
	m.list.Title = m.baseTitle

	return m, tea.Batch(cmds...)
}

// View renders the display model
func (m *ruleDisplayModel) View() string {
	if m.quitting {
		return ""
	}

	if m.showingPreview {
		// Render the preview overlay
		return m.previewHelper.RenderOverlay(m.list.View())
	}

	// Just render the list with filter indicator
	listView := m.list.View()

	// Add filter indicator if active
	filterIndicator := buildFilterIndicator(&m.list)
	if filterIndicator != "" {
		listView = listView + "\n" + filterIndicator
	}

	return appStyle.Render(listView)
}

// ruleItemDelegate implements list.ItemDelegate for custom rendering
type ruleItemDelegate struct {
	keys   *ruleDelegateKeyMap
	theme  ui.Theme
	styles ruleItemStyles
}

func newRuleItemDelegate(keys *ruleDelegateKeyMap) *ruleItemDelegate {
	return &ruleItemDelegate{
		keys:   keys,
		theme:  ui.DefaultTheme(),
		styles: createRuleItemStyles(),
	}
}

func newRuleDisplayDelegate() *ruleDisplayDelegate {
	return &ruleDisplayDelegate{
		theme:  ui.DefaultTheme(),
		styles: createRuleItemStyles(),
	}
}

func (d *ruleItemDelegate) Height() int {
	return 5 // Slightly smaller to fit more items and reduce blank space
}

func (d *ruleItemDelegate) Spacing() int {
	return 1
}

func (d *ruleItemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd {
	// We handle toggle in the main model Update, not here
	return nil
}

func (d *ruleItemDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	ruleItem, ok := item.(*ruleItem)
	if !ok {
		return
	}

	rule := ruleItem.rule
	title := rule.Title
	desc := rule.Description

	// Extract rule path from ID with local indicator
	rulePath := extractRulePathWithLocalIndicator(rule)

	// Build metadata lines
	basicMetadataLine, triggerLine, variablesLine := buildRuleMetadata(rule)

	if m.Width() <= 0 {
		return
	}

	// Determine selection and filter states
	isSelected := index == m.Index()
	isFiltered := m.FilterState() == list.Filtering || m.FilterState() == list.FilterApplied
	emptyFilter := m.FilterState() == list.Filtering && m.FilterValue() == ""

	// Get the current filter value for highlighting
	filterValue := ""
	if isFiltered && !emptyFilter {
		filterValue = m.FilterValue()
	}

	// Define colors for explicit styling
	selectedColor := lipgloss.Color(
		"#EE6FF8",
	) // Bright pink for selected titles
	selectedDescColor := lipgloss.Color(
		"#C084FC",
	) // More muted purple-pink for descriptions
	selectedMetaColor := lipgloss.Color(
		"#9CA3AF",
	) // Much more muted grey-pink for metadata
	borderColor := lipgloss.AdaptiveColor{Light: "#F793FF", Dark: "#AD58B4"} // Pink border colors
	theme := d.theme

	// Get checkbox content with appropriate colors based on state
	checkbox := "[ ]"
	var checkboxColor lipgloss.TerminalColor

	// Always show checkmark if the item is selected, regardless of hover state
	if ruleItem.selected {
		checkbox = "[✓]"
	}

	switch {
	case isSelected:
		// Pink when highlighted/hovered (always pink when hovering)
		checkboxColor = selectedColor // Pink for highlighted
	case ruleItem.selected:
		// Green when selected but not highlighted/hovered
		checkboxColor = d.theme.Success // Green for selected
	default:
		// Gray for normal state
		checkboxColor = d.theme.Muted // Gray for unselected
	}

	// Build the content lines for rendering
	lines := []string{}

	// Add rule path if it exists
	if rulePath != "" {
		lines = append(lines, rulePath)
	}

	// Add title and description
	lines = append(lines, title) // Title without checkbox
	lines = append(lines, desc)

	// Add basic metadata line if it has content
	if basicMetadataLine != "" {
		lines = append(lines, basicMetadataLine)
	}

	// Add trigger line if it has content
	if triggerLine != "" {
		lines = append(lines, triggerLine)
	}

	// Add variables line if it has content
	if variablesLine != "" {
		lines = append(lines, variablesLine)
	}

	// Apply styles and highlighting with separate checkbox handling
	var styledLines []string

	// Create checkbox style
	checkboxStyle := lipgloss.NewStyle().Foreground(checkboxColor).Bold(true)

	if emptyFilter {
		// Dimmed when filtering with empty query, but maintain proper padding for selected items
		checkboxStyled := checkboxStyle.Faint(true).Render(checkbox)

		config := StyleConfig{
			Theme:             &theme,
			SelectedColor:     selectedColor,
			SelectedDescColor: selectedDescColor,
			SelectedMetaColor: selectedMetaColor,
			BorderColor:       borderColor,
			DimmedPath:        d.styles.dimmedPath,
			DimmedDesc:        d.styles.dimmedDesc,
			DimmedMeta:        d.styles.dimmedMeta,
		}

		styledLines = renderStyledLinesForEmptyFilter(lines, rulePath, checkboxStyled, isSelected, config)
	} else {
		// Normal and filtered styling

		// Define filter-specific colors for consistent text color after highlighting
		filterTitleColor := lipgloss.AdaptiveColor{Light: "#1a1a1a", Dark: "#dddddd"} // White/dark for titles
		filterDescColor := lipgloss.Color("#C084FC")                                  // Purple for descriptions
		filterMetaColor := lipgloss.Color("#6B7280")                                  // Dark grey for metadata
		filterPathColor := filterMetaColor                                            // Same as metadata

		// Determine line indices based on whether path exists
		pathIndex := -1
		titleIndex := 0
		descIndex := 1

		if rulePath != "" {
			pathIndex = 0
			titleIndex = 1
			descIndex = 2
		}

		for i, line := range lines {
			switch i {
			case pathIndex:
				// Rule path line
				var pathStyle lipgloss.Style
				if isSelected {
					pathStyle = d.styles.selectedPath
				} else {
					if filterValue != "" {
						pathStyle = lipgloss.NewStyle().
							Foreground(filterPathColor).
							Padding(0, 0, 0, 2)
					} else {
						pathStyle = d.styles.normalPath
					}
				}

				if filterValue != "" {
					styledLines = append(styledLines, d.applyHighlights(line, filterValue, pathStyle, d.styles.matchHighlight))
				} else {
					styledLines = append(styledLines, pathStyle.Render(line))
				}
			case titleIndex:
				// Title with checkbox
				if isSelected {
					if filterValue != "" {
						// For selected items when filtering: pink base with white highlights
						// Style the checkbox with pink color
						styledCheckbox := lipgloss.NewStyle().Foreground(selectedColor).Bold(true).Render(checkbox)

						// Apply highlighting to title only
						highlightedTitle := d.applyTitleHighlighting(line, filterValue, selectedColor)

						// Create the combined content first
						combinedContent := styledCheckbox + " " + highlightedTitle

						// Apply border and padding without overriding internal colors
						structuralStyle := lipgloss.NewStyle().
							Border(lipgloss.ThickBorder(), false, false, false, true).
							BorderForeground(borderColor).
							Padding(0, 0, 0, 1)

						// Apply the structural style with preserved internal formatting
						fullLine := structuralStyle.Render(combinedContent)

						styledLines = append(styledLines, fullLine)
					} else {
						// No filter - create a complete pink style for the whole line
						fullSelectedStyle := lipgloss.NewStyle().
							Border(lipgloss.ThickBorder(), false, false, false, true).
							BorderForeground(borderColor).
							Foreground(selectedColor).
							Bold(true).
							Padding(0, 0, 0, 1)
						checkboxAndTitle := checkbox + " " + line
						styledLines = append(styledLines, fullSelectedStyle.Render(checkboxAndTitle))
					}
				} else {
					// Normal items - render checkbox separately
					checkboxStyled := checkboxStyle.Render(checkbox)

					// Normal items when filtering: white base with pink highlights
					if filterValue != "" {
						titleStyle = lipgloss.NewStyle().
							Foreground(filterTitleColor).
							Bold(true).
							Padding(0, 0, 0, 2)

						checkboxAndTitle := checkboxStyled + " " + line
						// Use pink highlighting for non-selected items
						pinkHighlight := lipgloss.NewStyle().Foreground(selectedColor).Bold(true)
						styledLines = append(styledLines, d.applyHighlights(checkboxAndTitle, filterValue, titleStyle, pinkHighlight))
					} else {
						titleStyle = d.styles.normalTitle
						checkboxAndTitle := checkboxStyled + " " + line
						styledLines = append(styledLines, titleStyle.Render(checkboxAndTitle))
					}
				}
			case descIndex:
				// Description
				var descStyle lipgloss.Style
				if isSelected {
					descStyle = d.styles.selectedDesc // Use selected style with border
				} else {
					// Use filter colors when filtering
					if filterValue != "" {
						descStyle = lipgloss.NewStyle().
							Foreground(filterDescColor).
							Padding(0, 0, 0, 2)
					} else {
						descStyle = d.styles.normalDesc
					}
				}

				if filterValue != "" {
					styledLines = append(styledLines, d.applyHighlights(line, filterValue, descStyle, d.styles.matchHighlight))
				} else {
					styledLines = append(styledLines, descStyle.Render(line))
				}
			default:
				// Metadata lines (basic metadata and trigger)
				var metadataStyle lipgloss.Style
				if isSelected {
					metadataStyle = d.styles.selectedMeta // Use selected metadata style with pink color
				} else {
					// Use filter colors when filtering
					if filterValue != "" {
						metadataStyle = lipgloss.NewStyle().
							Foreground(filterMetaColor).
							Padding(0, 0, 0, 2)
					} else {
						metadataStyle = d.styles.normalMeta // Use darker metadata style
					}
				}

				if filterValue != "" {
					styledLines = append(styledLines, d.applyHighlights(line, filterValue, metadataStyle, d.styles.matchHighlight))
				} else {
					styledLines = append(styledLines, metadataStyle.Render(line))
				}
			}
		}
	}

	// Output the final result (no background styling, borders handled by individual line styles)
	_, _ = fmt.Fprint(w, strings.Join(styledLines, "\n"))
}

// applyHighlights applies highlighting by finding matches within the specific text
func (d *ruleItemDelegate) applyHighlights(
	text, filterValue string,
	baseStyle, highlightStyle lipgloss.Style,
) string {
	return applyHighlightsGeneric(text, filterValue, baseStyle, highlightStyle)
}

// applyTitleHighlighting applies pink base with white highlights for selected titles
func (d *ruleItemDelegate) applyTitleHighlighting(
	title, filterValue string,
	baseColor lipgloss.TerminalColor,
) string {
	return applyTitleHighlightingGeneric(title, filterValue, baseColor)
}

// ruleDisplayDelegate implements list.ItemDelegate for read-only display (no checkboxes)
type ruleDisplayDelegate struct {
	theme  ui.Theme
	styles ruleItemStyles
}

func (d *ruleDisplayDelegate) Height() int {
	return 5 // Slightly smaller to fit more items and reduce blank space
}

func (d *ruleDisplayDelegate) Spacing() int {
	return 1
}

func (d *ruleDisplayDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd {
	return nil
}

func (d *ruleDisplayDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	ruleItem, ok := item.(*ruleItem)
	if !ok {
		return
	}

	rule := ruleItem.rule
	title := rule.Title
	desc := rule.Description

	// Extract rule path from ID with local indicator
	rulePath := extractRulePathWithLocalIndicator(rule)

	// Build metadata lines
	basicMetadataLine, triggerLine, variablesLine := buildRuleMetadata(rule)

	if m.Width() <= 0 {
		return
	}

	// Determine selection and filter states
	isSelected := index == m.Index()
	isFiltered := m.FilterState() == list.Filtering || m.FilterState() == list.FilterApplied
	emptyFilter := m.FilterState() == list.Filtering && m.FilterValue() == ""

	// Get the current filter value for highlighting
	filterValue := ""
	if isFiltered && !emptyFilter {
		filterValue = m.FilterValue()
	}

	// Build the content lines for rendering
	lines := []string{}

	// Add rule path if it exists
	if rulePath != "" {
		lines = append(lines, rulePath)
	}

	// Add title and description (no checkbox)
	lines = append(lines, title)
	lines = append(lines, desc)

	// Add basic metadata line if it has content
	if basicMetadataLine != "" {
		lines = append(lines, basicMetadataLine)
	}

	// Add trigger line if it has content
	if triggerLine != "" {
		lines = append(lines, triggerLine)
	}

	// Add variables line if it has content
	if variablesLine != "" {
		lines = append(lines, variablesLine)
	}

	// Apply styles and highlighting (similar to ruleItemDelegate but without checkboxes)
	var styledLines []string

	if emptyFilter {
		// Dimmed when filtering with empty query
		pathIndex := -1
		titleIndex := 0
		descIndex := 1

		if rulePath != "" {
			pathIndex = 0
			titleIndex = 1
			descIndex = 2
		}

		for i, line := range lines {
			switch i {
			case pathIndex:
				// Rule path line
				var pathStyle lipgloss.Style
				if isSelected {
					pathStyle = d.styles.selectedPath.Faint(true)
				} else {
					pathStyle = d.styles.dimmedPath
				}
				styledLines = append(styledLines, pathStyle.Render(line))
			case titleIndex:
				// Title line (no checkbox)
				var titleStyle lipgloss.Style
				if isSelected {
					titleStyle = d.styles.selectedTitle.Faint(true)
				} else {
					titleStyle = d.styles.dimmedTitle
				}
				styledLines = append(styledLines, titleStyle.Render(line))
			case descIndex:
				// Description
				var descStyle lipgloss.Style
				if isSelected {
					descStyle = d.styles.selectedDesc.Faint(true)
				} else {
					descStyle = d.styles.dimmedDesc
				}
				styledLines = append(styledLines, descStyle.Render(line))
			default:
				// Metadata
				var metadataStyle lipgloss.Style
				if isSelected {
					metadataStyle = d.styles.selectedMeta.Faint(true)
				} else {
					metadataStyle = d.styles.dimmedMeta
				}
				styledLines = append(styledLines, metadataStyle.Render(line))
			}
		}
	} else {
		// Normal and filtered styling (similar to original but without checkboxes)
		pathIndex := -1
		titleIndex := 0
		descIndex := 1

		if rulePath != "" {
			pathIndex = 0
			titleIndex = 1
			descIndex = 2
		}

		for i, line := range lines {
			switch i {
			case pathIndex:
				// Rule path line
				var pathStyle lipgloss.Style
				if isSelected {
					pathStyle = d.styles.selectedPath
				} else {
					pathStyle = d.styles.normalPath
				}

				if filterValue != "" {
					styledLines = append(styledLines, d.applyHighlights(line, filterValue, pathStyle, d.styles.matchHighlight))
				} else {
					styledLines = append(styledLines, pathStyle.Render(line))
				}
			case titleIndex:
				// Title line (no checkbox)
				if isSelected {
					if filterValue != "" {
						// For selected items when filtering: pink base with white highlights
						// Apply highlighting to title only
						highlightedTitle := d.applyTitleHighlighting(line, filterValue, lipgloss.Color("#EE6FF8"))

						// Create the styled line with border and padding
						structuralStyle := lipgloss.NewStyle().
							Border(lipgloss.ThickBorder(), false, false, false, true).
							BorderForeground(lipgloss.AdaptiveColor{Light: "#F793FF", Dark: "#AD58B4"}).
							Padding(0, 0, 0, 1)

						// Apply the structural style with preserved internal formatting
						fullLine := structuralStyle.Render(highlightedTitle)
						styledLines = append(styledLines, fullLine)
					} else {
						// No filter - use normal selected style
						styledLines = append(styledLines, d.styles.selectedTitle.Render(line))
					}
				} else {
					// Normal items - use generic highlighting or normal style
					if filterValue != "" {
						styledLines = append(styledLines, d.applyHighlights(line, filterValue, d.styles.normalTitle, d.styles.matchHighlight))
					} else {
						styledLines = append(styledLines, d.styles.normalTitle.Render(line))
					}
				}
			case descIndex:
				// Description
				var descStyle lipgloss.Style
				if isSelected {
					descStyle = d.styles.selectedDesc
				} else {
					descStyle = d.styles.normalDesc
				}

				if filterValue != "" {
					styledLines = append(styledLines, d.applyHighlights(line, filterValue, descStyle, d.styles.matchHighlight))
				} else {
					styledLines = append(styledLines, descStyle.Render(line))
				}
			default:
				// Metadata lines
				var metadataStyle lipgloss.Style
				if isSelected {
					metadataStyle = d.styles.selectedMeta
				} else {
					metadataStyle = d.styles.normalMeta
				}

				if filterValue != "" {
					styledLines = append(styledLines, d.applyHighlights(line, filterValue, metadataStyle, d.styles.matchHighlight))
				} else {
					styledLines = append(styledLines, metadataStyle.Render(line))
				}
			}
		}
	}

	// Output the final result
	_, _ = fmt.Fprint(w, strings.Join(styledLines, "\n"))
}

// applyHighlights applies highlighting by finding matches within the specific text (same as ruleItemDelegate)
func (d *ruleDisplayDelegate) applyHighlights(
	text, filterValue string,
	baseStyle, highlightStyle lipgloss.Style,
) string {
	return applyHighlightsGeneric(text, filterValue, baseStyle, highlightStyle)
}

// applyTitleHighlighting applies pink base with white highlights for selected titles (for ruleDisplayDelegate)
func (d *ruleDisplayDelegate) applyTitleHighlighting(
	title, filterValue string,
	baseColor lipgloss.TerminalColor,
) string {
	return applyTitleHighlightingGeneric(title, filterValue, baseColor)
}
