# Template Package

This package provides markdown-safe text template processing functionality for Contexture, wrapping Go's text/template with custom functions for string manipulation, formatting, and variable extraction.

## Purpose

The template package enables dynamic content generation within Contexture rules by processing Go-style templates with contextual variables. It provides a rich set of template functions while maintaining safety for markdown output by avoiding HTML escaping.

## Key Features

- **Markdown-Safe Rendering**: Uses text/template instead of html/template to prevent HTML escaping
- **Custom Function Library**: Extended set of functions for string manipulation, formatting, and array operations
- **Variable Extraction**: Automatic detection of template variables for validation and dependency analysis
- **Template Validation**: Syntax checking and parsing validation before rendering
- **Thread-Safe Operations**: Each render operation uses isolated template instances
- **Performance Optimization**: Pre-compiled regex patterns for efficient variable detection

## Template Functions

The package provides custom template functions including:
- **String Operations**: Case conversion, formatting, and manipulation functions
- **Formatting**: Text formatting and output styling functions  
- **Array Operations**: List processing and iteration helpers
- **Conditional Logic**: Enhanced conditional processing beyond standard Go templates

## Variable Detection

Advanced variable extraction capabilities:
- **Dot Variables**: Detection of `{{.Variable}}` patterns with nested field access
- **Action Variables**: Parsing of `{{if .Variable}}`, `{{range .Items}}` constructs
- **Function Variables**: Recognition of function calls that reference variables
- **Nested Access**: Support for complex variable paths like `{{.Config.Setting.Value}}`

### Template Processing Pipeline

```mermaid
flowchart TD
    START([Template Input]) --> VALIDATE[Parse and Validate Template]
    
    VALIDATE --> PARSECHECK{Valid Syntax?}
    
    PARSECHECK -->|No| PARSEERROR[Return Parse Error]
    PARSECHECK -->|Yes| EXTRACT[Extract Variables]
    
    EXTRACT --> VARIABLES[Identify Variable References]
    VARIABLES --> FUNCMAP[Apply Custom Function Map]
    
    FUNCMAP --> RENDER[Render Template]
    RENDER --> CONTEXT[Apply Variable Context]
    
    CONTEXT --> EXECUTE[Execute Template]
    EXECUTE --> EXECUTECHECK{Execution Success?}
    
    EXECUTECHECK -->|No| RUNTIMEERROR[Return Runtime Error]
    EXECUTECHECK -->|Yes| OUTPUT[Generate Output]
    
    OUTPUT --> SUCCESS[Return Rendered Content]
    
    PARSEERROR --> ERROR[Error Result]
    RUNTIMEERROR --> ERROR
    
    style VALIDATE fill:#e1f5fe
    style EXTRACT fill:#f3e5f5
    style RENDER fill:#e8f5e8
    style SUCCESS fill:#c8e6c9
    style ERROR fill:#ffcdd2
```

### Template Engine Architecture

```mermaid
graph TB
    subgraph "Template Engine"
        ENGINE[Template Engine]
        FUNCMAP[Function Map]
        PARSER[Template Parser]
        RENDERER[Template Renderer]
        EXTRACTOR[Variable Extractor]
    end
    
    subgraph "Custom Functions"
        STRINGOPS[String Operations]
        FORMATTING[Formatting Functions]
        ARRAYOPS[Array Operations]
        CONDITIONAL[Conditional Functions]
    end
    
    subgraph "Variable Detection"
        DOTVAR[Dot Variable Regex]
        ACTIONVAR[Action Variable Regex]
        FUNCVAR[Function Variable Regex]
        VARPARSER[Variable Parser]
    end
    
    subgraph "Template Processing"
        GOTMPL[Go text/template]
        VALIDATE[Syntax Validation]
        EXECUTE[Template Execution]
        RESULT[Rendered Output]
    end
    
    subgraph "Client Usage"
        RULES[Rule Processing]
        FORMATS[Format Generation]
        VARIABLES[Variable Context]
    end
    
    ENGINE --> PARSER
    ENGINE --> RENDERER
    ENGINE --> EXTRACTOR
    ENGINE --> FUNCMAP
    
    FUNCMAP --> STRINGOPS
    FUNCMAP --> FORMATTING
    FUNCMAP --> ARRAYOPS
    FUNCMAP --> CONDITIONAL
    
    EXTRACTOR --> DOTVAR
    EXTRACTOR --> ACTIONVAR
    EXTRACTOR --> FUNCVAR
    EXTRACTOR --> VARPARSER
    
    PARSER --> VALIDATE
    RENDERER --> EXECUTE
    VALIDATE --> GOTMPL
    EXECUTE --> GOTMPL
    GOTMPL --> RESULT
    
    RULES --> ENGINE
    FORMATS --> ENGINE
    VARIABLES --> ENGINE
    
    style ENGINE fill:#e1f5fe
    style FUNCMAP fill:#f3e5f5
    style EXTRACTOR fill:#e8f5e8
    style GOTMPL fill:#fff3e0
```

### Variable Extraction Process

```mermaid
sequenceDiagram
    participant Client as Client Code
    participant Engine as Template Engine
    participant Extractor as Variable Extractor
    participant Regex as Regex Patterns
    participant Parser as Variable Parser
    
    Client->>Engine: ExtractVariables(template)
    
    Engine->>Extractor: Extract from template string
    
    Extractor->>Regex: Apply dot variable pattern
    Regex-->>Extractor: {{.Variable}} matches
    
    Extractor->>Regex: Apply action variable pattern
    Regex-->>Extractor: {{if .Variable}} matches
    
    Extractor->>Regex: Apply function variable pattern
    Regex-->>Extractor: {{func .Variable}} matches
    
    Extractor->>Parser: Parse variable paths
    
    loop For each variable match
        Parser->>Parser: Extract variable name
        Parser->>Parser: Handle nested paths (e.g., .Config.Value)
        Parser->>Parser: Deduplicate variables
    end
    
    Parser-->>Extractor: Parsed variables list
    Extractor-->>Engine: Unique variable names
    Engine-->>Client: []string variable list
    
    note over Extractor: Supports complex patterns:<br/>• {{.Variable}}<br/>• {{.Config.Setting}}<br/>• {{if .Condition}}<br/>• {{range .Items}}<br/>• {{func .Variable}}
```

## Template Safety

- **Text-Only Output**: No HTML escaping ensures clean markdown generation
- **Safe Function Set**: All custom functions are designed for text output
- **Error Handling**: Comprehensive error reporting for template issues
- **Validation First**: Parse and validate before rendering to catch issues early

## Usage Within Project

This package is used by:
- **Rule Package**: Template processor uses this engine for rule content rendering
- **Format Package**: Various format implementations use templates for output generation

## API

- `NewEngine()`: Creates a template engine with all custom functions registered
- `Render(template, variables)`: Processes template with provided variable context
- `ParseAndValidate(template)`: Validates template syntax without rendering
- `ExtractVariables(template)`: Returns list of all variables referenced in template