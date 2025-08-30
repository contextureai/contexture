# Template Package

This package provides a markdown-safe text template processing engine. It wraps Go's `text/template` package and extends it with custom functions for string manipulation, formatting, and variable extraction.

## Features

- **Markdown-Safe Rendering**: Uses `text/template` to avoid HTML escaping, making it safe for generating markdown.
- **Custom Function Library**: Includes a rich set of functions for string manipulation, formatting, and array operations.
- **Variable Extraction**: Automatically detects template variables (e.g., `{{.Variable}}`) for validation and dependency analysis.
- **Template Validation**: Provides functions to check template syntax and parse errors before rendering.

## Variable Detection

The engine can extract variables from various constructs, including:
- **Dot Variables**: `{{.Variable}}`
- **Actions**: `{{if .Variable}}`, `{{range .Items}}`
- **Function Calls**: `{{someFunc .Variable}}`
- **Nested Paths**: `{{.Config.Setting.Value}}`

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

## Usage

This package is used by:
- `rule` package: For rendering rule content.
- `format` package: For generating formatted output.

## API

- `NewEngine() -> Engine`: Creates a new template engine with all custom functions registered.
- `Render(template, vars) -> string`: Renders a template with the given variables.
- `ParseAndValidate(template) -> error`: Validates the template syntax without rendering it.
- `ExtractVariables(template) -> []string`: Returns a list of all variables referenced in the template.