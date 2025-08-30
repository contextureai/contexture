# Git Package

This package provides secure Git repository operations with comprehensive error handling, configurable timeouts, authentication management, and retry logic for robust repository interactions.

## Purpose

The git package abstracts Git operations behind a clean interface, handling the complexity of authentication, network issues, and repository management. It ensures secure, reliable access to Git repositories while providing detailed progress reporting and error recovery.

## Key Features

- **Secure Operations**: Comprehensive URL validation and host allowlisting
- **Authentication Support**: HTTP basic auth, SSH keys, and token-based authentication
- **Timeout Management**: Configurable timeouts for clone and pull operations
- **Retry Logic**: Automatic retry for transient network failures
- **Progress Reporting**: Real-time progress updates for long-running operations
- **Repository Validation**: Checks for valid Git repositories and remote URLs
- **Commit Information**: Retrieval of commit metadata and file history

## Security Features

- **URL Validation**: Strict validation of repository URLs and schemes
- **Host Allowlisting**: Configurable list of allowed Git hosts
- **Scheme Restrictions**: Support for HTTPS, HTTP, SSH, and git protocols only
- **Authentication Abstraction**: Pluggable authentication providers

## Operation Types

- **Clone**: Full repository cloning with branch/tag specification
- **Pull**: Updates to existing repositories with conflict resolution
- **Commit Info**: Retrieval of commit metadata and file-specific history
- **File Access**: Reading files at specific commits or branches
- **Repository Validation**: Checking repository validity and remote access

## Usage Within Project

This package is used by:
- **Cache Package**: Repository caching relies on git operations for clone and pull
- **Rule Package**: Git fetcher uses this package for repository-based rule retrieval
- **Integration Tests**: End-to-end testing uses git operations for test repositories

## Interface Design

The package follows Go interface conventions with a clean `Repository` interface that:
- Enables easy testing through mock implementations
- Supports different Git backends (go-git, system git, etc.)
- Provides consistent error handling across all operations
- Allows for progress monitoring and cancellation

### Git Operations Architecture

```mermaid
graph TB
    subgraph "Client Layer"
        CLIENT[Git Client]
        CONFIG[Git Config]
        AUTH[Auth Provider]
        PROGRESS[Progress Handler]
    end
    
    subgraph "Repository Interface"
        REPO[Repository Interface]
        CLONE[Clone Operations]
        PULL[Pull Operations]
        INFO[Commit Info Operations]
        VALIDATE[Validation Operations]
    end
    
    subgraph "Security Layer"
        URLVAL[URL Validation]
        HOSTALLOW[Host Allowlisting]
        SCHEMECHECK[Scheme Validation]
        AUTHCHECK[Authentication Check]
    end
    
    subgraph "Network Layer"
        TIMEOUT[Timeout Management]
        RETRY[Retry Logic]
        TRANSPORT[Transport Layer]
        HTTPAUTH[HTTP Auth]
        SSHAUTH[SSH Auth]
    end
    
    subgraph "Storage Layer"
        FS[Filesystem]
        GITDIR[Git Directory]
        REFS[References]
        OBJECTS[Git Objects]
    end
    
    CLIENT --> REPO
    CONFIG --> CLIENT
    AUTH --> CLIENT
    PROGRESS --> CLIENT
    
    REPO --> CLONE
    REPO --> PULL
    REPO --> INFO
    REPO --> VALIDATE
    
    CLONE --> URLVAL
    PULL --> URLVAL
    VALIDATE --> URLVAL
    
    URLVAL --> HOSTALLOW
    URLVAL --> SCHEMECHECK
    URLVAL --> AUTHCHECK
    
    CLONE --> TIMEOUT
    PULL --> TIMEOUT
    TIMEOUT --> RETRY
    RETRY --> TRANSPORT
    
    TRANSPORT --> HTTPAUTH
    TRANSPORT --> SSHAUTH
    
    CLONE --> FS
    PULL --> GITDIR
    INFO --> REFS
    INFO --> OBJECTS
    
    style CLIENT fill:#e1f5fe
    style REPO fill:#f3e5f5
    style URLVAL fill:#e8f5e8
    style TIMEOUT fill:#fff3e0
```

