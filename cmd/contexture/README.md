# CLI Entry Point

This package provides the main entry point for the CLI application. It serves as a minimal bootstrap that delegates all functionality to the internal application structure.

## Purpose

This package follows Go conventions for command-line applications by providing a clean entry point that:
- Initializes the application via the `internal/app` package.
- Handles process exit codes.
- Maintains separation between the CLI interface and the application logic.

### Bootstrap Architecture

```mermaid
graph TB
    subgraph "Operating System"
        OS[OS Process]
        ARGS[Command Line Args]
        EXITCODE[Exit Code]
    end
    
    subgraph "cmd/contexture Package"
        MAIN[main function]
        OSARGS[os.Args]
        OSEXIT[os.Exit function]
    end
    
    subgraph "Internal Application"
        APPRUN[app.Run function]
        DEPENDENCIES[Dependencies]
        CLI[CLI Framework]
        COMMANDS[Command Actions]
    end
    
    subgraph "Exit Code Sources"
        SUCCESS[Success - 0]
        USAGE[Usage Error - 1]
        CONFIG[Config Error - 2]
        NETWORK[Network Error - 3]
        PERMISSION[Permission Error - 4]
    end
    
    OS --> MAIN
    ARGS --> OSARGS
    MAIN --> APPRUN
    OSARGS --> APPRUN
    
    APPRUN --> DEPENDENCIES
    APPRUN --> CLI
    APPRUN --> COMMANDS
    
    COMMANDS --> SUCCESS
    COMMANDS --> USAGE
    COMMANDS --> CONFIG
    COMMANDS --> NETWORK
    COMMANDS --> PERMISSION
    
    SUCCESS --> OSEXIT
    USAGE --> OSEXIT
    CONFIG --> OSEXIT
    NETWORK --> OSEXIT
    PERMISSION --> OSEXIT
    
    OSEXIT --> EXITCODE
    EXITCODE --> OS
    
    style MAIN fill:#e1f5fe
    style APPRUN fill:#f3e5f5
    style DEPENDENCIES fill:#e8f5e8
    style SUCCESS fill:#c8e6c9
```

### Execution Flow

```mermaid
sequenceDiagram
    participant OS as Operating System
    participant Main as main()
    participant App as app.Run()
    participant CLI as CLI Framework
    participant Cmd as Command
    participant Exit as os.Exit()
    
    OS->>Main: Launch process with args
    Main->>App: app.Run(os.Args)
    
    App->>App: Initialize dependencies
    App->>CLI: Create CLI application
    App->>CLI: Parse arguments
    
    alt Valid command
        CLI->>Cmd: Execute command action
        Cmd->>Cmd: Perform operation
        Cmd-->>CLI: Success (exit code 0)
        CLI-->>App: Success
        App-->>Main: 0
    else Invalid usage
        CLI-->>App: Usage error
        App-->>Main: 1
    else Configuration error
        Cmd-->>CLI: Config error
        CLI-->>App: Config error  
        App-->>Main: 2
    else Network error
        Cmd-->>CLI: Network error
        CLI-->>App: Network error
        App-->>Main: 3
    end
    
    Main->>Exit: os.Exit(code)
    Exit->>OS: Process termination
    
    note over Main: Minimal delegation:<br/>• No business logic<br/>• Clean error propagation<br/>• Standard exit codes
```

### Build and Deployment Process

```mermaid
flowchart TD
    START([Go Build Process]) --> LDFLAGS[Set ldflags Variables]
    
    LDFLAGS --> VERSION["-X internal/version.version=1.0.0"]
    LDFLAGS --> COMMIT["-X internal/version.commit=abc123"]
    LDFLAGS --> DATE["-X internal/version.date=2024-01-01"]
    LDFLAGS --> BUILDER["-X internal/version.builtBy=ci"]
    
    VERSION --> COMPILE[Go Compiler]
    COMMIT --> COMPILE
    DATE --> COMPILE
    BUILDER --> COMPILE
    
    COMPILE --> EMBED[Embed Version Info]
    EMBED --> BINARY[contexture Binary]
    
    BINARY --> DEPLOY[Deployment]
    
    DEPLOY --> USER[User Execution]
    USER --> RUNTIME[Runtime main function]
    RUNTIME --> APPRUN[Delegate to app.Run function]
    
    APPRUN --> VERSIONCMD{Version Command?}
    
    VERSIONCMD -->|Yes| SHOWVERSION[Display Embedded Version]
    VERSIONCMD -->|No| NORMALCMD[Execute Normal Command]
    
    SHOWVERSION --> SUCCESS[Exit 0]
    NORMALCMD --> SUCCESS
    
    style LDFLAGS fill:#e1f5fe
    style BINARY fill:#f3e5f5
    style RUNTIME fill:#e8f5e8
    style SUCCESS fill:#c8e6c9
```

## Design Principles

- **Minimal Responsibility**: This package contains only the necessary bootstrap code.
- **Delegation**: All business logic is handled by internal packages.
- **Clean Exit Handling**: The application uses appropriate process exit codes for different scenarios.

## Usage

This package is built into the final `contexture` binary.