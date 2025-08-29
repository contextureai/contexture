# Domain Package

This package contains the core business logic and domain models for Contexture, implementing Domain-Driven Design principles. It defines the essential entities, value objects, and interfaces that form the heart of the application's business rules.

## Purpose

The domain package serves as the central repository for all business logic, maintaining independence from infrastructure concerns. It establishes the fundamental concepts, constraints, and behaviors that define what Contexture does and how it operates.

## Key Components

- **Core Entities**: `Rule` and `Project` representing the primary business objects
- **Value Objects**: Type-safe structures like `RuleRef`, `FormatConfig`, and trigger types
- **Domain Interfaces**: Format operations and validation contracts
- **Business Constants**: Configuration limits, format types, and trigger definitions
- **Domain Services**: Configuration management and rule tree operations

## Domain-Driven Design Principles

- **Ubiquitous Language**: Consistent terminology used throughout the codebase
- **Rich Domain Models**: Entities contain behavior, not just data
- **Infrastructure Independence**: No dependencies on databases, frameworks, or external services
- **Business Rule Centralization**: All business logic resides within domain boundaries

## Core Entities

- **Rule**: Represents a contexture rule with content, metadata, triggers, and validation logic
- **Project**: Manages project configuration including rules, formats, and validation constraints
- **Rule Tree**: Hierarchical organization of rules for efficient processing

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

## Usage Within Project

This package is used extensively throughout the application:
- **Validation Package**: Validates domain entities and enforces business rules
- **Commands Package**: All CLI operations work with domain entities  
- **Format Package**: Format implementations use domain interfaces and entities
- **Rule Package**: Rule processing operations center around domain models
- **Project Package**: Configuration management operates on domain entities

## Design Benefits

- **Testability**: Pure domain logic enables comprehensive unit testing
- **Maintainability**: Business rules are centralized and clearly defined
- **Flexibility**: Infrastructure can change without affecting business logic
- **Clarity**: Domain concepts are explicitly modeled and named