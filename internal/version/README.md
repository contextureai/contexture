# Version Package

This package provides comprehensive build-time version information for the Contexture CLI application. It captures essential metadata about the binary including version, commit hash, build date, and runtime environment details.

## Purpose

The version package serves as the central source of truth for application versioning and build metadata. It enables the CLI to report detailed version information to users and supports debugging by providing build context.

## Key Features

- **Build-time Variables**: Version, commit hash, build date, and builder information set via ldflags during compilation
- **Runtime Information**: Automatically captures Go version and platform (OS/architecture)
- **Multiple Output Formats**: Short version string for basic usage, detailed output for comprehensive information
- **Default Fallbacks**: Graceful handling when build-time variables aren't set (defaults to "dev" and "unknown")

## Usage Within Project

This package is used by:
- **App Package**: For version commands and CLI metadata
- **Rule Variable Manager**: For template variable substitution in rules

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

- `Get()`: Returns complete version information as an `Info` struct
- `GetShort()`: Returns just the version string
- `Info.String()`: Formatted version string for display
- `Info.Detailed()`: Comprehensive version information including all metadata