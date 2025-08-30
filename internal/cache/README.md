# Cache Package

This package provides a simple, cross-session caching mechanism for Git repositories. It uses human-readable directory names and stores repositories in the system's temporary directory.

## Features

- **Cross-Session Persistence**: Repositories are cached between CLI invocations, improving performance by avoiding redundant clones.
- **Human-Readable Cache Keys**: Cache directories are named based on the repository URL and Git reference (e.g., `github.com_user_repo-main`).
- **Smart Updates**: Supports both retrieving from the cache and forcing an update via `git pull`.
- **URL Support**: Handles both HTTPS and SSH Git URLs.
- **Automatic Cleanup**: Automatically removes failed clone directories.

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

## Usage

This package is used by:
- The `rule` package for caching repositories when fetching rules.
- Integration tests to improve performance.

## API

- `NewSimpleCache(fs, repository) -> SimpleCache`: Creates a new cache instance.
- `GetRepository(ctx, repoURL, gitRef) -> string`: Returns the path to a cached repository, cloning it if it's not already cached.
- `GetRepositoryWithUpdate(ctx, repoURL, gitRef) -> string`: Forces an update of a cached repository by pulling the latest changes.