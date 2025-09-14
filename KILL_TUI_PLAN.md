# Plan to Remove Full-Screen Interactive TUI Support

## Executive Summary

The current implementation uses Bubble Tea's full-screen mode (`tea.WithAltScreen()`) for browsing and selecting rules. This creates integration testing challenges, complexity in maintenance, and is unnecessary for a CLI tool. This plan outlines a migration to simpler, non-fullscreen alternatives while maintaining a good user experience.

## Current Full-Screen TUI Usage Analysis

### Components Using Full-Screen Mode

1. **Rule Selector (`internal/tui/rule_selector.go`)**
   - `DisplayRules()` - Full-screen rule browser for `contexture rules list`
   - `SelectRules()` - Full-screen multi-select for `contexture rules add`
   - Uses `tea.WithAltScreen()` for both display and selection modes
   - Features: Filtering, preview overlay, keyboard navigation, checkbox selection

2. **File Browser (`internal/tui/file_browser.go`)**
   - `BrowseRules()` - Full-screen folder tree navigation for `contexture rules add`
   - Uses `tea.WithAltScreen()` for hierarchical navigation
   - Features: Folder/file tree, filtering, preview, multi-select

### Commands Affected

1. **`contexture rules list`** (`internal/commands/list.go`)
   - Uses `DisplayRules()` to show installed rules in full-screen mode
   - Could be replaced with simple formatted output

2. **`contexture rules add`** (`internal/commands/add.go`)
   - Uses `BrowseRules()` when no arguments provided (interactive mode)
   - Shows available rules in a full-screen file browser
   - Could use simpler selection methods

3. **`contexture init`** (`internal/commands/init.go`)
   - Already uses non-fullscreen `huh` forms (GOOD EXAMPLE TO FOLLOW)
   - Simple inline prompts that don't take over the terminal

## Problems with Current Approach

1. **Testing Complexity**
   - Full-screen TUI components are difficult to test
   - E2E tests cannot easily verify behavior
   - Integration tests become fragile

2. **User Experience Issues**
   - Takes over entire terminal
   - Can't see command history
   - Disorienting for quick operations
   - Harder to pipe/script

3. **Maintenance Burden**
   - Complex state management
   - Rendering logic for overlays
   - Keyboard event handling
   - Window resize management

## Proposed Alternative Approaches

### Option 1: Inline List with Simple Selection (Recommended)

Use Bubble Tea's list component without `WithAltScreen()`:

**For `contexture rules list`:**
```go
// Simple formatted output with optional pager
func DisplayRules(rules []*domain.Rule) error {
    // Use lipgloss for styling
    // Output formatted list
    // Optional: Use bubbles/paginator for long lists
    // Optional: Use less/more for paging
}
```

**Benefits:**
- Stays in terminal flow
- Can pipe output
- Scriptable
- Simple to test

**Example Output:**
```
Installed Rules (12 total):

📁 languages/
  📄 go/testing - Go Testing Best Practices
     Tags: go, testing, unit-tests
     Trigger: **/*_test.go
  
  📄 go/errors - Error Handling Patterns
     Tags: go, errors, best-practices

📁 security/
  📄 input-validation - Validate User Inputs
     Tags: security, validation
     Variables: strict_mode=true

Use 'contexture rules show <rule-id>' for details
```

### Option 2: Interactive Prompts with huh (Like init command)

**For `contexture rules add`:**
```go
// Use huh for interactive selection
func SelectRulesToAdd(rules []*domain.Rule) ([]string, error) {
    // Group rules by category
    // Use huh.NewMultiSelect for each category
    // Or use sequential prompts for navigation
}
```

**Benefits:**
- Consistent with init command
- Simple inline interaction
- Easy to test
- Accessible

### Option 3: Hybrid with Optional Table View

Use `bubbles/table` without full-screen for data-heavy views:

```go
// Non-fullscreen table for rule browsing
func ShowRuleTable(rules []*domain.Rule) error {
    // Create table model
    // Run without WithAltScreen()
    // Allow filtering and selection
}
```

**Benefits:**
- Rich data display
- Filtering capability
- Stays inline
- Better for many rules

## Implementation Strategy

### Phase 1: Create New Display Components

1. **Create `internal/ui/rules/` package**
   - `display.go` - Formatted rule display
   - `select.go` - Rule selection prompts
   - `table.go` - Optional table view
   - `format.go` - Shared formatting utilities

2. **Implement display alternatives**
   ```go
   // Simple display with categorization
   func DisplayRuleList(rules []*domain.Rule, options DisplayOptions) error
   
   // Interactive selection using huh
   func SelectRulesInteractive(rules []*domain.Rule) ([]string, error)
   
   // Compact table view (optional)
   func ShowRuleTable(rules []*domain.Rule) error
   ```

3. **Add formatting utilities**
   - Rule path extraction and display
   - Metadata formatting (tags, triggers, variables)
   - Tree structure rendering (for categories)
   - Color coding by status/type

### Phase 2: Update Commands

1. **Update `list` command**
   - Replace `DisplayRules()` with `DisplayRuleList()`
   - Add `--format` flag (list, table, json, yaml)
   - Add `--output` flag for file output
   - Keep filtering in command logic

2. **Update `add` command**
   - Replace `BrowseRules()` with `SelectRulesInteractive()`
   - Use categorized multi-select
   - Show preview inline when selecting
   - Add `--no-interactive` flag for scripting

### Phase 3: Remove Old TUI Components

