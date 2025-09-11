package tui

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/contextureai/contexture/internal/domain"
	"github.com/contextureai/contexture/internal/ui"
)

// FileBrowserItem represents an item in the file browser (folder or rule)
type FileBrowserItem struct {
	node     *domain.RuleNode
	rule     *domain.Rule
	selected bool
}

// FilterValue implements list.Item
func (i *FileBrowserItem) FilterValue() string {
	if i.rule != nil {
		// Include rule details for filtering
		var parts []string
		parts = append(parts, i.rule.Title)
		if i.rule.Description != "" {
			parts = append(parts, i.rule.Description)
		}
		parts = append(parts, strings.Join(i.rule.Tags, " "))
		parts = append(parts, strings.Join(i.rule.Frameworks, " "))
		parts = append(parts, strings.Join(i.rule.Languages, " "))
		parts = append(parts, i.node.RuleID) // Include the rule path
		return strings.Join(parts, " ")
	}
	return i.node.Name // For folders, just use the name
}

// Title implements list.Item
func (i *FileBrowserItem) Title() string {
	if i.node.Type == domain.RuleNodeTypeFolder {
		return "📁 " + i.node.Name
	}
	return "📄 " + i.node.Name
}

// Description implements list.Item
func (i *FileBrowserItem) Description() string {
	if i.rule != nil {
		return i.rule.Description
	}
	return "" // Folders don't have descriptions
}

// IsFolder returns true if this item is a folder
func (i *FileBrowserItem) IsFolder() bool {
	return i.node.Type == domain.RuleNodeTypeFolder
}

// IsRule returns true if this item is a rule
func (i *FileBrowserItem) IsRule() bool {
	return i.node.Type == domain.RuleNodeTypeRule
}

// FileBrowser provides a folder-based navigation interface for rules
type FileBrowser struct{}

// NewFileBrowser creates a new file browser
func NewFileBrowser() *FileBrowser {
	return &FileBrowser{}
}

// BrowseRules shows a file browser interface for rule selection
func (fb *FileBrowser) BrowseRules(ruleTree *domain.RuleNode, allRules []*domain.Rule, title string) ([]string, error) {
	// Create a map of rule IDs to rules for quick lookup
	ruleMap := make(map[string]*domain.Rule)
	for _, rule := range allRules {
		// Extract rule display path from the rule's file path or ID (includes source for custom rules)
		rulePath := domain.ExtractRuleDisplayPath(rule.ID)
		if rulePath == "" {
			rulePath = rule.FilePath
		}
		ruleMap[rulePath] = rule
	}

	// Create the model
	model := newFileBrowserModel(ruleTree, ruleMap, title)

	// Run the program
	p := tea.NewProgram(model, tea.WithAltScreen())
	finalModel, err := p.Run()
	if err != nil {
		return nil, fmt.Errorf("failed to run file browser: %w", err)
	}

	// Extract selected rules
	m, ok := finalModel.(*fileBrowserModel)
	if !ok || m.quitting {
		return nil, nil // User cancelled
	}

	return m.chosen, nil
}

// fileBrowserModel is the Bubble Tea model for file browsing
type fileBrowserModel struct {
	list           list.Model
	keys           *fileBrowserKeyMap
	delegateKeys   *fileBrowserDelegateKeyMap
	quitting       bool
	chosen         []string
	showingPreview bool
	previewRule    *domain.Rule
	previewHelper  *RulePreviewHelper
	baseTitle      string
	ruleTree       *domain.RuleNode
	currentNode    *domain.RuleNode
	ruleMap        map[string]*domain.Rule
	isFiltering    bool
	allRulesItems  []list.Item // Flat list for filtering
	folderItems    []list.Item // Cached folder items to avoid rebuilding
}

