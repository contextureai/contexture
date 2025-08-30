# Errors Package

This package provides a unified error handling system for the application. It defines structured error types with user-friendly messaging and consistent error categorization.

## Features

- **Structured Error Types**: Errors include operation context, error kinds, and exit codes.
- **User-Friendly Messaging**: Provides clear, actionable error messages with helpful suggestions.
- **Error Classification**: Categorizes errors into domains like `validation`, `network`, and `permission`.
- **Exit Code Management**: Defines standardized exit codes for different error types.
- **Retryable Error Detection**: Identifies transient failures that can be retried.
- **Terminal-Aware Display**: Supports color-coded error output.

## Error Categories

Errors are classified into several kinds, including:
- `Validation`: Data validation failures.
- `Network`: Connection issues and timeouts.
- `Permission`: Access denied and authorization failures.
- `Configuration`: Invalid configuration files or settings.
- `Repository`: Git repository operation failures.
- `NotFound`: Missing resources or files.

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

## Usage

This package is used throughout the application to ensure consistent error handling.

## API

- `Wrap(err, op) -> error`: Adds operational context to an existing error.
- `Validation(field, message) -> error`: Creates a field-specific validation error.
- `Display(err)`: Renders a user-friendly error message, with colors and suggestions if in a terminal.
- `IsRetryable(err) -> bool`: Checks if an error represents a transient failure.
- The `Error` type implements the standard `Error()`, `Unwrap()`, and `ExitCode()` methods.