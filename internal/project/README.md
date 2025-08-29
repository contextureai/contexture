# Project Package

This package provides comprehensive project configuration management for the Contexture CLI, implementing the Repository pattern to handle configuration persistence, rule matching, and validation operations.

## Purpose

The project package manages all aspects of Contexture project configuration, from loading and saving configuration files to validating project structure and matching rule references. It serves as the bridge between the CLI commands and the underlying configuration data.

## Architecture

The package follows clean architecture principles with distinct interfaces:
- **ConfigRepository**: Handles file I/O operations and configuration persistence
- **RuleMatcher**: Manages rule ID parsing and pattern matching logic
- **ConfigValidator**: Validates project configuration and rule references
- **Manager**: High-level orchestration of configuration operations

### Repository Pattern Architecture

```mermaid
graph TB
    subgraph "Project Package"
        MGR[Manager]
        REPO[ConfigRepository]
        MATCHER[RuleMatcher]
        VALIDATOR[ConfigValidator]
        HOME[HomeDirectoryProvider]
    end
    
    subgraph "Interfaces"
        IREPO[ConfigRepository Interface]
        IMATCHER[RuleMatcher Interface]
        IVALIDATOR[ConfigValidator Interface]
        IHOME[HomeDirectoryProvider Interface]
    end
    
    subgraph "Implementation"
        FSREPO[FilesystemRepository]
        REGEXMATCHER[RegexMatcher]
        DOMAINVAL[DomainValidator]
        OSHOME[OSHomeProvider]
    end
    
    subgraph "External Dependencies"
        FS[Filesystem]
        YAML[YAML Parser]
        DOMAIN[Domain Package]
        VAL[Validation Package]
    end
    
    subgraph "Client Code"
        COMMANDS[Commands Package]
        APP[App Package]
    end
    
    MGR --> IREPO
    MGR --> IMATCHER
    MGR --> IVALIDATOR
    MGR --> IHOME
    
    IREPO -.-> FSREPO
    IMATCHER -.-> REGEXMATCHER
    IVALIDATOR -.-> DOMAINVAL
    IHOME -.-> OSHOME
    
    FSREPO --> FS
    FSREPO --> YAML
    REGEXMATCHER --> DOMAIN
    DOMAINVAL --> VAL
    
    COMMANDS --> MGR
    APP --> MGR
    
    style MGR fill:#e1f5fe
    style IREPO fill:#f3e5f5
    style IMATCHER fill:#e8f5e8
    style IVALIDATOR fill:#fff3e0
```

### Configuration Loading and Validation Flow

```mermaid
sequenceDiagram
    participant Client as Client Code
    participant Mgr as Manager
    participant Repo as ConfigRepository
    participant Val as ConfigValidator
    participant Match as RuleMatcher
    participant FS as Filesystem
    
    Client->>Mgr: LoadConfig(path)
    
    Mgr->>Repo: Exists(path)
    Repo->>FS: FileExists
    FS-->>Repo: Exists Result
    Repo-->>Mgr: File Status
    
    alt File Exists
        Mgr->>Repo: Load(path)
        Repo->>FS: ReadFile
        FS-->>Repo: YAML Content
        Repo->>Repo: Parse YAML
        Repo-->>Mgr: Project Config
        
        Mgr->>Val: ValidateProject(config)
        
        loop For each rule reference
            Val->>Match: ValidateRuleRef(ruleRef)
            Match-->>Val: Validation Result
        end
        
        Val-->>Mgr: Validation Result
        
        alt Validation Success
            Mgr-->>Client: Valid Project Config
        else Validation Failure
            Mgr-->>Client: Validation Errors
        end
        
    else File Not Found
        Mgr-->>Client: File Not Found Error
    end
```

### Rule ID Matching Logic