### Clone Operation Flow

```mermaid
sequenceDiagram
    participant Client as Client Code
    participant GitClient as Git Client
    participant Validator as URL Validator
    participant Auth as Auth Provider
    participant Transport as Transport Layer
    participant Progress as Progress Handler
    participant FS as Filesystem
    
    Client->>GitClient: Clone(url, localPath, options)
    
    GitClient->>Validator: ValidateURL(url)
    Validator->>Validator: Check scheme (https, ssh, git)
    Validator->>Validator: Validate host allowlist
    Validator-->>GitClient: Validation result
    
    alt Invalid URL
        GitClient-->>Client: ErrInvalidURL
    else Valid URL
        GitClient->>Auth: GetAuth(url)
        Auth-->>GitClient: Auth method
        
        GitClient->>Progress: OnProgress("Starting clone")
        
        GitClient->>Transport: Configure transport
        Transport->>Transport: Set auth method
        Transport->>Transport: Set timeout
        
        loop Retry logic (max 3 times)
            Transport->>Transport: Attempt clone
            
            alt Network error
                Transport->>Transport: Wait and retry
            else Success
                Transport-->>GitClient: Clone successful
            else Permanent error
                Transport-->>GitClient: Clone failed
            end
        end
        
        alt Clone successful
            GitClient->>FS: Validate git directory
            FS-->>GitClient: Repository valid
            GitClient->>Progress: OnComplete()
            GitClient-->>Client: Success
        else Clone failed
            GitClient->>Progress: OnError(err)
            GitClient-->>Client: Error
        end
    end
```

### Authentication Flow

```mermaid
flowchart TD
    START([Git Operation]) --> GETAUTH[Get Authentication Method]
    
    GETAUTH --> URLCHECK{Check URL Scheme}
    
    URLCHECK -->|HTTPS| HTTPAUTH[HTTP Authentication]
    URLCHECK -->|SSH| SSHAUTH[SSH Authentication]
    URLCHECK -->|git://| NOAUTH[No Authentication]
    
    HTTPAUTH --> TOKENCHECK{Token Available?}
    TOKENCHECK -->|Yes| USETOKEN[Use Token Auth]
    TOKENCHECK -->|No| BASICCHECK{Basic Auth?}
    
    BASICCHECK -->|Yes| USEBASIC[Use Basic Auth]
    BASICCHECK -->|No| ANONYMOUS[Anonymous Access]
    
    SSHAUTH --> KEYCHECK{SSH Key Available?}
    KEYCHECK -->|Yes| USEKEY[Use SSH Key]
    KEYCHECK -->|No| AGENT[Try SSH Agent]
    
    AGENT --> AGENTCHECK{Agent Available?}
    AGENTCHECK -->|Yes| USEAGENT[Use SSH Agent]
    AGENTCHECK -->|No| SSHFAIL[SSH Auth Failed]
    
    USETOKEN --> SUCCESS[Authentication Ready]
    USEBASIC --> SUCCESS
    ANONYMOUS --> SUCCESS
    USEKEY --> SUCCESS
    USEAGENT --> SUCCESS
    NOAUTH --> SUCCESS
    
    SSHFAIL --> ERROR[Authentication Error]
    
    SUCCESS --> OPERATION[Perform Git Operation]
    ERROR --> FAIL[Operation Failed]
    
    style GETAUTH fill:#e1f5fe
    style HTTPAUTH fill:#f3e5f5
    style SSHAUTH fill:#e8f5e8
    style SUCCESS fill:#c8e6c9
    style ERROR fill:#ffcdd2
```

## API

- `NewClient(fs, config)`: Creates a new Git client with filesystem and configuration
- Repository interface includes `Clone()`, `Pull()`, `GetLatestCommitHash()`, `ValidateURL()`
- Configuration supports timeouts, authentication, progress handlers, and security settings