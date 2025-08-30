# UI Package

This package provides a centralized theming and styling system for all user interface components in the application. It uses the `lipgloss` framework to create a consistent visual design with adaptive colors.

## Theming System

- **Adaptive Colors**: Automatically detects and adjusts for light/dark terminal themes.
- **Semantic Colors**: Maps colors to meanings such as `success`, `error`, `warning`, and `info`.
- **Consistent Palette**: Uses a coordinated color scheme for a consistent appearance.
- **Terminal Compatibility**: Provides graceful fallbacks for terminals with limited color support.

## Component Library

The package includes a library of pre-styled components:

- **Text Components**: `Headers`, `Status Messages` (Success, Warning, Error, Info), and `Body Text`.
- **Layout Components**: `Cards` for bordered content, `Dividers`, and `Sidebars`.
- **Utility Components**: `Loading Indicators`, `Banners`, and `Status Indicators`.

## Icon System

A consistent set of icons is used across all components for statuses and navigation, including:
- **Status Icons**: `✓` (success), `✗` (error), `⚠` (warning), `ⓘ` (info)
- **Navigation Icons**: `▶` (expand), `◀` (collapse)

### UI Component Hierarchy

```mermaid
graph TB
    subgraph "Theme System"
        THEME[Default Theme]
        STYLES[Styles Instance]
        ADAPTIVE[Adaptive Colors]
        LIPGLOSS[Lipgloss Framework]
    end
    
    subgraph "Text Components"
        HEADER[Header Component]
        SUCCESS[Success Message]
        ERROR[Error Message]
        WARNING[Warning Message]
        INFO[Info Message]
        MUTED[Muted Text]
    end
    
    subgraph "Layout Components"
        CARD[Card Container]
        DIVIDER[Divider Separator]
        SIDEBAR[Sidebar Panel]
        BANNER[Banner Message]
    end
    
    subgraph "Interactive Components"
        MENU[Menu Interface]
        PROGRESS[Progress Bar]
        NOTIFICATION[Notification]
        LOADING[Loading Indicator]
        STATUS[Status Indicator]
    end
    
    subgraph "Client Integration"
        TUI[TUI Package]
        CLI[CLI Package]
        COMMANDS[Commands Package]
        FORMAT[Format Package]
    end
    
    THEME --> STYLES
    STYLES --> ADAPTIVE
    ADAPTIVE --> LIPGLOSS
    
    STYLES --> HEADER
    STYLES --> SUCCESS
    STYLES --> ERROR
    STYLES --> WARNING
    STYLES --> INFO
    STYLES --> MUTED
    
    STYLES --> CARD
    STYLES --> DIVIDER
    STYLES --> SIDEBAR
    STYLES --> BANNER
    
    STYLES --> MENU
    STYLES --> PROGRESS
    STYLES --> NOTIFICATION
    STYLES --> LOADING
    STYLES --> STATUS
    
    TUI --> HEADER
    TUI --> SUCCESS
    TUI --> CARD
    TUI --> MENU
    
    CLI --> SUCCESS
    CLI --> ERROR
    CLI --> WARNING
    CLI --> INFO
    
    COMMANDS --> SUCCESS
    COMMANDS --> ERROR
    COMMANDS --> PROGRESS
    COMMANDS --> STATUS
    
    FORMAT --> BANNER
    FORMAT --> DIVIDER
    
    style THEME fill:#e1f5fe
    style STYLES fill:#f3e5f5
    style LIPGLOSS fill:#e8f5e8
    style TUI fill:#fff3e0
```

### Theme Application Flow

```mermaid
sequenceDiagram
    participant Client as Client Code
    participant Theme as Theme System
    participant Styles as Styles Instance
    participant Lipgloss as Lipgloss
    participant Terminal as Terminal
    
    Client->>Theme: DefaultTheme()
    Theme-->>Client: Theme with adaptive colors
    
    Client->>Styles: NewStyles(theme)
    Styles-->>Client: Styles instance
    
    Client->>Styles: Header("Section Title")
    
    Styles->>Theme: Get primary color
    Theme->>Terminal: Detect terminal capabilities
    Terminal-->>Theme: Color support info
    Theme-->>Styles: Adaptive color value
    
    Styles->>Lipgloss: Create style with color
    Lipgloss->>Lipgloss: Apply formatting (bold, padding)
    Lipgloss->>Lipgloss: Add icon prefix "▶"
    Lipgloss-->>Styles: Styled component
    
    Styles-->>Client: Rendered header text
    
    note over Theme: Adaptive colors automatically<br/>adjust for light/dark terminals
```

### Component Styling System

```mermaid
flowchart TD
    START([Component Request]) --> GETTHEME[Get Default Theme]
    
    GETTHEME --> CREATESTYLES[Create Styles Instance]
    CREATESTYLES --> COMPONENT{Component Type?}
    
    COMPONENT -->|Header| HEADER_STYLE[Apply Header Style]
    COMPONENT -->|Status| STATUS_STYLE[Apply Status Style]
    COMPONENT -->|Card| CARD_STYLE[Apply Card Style]
    COMPONENT -->|Menu| MENU_STYLE[Apply Menu Style]
    
    HEADER_STYLE --> HEADER_CONFIG[Bold + Primary Color + Icon]
    STATUS_STYLE --> STATUS_CONFIG[Semantic Color + Icon]
    CARD_STYLE --> CARD_CONFIG[Border + Padding + Background]
    MENU_STYLE --> MENU_CONFIG[Border + Selection Colors]
    
    HEADER_CONFIG --> LIPGLOSS_RENDER[Lipgloss Render]
    STATUS_CONFIG --> LIPGLOSS_RENDER
    CARD_CONFIG --> LIPGLOSS_RENDER
    MENU_CONFIG --> LIPGLOSS_RENDER
    
    LIPGLOSS_RENDER --> TERMINAL_CHECK{Terminal Capabilities?}
    
    TERMINAL_CHECK -->|Color Support| COLORED[Apply Colors]
    TERMINAL_CHECK -->|No Color| PLAIN[Plain Text]
    
    COLORED --> OUTPUT[Styled Component]
    PLAIN --> OUTPUT
    
    OUTPUT --> SUCCESS([Component Ready])
    
    style GETTHEME fill:#e1f5fe
    style COMPONENT fill:#f3e5f5
    style LIPGLOSS_RENDER fill:#e8f5e8
    style SUCCESS fill:#c8e6c9
```

## Lipgloss Integration

This package is built on the `lipgloss` styling framework, using it for:
- **Style Composition**: Creating reusable style definitions.
- **Layout**: Managing padding, margins, and alignment.
- **Color Management**: Handling adaptive colors.

## Usage

This package is used by:
- `tui` package: For theming interactive components.
- `cli` package: For styling help and command output.
- `commands` package: For formatting command output and status messages.

## API

- `DefaultTheme() -> Theme`: Returns the default adaptive theme.
- `NewStyles(theme) -> Styles`: Creates a `Styles` instance with rendering functions based on the provided theme.
- **Styled Text Functions**: `Header(text)`, `Success(text)`, `Error(text)`, etc., which return styled strings with appropriate icons and colors.
- **Component Functions**: Functions for creating card, divider, and other UI components.