func newFileBrowserModel(ruleTree *domain.RuleNode, ruleMap map[string]*domain.Rule, title string) *fileBrowserModel {
	var (
		delegateKeys = newFileBrowserDelegateKeyMap()
		listKeys     = newFileBrowserKeyMap()
	)

	theme := ui.DefaultTheme()

	// Setup list with custom delegate
	delegate := newFileBrowserItemDelegate(delegateKeys)

	// Start at root and build initial items
	initialItems := buildItemsForNode(ruleTree, ruleMap)

	browserList := list.New(initialItems, delegate, 100, 30)
	browserList.Title = title
	browserList.Styles.Title = titleStyle
	browserList.SetShowStatusBar(false)  // Disable built-in status bar
	browserList.SetShowPagination(false) // Hide pagination numbers

	// Configure styles similar to rule selector
	purpleColor := lipgloss.Color("#C084FC")

	browserList.Styles.StatusBar = lipgloss.NewStyle().
		Foreground(theme.Muted).
		Background(lipgloss.Color(""))
	browserList.Styles.StatusEmpty = lipgloss.NewStyle().
		Foreground(lipgloss.Color("")).
		Background(lipgloss.Color("")).
		Width(0).
		Height(0)

	// Filter styles
	browserList.Styles.StatusBarActiveFilter = lipgloss.NewStyle().
		Foreground(purpleColor).
		Bold(true).
		Background(lipgloss.Color(""))
	browserList.Styles.StatusBarFilterCount = lipgloss.NewStyle().
		Foreground(lipgloss.Color("")).
		Background(lipgloss.Color("")).
		Width(0).
		Height(0)

	browserList.Styles.FilterPrompt = lipgloss.NewStyle().
		Foreground(purpleColor).
		Bold(true).
		Background(lipgloss.Color(""))
	browserList.Styles.FilterCursor = lipgloss.NewStyle().
		Foreground(purpleColor)
	browserList.FilterInput.PromptStyle = lipgloss.NewStyle().
		Foreground(purpleColor).
		Bold(true).
		Background(lipgloss.Color(""))

	browserList.Styles.NoItems = lipgloss.NewStyle().
		Foreground(theme.Muted).
		Background(lipgloss.Color(""))
	browserList.Styles.HelpStyle = lipgloss.NewStyle().
		Foreground(theme.Muted).
		Background(lipgloss.Color("")).
		Margin(0).
		Padding(0)

	// Set to show help and enable filtering
	browserList.SetShowHelp(true)
	browserList.SetFilteringEnabled(true)

	// Customize key map - only disable quit keys, keep filter management
	keyMap := list.DefaultKeyMap()
	keyMap.Quit = key.NewBinding()
	keyMap.ForceQuit = key.NewBinding()
	browserList.KeyMap = keyMap

	// Add custom keys to help
	browserList.AdditionalShortHelpKeys = func() []key.Binding {
		return []key.Binding{
			listKeys.enter,
			listKeys.back,
			listKeys.preview,
			listKeys.quit,
		}
	}

	browserList.AdditionalFullHelpKeys = func() []key.Binding {
		return []key.Binding{
			listKeys.toggle,
			listKeys.enter,
			listKeys.back,
			listKeys.choose,
			listKeys.preview,
			listKeys.quit,
		}
	}

	// Initialize preview helper
	previewHelper := NewRulePreviewHelper()

	// Build flat list for filtering
	allRulesItems := buildAllRulesItems(ruleTree, ruleMap)

	// Initialize folder items cache with initial folder items
	folderItems := buildItemsForNode(ruleTree, ruleMap)

	return &fileBrowserModel{
		list:          browserList,
		keys:          listKeys,
		delegateKeys:  delegateKeys,
		previewHelper: previewHelper,
		baseTitle:     title,
		ruleTree:      ruleTree,
		currentNode:   ruleTree,
		ruleMap:       ruleMap,
		allRulesItems: allRulesItems,
		folderItems:   folderItems,
	}
}