```mermaid
flowchart TD
    START([Rule ID Input]) --> PARSE[Parse Rule ID Format]
    
    PARSE --> CHECK{Check Format Type}
    
    CHECK -->|Simple ID| SIMPLE[Simple String Match]
    CHECK -->|Contexture Format| CONTEXTMATCH[Parse Contexture Format]
    CHECK -->|Repository Format| REPOMATCH[Parse Repository Format]
    
    SIMPLE --> EXACT{Exact Match?}
    EXACT -->|Yes| MATCH[Return Match]
    EXACT -->|No| NOMATCH[Return No Match]
    
    CONTEXTMATCH --> EXTRACT[Extract Path Component]
    EXTRACT --> PATHCHECK{Path Match?}
    PATHCHECK -->|Yes| MATCH
    PATHCHECK -->|No| NOMATCH
    
    REPOMATCH --> REPOEXTRACT[Extract Repository & Path]
    REPOEXTRACT --> REPOCHECK{Repository & Path Match?}
    REPOCHECK -->|Yes| BRANCHCHECK{Check Branch/Ref}
    REPOCHECK -->|No| NOMATCH
    
    BRANCHCHECK -->|Match| MATCH
    BRANCHCHECK -->|No Match| NOMATCH
    
    MATCH --> SUCCESS([Successful Match])
    NOMATCH --> FAIL([No Match Found])
    
    style CONTEXTMATCH fill:#e1f5fe
    style REPOMATCH fill:#f3e5f5
    style EXTRACT fill:#e8f5e8
    style PATHCHECK fill:#fff3e0
```

### Configuration Persistence Flow

```mermaid
flowchart TD
    START([Save Configuration]) --> VALIDATE[Validate Configuration]
    
    VALIDATE --> BACKUP{Backup Exists?}
    BACKUP -->|Yes| CREATEBACKUP[Create Backup Copy]
    BACKUP -->|No| SERIALIZE[Serialize to YAML]
    
    CREATEBACKUP --> SERIALIZE
    
    SERIALIZE --> ATOMIC[Write Atomically]
    ATOMIC --> TEMPFILE[Write to Temp File]
    
    TEMPFILE --> RENAME{Rename to Target}
    
    RENAME -->|Success| CLEANUP[Cleanup Temp Files]
    RENAME -->|Failure| RESTORE[Restore from Backup]
    
    CLEANUP --> SUCCESS[Save Complete]
    
    RESTORE --> ERROR[Save Failed]
    
    style VALIDATE fill:#e1f5fe
    style ATOMIC fill:#f3e5f5
    style TEMPFILE fill:#e8f5e8
    style SUCCESS fill:#c8e6c9
    style ERROR fill:#ffcdd2
```

## Key Features

- **Atomic Configuration Updates**: Safe configuration saving with atomic file operations
- **Rule ID Matching**: Sophisticated pattern matching for different rule ID formats
- **Path Extraction**: Parsing of complex rule reference formats including contexture-specific patterns
- **Validation Integration**: Deep validation of project structure and rule constraints
- **Thread Safety**: All operations are safe for concurrent use
- **Home Directory Support**: Automatic resolution of home directory paths

## Configuration Management

- **YAML-Based**: Project configurations stored as YAML files
- **Schema Validation**: Strict validation of configuration structure and content
- **Format Management**: Validation of output format configurations
- **Rule Uniqueness**: Ensures unique rule IDs across project configurations

## Rule Matching Logic

The package handles complex rule ID formats:
- Simple rule IDs (e.g., `my-rule`)
- Full contexture format (e.g., `[contexture:path/to/rule]`)
- Repository-based format (e.g., `[contexture(source):path,branch]`)
- Path extraction from formatted rule references

## Usage Within Project

This package is used by:
- **Commands Package**: All configuration-related CLI commands use this package
- **App Package**: Application initialization loads project configuration
- **Validation Package**: Project validation delegates to this package's validators

## API

- `NewManager(fs)`: Creates a project manager with filesystem dependencies
- `LoadConfig(path)`: Loads and validates project configuration from file
- `SaveConfig(config, path)`: Atomically saves configuration to file
- `MatchRule(ruleID, targetID)`: Determines if rule IDs match across different formats
- `ExtractPath(ruleID)`: Extracts file path from formatted rule references