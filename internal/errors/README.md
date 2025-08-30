# Errors Package

This package provides unified error handling for the Contexture CLI with structured error types, user-friendly messaging, and consistent error categorization across the entire application.

## Purpose

The errors package standardizes error handling throughout Contexture, providing rich error information that helps both users and developers understand what went wrong and how to fix it. It bridges the gap between technical errors and user-friendly messaging.

## Key Features

- **Structured Error Types**: Rich error objects with operation context, error kinds, and exit codes
- **User-Friendly Messaging**: Clear, actionable error messages with helpful suggestions
- **Error Classification**: Categorizes errors into domains (validation, network, permission, etc.)
- **Exit Code Management**: Standardized exit codes for different error types
- **Retryable Error Detection**: Automatic identification of transient failures
- **Terminal-Aware Display**: Color-coded error output with proper terminal detection

## Error Categories

- **Validation**: Data validation failures with field-specific context
- **Network**: Connection issues, timeouts, and network-related failures  
- **Permission**: Access denied and authorization failures
- **Configuration**: Invalid configuration files or settings
- **Repository**: Git repository operations and access issues
- **Format**: Template parsing and rendering errors
- **Not Found**: Missing resources or files

### Error Type Hierarchy

```mermaid
classDiagram
    class Error {
        +string Op
        +ErrorKind Kind
        +ErrorCode Code
        +error Err
        +string Message
        +[]string Suggestions
        +string Field
        +Error() string
        +Unwrap() error
        +ExitCode() int
        +IsRetryable() bool
    }
    
    class ErrorKind {
        <<enumeration>>
        KindOther
        KindNotFound
        KindValidation
        KindPermission
        KindNetwork
        KindConfig
        KindFormat
        KindRepository
        KindTimeout
        KindCanceled
    }
    
    class ErrorCode {
        <<enumeration>>
        ExitSuccess
        ExitError
        ExitUsageError
        ExitConfigError
        ExitPermError
        ExitNetworkError
        ExitNotFound
        ExitValidation
        ExitFormat
    }
    
    Error --> ErrorKind
    Error --> ErrorCode
    
    note for Error "Implements error interface\nProvides user-friendly display\nSupports error chaining"
```

### Error Processing Pipeline

```mermaid
flowchart TD
    START([Error Occurs]) --> DETECT[Detect Error Type]
    
    DETECT --> CLASSIFY{Classify Error}
    
    CLASSIFY -->|System Error| WRAP[Wrap with Context]
    CLASSIFY -->|Business Error| DOMAIN[Create Domain Error]
    CLASSIFY -->|Validation Error| VALIDATE[Create Validation Error]
    
    WRAP --> ENHANCE[Enhance with Metadata]
    DOMAIN --> ENHANCE
    VALIDATE --> ENHANCE
    
    ENHANCE --> SUGGESTIONS[Add Suggestions]
    SUGGESTIONS --> EXITCODE[Determine Exit Code]
    
    EXITCODE --> DISPLAY{Display Mode}
    
    DISPLAY -->|Terminal| COLOR[Apply Color Formatting]
    DISPLAY -->|Non-Terminal| PLAIN[Plain Text Output]
    
    COLOR --> OUTPUT[Display to User]
    PLAIN --> OUTPUT
    
    OUTPUT --> LOG{Debug Mode?}
    
    LOG -->|Yes| DETAILS[Show Technical Details]
    LOG -->|No| END[Error Displayed]
    
    DETAILS --> END
    
    style DETECT fill:#e1f5fe
    style CLASSIFY fill:#f3e5f5
    style ENHANCE fill:#e8f5e8
    style DISPLAY fill:#fff3e0
    style END fill:#c8e6c9
```

### Error Flow Through System

```mermaid
sequenceDiagram
    participant Comp as Component
    participant Err as Error Package
    participant Display as Display System
    participant User as User
    
    Comp->>Err: Original error occurs
    
    alt System Error
        Err->>Err: Wrap(err, operation)
        Err->>Err: Detect error kind
        Err->>Err: Map to exit code
    else Business Logic Error
        Err->>Err: Create structured error
        Err->>Err: Add field context
        Err->>Err: Set validation code
    end
    
    Err->>Err: Add helpful suggestions
    Err->>Err: Determine retryable status
    
    Comp->>Display: Display(error)
    
    Display->>Display: Check terminal capabilities
    
    alt Color Terminal
        Display->>Display: Apply color formatting
        Display->>Display: Add visual hierarchy
    else Plain Terminal
        Display->>Display: Use plain text
    end
    
    Display->>Display: Format error message
    Display->>Display: Add suggestions section
    
    alt Debug Mode
        Display->>Display: Include technical details
    end
    
    Display->>User: Show formatted error
    
    note over Err: Features:<br/>• Error chaining<br/>• Context preservation<br/>• Automatic classification<br/>• User-friendly messaging
```

## Error Enhancement

- **Context Wrapping**: Preserves error chains while adding operational context
- **Suggestions**: Automated helpful suggestions based on error type
- **Debug Information**: Technical details available in debug mode
- **Field-Level Validation**: Specific field validation with error codes

## Usage Within Project

This package is used universally throughout the application:
- **Commands Package**: All CLI operations use structured error handling
- **Domain Package**: Business rule violations generate domain-specific errors
- **Validation Package**: Field validation errors with detailed context
- **Git Package**: Repository operation failures with retry logic
- **Format Package**: Template processing errors with context

## API

- `Wrap(err, op)`: Adds operational context to existing errors
- `Validation(field, message)`: Creates field-specific validation errors
- `Display(err)`: User-friendly error display with colors and suggestions
- `IsRetryable(err)`: Detects if an error represents a transient failure
- Error types implement proper `Error()`, `Unwrap()`, and `ExitCode()` methods