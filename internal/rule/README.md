# Rule Package

This package handles rule processing, including fetching, parsing, and validating rules from various sources.

## Architecture

The package is designed with a component-based architecture:
- **Fetcher**: A composite fetcher that retrieves rules from local files or Git repositories, with support for caching.
- **Parser**: Parses rule content, extracting frontmatter and metadata.
- **Processor**: Processes templates, substitutes variables, and manages context.
- **Validator**: Validates the rule's structure, content, and business logic.

### Component Interaction Diagram

```mermaid
graph TB
    subgraph "Commands Layer"
        CMD[Command Actions]
    end
    
    subgraph "Rule Package"
        CF[CompositeFetcher]
        LF[LocalFetcher] 
        GF[GitFetcher]
        P[Parser]
        PR[Processor]
        V[Validator]
        TP[TemplateProcessor]
        VM[VariableManager]
        IDP[IDParser]
    end
    
    subgraph "External Dependencies"
        CACHE[Cache Package]
        GIT[Git Package]
        FS[Filesystem]
        TEMPLATE[Template Package]
        DOMAIN[Domain Package]
        VAL[Validation Package]
    end
    
    CMD --> CF
    CF --> LF
    CF --> GF
    CF --> IDP
    
    LF --> FS
    GF --> CACHE
    GF --> GIT
    CACHE --> GIT
    
    CF --> P
    P --> V
    P --> VAL
    V --> DOMAIN
    
    P --> PR
    PR --> TP
    PR --> VM
    TP --> TEMPLATE
    VM --> TEMPLATE
    
    style CF fill:#e1f5fe
    style P fill:#f3e5f5
    style PR fill:#e8f5e8
    style V fill:#fff3e0
```

### Rule Processing Pipeline

```mermaid
sequenceDiagram
    participant Client as Client Code
    participant CF as CompositeFetcher
    participant LF as LocalFetcher
    participant GF as GitFetcher
    participant P as Parser
    participant V as Validator
    participant PR as Processor
    participant TP as TemplateProcessor
    
    Client->>CF: FetchRule(ruleID)
    
    alt Local Rule ID
        CF->>LF: FetchRule(ruleID)
        LF->>P: ParseRule(content)
        P->>V: ValidateRule(rule)
        V-->>P: Validation Result
        P-->>CF: Parsed Rule
    else Git Rule ID
        CF->>GF: FetchRule(ruleID)
        GF->>P: ParseRule(content)
        P->>V: ValidateRule(rule)
        V-->>P: Validation Result
        P-->>CF: Parsed Rule
    end
    
    CF-->>Client: Rule
    
    Client->>PR: ProcessRule(rule, context)
    PR->>TP: ProcessTemplate(content, vars)
    TP-->>PR: Processed Content
    PR-->>Client: Processed Rule
```

### Fetching Strategy Flow

```mermaid
flowchart TD
    START([Rule ID Input]) --> CHECK{Check Rule ID Format}
    
    CHECK -->|Local Path| LOCAL[Use LocalFetcher]
    CHECK -->|Contexture Format| PARSE[Parse Rule ID]
    CHECK -->|Simple ID| DEFAULT[Use Default Repository]
    
    PARSE --> EXTRACT{Extract Source Info}
    EXTRACT -->|Has Repository| REPO[Use GitFetcher with Custom Repo]
    EXTRACT -->|No Repository| DEFAULT
    
    LOCAL --> READFILE[Read from Filesystem]
    REPO --> GITFETCH[Fetch from Git Repository]
    DEFAULT --> GITFETCH
    
    READFILE --> PARSECONTENT[Parse Rule Content]
    GITFETCH --> CACHE{Check Cache}
    CACHE -->|Cache Hit| CACHED[Use Cached Repository]
    CACHE -->|Cache Miss| CLONE[Clone Repository]
    
    CACHED --> PARSECONTENT
    CLONE --> PARSECONTENT
    
    PARSECONTENT --> VALIDATE[Validate Rule]
    VALIDATE --> RESULT([Return Parsed Rule])
    
    style LOCAL fill:#e3f2fd
    style GITFETCH fill:#e8f5e8
    style CACHE fill:#fff3e0
    style VALIDATE fill:#fce4ec
```

## Features

- **Multi-Source Support**: Handles rules from both local files and remote Git repositories.
- **Template Processing**: Uses Go templates with custom functions for dynamic content generation.
- **Variable Management**: Supports context-aware variable substitution.
- **Repository Caching**: Caches Git repositories for improved performance.
- **Rule ID Parsing**: Parses various rule ID formats.
- **Attribution Generation**: Automatically generates attribution for rule sources.

## Usage

This package is used by:
- The `commands` package for all rule-related operations.
- The `format` package, which uses processed rules to generate output.

## API

- **Interfaces**: Provides `Fetcher`, `Parser`, and `Processor` interfaces.
- **Factory Functions**: `NewFetcher()`, `NewParser()`, and `NewProcessor()` for creating instances of the components.