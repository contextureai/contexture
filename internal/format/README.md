# Format Package

This package manages output format implementations for different AI assistant platforms, providing a registry system for format discovery, a builder pattern for format instantiation, and base utilities for format development.

## Purpose

The format package enables Contexture to generate output files tailored to specific AI assistant platforms (Claude, Cursor, Windsurf) while maintaining a consistent interface for format operations. It provides the infrastructure for format extensibility and management.

## Architecture

### Registry Pattern
- **Format Registry**: Central registry for discovering and managing available formats
- **Handler Interface**: UI integration for format selection and configuration
- **Dynamic Registration**: Support for runtime format registration and discovery

### Builder Pattern
- **Format Builder**: Factory for creating format implementations with configuration
- **Constructor Functions**: Type-safe format instantiation with validation
- **Options Support**: Flexible configuration through option maps

### Base Infrastructure
- **Common Interfaces**: Shared contracts for all format implementations
- **Directory Management**: Utilities for output directory creation and management
- **UI Integration**: Format-specific user interface components

### Registry and Builder Pattern

```mermaid
graph TB
    subgraph "Format Package"
        REG[Registry]
        BUILDER[Builder]
        DIRMGR[DirectoryManager]
        INTERFACES[Interfaces]
    end
    
    subgraph "Format Implementations"
        CLAUDE[Claude Format]
        CURSOR[Cursor Format]
        WINDSURF[Windsurf Format]
        CHAND[Claude Handler]
        CURHAND[Cursor Handler]
        WINHAND[Windsurf Handler]
    end
    
    subgraph "External Dependencies"
        FS[Filesystem]
        DOMAIN[Domain Package]
        UI[UI Package]
        TEMPLATE[Template Package]
    end
    
    subgraph "Client Code"
        CMD[Commands]
        BUILD[Build Process]
    end
    
    REG --> CLAUDE
    REG --> CURSOR
    REG --> WINDSURF
    REG --> CHAND
    REG --> CURHAND
    REG --> WINHAND
    REG --> BUILDER
    REG --> DIRMGR
    
    BUILDER --> CLAUDE
    BUILDER --> CURSOR
    BUILDER --> WINDSURF
    
    DIRMGR --> FS
    
    CLAUDE --> DOMAIN
    CURSOR --> DOMAIN
    WINDSURF --> DOMAIN
    
    CHAND --> UI
    CURHAND --> UI
    WINHAND --> UI
    
    CMD --> REG
    BUILD --> BUILDER
    
    CLAUDE --> TEMPLATE
    CURSOR --> TEMPLATE
    WINDSURF --> TEMPLATE
    
    style REG fill:#e1f5fe
    style BUILDER fill:#f3e5f5
    style DIRMGR fill:#e8f5e8
    style INTERFACES fill:#fff3e0
```

### Format Processing Flow

```mermaid
sequenceDiagram
    participant Build as Build Command
    participant Reg as Registry
    participant Builder as Builder
    participant Format as Format Implementation
    participant DirMgr as DirectoryManager
    participant FS as Filesystem
    
    Build->>Reg: GetDefaultRegistry(fs)
    Reg-->>Build: Registry Instance
    
    Build->>Reg: CreateFormat(formatType, fs, options)
    Reg->>Builder: Build(formatType, fs, options)
    
    Builder->>Builder: Lookup Constructor
    Builder->>Format: constructor(fs, options)
    Format-->>Builder: Format Instance
    Builder-->>Reg: Format Instance
    Reg-->>Build: Format Instance
    
    Build->>Format: Generate(rules, context)
    
    Format->>DirMgr: CreateOutputDirectory()
    DirMgr->>FS: MkdirAll()
    FS-->>DirMgr: Directory Created
    DirMgr-->>Format: Output Path
    
    Format->>Format: ProcessRules(rules)
    Format->>Format: RenderTemplate(content)
    
    Format->>FS: WriteFile(output)
    FS-->>Format: File Written
    
    Format-->>Build: Generation Complete
```

### Format Selection and Configuration

```mermaid
flowchart TD
    START([Build Process]) --> LOADCONFIG[Load Project Config]
    
    LOADCONFIG --> GETFORMATS[Get Enabled Formats]
    GETFORMATS --> REGISTRY[Get Format Registry]
    
    REGISTRY --> CHECKFORMATS{Validate Format Types}
    
    CHECKFORMATS -->|Valid| CREATEFORMATS[Create Format Instances]
    CHECKFORMATS -->|Invalid| ERROR[Format Not Supported Error]
    
    CREATEFORMATS --> LOOP{For Each Format}
    
    LOOP --> GETBUILDER[Get Format Builder]
    GETBUILDER --> BUILDFORMAT[Build Format Instance]
    
    BUILDFORMAT --> CONFIGURE[Apply Configuration Options]
    CONFIGURE --> VALIDATE[Validate Format Config]
    
    VALIDATE -->|Valid| GENERATE[Generate Output]
    VALIDATE -->|Invalid| CONFIGERROR[Configuration Error]
    
    GENERATE --> CREATEDIR[Create Output Directory]
    CREATEDIR --> PROCESS[Process Rules]
    PROCESS --> RENDER[Render Templates]
    RENDER --> WRITE[Write Files]
    
    WRITE --> NEXTFORMAT{More Formats?}
    NEXTFORMAT -->|Yes| LOOP
    NEXTFORMAT -->|No| SUCCESS[Build Complete]
    
    ERROR --> FAIL[Build Failed]
    CONFIGERROR --> FAIL
    
    style REGISTRY fill:#e1f5fe
    style BUILDFORMAT fill:#f3e5f5
    style GENERATE fill:#e8f5e8
    style SUCCESS fill:#c8e6c9
    style FAIL fill:#ffcdd2
```

## Supported Formats

### Built-in Formats
- **Claude**: AI assistant format optimized for Claude's interface and capabilities
- **Cursor**: Code editor format tailored for Cursor's AI features
- **Windsurf**: Development environment format designed for Windsurf workflows

### Format Capabilities
- **Single/Multi-file Output**: Support for both single-file and directory-based outputs
- **File Extension Management**: Automatic file extension handling based on format type
- **Template Processing**: Integration with template engine for dynamic content generation
- **Configuration Validation**: Format-specific configuration validation and constraints

## Directory Management

- **Output Directory Creation**: Automatic creation of format-specific output directories
- **File Organization**: Structured organization of generated files
- **Path Resolution**: Intelligent path resolution for various output scenarios
- **Cleanup Support**: Utilities for managing generated file lifecycles

## Usage Within Project

This package is used by:
- **Commands Package**: Build command uses formats for output generation
- **Domain Package**: Format interfaces define contracts for format implementations
- **Project Package**: Configuration management includes format settings

## API

### Registry Operations
- `NewRegistry(fs)`: Creates format registry with filesystem support
- `GetDefaultRegistry(fs)`: Returns registry with all built-in formats
- `Register(formatType, handler)`: Adds format to registry
- `CreateFormat(type, fs, options)`: Creates format implementation

### Builder Operations
- `NewBuilder()`: Creates format builder with built-in formats
- `Register(type, constructor)`: Adds constructor function
- `Build(type, fs, options)`: Instantiates format with configuration
- `GetSupportedFormats()`: Lists available format types