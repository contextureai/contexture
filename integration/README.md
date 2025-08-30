# Integration Test Suite

This package provides integration tests for the application, focusing on the interactions between internal packages and core functionality, without the overhead of full end-to-end binary execution.

## Purpose

These tests serve as a middle layer between unit and end-to-end tests by:
- Testing the interactions between multiple internal packages.
- Validating core business logic workflows without the CLI overhead.
- Providing faster feedback than full e2e tests while offering broader coverage than unit tests.

## Test Focus Areas

- **Project Management**: Configuration loading, validation, and persistence.
- **Rule Processing**: Integration between fetchers, parsers, processors, and validators.
- **Git Operations**: Repository access, caching, and rule fetching.
- **Format Generation**: End-to-end format processing.

## Testing Approach

- **In-Memory Filesystem**: Uses `afero.MemMapFs` for filesystem operations to avoid side effects and ensure fast, repeatable tests.
- **Real Components**: Tests the actual component implementations rather than mocks to validate real integration points.
- **Realistic Scenarios**: Tests common user workflows and edge cases at the component level.

### Integration Test Architecture

```mermaid
graph TB
    subgraph "Integration Test Layer"
        IT[Integration Tests]
        TH[Test Helpers]
        TD[Test Data]
        FS[afero.MemMapFs]
    end
    
    subgraph "Internal Packages Under Test"
        PM[internal/project]
        DM[internal/domain]
        GIT[internal/git]
        RULE[internal/rule]
        FMT[internal/format]
        VAL[internal/validation]
        CACHE[internal/cache]
        TPL[internal/template]
    end
    
    subgraph "Test Scenarios"
        CONFIG[Config Management]
        RULEPROC[Rule Processing]
        GITOPS[Git Operations]
        FMTGEN[Format Generation]
    end
    
    subgraph "Validation Layer"
        ASSERT[Assertions]
        EXPECT[Expectations]
        VERIFY[State Verification]
    end
    
    IT --> TH
    IT --> TD
    IT --> FS
    
    TH --> PM
    TH --> DM
    TH --> GIT
    TH --> RULE
    TH --> FMT
    TH --> VAL
    TH --> CACHE
    TH --> TPL
    
    PM --> CONFIG
    DM --> CONFIG
    VAL --> CONFIG
    
    RULE --> RULEPROC
    TPL --> RULEPROC
    VAL --> RULEPROC
    
    GIT --> GITOPS
    CACHE --> GITOPS
    RULE --> GITOPS
    
    FMT --> FMTGEN
    RULE --> FMTGEN
    TPL --> FMTGEN
    
    CONFIG --> ASSERT
    RULEPROC --> ASSERT
    GITOPS --> ASSERT
    FMTGEN --> ASSERT
    
    ASSERT --> EXPECT
    EXPECT --> VERIFY
    
    FS --> PM
    FS --> GIT
    FS --> CACHE
    
    style IT fill:#e1f5fe
    style PM fill:#f3e5f5
    style FS fill:#e8f5e8
    style ASSERT fill:#fff3e0
```

### Project Management Integration Flow

```mermaid
sequenceDiagram
    participant Test as Integration Test
    participant FS as afero.MemMapFs
    participant PM as ProjectManager
    participant Config as domain.Config
    participant Val as Validator
    participant Persist as Persistence
    
    Note over Test: Setup Phase
    Test->>FS: Create in-memory filesystem
    Test->>PM: NewManager(fs)
    Test->>FS: MkdirAll(workingDir)
    
    Note over Test: Config Initialization
    Test->>PM: InitConfig(dir, formats, location)
    PM->>Config: Create new config
    PM->>Val: Validate config structure
    Val-->>PM: Validation result
    PM->>Persist: Save config to filesystem
    Persist->>FS: Write .contexture.yaml
    PM-->>Test: Created config
    
    Note over Test: Config Loading
    Test->>PM: LoadConfig(workingDir)
    PM->>FS: Locate config file
    FS-->>PM: Config file path
    PM->>FS: Read config content
    FS-->>PM: YAML content
    PM->>Config: Parse YAML to config
    Config->>Val: Validate parsed config
    Val-->>Config: Validation result
    PM-->>Test: ConfigResult{config, location}
    
    Note over Test: Rule Management
    Test->>PM: AddRule(config, ruleRef)
    PM->>Config: Add rule to rules list
    Config->>Config: Check for duplicates
    Config->>Val: Validate rule reference
    Val-->>Config: Rule validation
    Config-->>PM: Updated config
    PM-->>Test: Success
    
    Test->>PM: HasRule(config, ruleID)
    PM->>Config: Search rules by ID
    Config-->>PM: Boolean result
    PM-->>Test: Rule exists
    
    Test->>PM: RemoveRule(config, ruleID)
    PM->>Config: Remove rule from list
    Config->>Config: Find and remove rule
    Config-->>PM: Updated config
    PM-->>Test: Success/Error
    
    Note over Test: Config Persistence
    Test->>PM: SaveConfig(config, location, dir)
    PM->>Config: Clean config for saving
    Config->>Config: Remove default values
    PM->>Persist: Write cleaned config
    Persist->>FS: Write YAML to file
    FS-->>Persist: Write result
    PM-->>Test: Save result
    
    Note over Test: Validation
    Test->>Test: Assert config state
    Test->>FS: Verify file existence
    Test->>FS: Read and validate content
```

### Component Interaction Matrix