// buildItemsForNode creates list items for the current node
func buildItemsForNode(node *domain.RuleNode, ruleMap map[string]*domain.Rule) []list.Item {
	if node.Type != domain.RuleNodeTypeFolder {
		return []list.Item{}
	}

	children := node.GetChildren()
	items := make([]list.Item, 0, len(children))

	for _, child := range children {
		var rule *domain.Rule
		if child.Type == domain.RuleNodeTypeRule {
			rule = ruleMap[child.RuleID]
		}

		items = append(items, &FileBrowserItem{
			node: child,
			rule: rule,
		})
	}

	return items
}

// buildAllRulesItems creates a flat list of all rules for filtering
func buildAllRulesItems(node *domain.RuleNode, ruleMap map[string]*domain.Rule) []list.Item {
	allRules := node.GetAllRules()
	items := make([]list.Item, 0, len(allRules))

	for _, ruleNode := range allRules {
		rule := ruleMap[ruleNode.RuleID]
		if rule != nil {
			items = append(items, &FileBrowserItem{
				node: ruleNode,
				rule: rule,
			})
		}
	}

	return items
}

// fileBrowserKeyMap defines keyboard shortcuts for the file browser
type fileBrowserKeyMap struct {
	toggle  key.Binding
	enter   key.Binding
	back    key.Binding
	choose  key.Binding
	preview key.Binding
	filter  key.Binding
	quit    key.Binding
}

func newFileBrowserKeyMap() *fileBrowserKeyMap {
	return &fileBrowserKeyMap{
		toggle: key.NewBinding(
			key.WithKeys(" "),
			key.WithHelp("space", "toggle rule"),
		),
		enter: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "open folder/confirm"),
		),
		back: key.NewBinding(
			key.WithKeys("backspace"),
			key.WithHelp("backspace", "go back"),
		),
		choose: key.NewBinding(
			key.WithKeys("ctrl+enter"),
			key.WithHelp("ctrl+enter", "confirm selection"),
		),
		preview: key.NewBinding(
			key.WithKeys("p"),
			key.WithHelp("p", "preview rule"),
		),
		filter: key.NewBinding(
			key.WithKeys("/"),
			key.WithHelp("/", "filter rules"),
		),
		quit: key.NewBinding(
			key.WithKeys("q", "esc", "ctrl+c"),
			key.WithHelp("q/esc", "quit"),
		),
	}
}

// fileBrowserDelegateKeyMap defines keyboard shortcuts for individual items
type fileBrowserDelegateKeyMap struct {
	toggle key.Binding
	enter  key.Binding
}

func newFileBrowserDelegateKeyMap() *fileBrowserDelegateKeyMap {
	return &fileBrowserDelegateKeyMap{
		toggle: key.NewBinding(
			key.WithKeys(" "),
			key.WithHelp("space", "toggle"),
		),
		enter: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "open"),
		),
	}
}

