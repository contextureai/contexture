# Validation Package

This package provides centralized validation logic for all Contexture types using a consistent, robust approach built on the validator/v10 library with custom validation tags and comprehensive error formatting.

## Purpose

The validation package serves as the single source of truth for all data validation within the Contexture CLI. It ensures data integrity, provides meaningful error messages, and maintains consistent validation rules across the entire application.

## Key Features

- **Comprehensive Rule Validation**: Validates rule structure, content, metadata, and business logic constraints
- **Project Configuration Validation**: Ensures project configs are well-formed with valid formats and unique rule IDs  
- **Custom Validation Tags**: Extended validator with domain-specific tags like `ruleref`, `ruleid`, `formattype`, and `giturl`
- **Batch Processing**: Efficiently validates multiple rules with detailed result reporting
- **Context-Aware**: Supports validation with cancellation context for long-running operations
- **Structured Error Handling**: Consistent error formatting with field-specific messages and error codes

## Validation Capabilities

- **Rules**: Content validation, ID format checking, tag uniqueness, trigger configuration
- **Projects**: Format configuration validation, rule ID uniqueness, enabled format requirements
- **Rule References**: ID format and structure validation
- **Git URLs**: Repository URL format validation
- **Format Configurations**: Type validation and configuration structure

### Validation System Architecture

```mermaid
graph TB
    subgraph "Validation Package"
        VALIDATOR[Default Validator]
        CUSTOMTAGS[Custom Tag Registry]
        CONSTRAINTS[Business Constraints]
        FORMATTER[Error Formatter]
    end
    
    subgraph "Validator/v10 Library"
        V10[validator/v10]
        STRUCTVAL[Struct Validation]
        FIELDVAL[Field Validation]
        TAGVAL[Tag Validation]
    end
    
    subgraph "Custom Validators"
        RULEID[Rule ID Validator]
        RULEREF[Rule Ref Validator]
        FORMATTYPE[Format Type Validator]
        GITURL[Git URL Validator]
        CONTEXTPATH[Context Path Validator]
    end
    
    subgraph "Domain Integration"
        DOMAIN[Domain Package]
        ERRORS[Errors Package]
        CONSTANTS[Business Constants]
    end
    
    subgraph "Client Usage"
        PROJECT[Project Validation]
        RULE[Rule Validation]
        CONFIG[Config Validation]
        BATCH[Batch Validation]
    end
    
    VALIDATOR --> V10
    VALIDATOR --> CUSTOMTAGS
    VALIDATOR --> FORMATTER
    VALIDATOR --> CONSTRAINTS
    
    CUSTOMTAGS --> RULEID
    CUSTOMTAGS --> RULEREF
    CUSTOMTAGS --> FORMATTYPE
    CUSTOMTAGS --> GITURL
    CUSTOMTAGS --> CONTEXTPATH
    
    V10 --> STRUCTVAL
    V10 --> FIELDVAL
    V10 --> TAGVAL
    
    VALIDATOR --> DOMAIN
    FORMATTER --> ERRORS
    CONSTRAINTS --> CONSTANTS
    
    PROJECT --> VALIDATOR
    RULE --> VALIDATOR
    CONFIG --> VALIDATOR
    BATCH --> VALIDATOR
    
    style VALIDATOR fill:#e1f5fe
    style CUSTOMTAGS fill:#f3e5f5
    style V10 fill:#e8f5e8
    style DOMAIN fill:#fff3e0
```

### Rule Validation Pipeline