```mermaid
graph LR
    subgraph "Core Domain"
        CONFIG[Config Management]
        RULES[Rule References]
        FORMATS[Format Definitions]
    end
    
    subgraph "Storage Layer"
        FILESYSTEM[File Operations]
        GITREPO[Git Repositories]
        CACHE[Cache Storage]
    end
    
    subgraph "Processing Layer"
        VALIDATION[Rule Validation]
        TEMPLATE[Template Processing]
        GENERATION[Format Generation]
    end
    
    subgraph "Integration Points"
        direction TB
        I1["project ↔ domain<br/>Config CRUD"]
        I2["project ↔ validation<br/>Config Validation"]
        I3["rule ↔ git<br/>Remote Fetching"]
        I4["rule ↔ cache<br/>Repository Caching"]
        I5["format ↔ template<br/>Content Processing"]
        I6["validation ↔ domain<br/>Business Rules"]
        I7["cache ↔ git<br/>Repository Operations"]
        I8["template ↔ rule<br/>Content Templates"]
    end
    
    CONFIG --> I1
    CONFIG --> I2
    RULES --> I3
    RULES --> I4
    FORMATS --> I5
    
    FILESYSTEM --> I1
    GITREPO --> I3
    GITREPO --> I7
    CACHE --> I4
    CACHE --> I7
    
    VALIDATION --> I2
    VALIDATION --> I6
    TEMPLATE --> I5
    TEMPLATE --> I8
    GENERATION --> I5
    
    I1 --> FILESYSTEM
    I3 --> RULES
    I4 --> CACHE
    I5 --> GENERATION
    I6 --> VALIDATION
    I7 --> CACHE
    I8 --> TEMPLATE
    
    style I1 fill:#ffebee
    style I2 fill:#f3e5f5
    style I3 fill:#e8f5e8
    style I4 fill:#fff3e0
    style I5 fill:#e1f5fe
    style I6 fill:#fce4ec
    style I7 fill:#f9fbe7
    style I8 fill:#e0f2f1
```

### Test Execution and Validation Flow

```mermaid
flowchart TD
    START([Integration Test Start]) --> SETUP[Setup In-Memory Environment]
    
    SETUP --> INITFS[Initialize afero.MemMapFs]
    INITFS --> CREATEDIR[Create Test Directory Structure]
    CREATEDIR --> INITPM[Initialize ProjectManager]
    
    INITPM --> TESTTYPE{Test Type}
    
    TESTTYPE -->|Config Test| CONFIGFLOW[Config Management Flow]
    TESTTYPE -->|Rule Test| RULEFLOW[Rule Processing Flow]
    TESTTYPE -->|Format Test| FORMATFLOW[Format Generation Flow]
    TESTTYPE -->|Error Test| ERRORFLOW[Error Handling Flow]
    
    CONFIGFLOW --> INITCONFIG[InitConfig]
    INITCONFIG --> LOADCONFIG[LoadConfig]
    LOADCONFIG --> VALIDATE1[Validate Config State]
    
    RULEFLOW --> ADDRULE[AddRule]
    ADDRULE --> HASRULE[HasRule]
    HASRULE --> REMOVERULE[RemoveRule]
    REMOVERULE --> VALIDATE2[Validate Rule State]
    
    FORMATFLOW --> ENABLEFORMAT[Enable Formats]
    ENABLEFORMAT --> GETFORMATS[GetEnabledFormats]
    GETFORMATS --> FORMATCONFIG[Format Configuration]
    FORMATCONFIG --> VALIDATE3[Validate Format State]
    
    ERRORFLOW --> INTRODUCEERROR[Introduce Error Condition]
    INTRODUCEERROR --> TESTERROR[Test Error Response]
    TESTERROR --> VALIDATEERROR[Validate Error Handling]
    
    VALIDATE1 --> PERSIST[Save Configuration]
    VALIDATE2 --> PERSIST
    VALIDATE3 --> PERSIST
    VALIDATEERROR --> PERSIST
    
    PERSIST --> RELOAD[Reload and Verify]
    RELOAD --> ASSERT[Assert Expected State]
    
    ASSERT --> CLEANUP[Cleanup Test Environment]
    CLEANUP --> END([Test Complete])
    
    subgraph "Validation Strategies"
        V1["State Assertions<br/>• Config structure<br/>• Rule presence<br/>• Format settings"]
        V2["Persistence Verification<br/>• File system state<br/>• Config file content<br/>• Data integrity"]
        V3["Integration Verification<br/>• Component interactions<br/>• Data flow correctness<br/>• Error propagation"]
    end
    
    subgraph "Error Scenarios"
        E1["Configuration Errors<br/>• Invalid structure<br/>• Missing fields<br/>• Type mismatches"]
        E2["Rule Management Errors<br/>• Duplicate rules<br/>• Missing rules<br/>• Invalid references"]
        E3["Persistence Errors<br/>• Write failures<br/>• Read errors<br/>• Directory issues"]
    end
    
    style START fill:#e1f5fe
    style TESTTYPE fill:#f3e5f5
    style PERSIST fill:#e8f5e8
    style END fill:#fff3e0
```

## Relationship to Other Tests

- **Compared to Unit Tests**: These tests have a broader scope, focusing on the interactions between components rather than isolated behavior.
- **Compared to E2E Tests**: These tests are faster as they don't involve CLI overhead. They focus on business logic rather than the user interface.