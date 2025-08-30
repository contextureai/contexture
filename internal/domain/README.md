# Domain Package

This package contains the core business logic and domain models for the application, following Domain-Driven Design (DDD) principles. It defines the central entities, value objects, and business rules.

## Core Components

- **Entities**: The primary business objects, `Rule` and `Project`.
- **Value Objects**: Type-safe structures for concepts like `RuleRef`, `FormatConfig`, and `RuleTrigger`.
- **Business Constants**: Defines configuration limits, format types, and trigger types.

## Domain-Driven Design Principles

- **Infrastructure Independence**: The domain package has no dependencies on infrastructure concerns like databases or frameworks.
- **Centralized Business Rules**: All business logic is located within this package.

### Domain Model Relationships

```mermaid
erDiagram
    Project ||--o{ RuleRef : contains
    Project ||--o{ FormatConfig : has
    Project {
        string Name
        string Description
        string Version
        RuleRef[] Rules
        FormatConfig[] Formats
    }
    
    RuleRef ||--|| Rule : references
    RuleRef {
        string ID
        string Source
        string Ref
    }
    
    Rule ||--o| RuleTrigger : has
    Rule ||--o{ string : has-tags
    Rule {
        string ID
        string Title
        string Description
        string Content
        string[] Tags
        RuleTrigger Trigger
        map Variables
    }
    
    RuleTrigger {
        TriggerType Type
        string[] Globs
        string[] Extensions
        string Model
    }
    
    FormatConfig {
        FormatType Type
        bool Enabled
        string OutputPath
        map Options
    }
    
    ValidationResult ||--o{ Error : contains
    ValidationResult ||--o{ ValidationWarning : contains
    ValidationResult {
        bool Valid
        Error[] Errors
        ValidationWarning[] Warnings
    }
    
    ProcessedRule ||--|| Rule : processes
    ProcessedRule {
        string ProcessedContent
        string Attribution
        map ResolvedVariables
        Rule OriginalRule
    }
```

### Business Logic Flow

```mermaid
flowchart TB
    subgraph "Domain Layer"
        RULE[Rule Entity]
        PROJECT[Project Entity]
        RULEREF[RuleRef Value Object]
        FORMATCONFIG[FormatConfig Value Object]
        TRIGGER[RuleTrigger Value Object]
    end
    
    subgraph "Business Operations"
        VALIDATE[Validation Logic]
        PROCESS[Processing Logic]
        TRANSFORM[Transformation Logic]
    end
    
    subgraph "Value Objects & Constraints"
        CONSTANTS[Business Constants]
        LIMITS[Configuration Limits]
        TYPES[Type Definitions]
        PATTERNS[Pattern Validation]
    end
    
    subgraph "External Interactions"
        VAL_PKG[Validation Package]
        RULE_PKG[Rule Package]
        FORMAT_PKG[Format Package]
        CMD_PKG[Commands Package]
    end
    
    PROJECT --> RULEREF
    PROJECT --> FORMATCONFIG
    RULE --> TRIGGER
    
    RULEREF --> VALIDATE
    RULE --> VALIDATE
    PROJECT --> VALIDATE
    
    RULE --> PROCESS
    PROCESS --> TRANSFORM
    
    VALIDATE --> CONSTANTS
    VALIDATE --> LIMITS
    VALIDATE --> TYPES
    VALIDATE --> PATTERNS
    
    VAL_PKG --> VALIDATE
    RULE_PKG --> PROCESS
    FORMAT_PKG --> FORMATCONFIG
    CMD_PKG --> PROJECT
    CMD_PKG --> RULE
    
    style RULE fill:#e1f5fe
    style PROJECT fill:#f3e5f5
    style VALIDATE fill:#e8f5e8
    style CONSTANTS fill:#fff3e0
```

### Validation and Business Rules

```mermaid
sequenceDiagram
    participant Client as Client Code
    participant Rule as Rule Entity
    participant Trigger as RuleTrigger
    participant Validator as Domain Validator
    participant Constants as Business Constants
    
    Client->>Rule: Create/Update Rule
    Rule->>Rule: Apply Business Rules
    
    Rule->>Trigger: Validate Trigger
    Trigger->>Constants: Check Trigger Types
    Constants-->>Trigger: Valid Types
    Trigger-->>Rule: Validation Result
    
    Rule->>Validator: Validate Content
    Validator->>Constants: Check Length Limits
    Validator->>Constants: Check Format Rules
    Constants-->>Validator: Constraints
    Validator-->>Rule: Content Validation
    
    Rule->>Rule: Validate Tags
    Rule->>Constants: Check Tag Limits
    Constants-->>Rule: Tag Constraints
    
    Rule->>Rule: Check Uniqueness Rules
    Rule-->>Client: Validation Result
    
    alt Validation Success
        Client->>Rule: Proceed with Operation
        Rule-->>Client: Success
    else Validation Failure
        Rule-->>Client: Business Rule Violations
    end
```

## Usage

This package is used by most other packages in the application, including:
- `validation`: For validating domain entities.
- `commands`: For working with domain entities in CLI operations.
- `format`: For using domain interfaces and entities in format implementations.
- `rule`: For processing domain models.
- `project`: For managing configurations of domain entities.