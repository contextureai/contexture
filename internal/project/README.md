# Project Package

This package manages project configurations. It uses a repository pattern to handle configuration persistence, rule matching, and validation.

## Architecture

The package is designed with a clean architecture, using interfaces for its core components:
- **`ConfigRepository`**: Handles loading and saving configuration files.
- **`RuleMatcher`**: Implements logic for parsing and matching rule IDs.
- **`ConfigValidator`**: Validates the project configuration and rule references.
- **`Manager`**: Orchestrates the configuration operations.

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

## Features

- **Atomic Configuration Updates**: Ensures safe configuration saving with atomic file operations.
- **Rule ID Matching**: Supports pattern matching for various rule ID formats (simple, contexture, and repository-based).
- **Path Extraction**: Parses complex rule references to extract file paths.
- **Validation Integration**: Provides deep validation of project structure and rule constraints.
- **Home Directory Support**: Automatically resolves home directory paths (e.g., `~/`).

## Usage

This package is used by:
- The `commands` package for all configuration-related CLI commands.
- The `app` package for loading the project configuration on initialization.

## API

- `NewManager(fs) -> Manager`: Creates a new project manager.
- `LoadConfig(path) -> *Project`: Loads and validates a project configuration from a file.
- `SaveConfig(config, path) -> error`: Atomically saves a configuration to a file.
- `MatchRule(ruleID, targetID) -> bool`: Checks if two rule IDs match, accounting for different formats.
- `ExtractPath(ruleID) -> string`: Extracts the file path from a formatted rule reference.