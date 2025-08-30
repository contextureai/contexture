# Dependencies Package

This package provides minimal dependency injection for the Contexture CLI application. It manages core dependencies like filesystem operations and context propagation in a clean, testable way without heavy dependency injection frameworks.

## Purpose

The dependencies package serves as a lightweight container for essential application dependencies, enabling clean separation of concerns and facilitating testing through dependency substitution. It provides a simple alternative to complex DI frameworks while maintaining testability and flexibility.

## Key Features

- **Minimal Design**: Contains only essential dependencies (filesystem and context)
- **Production Defaults**: Real filesystem operations for normal application usage
- **Testing Support**: In-memory filesystem variant to avoid side effects during tests
- **Immutable Updates**: Methods return new instances rather than modifying existing ones
- **Context Propagation**: Proper context handling for cancellation and request scoping

## Dependencies Managed

- **Filesystem Operations**: Uses `afero.Fs` interface for file system abstraction
- **Application Context**: Manages `context.Context` for lifecycle and cancellation

## Usage Within Project

This package is used extensively throughout the application:
- **Commands Package**: All CLI commands receive dependencies for file operations
- **App Package**: Main application initialization and dependency setup
- **Actions**: Application actions use dependencies for their operations

### Dependency Flow Architecture

```mermaid
graph TB
    subgraph "Application Initialization"
        MAIN[Main Function]
        APP[App Package]
        DEPS[Dependencies]
    end
    
    subgraph "Core Dependencies"
        FS[Filesystem Interface]
        CTX[Context]
        PRODFS[Production FileSystem]
        TESTFS[Test FileSystem]
    end
    
    subgraph "Consumer Packages"
        CMD[Commands Package]
        PROJ[Project Package]
        RULE[Rule Package]
        FORMAT[Format Package]
        CACHE[Cache Package]
    end
    
    subgraph "Operations"
        FILEOPS[File Operations]
        CONFIGOPS[Config Operations]
        RULEOPS[Rule Operations]
        BUILDOPS[Build Operations]
    end
    
    MAIN --> APP
    APP --> DEPS
    
    DEPS --> FS
    DEPS --> CTX
    
    FS -.->|Production| PRODFS
    FS -.->|Testing| TESTFS
    
    DEPS --> CMD
    DEPS --> PROJ
    DEPS --> RULE
    DEPS --> FORMAT
    DEPS --> CACHE
    
    CMD --> FILEOPS
    PROJ --> CONFIGOPS
    RULE --> RULEOPS
    FORMAT --> BUILDOPS
    
    style DEPS fill:#e1f5fe
    style FS fill:#f3e5f5
    style CTX fill:#e8f5e8
    style CMD fill:#fff3e0
```

### Dependency Injection Pattern

```mermaid
sequenceDiagram
    participant Main as main()
    participant App as App.Run()
    participant Deps as Dependencies
    participant Cmd as Command
    participant FS as Filesystem
    
    Main->>App: os.Args
    App->>Deps: New(context)
    
    alt Production Mode
        Deps->>FS: afero.NewOsFs()
    else Test Mode
        Deps->>FS: afero.NewMemMapFs()
    end
    
    FS-->>Deps: Filesystem Instance
    Deps-->>App: Dependencies
    
    App->>Cmd: Execute(deps)
    
    Cmd->>Deps: Access filesystem
    Deps->>FS: Delegate operation
    FS-->>Deps: Operation result
    Deps-->>Cmd: Result
    
    Cmd->>Deps: Access context
    Deps-->>Cmd: Context for cancellation
    
    note over Deps: Immutable pattern:<br/>WithContext() creates<br/>new instance
```

### Dependencies Lifecycle

```mermaid
stateDiagram-v2
    [*] --> Created: New() or NewForTesting()
    
    Created --> InUse: Passed to commands
    
    InUse --> ContextUpdated: WithContext()
    InUse --> FilesystemUpdated: WithFS()
    
    ContextUpdated --> InUse: New instance created
    FilesystemUpdated --> InUse: New instance created
    
    InUse --> Cleanup: Application shutdown
    Cleanup --> [*]
    
    note right of Created
        Production: Real filesystem
        Testing: In-memory filesystem
    end note
    
    note right of ContextUpdated
        Immutable pattern:
        Original unchanged
    end note
```

## API

- `New(ctx)`: Creates production dependencies with real filesystem
- `NewForTesting(ctx)`: Creates test dependencies with in-memory filesystem  
- `WithContext(ctx)`: Returns new instance with different context
- `WithFS(fs)`: Returns new instance with different filesystem implementation

## Testing Benefits

The testing variant uses an in-memory filesystem (`afero.NewMemMapFs()`) which provides:
- No file system side effects during tests
- Faster test execution
- Isolated test environments
- Deterministic test behavior