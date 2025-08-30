# CLI Package

This package provides utilities for the command-line interface, including custom help formatting, command rendering, and colored output. It enhances the `urfave/cli` library to improve the user experience.

## Features

- **Custom Help Templates**: Provides enhanced help formatting with better organization and readability.
- **Color Support**: Uses terminal-aware colored output to improve visual hierarchy.
- **Consistent Theming**: Integrates with the internal `ui` package for consistent styling.
- **Enhanced Formatting**: Improves the presentation of commands and options.

### Help System Architecture

```mermaid
graph TB
    subgraph "CLI Framework"
        URFAVE[urfave/cli]
        COMMAND[CLI Commands]
        FLAGS[Global Flags]
    end
    
    subgraph "CLI Package"
        HELPPRINTER[HelpPrinter]
        TEMPLATES[Help Templates]
        FORMATTER[Text Formatter]
        COLORIZER[Color Support]
    end
    
    subgraph "UI Integration"
        THEME[UI Theme System]
        TERMINAL[Terminal Detection]
        STYLES[Text Styling]
    end
    
    subgraph "Output"
        STDOUT[Formatted Help Output]
        COLORS[Colored Text]
        LAYOUT[Structured Layout]
    end
    
    URFAVE --> HELPPRINTER
    COMMAND --> HELPPRINTER
    FLAGS --> HELPPRINTER
    
    HELPPRINTER --> TEMPLATES
    HELPPRINTER --> FORMATTER
    HELPPRINTER --> COLORIZER
    
    TEMPLATES --> THEME
    FORMATTER --> STYLES
    COLORIZER --> TERMINAL
    
    THEME --> STDOUT
    STYLES --> LAYOUT
    TERMINAL --> COLORS
    
    STDOUT --> LAYOUT
    COLORS --> LAYOUT
    
    style HELPPRINTER fill:#e1f5fe
    style TEMPLATES fill:#f3e5f5
    style FORMATTER fill:#e8f5e8
    style THEME fill:#fff3e0
```

### Help Rendering Pipeline

```mermaid
sequenceDiagram
    participant User as User
    participant CLI as CLI Framework
    participant Help as HelpPrinter
    participant Template as Template Engine
    participant Theme as UI Theme
    participant Term as Terminal
    
    User->>CLI: --help flag
    CLI->>Help: Print(writer, template, data)
    
    Help->>Term: Check terminal capabilities
    Term-->>Help: Color support status
    
    Help->>Template: Render help template
    Template->>Template: Process template data
    Template->>Template: Apply formatting rules
    Template-->>Help: Rendered content
    
    Help->>Theme: Apply color styling
    Theme->>Theme: Select adaptive colors
    Theme-->>Help: Styled content
    
    Help->>Help: Format layout structure
    Help->>Help: Add visual hierarchy
    
    Help->>Term: Write formatted output
    Term-->>User: Enhanced help display
    
    note over Help: Features:<br/>• Custom templates<br/>• Color detection<br/>• Structured layout<br/>• Visual hierarchy
```

### Template System Integration

```mermaid
flowchart TD
    START([Help Request]) --> DETECT[Detect Terminal Capabilities]
    
    DETECT --> LOADTEMPLATE[Load Help Template]
    LOADTEMPLATE --> PROCESSDATA[Process Command Data]
    
    PROCESSDATA --> APPLYTEMPLATE[Apply Template Formatting]
    APPLYTEMPLATE --> ADDCOLORS{Colors Supported?}
    
    ADDCOLORS -->|Yes| COLORIZE[Apply Color Styling]
    ADDCOLORS -->|No| PLAINTEXT[Use Plain Text]
    
    COLORIZE --> LAYOUT[Apply Layout Structure]
    PLAINTEXT --> LAYOUT
    
    LAYOUT --> HIERARCHY[Add Visual Hierarchy]
    HIERARCHY --> OUTPUT[Generate Final Output]
    
    OUTPUT --> SUCCESS([Display Help])
    
    style DETECT fill:#e1f5fe
    style APPLYTEMPLATE fill:#f3e5f5
    style COLORIZE fill:#e8f5e8
    style LAYOUT fill:#fff3e0
    style SUCCESS fill:#c8e6c9
```

## Usage

This package is used by the `app` package to configure the CLI application's help rendering and user interaction elements.

## API

- `NewHelpPrinter() -> HelpPrinter`: Creates a new help printer with enhanced formatting.
- `Print(writer, template, data)`: Renders help content using a custom template and formatting.
- `AppHelpTemplate`: A custom template for the application-level help display.