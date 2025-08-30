# Cache Package

This package provides a simple repository caching mechanism for Contexture, enabling cross-session persistence of Git repositories with human-readable cache directory names.

## Purpose

The cache package improves performance by storing cloned repositories locally, avoiding redundant clone operations across CLI sessions. It manages repository updates intelligently and provides clean, predictable cache keys for easy maintenance.

## Key Features

- **Cross-Session Persistence**: Repositories remain cached between CLI invocations
- **Human-Readable Cache Keys**: Generated from repository URLs and Git references (e.g., `github.com_user_repo-main`)
- **Smart Updates**: Supports both cached retrieval and forced updates with git pull
- **URL Format Support**: Handles both HTTPS and SSH Git URLs with proper parsing
- **Automatic Cleanup**: Failed clones are cleaned up automatically
- **Temp Directory Storage**: Uses system temp directory for cache storage

## Cache Key Generation

The package creates predictable cache directory names by:
- Parsing repository URLs (both `https://` and `git@` formats)
- Extracting hostname and path components
- Sanitizing special characters to filesystem-safe names
- Appending the Git reference (branch/tag/commit)

### Cache Operations Flow

```mermaid
flowchart TD
    START([Repository Request]) --> CHECK{Cache Exists?}
    
    CHECK -->|No| CLONE[Clone Repository]
    CHECK -->|Yes| UPDATE{Update Requested?}
    
    UPDATE -->|No| HIT[Cache Hit - Return Path]
    UPDATE -->|Yes| PULL[Pull Updates]
    
    CLONE --> MKDIR[Create Cache Directory]
    MKDIR --> GITCLONE[Git Clone Operation]
    GITCLONE --> VALIDATE{Clone Success?}
    
    VALIDATE -->|Success| CACHED[Repository Cached]
    VALIDATE -->|Failure| CLEANUP[Cleanup Failed Clone]
    CLEANUP --> ERROR[Return Error]
    
    PULL --> PULLVALIDATE{Pull Success?}
    PULLVALIDATE -->|Success| UPDATED[Repository Updated]
    PULLVALIDATE -->|Failure| WARNING[Log Warning - Use Cached]
    
    WARNING --> HIT
    UPDATED --> HIT
    CACHED --> HIT
    HIT --> SUCCESS([Return Cache Path])
    
    style CHECK fill:#e1f5fe
    style UPDATE fill:#f3e5f5
    style CLONE fill:#e8f5e8
    style PULL fill:#fff3e0
    style SUCCESS fill:#c8e6c9
    style ERROR fill:#ffcdd2
```

### Cache Integration Architecture

```mermaid
graph TB
    subgraph "Client Layer"
        RULE[Rule Fetcher]
        TESTS[Integration Tests]
    end
    
    subgraph "Cache Package"
        CACHE[SimpleCache]
        KEYGEN[Cache Key Generator]
        VALIDATOR[Repository Validator]
    end
    
    subgraph "Dependencies"
        GIT[Git Repository Interface]
        FS[Filesystem Interface]
        TEMPDIR[System Temp Directory]
    end
    
    subgraph "Storage"
        CACHEDIR[Cache Directories]
        REPOS[Cached Repositories]
    end
    
    RULE --> CACHE
    TESTS --> CACHE
    
    CACHE --> KEYGEN
    CACHE --> VALIDATOR
    CACHE --> GIT
    CACHE --> FS
    
    KEYGEN --> TEMPDIR
    GIT --> REPOS
    FS --> CACHEDIR
    VALIDATOR --> REPOS
    
    style CACHE fill:#e1f5fe
    style KEYGEN fill:#f3e5f5
    style VALIDATOR fill:#e8f5e8
    style GIT fill:#fff3e0
```

### Cache Key Generation Process

```mermaid
sequenceDiagram
    participant Client as Client Code
    participant Cache as SimpleCache
    participant KeyGen as Key Generator
    participant FS as Filesystem
    
    Client->>Cache: GetRepository(url, ref)
    Cache->>KeyGen: generateCacheKey(url, ref)
    
    alt HTTPS URL
        KeyGen->>KeyGen: Parse URL components
        KeyGen->>KeyGen: Extract host and path
        KeyGen->>KeyGen: Sanitize path
        KeyGen->>KeyGen: Format: host_path-ref
    else SSH URL
        KeyGen->>KeyGen: Parse git@ format
        KeyGen->>KeyGen: Extract host and repo
        KeyGen->>KeyGen: Remove .git suffix
        KeyGen->>KeyGen: Format: host_repo-ref
    else Fallback
        KeyGen->>KeyGen: Sanitize entire URL
        KeyGen->>KeyGen: Replace special chars
        KeyGen->>KeyGen: Format: sanitized-ref
    end
    
    KeyGen-->>Cache: Cache Key
    Cache->>FS: Check cache path exists
    FS-->>Cache: Exists result
    Cache-->>Client: Repository path
    
    note over KeyGen: Examples:<br/>github.com_user_repo-main<br/>gitlab.com_org_project-develop
```

## Usage Within Project

This package is used by:
- **Rule Package**: Git fetcher uses caching for repository-based rule retrieval
- **Integration Tests**: Repository operations leverage caching for test performance

## API

- `NewSimpleCache(fs, repository)`: Creates a cache instance with filesystem and git repository dependencies
- `GetRepository(ctx, repoURL, gitRef)`: Returns cached repository path or clones if not cached
- `GetRepositoryWithUpdate(ctx, repoURL, gitRef)`: Forces updates to cached repositories with latest changes