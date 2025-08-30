# Git Package

This package provides an interface for Git repository operations, including cloning, pulling, and retrieving commit information. It includes features for security, authentication, and error handling.

## Features

- **Secure Operations**: Enforces URL validation and a configurable host allowlist.
- **Authentication**: Supports HTTP basic auth, SSH keys, and token-based authentication.
- **Resilience**: Implements configurable timeouts and automatic retries for transient network failures.
- **Progress Reporting**: Provides real-time progress updates for long-running operations like `clone` and `pull`.
- **Repository Validation**: Includes functions to check for valid Git repositories and remote URLs.
- **Commit Information**: Allows for retrieval of commit metadata and file history.

## Usage

This package is used by:
- `cache` package: For cloning and pulling repositories for caching.
- `rule` package: For fetching rules from Git repositories.

## Interface Design

The core of the package is the `Repository` interface, which abstracts Git operations. This design allows for mock implementations in tests and supports different backend Git libraries.

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

- `NewClient(fs, config) -> Client`: Creates a new Git client.
- The `Repository` interface provides methods such as `Clone()`, `Pull()`, `GetLatestCommitHash()`, and `ValidateURL()`.
- The configuration struct allows for setting timeouts, authentication methods, progress handlers, and security options.