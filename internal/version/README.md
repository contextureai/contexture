# Version Package

This package provides build-time version information for the application. It captures metadata including the version, commit hash, build date, and runtime environment details, which are embedded into the binary at compile time via `ldflags`.

## Features

- **Build-time Variables**: Captures version, commit hash, build date, and builder information.
- **Runtime Information**: Automatically includes the Go version and platform (OS/architecture).
- **Multiple Output Formats**: Provides both a short version string and a detailed output.
- **Default Fallbacks**: If build-time variables are not set, it defaults to `dev` for the version and `unknown` for other fields.

## Usage

This package is primarily used by:
- The `app` package for the `version` command and other CLI metadata.
- The rule variable manager for template variable substitution.

### Version Information Flow

```mermaid
graph TB
    subgraph "Build Process"
        LDFLAGS[Build ldflags]
        COMPILE[Go Compilation]
        BINARY[Binary Output]
    end
    
    subgraph "Version Package"
        VERSION[Version Variables]
        COMMIT[Commit Hash]
        DATE[Build Date]
        BUILDER[Builder Info]
        RUNTIME[Runtime Info]
        INFO[Version Info Struct]
    end
    
    subgraph "Runtime Detection"
        GOVERSION[Go Version]
        GOOS[Operating System]
        GOARCH[Architecture]
        PLATFORM[Platform String]
    end
    
    subgraph "Application Usage"
        CLI[CLI Version Command]
        TEMPLATE[Template Variables]
        DEBUG[Debug Information]
        HELP[Help Display]
    end
    
    LDFLAGS --> VERSION
    LDFLAGS --> COMMIT
    LDFLAGS --> DATE
    LDFLAGS --> BUILDER
    
    COMPILE --> BINARY
    BINARY --> INFO
    
    INFO --> RUNTIME
    RUNTIME --> GOVERSION
    RUNTIME --> GOOS
    RUNTIME --> GOARCH
    RUNTIME --> PLATFORM
    
    INFO --> CLI
    INFO --> TEMPLATE
    INFO --> DEBUG
    INFO --> HELP
    
    style VERSION fill:#e1f5fe
    style INFO fill:#f3e5f5
    style RUNTIME fill:#e8f5e8
    style CLI fill:#fff3e0
```

### Build-time Integration

```mermaid
sequenceDiagram
    participant Build as Build System
    participant LDFLAGS as ldflags
    participant Go as Go Compiler
    participant Binary as Binary
    participant Version as Version Package
    participant User as User
    
    Build->>LDFLAGS: Set version variables
    LDFLAGS->>LDFLAGS: -X version.version=1.0.0
    LDFLAGS->>LDFLAGS: -X version.commit=abc123
    LDFLAGS->>LDFLAGS: -X version.date=2024-01-01
    LDFLAGS->>LDFLAGS: -X version.builtBy=ci-system
    
    Build->>Go: go build -ldflags
    Go->>Binary: Compile with embedded values
    
    User->>Binary: contexture --version
    Binary->>Version: Get()
    
    Version->>Version: Read embedded variables
    Version->>Version: Detect runtime info
    Version->>Version: Create Info struct
    Version-->>Binary: Complete version info
    
    Binary-->>User: Formatted version display
    
    note over Version: Falls back to defaults<br/>if ldflags not set:<br/>• version: "dev"<br/>• commit: "unknown"
```

## API

- `Get() -> Info`: Returns a complete `Info` struct containing all version information.
- `GetShort() -> string`: Returns only the version string.
- `Info.String() -> string`: Returns a formatted string of the version information for display.
- `Info.Detailed() -> string`: Returns a comprehensive, multi-line string containing all version metadata.