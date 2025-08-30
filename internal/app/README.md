# App Package

This package provides the main application structure for the CLI. It serves as the composition root, initializing dependencies, configuring the CLI framework, and managing the application lifecycle.

## Features

- **Application Lifecycle Management**: Handles the setup and teardown of application resources.
- **Dependency Injection**: Provides a central point for creating dependencies and distributing them to commands.
- **CLI Framework Integration**: Configures and sets up the `urfave/cli` framework.
- **Command Orchestration**: Registers and organizes all CLI commands and subcommands.
- **Error Handling**: Implements unified error display and exit code management.
- **Testable Actions**: Wraps command actions to enable comprehensive testing.

### Application Architecture

```mermaid
graph TB
    subgraph "Entry Point"
        MAIN[cmd/contexture/main.go]
    end
    
    subgraph "App Package"
        RUN[Run Function]
        APP[Application]
        ACTIONS[CommandActions]
        BUILDER[CLI Builder]
    end
    
    subgraph "CLI Framework"
        CLI[urfave/cli]
        HELP[Help System]
        FLAGS[Global Flags]
    end
    
    subgraph "Dependencies"
        DEPS[Dependencies]
        FS[Filesystem]
        CTX[Context]
    end
    
    subgraph "Commands"
        INIT[InitAction]
        ADD[AddAction]
        REMOVE[RemoveAction]
        LIST[ListAction]
        UPDATE[UpdateAction]
        BUILD[BuildAction]
        CONFIG[ConfigAction]
    end
    
    subgraph "Core Services"
        COMMANDS[Commands Package]
        ERRORS[Error Handling]
        VERSION[Version Info]
    end
    
    MAIN --> RUN
    RUN --> APP
    RUN --> DEPS
    
    APP --> ACTIONS
    APP --> BUILDER
    
    BUILDER --> CLI
    BUILDER --> HELP
    BUILDER --> FLAGS
    
    ACTIONS --> INIT
    ACTIONS --> ADD
    ACTIONS --> REMOVE
    ACTIONS --> LIST
    ACTIONS --> UPDATE
    ACTIONS --> BUILD
    ACTIONS --> CONFIG
    
    DEPS --> FS
    DEPS --> CTX
    
    APP --> COMMANDS
    APP --> ERRORS
    APP --> VERSION
    
    style APP fill:#e1f5fe
    style ACTIONS fill:#f3e5f5
    style DEPS fill:#e8f5e8
    style COMMANDS fill:#fff3e0
```

### Dependency Injection Flow

```mermaid
sequenceDiagram
    participant Main as main()
    participant Run as Run()
    participant App as Application
    participant Deps as Dependencies
    participant Actions as CommandActions
    participant Cmd as Commands
    
    Main->>Run: os.Args
    Run->>Deps: New(context)
    Deps-->>Run: Dependencies Instance
    
    Run->>App: New(deps)
    App->>Actions: NewCommandActions(deps)
    Actions-->>App: Action Wrappers
    App-->>Run: Application Instance
    
    Run->>App: Execute(ctx, args)
    App->>App: buildCLIApp()
    App->>App: buildCommands()
    
    loop For each command
        App->>Actions: Get Action Wrapper
        Actions->>Cmd: Call Command Implementation
        Cmd-->>Actions: Result
        Actions-->>App: Wrapped Result
    end
    
    App-->>Run: Execution Result
    Run-->>Main: Exit Code
```

### Command Registration Process

```mermaid
flowchart TD
    START([Application Start]) --> CREATE[Create Dependencies]
    
    CREATE --> NEWAPP[New Application Instance]
    NEWAPP --> ACTIONS[Create CommandActions]
    
    ACTIONS --> BUILD[Build CLI App]
    BUILD --> COMMANDS[Build Commands]
    
    COMMANDS --> INIT_CMD[Register init command]
    COMMANDS --> RULES_CMD[Register rules command]
    COMMANDS --> BUILD_CMD[Register build command]
    COMMANDS --> CONFIG_CMD[Register config command]
    
    INIT_CMD --> SUBCOMMANDS[Add Subcommands]
    RULES_CMD --> SUBCOMMANDS
    BUILD_CMD --> SUBCOMMANDS
    CONFIG_CMD --> SUBCOMMANDS
    
    SUBCOMMANDS --> ADD_SUB[add subcommand]
    SUBCOMMANDS --> REMOVE_SUB[remove subcommand]
    SUBCOMMANDS --> LIST_SUB[list subcommand]
    SUBCOMMANDS --> UPDATE_SUB[update subcommand]
    
    ADD_SUB --> FLAGS[Set Flags & Options]
    REMOVE_SUB --> FLAGS
    LIST_SUB --> FLAGS
    UPDATE_SUB --> FLAGS
    
    FLAGS --> HELP[Configure Help]
    HELP --> READY[CLI Ready for Execution]
    
    style NEWAPP fill:#e1f5fe
    style ACTIONS fill:#f3e5f5
    style COMMANDS fill:#e8f5e8
    style FLAGS fill:#fff3e0
```

## Usage

This package is used by the main entry point (`cmd/contexture/main.go`) to run the application. It is also used in integration tests.

## API

- `New(deps) -> Application`: Creates a new application instance with the provided dependencies.
- `Run(args) -> int`: The main entry point for the application. It returns an exit code.
- `Execute(ctx, args) -> error`: Executes the CLI application with a given context and arguments.