# Rule Package

This package provides comprehensive rule processing functionality for Contexture, implementing a modular architecture with separate components for fetching, parsing, processing, and validating rules from various sources.

## Purpose

The rule package serves as the core engine for rule management, handling everything from retrieving rules from local files or Git repositories to processing templates and validating rule structure. It provides a clean interface for rule operations while supporting multiple data sources and processing requirements.

## Architecture

### Component-Based Design
- **Fetcher**: Retrieves rules from local files and Git repositories with intelligent caching
- **Parser**: Processes rule content including frontmatter parsing and metadata extraction
- **Processor**: Handles template processing, variable substitution, and context management
- **Validator**: Validates rule structure, content, and business logic constraints

### Rule Sources
- **Local Files**: Direct access to filesystem-based rules for development workflows
- **Git Repositories**: Remote rule fetching with repository caching and branch/commit support
- **Composite Fetching**: Intelligent routing between local and Git sources based on rule ID format

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

## Key Features

- **Multi-Source Support**: Seamless handling of both local and remote rule sources
- **Template Processing**: Dynamic content generation using Go templates with custom functions
- **Variable Management**: Context-aware variable substitution and dependency tracking
- **Repository Caching**: Efficient Git repository caching for improved performance
- **Rule ID Parsing**: Sophisticated parsing of complex rule reference formats
- **Attribution Generation**: Automatic attribution information for rule sources and metadata

## Rule Processing Pipeline

1. **Fetching**: Retrieve rule content from configured sources
2. **Parsing**: Extract frontmatter metadata and rule body content
3. **Validation**: Verify rule structure and business rule compliance
4. **Processing**: Apply template processing with variable context
5. **Attribution**: Generate source attribution for tracking and compliance

## Rule ID Formats

The package supports various rule ID formats:
- Simple paths: `path/to/rule.md`
- Contexture format: `[contexture:path/to/rule]`
- Repository format: `[contexture(repo-url):path,branch]`
- Local references: Direct filesystem paths

## Usage Within Project

This package is used by:
- **Commands Package**: All rule-related CLI commands use this package for rule operations
- **Format Package**: Rule processing provides processed rules for format generation
- **Build System**: Template processing and rule compilation for output generation

## API

### Core Interfaces
- `Fetcher`: `FetchRule()`, `FetchRules()`, `ParseRuleID()`, `ListAvailableRules()`
- `Parser`: `ParseRule()`, `ParseContent()`, `ValidateRule()`
- `Processor`: `ProcessRule()`, `ProcessRules()`, `ProcessTemplate()`, `GenerateAttribution()`

### Factory Functions
- `NewFetcher(fs, repository, config)`: Creates composite fetcher with Git and local support
- `NewParser()`: Creates rule parser with frontmatter support
- `NewProcessor(templateEngine, validator)`: Creates rule processor with template engine