// Init initializes the model
func (m *fileBrowserModel) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (m *fileBrowserModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		h, v := appStyle.GetFrameSize()
		// Give the list more height to reduce blank space at bottom
		availableHeight := msg.Height - v
		m.list.SetSize(msg.Width-h, availableHeight)
		if m.showingPreview {
			m.previewHelper.UpdateSize(msg.Width, msg.Height)
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
				cmd := m.previewHelper.Update(msg)
				return m, cmd
			}
		}

		// Simple filtering logic - intercept problematic cases before they reach the list
		filterState := m.list.FilterState()

		// Intercept escape key when in filtering mode - ensure we return to folder view
		if filterState == list.Filtering && key.Matches(msg, m.keys.quit) {
			// First let the list handle the escape to clear its internal filter state
			newListModel, cmd := m.list.Update(msg)
			m.list = newListModel
			cmds = append(cmds, cmd)

			// Then manually transition back to folder view
			m.isFiltering = false
			m.list.SetItems(m.folderItems)
			m.updateTitle()
			return m, tea.Batch(cmds...)
		}

		// For all other filtering cases, let the list handle it
		if filterState == list.Filtering || (filterState == list.FilterApplied && m.list.FilterValue() != "") {
			// Switch to filter mode if not already
			if !m.isFiltering {
				m.isFiltering = true
				m.list.SetItems(m.allRulesItems)
			}

			// If filter is applied with content, intercept our custom keys
			if filterState == list.FilterApplied && m.list.FilterValue() != "" {
				switch {
				case key.Matches(msg, m.keys.toggle):
					selectedItem := m.list.SelectedItem()
					if selectedItem != nil {
						if item, ok := selectedItem.(*FileBrowserItem); ok && item.IsRule() {
							item.selected = !item.selected
							return m, nil
						}
					}
					return m, nil
				case key.Matches(msg, m.keys.enter):
					// In filter mode, add only selected (checked) rules, not the hovered rule
					m.chosen = []string{}
					for _, item := range m.allRulesItems {
						if browserItem, ok := item.(*FileBrowserItem); ok && browserItem.selected && browserItem.IsRule() {
							m.chosen = append(m.chosen, browserItem.node.RuleID)
						}
					}
					return m, tea.Quit
				case key.Matches(msg, m.keys.choose):
					m.chosen = []string{}
					for _, item := range m.allRulesItems {
						if browserItem, ok := item.(*FileBrowserItem); ok && browserItem.selected && browserItem.IsRule() {
							m.chosen = append(m.chosen, browserItem.node.RuleID)
						}
					}
					return m, tea.Quit
				}
			}

			// Pass to list
			newListModel, cmd := m.list.Update(msg)
			m.list = newListModel
			cmds = append(cmds, cmd)
			return m, tea.Batch(cmds...)
		}

		// Handle transition back from empty filter
		if filterState == list.FilterApplied && m.list.FilterValue() == "" && m.isFiltering {
			m.isFiltering = false
			m.list.SetItems(buildItemsForNode(m.currentNode, m.ruleMap))
			m.updateTitle()
			return m, nil
		}

		switch {
		case key.Matches(msg, m.keys.preview):
			selectedItem := m.list.SelectedItem()
			if selectedItem != nil {
				if item, ok := selectedItem.(*FileBrowserItem); ok && item.IsRule() {
					m.showingPreview = true
					m.previewRule = item.rule
					m.previewHelper.SetupPreview(item.rule)
					return m, nil
				}
			}
			return m, nil

		case key.Matches(msg, m.keys.toggle):
			// Toggle rule selection (works in both filtering and folder modes)
			selectedItem := m.list.SelectedItem()
			if selectedItem != nil {
				if item, ok := selectedItem.(*FileBrowserItem); ok && item.IsRule() {
					item.selected = !item.selected
					return m, nil
				}
			}
			return m, nil

		case key.Matches(msg, m.keys.enter):
			selectedItem := m.list.SelectedItem()
			if selectedItem != nil {
				if item, ok := selectedItem.(*FileBrowserItem); ok {
					if item.IsFolder() && !m.isFiltering {
						// Navigate into folder
						m.currentNode = item.node
						m.updateFolderItems()
						m.list.SetItems(m.folderItems)
						m.list.ResetSelected()
						// Update title immediately
						m.updateTitle()
						return m, nil
					} else if item.IsRule() {
						// For rules - confirm selection and add only selected (checked) rules
						m.chosen = []string{}

						var itemsToCheck []list.Item
						if m.isFiltering {
							itemsToCheck = m.allRulesItems
						} else {
							itemsToCheck = m.list.Items()
						}

						for _, listItem := range itemsToCheck {
							if browserItem, ok := listItem.(*FileBrowserItem); ok && browserItem.selected && browserItem.IsRule() {
								m.chosen = append(m.chosen, browserItem.node.RuleID)
							}
						}

						return m, tea.Quit
					}
				}
			}
			return m, nil

		case key.Matches(msg, m.keys.back):
			if !m.isFiltering && m.currentNode.Path != "" {
				// Navigate up one level
				parentPath := m.currentNode.GetParentPath()
				parentNode := m.ruleTree.FindNodeByPath(parentPath)
				if parentNode != nil {
					m.currentNode = parentNode
					m.updateFolderItems()
					m.list.SetItems(m.folderItems)
					m.list.ResetSelected()
					// Update title immediately
					m.updateTitle()
					return m, nil
				}
			}
			return m, nil

		case key.Matches(msg, m.keys.choose):
			// Collect selected rules
			m.chosen = []string{}

			var itemsToCheck []list.Item
			if m.isFiltering {
				itemsToCheck = m.allRulesItems
			} else {
				itemsToCheck = m.list.Items()
			}

			for _, item := range itemsToCheck {
				if browserItem, ok := item.(*FileBrowserItem); ok && browserItem.selected && browserItem.IsRule() {
					m.chosen = append(m.chosen, browserItem.node.RuleID)
				}
			}
			return m, tea.Quit

		case key.Matches(msg, m.keys.quit):
			m.quitting = true
			return m, tea.Quit
		}
	}

	// Update the list
	newListModel, cmd := m.list.Update(msg)
	m.list = newListModel
	cmds = append(cmds, cmd)

	// Update title with breadcrumb
	m.updateTitle()

	return m, tea.Batch(cmds...)
}