```mermaid
sequenceDiagram
    participant Client as Client Code
    participant Val as Validator
    participant V10 as validator/v10
    participant Custom as Custom Validators
    participant Domain as Domain Logic
    participant Result as ValidationResult
    
    Client->>Val: ValidateRule(rule)
    
    Val->>Result: Create ValidationResult
    
    alt Rule is nil
        Val->>Result: AddError("rule", "cannot be nil")
        Val-->>Client: Invalid result
    else Rule exists
        Val->>V10: Struct(rule)
        
        loop For each field with tags
            V10->>Custom: Call custom validator
            Custom->>Custom: Apply business rules
            Custom-->>V10: Validation result
        end
        
        V10-->>Val: Struct validation result
        
        alt Struct validation failed
            Val->>Val: Parse validation errors
            Val->>Result: AddError for each failure
        end
        
        Val->>Domain: Apply business rules
        
        Domain->>Domain: Check content not empty
        Domain->>Domain: Validate tag uniqueness
        Domain->>Domain: Check rule ID format
        Domain->>Domain: Validate trigger configuration
        
        Domain-->>Val: Business rule results
        
        loop For each business rule violation
            Val->>Result: AddError with context
        end
        
        Val-->>Client: Complete validation result
    end
    
    note over Val: Validation includes:<br/>• Struct tag validation<br/>• Custom business rules<br/>• Field-specific checks<br/>• Error code generation
```

### Custom Tag System

```mermaid
flowchart TD
    START([Validation Request]) --> REGISTER[Register Custom Tags]
    
    REGISTER --> RULEID_TAG[ruleid tag]
    REGISTER --> RULEREF_TAG[ruleref tag]
    REGISTER --> FORMAT_TAG[formattype tag]
    REGISTER --> GITURL_TAG[giturl tag]
    REGISTER --> PATH_TAG[contexturepath tag]
    
    RULEID_TAG --> RULEID_LOGIC{Rule ID Format Check}
    RULEID_LOGIC -->|Valid| RULEID_OK[Rule ID Valid]
    RULEID_LOGIC -->|Invalid| RULEID_ERROR[Rule ID Error]
    
    RULEREF_TAG --> RULEREF_LOGIC{Rule Ref Structure}
    RULEREF_LOGIC -->|Valid| RULEREF_OK[Rule Ref Valid]
    RULEREF_LOGIC -->|Invalid| RULEREF_ERROR[Rule Ref Error]
    
    FORMAT_TAG --> FORMAT_LOGIC{Format Type Check}
    FORMAT_LOGIC -->|Valid| FORMAT_OK[Format Valid]
    FORMAT_LOGIC -->|Invalid| FORMAT_ERROR[Format Error]
    
    GITURL_TAG --> GITURL_LOGIC{Git URL Format}
    GITURL_LOGIC -->|Valid| GITURL_OK[Git URL Valid]
    GITURL_LOGIC -->|Invalid| GITURL_ERROR[Git URL Error]
    
    PATH_TAG --> PATH_LOGIC{Path Validation}
    PATH_LOGIC -->|Valid| PATH_OK[Path Valid]
    PATH_LOGIC -->|Invalid| PATH_ERROR[Path Error]
    
    RULEID_OK --> SUCCESS[Validation Success]
    RULEREF_OK --> SUCCESS
    FORMAT_OK --> SUCCESS
    GITURL_OK --> SUCCESS
    PATH_OK --> SUCCESS
    
    RULEID_ERROR --> FAILURE[Validation Failure]
    RULEREF_ERROR --> FAILURE
    FORMAT_ERROR --> FAILURE
    GITURL_ERROR --> FAILURE
    PATH_ERROR --> FAILURE
    
    style REGISTER fill:#e1f5fe
    style RULEID_LOGIC fill:#f3e5f5
    style FORMAT_LOGIC fill:#e8f5e8
    style SUCCESS fill:#c8e6c9
    style FAILURE fill:#ffcdd2
```

## Usage Within Project

This package is used by:
- **Project Package**: For validating project configuration files
- **Rule Package**: For rule validation and parsing operations  
- **Parser Package**: For validating parsed rule data

## API

- `NewValidator()`: Creates a new validator instance with custom tags registered
- `ValidateRule()`: Validates a single rule with detailed results
- `ValidateRules()`: Batch validates multiple rules
- `ValidateProject()`: Validates complete project configurations
- `ValidateRuleRef()`, `ValidateRuleID()`, `ValidateGitURL()`: Specific validation functions