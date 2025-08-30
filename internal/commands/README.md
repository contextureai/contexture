# Commands Package

This package implements the actions for all CLI commands. It serves as the business logic layer that connects the CLI framework to the underlying domain operations.

## Commands

### Project Management
- `init`: Initializes a new project with a default configuration.
- `config`: Manages project configuration.

### Rule Operations
- `add`: Adds new rules to the project from local files or Git repositories.
- `remove`: Removes rules from the project configuration.
- `list`: Lists the rules in the project.
- `update`: Updates existing rules from their sources.

### Build System
- `build`: Generates output files in the configured formats.

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

## Architecture

- **Command Actions**: Each command is implemented as a separate action with a clear separation of concerns.
- **Shared Utilities**: Common functionality, such as configuration loading and validation, is shared across commands.
- **Error Handling**: All commands use the centralized error handling package to provide consistent, user-friendly error messages and appropriate exit codes.

## Usage

This package is used by the `app` package, which delegates all command execution to the actions in this package.