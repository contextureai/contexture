# CLI Package

This package provides command-line interface utilities for the Contexture application, including enhanced help formatting, command rendering, and user interaction components with color support and improved usability.

## Purpose

The cli package enhances the standard urfave/cli experience by providing customized help templates, colored output, and improved formatting for better user experience. It bridges the gap between the CLI framework and user-facing presentation.

## Key Features

- **Custom Help Templates**: Enhanced help formatting with better organization and readability
- **Color Support**: Terminal-aware colored output for improved visual hierarchy
- **Consistent Theming**: Integration with internal UI components for unified styling
- **Enhanced Formatting**: Improved command and option presentation
- **User Experience**: Focus on clarity and ease of use for CLI interactions

## Help System Enhancement

- **Template Customization**: Custom help templates that improve upon urfave/cli defaults
- **Structured Layout**: Better organization of commands, flags, and descriptions
- **Visual Hierarchy**: Use of colors and formatting to guide user attention
- **Contextual Help**: Relevant help information displayed at the right time

## Integration Points

- **urfave/cli Framework**: Seamless integration with the CLI framework
- **Terminal Detection**: Automatic color support detection based on terminal capabilities
- **Theme Consistency**: Coordinated styling across all user-facing components

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

## Usage Within Project

This package is used by:
- **App Package**: Main application uses CLI utilities for help rendering and user interaction
- **Command Framework**: All commands benefit from enhanced help and formatting capabilities

## API

- `NewHelpPrinter()`: Creates a new help printer with enhanced formatting
- `Print(writer, template, data)`: Renders help content with custom formatting
- `AppHelpTemplate`: Custom template for application-level help display
- Color and formatting utilities for consistent CLI presentation