// View renders the model
func (m *fileBrowserModel) View() string {
	if m.quitting {
		return ""
	}

	// Count selected items
	selectedCount := 0
	var itemsToCheck []list.Item
	if m.isFiltering {
		itemsToCheck = m.allRulesItems
	} else {
		itemsToCheck = m.list.Items()
	}

	for _, item := range itemsToCheck {
		if browserItem, ok := item.(*FileBrowserItem); ok && browserItem.selected && browserItem.IsRule() {
			selectedCount++
		}
	}

	// Render list
	listView := m.list.View()

	// Add filter indicator if active
	filterIndicator := buildFilterIndicator(&m.list)
	if filterIndicator != "" {
		listView = listView + "\n" + filterIndicator
	}

	// Add selected count if any items are selected
	if selectedCount > 0 {
		countText := fmt.Sprintf("%d selected", selectedCount)
		statusMsg := statusMessageStyle(countText)
		listView = statusMsg + "\n" + listView
	}

	if m.showingPreview {
		return m.previewHelper.RenderOverlay(listView)
	}

	return appStyle.Render(listView)
}

// updateFolderItems updates the cached folder items for the current node
func (m *fileBrowserModel) updateFolderItems() {
	m.folderItems = buildItemsForNode(m.currentNode, m.ruleMap)
}

// updateTitle updates the list title with current path breadcrumb
func (m *fileBrowserModel) updateTitle() {
	// Show path on separate line below title with slashes
	if m.currentNode.Path == "" {
		// At root
		m.list.Title = m.baseTitle + "\n/"
	} else {
		// Show actual path
		m.list.Title = m.baseTitle + "\n/" + m.currentNode.Path
	}
}

// fileBrowserItemDelegate implements list.ItemDelegate for file browser items
type fileBrowserItemDelegate struct {
	keys   *fileBrowserDelegateKeyMap
	theme  ui.Theme
	styles ruleItemStyles
}

func newFileBrowserItemDelegate(keys *fileBrowserDelegateKeyMap) *fileBrowserItemDelegate {
	return &fileBrowserItemDelegate{
		keys:   keys,
		theme:  ui.DefaultTheme(),
		styles: createRuleItemStyles(),
	}
}

func (d *fileBrowserItemDelegate) Height() int {
	return 5 // Slightly smaller to fit more items and reduce blank space
}

func (d *fileBrowserItemDelegate) Spacing() int {
	return 1 // Match spacing used in rm/ls commands
}

