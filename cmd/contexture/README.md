# Contexture CLI Entry Point

This package provides the main entry point for the Contexture CLI application. It serves as the minimal bootstrap that delegates to the internal application structure for all functionality.

## Purpose

The cmd/contexture package follows Go conventions for command-line applications by providing a clean entry point that:
- Initializes the application through the internal app package
- Handles process exit codes appropriately
- Maintains separation between the CLI interface and application logic

## Architecture

The package implements a minimal main function that:
1. Delegates all functionality to `internal/app.Run()`
2. Passes command-line arguments directly to the application
3. Exits with the appropriate status code based on execution results

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

- **Minimal Responsibility**: Contains only the necessary bootstrap code
- **Delegation Pattern**: All business logic is handled by internal packages
- **Clean Exit Handling**: Proper process exit codes for different scenarios
- **Go Conventions**: Follows standard Go project layout for CLI applications

## Files

- `main.go`: The application entry point with minimal bootstrap logic
- `main_test.go`: Basic tests for the main function behavior

## Usage

This package is built into the final `contexture` binary that users interact with. The binary can be invoked with various commands and flags:

```bash
contexture --help
contexture init
contexture rules add [rule-id]
contexture build
```

## Relationship to Internal Packages

The cmd package acts as the public interface that users interact with, while delegating all functionality to:
- **internal/app**: Application structure and CLI framework setup
- **internal/commands**: Individual command implementations
- **internal/dependencies**: Dependency injection and resource management

## Build Process

This package is the build target for creating the final CLI binary:
- Compiled to produce the `contexture` executable
- All internal packages are bundled into the final binary
- No external runtime dependencies required