1. **Deprecate full-screen components**
   - Mark as deprecated in v1.x
   - Remove in v2.0

2. **Clean up dependencies**
   - Remove complex TUI models
   - Simplify event handling
   - Remove preview overlays

### Phase 4: Improve Testing

1. **Add comprehensive tests**
   - Unit tests for display formatting
   - Integration tests for selection
   - E2E tests for complete workflows

2. **Test scenarios**
   - Empty rule lists
   - Large rule lists (100+ rules)
   - Filtered views
   - Error cases

## Example Implementations

### Simple Rule Display
```go
func DisplayRuleList(rules []*domain.Rule, options DisplayOptions) error {
    if len(rules) == 0 {
        fmt.Println("No rules found.")
        return nil
    }

    // Group by category
    categories := groupRulesByCategory(rules)
    
    // Create styles
    headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#EE6FF8"))
    pathStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#666"))
    titleStyle := lipgloss.NewStyle().Bold(true)
    tagStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#C084FC"))
    
    fmt.Printf("%s (%d total)\n\n", headerStyle.Render("Installed Rules"), len(rules))
    
    for category, categoryRules := range categories {
        fmt.Printf("📁 %s\n", category)
        for _, rule := range categoryRules {
            fmt.Printf("  📄 %s - %s\n", 
                pathStyle.Render(rule.Path),
                titleStyle.Render(rule.Title))
            
            if rule.Description != "" {
                fmt.Printf("     %s\n", rule.Description)
            }
            
            if len(rule.Tags) > 0 {
                fmt.Printf("     %s: %s\n", 
                    tagStyle.Render("Tags"),
                    strings.Join(rule.Tags, ", "))
            }
            
            if options.ShowTriggers && rule.Trigger != nil {
                fmt.Printf("     Trigger: %s\n", formatTrigger(rule.Trigger))
            }
            
            fmt.Println()
        }
    }
    
    return nil
}
```

### Interactive Selection with huh
```go
func SelectRulesInteractive(availableRules []*domain.Rule) ([]string, error) {
    // Group rules by category for better organization
    categories := groupRulesByCategory(availableRules)
    
    var selectedRules []string
    
    for category, rules := range categories {
        var categorySelections []string
        
        // Create options for this category
        options := make([]huh.Option[string], len(rules))
        for i, rule := range rules {
            label := fmt.Sprintf("%s - %s", rule.Path, rule.Title)
            options[i] = huh.NewOption(label, rule.ID)
        }
        
        // Create form for this category
        form := huh.NewForm(
            huh.NewGroup(
                huh.NewMultiSelect[string]().
                    Title(fmt.Sprintf("Select rules from %s", category)).
                    Options(options...).
                    Value(&categorySelections),
            ),
        )
        
        if err := form.Run(); err != nil {
            return nil, err
        }
        
        selectedRules = append(selectedRules, categorySelections...)
    }
    
    return selectedRules, nil
}
```

### Compact Table View (Optional)
```go
func ShowRuleTable(rules []*domain.Rule) error {
    // Use bubbles/table WITHOUT WithAltScreen
    columns := []table.Column{
        {Title: "Rule", Width: 30},
        {Title: "Description", Width: 40},
        {Title: "Tags", Width: 20},
        {Title: "Trigger", Width: 15},
    }
    
    rows := make([]table.Row, len(rules))
    for i, rule := range rules {
        rows[i] = table.Row{
            rule.Path,
            truncate(rule.Description, 40),
            strings.Join(rule.Tags, ", "),
            formatTriggerShort(rule.Trigger),
        }
    }
    
    t := table.New(
        table.WithColumns(columns),
        table.WithRows(rows),
        table.WithFocused(true),
        table.WithHeight(min(len(rows)+1, 20)),
    )
    
    // Run inline without taking over screen
    p := tea.NewProgram(t) // Note: NO WithAltScreen()
    _, err := p.Run()
    return err
}
```

## Migration Path for Users

1. **Version 1.x (Deprecation)**
   - Add `--legacy-ui` flag to use old TUI
   - Default to new inline display
   - Show deprecation notice when legacy UI used

2. **Version 2.0 (Removal)**
   - Remove full-screen TUI completely
   - Clean, simple interface only

## Benefits of This Approach

1. **Better Developer Experience**
   - Scriptable and pipeable
   - Easier to test
   - Simpler to debug
   - CI/CD friendly

2. **Improved User Experience**
   - Stays in terminal context
   - Faster for quick operations
   - Consistent with other CLI tools
   - Accessible

3. **Reduced Complexity**
   - Less code to maintain
   - Fewer edge cases
   - Simpler state management
   - Easier contributions

4. **Better Integration**
   - Works with terminal multiplexers
   - Compatible with SSH sessions
   - Plays nice with shell history
   - Supports output redirection

## Timeline

- **Week 1**: Implement new display components
- **Week 2**: Update commands to use new components
- **Week 3**: Add comprehensive tests
- **Week 4**: Deprecate old components, documentation
- **Release**: Version 1.5 with both options
- **3 months later**: Version 2.0 with old TUI removed

## Success Metrics

- Reduction in TUI-related bug reports
- Improved test coverage (>80%)
- Faster CI/CD runs
- Positive user feedback
- Reduced code complexity metrics

## Conclusion

Moving away from full-screen TUI to simpler, inline interactions will make Contexture more maintainable, testable, and user-friendly. The proposed alternatives maintain the benefits of interactive selection while avoiding the complexity and issues of full-screen terminal takeover.