func (d *fileBrowserItemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd {
	return nil
}

func (d *fileBrowserItemDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	browserItem, ok := item.(*FileBrowserItem)
	if !ok {
		return
	}

	isSelected := index == m.Index()
	isFiltered := m.FilterState() == list.Filtering || m.FilterState() == list.FilterApplied
	emptyFilter := m.FilterState() == list.Filtering && m.FilterValue() == ""
	filterValue := ""
	if isFiltered && !emptyFilter {
		filterValue = m.FilterValue()
	}

	if browserItem.IsFolder() {
		// Render folders simply with consistent styling
		folderName := "📁 " + browserItem.node.Name

		var folderStyle lipgloss.Style
		switch {
		case isSelected:
			// Hovered folder in pink
			folderStyle = lipgloss.NewStyle().
				Foreground(primaryPink).
				Bold(true).
				Padding(0, 0, 0, 2)
		case emptyFilter:
			folderStyle = lipgloss.NewStyle().
				Foreground(d.theme.Muted).
				Padding(0, 0, 0, 2).
				Faint(true)
		default:
			// Default folders in purple
			folderStyle = lipgloss.NewStyle().
				Foreground(secondaryPurple).
				Bold(true).
				Padding(0, 0, 0, 2)
		}

		_, _ = fmt.Fprint(w, folderStyle.Render(folderName))
		return
	}

	// For rules, use the exact same format as rule selector
	rule := browserItem.rule
	if rule == nil {
		return
	}

	title := rule.Title
	desc := rule.Description

	// Extract rule display path from ID (includes source for custom rules)
	rulePath := domain.ExtractRuleDisplayPath(rule.ID)

	// Build metadata lines (same as rule selector)
	basicMetadataLine, triggerLine, variablesLine := buildRuleMetadata(rule)

	if m.Width() <= 0 {
		return
	}

	// Define colors for explicit styling (same as rule selector)
	selectedColor := lipgloss.Color("#EE6FF8")                               // Bright pink for selected titles
	selectedDescColor := lipgloss.Color("#C084FC")                           // More muted purple-pink for descriptions
	selectedMetaColor := lipgloss.Color("#9CA3AF")                           // Much more muted grey-pink for metadata
	borderColor := lipgloss.AdaptiveColor{Light: "#F793FF", Dark: "#AD58B4"} // Pink border colors
	theme := d.theme

	// Get checkbox content with appropriate colors based on state (same as rule selector)
	checkbox := "[ ]"
	var checkboxColor lipgloss.TerminalColor

	// Always show checkmark if the item is selected, regardless of hover state
	if browserItem.selected {
		checkbox = "[✓]"
	}

	switch {
	case isSelected:
		// Pink when highlighted/hovered (always pink when hovering)
		checkboxColor = selectedColor // Pink for highlighted
	case browserItem.selected:
		// Green when selected but not highlighted/hovered
		checkboxColor = d.theme.Success // Green for selected
	default:
		// Gray for normal state
		checkboxColor = d.theme.Muted // Gray for unselected
	}

	// Build the content lines for rendering (same as rule selector)
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

	// Apply styles and highlighting with separate checkbox handling (same as rule selector)
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
		// Normal and filtered styling (same as rule selector)

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
					styledLines = append(styledLines, applyHighlightsGeneric(line, filterValue, pathStyle, d.styles.matchHighlight))
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
						highlightedTitle := applyTitleHighlightingGeneric(line, filterValue, selectedColor)

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
						titleStyle := lipgloss.NewStyle().
							Foreground(filterTitleColor).
							Bold(true).
							Padding(0, 0, 0, 2)

						checkboxAndTitle := checkboxStyled + " " + line
						// Use pink highlighting for non-selected items
						pinkHighlight := lipgloss.NewStyle().Foreground(selectedColor).Bold(true)
						styledLines = append(styledLines, applyHighlightsGeneric(checkboxAndTitle, filterValue, titleStyle, pinkHighlight))
					} else {
						titleStyle := d.styles.normalTitle
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
					styledLines = append(styledLines, applyHighlightsGeneric(line, filterValue, descStyle, d.styles.matchHighlight))
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
					styledLines = append(styledLines, applyHighlightsGeneric(line, filterValue, metadataStyle, d.styles.matchHighlight))
				} else {
					styledLines = append(styledLines, metadataStyle.Render(line))
				}
			}
		}
	}

	// Output the final result (no background styling, borders handled by individual line styles)
	_, _ = fmt.Fprint(w, strings.Join(styledLines, "\n"))
}
