# TUI Package

The `tui` package provides terminal user interface components for interactive command-line experiences. It uses the `charmbracelet` UI library ecosystem to implement UI elements like rule selectors, file browsers, and prompts.

## Key Components & Features

- **Rule Selector**: A multi-selection interface for choosing rules, with a real-time preview panel. Supports keyboard navigation and clear visual feedback for selections.
- **File Browser**: A component for navigating and selecting files from the filesystem.
- **Prompt System**: A system for creating configurable user input prompts with validation.
- **Visual Rendering**: Provides rich formatting for rule previews, including syntax highlighting and metadata. It uses a shared styling system with adaptive colors for light and dark terminal themes.
- **Layouts**: Responsive layouts that adapt to different terminal sizes.

## Integration

The TUI components integrate with other internal packages:
- **`ui` package**: For a consistent theme, styles, and adaptive colors across all components.
- **`commands` package**: The interactive commands use TUI components for user selections and input.

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

This package is built on the charmbracelet ecosystem:
- **`huh`**: For form and prompt components.
- **`lipgloss`**: For styling and layout.

## API

The package exposes the following main components:

- `RuleSelector`: A multi-selection component with preview capabilities.
- `FileBrowser`: A file system navigation and selection component.
- `SelectOptions`: A configurable selection prompt with validation.

It also provides rendering and styling utilities.

## Error Handling

- `HandleFormError(err)`: A function to convert `charmbracelet/huh` errors into user-friendly messages.
- `ErrUserCancelled`: A standard error returned when the user cancels an operation (e.g., by pressing ESC).