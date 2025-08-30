# TUI Package

This package provides terminal user interface components for interactive user experiences in Contexture, implementing rich UI elements like rule selectors, file browsers, and prompts using the charmbracelet UI library ecosystem.

## Purpose

The tui package bridges the gap between the command-line interface and user-friendly interactive experiences. It provides sophisticated terminal UI components that make rule selection, file browsing, and user input both intuitive and visually appealing.

## Key Components

### Interactive Selectors
- **Rule Selector**: Multi-selection interface for choosing rules with preview capabilities
- **File Browser**: Navigate and select files from the filesystem with visual hierarchy
- **Prompt System**: Configurable prompts for user input with validation and error handling

### Visual Rendering
- **Rule Preview**: Rich formatting for rule content with syntax highlighting and metadata
- **Shared Styling**: Consistent visual styling across all TUI components
- **Adaptive Colors**: Theme-aware colors that adapt to light/dark terminal environments

## User Experience Features

- **Real-Time Preview**: Live preview of rules during selection process
- **Keyboard Navigation**: Full keyboard support with intuitive navigation patterns
- **Visual Feedback**: Clear indication of selection state and user actions
- **Error Handling**: Graceful handling of user cancellation and input errors
- **Responsive Layout**: Adaptive layouts that work across different terminal sizes

## Integration with UI System

- **Theme Integration**: Seamless integration with the internal UI theme system
- **Consistent Styling**: Unified visual language across all interactive components  
- **Icon Usage**: Contextual icons and visual indicators for enhanced usability
- **Color Coordination**: Coordinated color usage for status, selection, and emphasis

### TUI Component Architecture

```mermaid
graph TB
    subgraph "TUI Package"
        RULESELECTOR[Rule Selector]
        FILEBROWSER[File Browser]
        PROMPT[Prompt System]
        SHARED[Shared Styling]
        FORMATTER[Rule Formatter]
    end
    
    subgraph "Charmbracelet Libraries"
        HUH[huh Forms]
        LIPGLOSS[lipgloss Styling]
        BUBBLETEA[Bubble Tea Framework]
    end
    
    subgraph "UI Integration"
        THEME[UI Theme System]
        STYLES[UI Styles]
        COLORS[Adaptive Colors]
    end
    
    subgraph "Domain Integration"
        RULES[Rule Entities]
        CONFIG[Configuration]
        VALIDATION[Input Validation]
    end
    
    subgraph "Client Commands"
        ADD[Add Command]
        REMOVE[Remove Command]
        LIST[List Command]
        UPDATE[Update Command]
        INIT[Init Command]
    end
    
    RULESELECTOR --> HUH
    FILEBROWSER --> HUH
    PROMPT --> HUH
    
    SHARED --> LIPGLOSS
    FORMATTER --> LIPGLOSS
    RULESELECTOR --> LIPGLOSS
    
    SHARED --> THEME
    FORMATTER --> STYLES
    RULESELECTOR --> COLORS
    
    RULESELECTOR --> RULES
    FILEBROWSER --> CONFIG
    PROMPT --> VALIDATION
    
    ADD --> RULESELECTOR
    ADD --> FILEBROWSER
    REMOVE --> RULESELECTOR
    LIST --> RULESELECTOR
    UPDATE --> RULESELECTOR
    INIT --> PROMPT
    
    style RULESELECTOR fill:#e1f5fe
    style SHARED fill:#f3e5f5
    style HUH fill:#e8f5e8
    style THEME fill:#fff3e0
```

### Interactive Selection Flow

```mermaid
sequenceDiagram
    participant Cmd as Command
    participant TUI as TUI Component
    participant Huh as Huh Forms
    participant User as User
    participant Theme as UI Theme
    participant Rules as Rule Data
    
    Cmd->>TUI: ShowRuleSelector(rules)
    TUI->>Theme: Get theme colors
    Theme-->>TUI: Adaptive colors
    
    TUI->>Rules: Format rule display
    Rules-->>TUI: Formatted rule list
    
    TUI->>Huh: Create multi-select form
    Huh->>User: Display interactive list
    
    User->>Huh: Navigate with arrow keys
    User->>Huh: Toggle selection with space
    User->>Huh: Preview rule with enter
    
    TUI->>TUI: Render rule preview
    TUI->>User: Show formatted preview
    
    User->>Huh: Confirm selection
    Huh-->>TUI: Selected rules
    
    alt User cancels
        User->>Huh: Press ESC
        Huh->>TUI: ErrUserAborted
        TUI-->>Cmd: ErrUserCancelled
    else User confirms
        TUI-->>Cmd: Selected rule list
    end
    
    note over TUI: Features:<br/>• Live preview<br/>• Multi-selection<br/>• Keyboard navigation<br/>• Theme integration
```

### Component Integration Pattern

```mermaid
flowchart TD
    START([User Interaction Needed]) --> COMPONENT{Component Type?}
    
    COMPONENT -->|Rule Selection| RULESELECTOR[Rule Selector]
    COMPONENT -->|File Navigation| FILEBROWSER[File Browser] 
    COMPONENT -->|Input Prompt| PROMPT[Prompt Form]
    
    RULESELECTOR --> LOADRULES[Load Available Rules]
    FILEBROWSER --> LOADFILES[Load File System]
    PROMPT --> VALIDATION[Setup Validation]
    
    LOADRULES --> THEMING[Apply UI Theming]
    LOADFILES --> THEMING
    VALIDATION --> THEMING
    
    THEMING --> RENDER[Render Interactive UI]
    RENDER --> USERINTERACTION[User Interaction Loop]
    
    USERINTERACTION --> KEYPRESS{Key Press Event}
    
    KEYPRESS -->|Navigation| NAVIGATE[Update Selection]
    KEYPRESS -->|Toggle| TOGGLE[Toggle Selection]
    KEYPRESS -->|Preview| PREVIEW[Show Preview]
    KEYPRESS -->|Submit| SUBMIT[Confirm Selection]
    KEYPRESS -->|Cancel| CANCEL[User Cancellation]
    
    NAVIGATE --> USERINTERACTION
    TOGGLE --> USERINTERACTION
    PREVIEW --> SHOWPREVIEW[Display Rule Preview]
    SHOWPREVIEW --> USERINTERACTION
    
    SUBMIT --> SUCCESS[Return Selection]
    CANCEL --> CANCELLED[Return Cancellation]
    
    style COMPONENT fill:#e1f5fe
    style THEMING fill:#f3e5f5
    style USERINTERACTION fill:#e8f5e8
    style SUCCESS fill:#c8e6c9
    style CANCELLED fill:#ffcdd2
```

## Charmbracelet Integration

Built on the charmbracelet ecosystem:
- **huh**: Form and prompt components for user interaction
- **lipgloss**: Styling and layout system for terminal rendering
- **Shared Patterns**: Common patterns for terminal UI development

## Usage Within Project

This package is used by:
- **Commands Package**: Interactive commands use TUI components for user selection and input
- **Rule Selection**: Commands that require rule selection use the rule selector component
- **Configuration**: Interactive configuration commands use prompts and selectors

## API

### Selection Components
- `RuleSelector`: Multi-selection interface with preview capabilities
- `FileBrowser`: File system navigation with selection support
- `SelectOptions`: Configurable selection prompts with validation

### Rendering Utilities
- Shared styling functions for consistent visual presentation
- Rule rendering utilities with formatting and metadata display
- Color and theme management for adaptive terminal environments

### Error Handling
- `HandleFormError(err)`: Converts library errors to user-friendly messages
- `ErrUserCancelled`: Standard error for user-initiated cancellation
- Graceful degradation for unsupported terminal features