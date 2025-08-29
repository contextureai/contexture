# Commands Package

This package implements all CLI command actions for the Contexture application, providing the core functionality that users interact with through the command-line interface. It serves as the business logic layer between the CLI framework and the underlying domain operations.

## Purpose

The commands package translates user intentions expressed through CLI arguments and flags into concrete operations on projects, rules, and configurations. It orchestrates domain services while providing user feedback and error handling.

## Available Commands

### Project Management
- **init**: Initialize new Contexture projects with default configurations
- **config**: Manage project configuration with subcommands for formats and settings

### Rule Operations
- **add**: Add new rules to projects from various sources (local files, repositories)
- **remove**: Remove existing rules from project configurations
- **list**: Display project rules with filtering and formatting options
- **update**: Update existing rules from their original sources

### Build System
- **build**: Generate output files in configured formats (Claude, Cursor, Windsurf)

## Key Features

- **Unified Error Handling**: Consistent error reporting across all commands
- **Configuration Management**: Loading and validation of project configurations
- **Rule Processing**: Integration with rule fetchers, validators, and processors
- **Format Integration**: Coordination with format implementations for output generation
- **User Interaction**: Terminal UI components for user prompts and selections
- **Shared Utilities**: Common functionality reused across command implementations

## Architecture

- **Command Actions**: Individual command implementations with clear separation of concerns
- **Shared Utilities**: Common functionality for configuration loading, validation, and user interaction
- **Config Loaders**: Specialized utilities for loading and managing project configurations
- **Test Helpers**: Support utilities for comprehensive command testing

### Command Flow Architecture

```mermaid
graph TB
    subgraph "User Interface"
        CLI[CLI Framework]
        TUI[Terminal UI]
    end
    
    subgraph "Commands Package"
        INIT[InitAction]
        ADD[AddAction]
        REMOVE[RemoveAction]
        LIST[ListAction]
        UPDATE[UpdateAction]
        BUILD[BuildAction]
        CONFIG[ConfigAction]
        SHARED[Shared Utilities]
        LOADER[Config Loader]
    end
    
    subgraph "Core Services"
        RULE[Rule Package]
        PROJECT[Project Package]
        FORMAT[Format Package]
        DOMAIN[Domain Package]
        VAL[Validation Package]
        DEPS[Dependencies]
    end
    
    CLI --> INIT
    CLI --> ADD
    CLI --> REMOVE
    CLI --> LIST
    CLI --> UPDATE
    CLI --> BUILD
    CLI --> CONFIG
    
    INIT --> SHARED
    ADD --> SHARED
    REMOVE --> SHARED
    LIST --> SHARED
    UPDATE --> SHARED
    BUILD --> SHARED
    CONFIG --> SHARED
    
    SHARED --> LOADER
    SHARED --> TUI
    LOADER --> PROJECT
    
    ADD --> RULE
    REMOVE --> RULE
    LIST --> RULE
    UPDATE --> RULE
    BUILD --> RULE
    BUILD --> FORMAT
    
    INIT --> PROJECT
    CONFIG --> PROJECT
    
    SHARED --> VAL
    SHARED --> DOMAIN
    
    ALL_COMMANDS -.-> DEPS
    
    style SHARED fill:#e1f5fe
    style LOADER fill:#f3e5f5
    style RULE fill:#e8f5e8
    style PROJECT fill:#fff3e0
```

### Build Command Flow

```mermaid
sequenceDiagram
    participant User
    participant BuildCmd as BuildAction
    participant ConfigLdr as ConfigLoader
    participant RulePkg as Rule Package
    participant FmtPkg as Format Package
    participant TUI as Terminal UI
    
    User->>BuildCmd: contexture build
    
    BuildCmd->>ConfigLdr: LoadConfig()
    ConfigLdr-->>BuildCmd: Project Config
    
    BuildCmd->>RulePkg: FetchRules(ruleRefs)
    
    loop For each rule
        RulePkg->>RulePkg: FetchRule()
        RulePkg->>RulePkg: ProcessRule()
    end
    
    RulePkg-->>BuildCmd: Processed Rules
    
    loop For each format
        BuildCmd->>FmtPkg: CreateFormat(type)
        FmtPkg-->>BuildCmd: Format Instance
        
        BuildCmd->>FmtPkg: Generate(rules)
        FmtPkg-->>BuildCmd: Generated Content
        
        BuildCmd->>BuildCmd: WriteToFile()
    end
    
    BuildCmd->>TUI: ShowSuccess()
    TUI-->>User: "Build completed successfully"
```

### Add/Update Command Flow

```mermaid
flowchart TD
    START([Command Invocation]) --> LOAD[Load Project Config]
    
    LOAD --> CHECK{Rules Specified?}
    
    CHECK -->|Yes| DIRECT[Process Specified Rules]
    CHECK -->|No| INTERACTIVE[Interactive Rule Selection]
    
    INTERACTIVE --> FETCH[Fetch Available Rules]
    FETCH --> DISPLAY[Display Rule Selector UI]
    DISPLAY --> SELECT[User Selects Rules]
    SELECT --> DIRECT
    
    DIRECT --> PROCESS[Process Each Rule]
    
    PROCESS --> VALIDATE[Validate Rule]
    VALIDATE --> DUPLICATE{Check Duplicates}
    
    DUPLICATE -->|Exists| CONFLICT[Handle Conflict]
    DUPLICATE -->|New| ADD_RULE[Add to Configuration]
    
    CONFLICT --> PROMPT{Update Existing?}
    PROMPT -->|Yes| UPDATE_RULE[Update Rule Reference]
    PROMPT -->|No| SKIP[Skip Rule]
    
    UPDATE_RULE --> ADD_RULE
    SKIP --> NEXT{More Rules?}
    ADD_RULE --> NEXT
    
    NEXT -->|Yes| PROCESS
    NEXT -->|No| SAVE[Save Configuration]
    
    SAVE --> SUCCESS([Command Complete])
    
    style INTERACTIVE fill:#e3f2fd
    style VALIDATE fill:#e8f5e8
    style CONFLICT fill:#fff3e0
    style SAVE fill:#fce4ec
```

## Integration Points

Commands integrate with various internal packages:
- **Domain Package**: Operations on rules and project entities
- **Project Package**: Configuration persistence and management
- **Format Package**: Output generation in multiple formats
- **Validation Package**: Data validation and constraint checking
- **Rule Package**: Rule processing and template rendering

## Usage Within Project

This package is used by:
- **App Package**: Application structure delegates all command execution to this package
- **Integration Tests**: End-to-end testing exercises command functionality

## Error Handling

All commands provide consistent error handling with:
- Structured error types with appropriate exit codes
- User-friendly error messages with actionable suggestions
- Validation feedback with field-specific context
- Recovery guidance for common error scenarios