# Contexture Rule Capabilities Specification

## Overview

This document provides a comprehensive technical specification of all rule capabilities supported by the Contexture CLI. It covers rule structure, metadata, template engine features, validation requirements, and processing capabilities. This specification is designed to be used by LLMs and developers to generate fully valid rules that utilize all features of the Contexture system.

## Table of Contents

1. [Rule Structure](#rule-structure)
2. [Rule Metadata](#rule-metadata)
3. [Template Engine](#template-engine)
4. [Validation Requirements](#validation-requirements)
5. [Rule Processing](#rule-processing)
6. [Complete Examples](#complete-examples)

## Rule Structure

### File Format

Rules are stored as Markdown files (`.md` extension) with optional YAML frontmatter.

```markdown
---
title: Rule Title
description: Brief description of the rule
tags: [tag1, tag2, tag3]
trigger:
  type: manual
languages: [typescript, javascript]
frameworks: [react, nextjs]
variables:
  customVar: value
---

# Rule Content

The main content of the rule goes here.
```

### Frontmatter Fields

All frontmatter fields are optional when parsing, but certain fields are required for validation.

| Field | Type | Required | Description | Constraints |
|-------|------|----------|-------------|------------|
| `title` | string | Yes | Rule title | Max 80 characters |
| `description` | string | Yes | Brief description | Max 200 characters |
| `tags` | string[] | Yes | Rule categories | 1-10 tags |
| `trigger` | object/string | No | Activation trigger | See [Trigger Types](#trigger-types) |
| `languages` | string[] | No | Applicable languages | Any programming languages |
| `frameworks` | string[] | No | Applicable frameworks | Any frameworks/libraries |
| `variables` | map | No | Custom variables | Key-value pairs for templates |

## Rule Metadata

### Rule Identification

Rules are identified using a structured format:

#### Full Format
```
[contexture(source):path/to/rule,ref]{variables}
```

#### Standard Format
```
[contexture:path/to/rule]
```

#### Simple Format
```
path/to/rule
```

#### Examples
```
# Simple format (uses default repository)
core/security/input-validation

# Standard format
[contexture:core/security/input-validation]

# Full format with custom source
[contexture(https://github.com/myorg/rules.git):custom/rule,main]

# With variables (JSON5 format)
[contexture:template/component]{"componentName": "Button", "async": true}
```

### Trigger Types

Rules support four trigger types that determine when they are activated:

#### 1. Always Active (`always`)

```yaml
---
trigger: always
# OR
trigger:
  type: always
---
```

Rules are permanently included in the context.

#### 2. Manual Activation (`manual`)

```yaml
---
trigger: manual
# OR
trigger:
  type: manual
---
```

Rules are activated only when explicitly requested (default).

#### 3. Model Decision (`model`)

```yaml
---
description: When working with API endpoints or REST services
trigger:
  type: model
---
```

AI determines activation based on the rule's description field.

#### 4. File Pattern Matching (`glob`)

```yaml
---
trigger:
  type: glob
  globs: ["*.ts", "*.tsx", "src/**/*.js"]
---
```

Rules activate when referenced files match the patterns.

## Template Engine

The Contexture CLI uses Go's `text/template` engine with custom functions for dynamic content generation.

### Template Syntax

#### Basic Variable Substitution

```go
{{.variableName}}
{{.nested.field}}
{{.arrayIndex 0}}
```

#### Conditional Logic

```go
{{if .condition}}
  Content when true
{{else}}
  Content when false
{{end}}

{{if eq .type "component"}}
  Component-specific content
{{else if eq .type "service"}}
  Service-specific content
{{else}}
  Default content
{{end}}
```

#### Iteration

```go
{{range .items}}
  - {{.name}}: {{.description}}
{{end}}

{{range $index, $item := .items}}
  {{$index}}. {{$item}}
{{end}}
```

#### With Blocks

```go
{{with .complexObject}}
  Title: {{.title}}
  Description: {{.description}}
{{end}}
```

### Built-in Variables

These variables are automatically available in all templates:

#### Contexture Metadata

```go
{{.contexture.version}}        # CLI version (e.g., "1.2.3" or "dev")
{{.contexture.engine}}         # Always "go"
{{.contexture.build.version}}  # Full version string
{{.contexture.build.commit}}   # Git commit hash
{{.contexture.build.date}}     # Build timestamp
{{.contexture.build.by}}       # Builder identifier
{{.contexture.build.goVersion}}# Go version used
{{.contexture.build.platform}} # Target platform
```

#### Date/Time Variables

```go
{{.now}}        # Current datetime (2006-01-02 15:04:05)
{{.date}}       # Current date (2006-01-02)
{{.time}}       # Current time (15:04:05)
{{.datetime}}   # Full datetime (2006-01-02 15:04:05)
{{.timestamp}}  # Unix timestamp
{{.year}}       # Current year
```

#### Rule Metadata (when processing rules)

```go
{{.rule.id}}          # Rule identifier
{{.rule.title}}       # Rule title
{{.rule.description}} # Rule description
{{.rule.tags}}        # Array of tags
{{.rule.languages}}   # Array of languages
{{.rule.frameworks}}  # Array of frameworks
{{.rule.source}}      # Source repository
{{.rule.ref}}         # Source ref (branch/tag/commit)
{{.rule.filepath}}    # File path
{{.rule.trigger.type}}       # Trigger type
{{.rule.trigger.globs}}      # Glob patterns (if applicable)
{{.rule.trigger.description}}# Trigger description
```

### Custom Template Functions

#### String Manipulation

| Function | Description | Example |
|----------|-------------|---------|
| `slugify` | Convert to URL-friendly slug | `{{slugify "Hello World"}}` → `hello-world` |
| `camelcase` | Convert to camelCase | `{{camelcase "hello-world"}}` → `helloWorld` |
| `pascalcase` | Convert to PascalCase | `{{pascalcase "hello-world"}}` → `HelloWorld` |
| `snakecase` | Convert to snake_case | `{{snakecase "HelloWorld"}}` → `hello_world` |
| `kebabcase` | Convert to kebab-case | `{{kebabcase "HelloWorld"}}` → `hello-world` |
| `titlecase` | Convert to Title Case | `{{titlecase "hello world"}}` → `Hello World` |
| `lower` | Convert to lowercase | `{{lower "HELLO"}}` → `hello` |
| `upper` | Convert to uppercase | `{{upper "hello"}}` → `HELLO` |
| `trim` | Trim whitespace | `{{trim "  hello  "}}` → `hello` |
| `replace` | Replace all occurrences | `{{replace "hello" "l" "r"}}` → `herro` |

#### Array Functions

| Function | Description | Example |
|----------|-------------|---------|
| `join` | Join array with separator | `{{join .items ", "}}` → `item1, item2, item3` |
| `join_and` | Join with commas and "and" | `{{join_and .tags}}` → `tag1, tag2, and tag3` |
| `unique` | Remove duplicates | `{{unique .items}}` → unique items only |
| `len` | Get length | `{{len .items}}` → number of items |

#### Formatting Functions

| Function | Description | Example |
|----------|-------------|---------|
| `indent` | Indent lines | `{{indent .content 2}}` → indented by 2 spaces |
| `default_if_empty` | Provide default value | `{{default_if_empty .value "N/A"}}` |

### Template Examples

#### Dynamic Component Generation

```markdown
---
title: React Component Template
variables:
  componentName: Button
  props:
    - name: onClick
      type: function
    - name: disabled
      type: boolean
---

# {{pascalcase .componentName}} Component

\`\`\`typescript
import React from 'react';

interface {{pascalcase .componentName}}Props {
{{range .props}}
  {{.name}}{{if eq .type "function"}}?: () => void{{else if eq .type "boolean"}}?: boolean{{else}}: {{.type}}{{end}};
{{end}}
}

export const {{pascalcase .componentName}}: React.FC<{{pascalcase .componentName}}Props> = ({
{{range .props}}
  {{.name}},
{{end}}
}) => {
  return (
    <{{lower .componentName}}>
      {/* Implementation */}
    </{{lower .componentName}}>
  );
};
\`\`\`
```

#### Conditional Framework Support

```markdown
---
title: Framework-Specific Configuration
variables:
  framework: react
---

# {{titlecase .framework}} Configuration

{{if eq .framework "react"}}
## React Setup
- Install React: `npm install react react-dom`
- Configure JSX in tsconfig.json
- Set up React DevTools
{{else if eq .framework "vue"}}
## Vue Setup
- Install Vue: `npm install vue`
- Configure Vue compiler options
- Set up Vue DevTools
{{else if eq .framework "angular"}}
## Angular Setup
- Install Angular CLI: `npm install -g @angular/cli`
- Generate new project: `ng new project-name`
- Configure Angular modules
{{else}}
## Generic Setup
- Install framework dependencies
- Configure build tools
- Set up development environment
{{end}}
```

## Validation Requirements

### Required Fields

Rules must satisfy these validation requirements:

1. **Title**: Required, max 80 characters
2. **Description**: Required, max 200 characters
3. **Tags**: Required, 1-10 tags
4. **Content**: Required, non-empty after template processing
5. **ID**: Required, valid rule ID format

### Rule ID Validation

Valid rule ID formats:
- Simple: `category/subcategory/rule-name`
- Standard: `[contexture:category/subcategory/rule-name]`
- Full: `[contexture(source):path,ref]`

Invalid characters in simple format: `! @ # $ % ^ & * ( ) + = { } [ ] | \ : ; " ' < > ? , [space]`

### Trigger Validation

- **Type**: Must be one of: `always`, `manual`, `model`, `glob`
- **Globs**: Required when type is `glob`, must be non-empty array
- **Description**: Optional for all types, recommended for `model` type

### Content Validation

- Content cannot be empty after trimming whitespace
- Content must be valid Markdown
- Template syntax must be valid Go template syntax
- All referenced variables must be available or provided

## Rule Processing

### Processing Pipeline

1. **Parsing**: Extract frontmatter and content
2. **Validation**: Validate structure and required fields
3. **Variable Resolution**: Merge variables from multiple sources
4. **Template Processing**: Render templates with variables
5. **Attribution**: Generate attribution text
6. **Output Generation**: Format for target systems

### Variable Precedence

Variables are merged in this order (later sources override earlier):

1. Global variables (lowest precedence)
2. Context variables
3. Rule-specific variables
4. Built-in variables (highest precedence for system vars)

## Complete Examples

### Example 1: Security Rule with Triggers

```markdown
---
title: Input Validation Security
description: Enforce input validation for all user-provided data
tags: [security, validation, best-practices]
trigger:
  type: glob
  globs: ["**/api/**/*.ts", "**/controllers/**/*.ts"]
languages: [typescript, javascript]
frameworks: [express, nestjs]
---

# Input Validation Standards

## Required Validations

All API endpoints must validate input using a schema validation library.

### Implementation

\`\`\`typescript
import { z } from 'zod';

const UserSchema = z.object({
  email: z.string().email(),
  age: z.number().min(0).max(120),
  name: z.string().min(1).max(100)
});

export function validateUser(data: unknown) {
  return UserSchema.parse(data);
}
\`\`\`

## Security Considerations

- Never trust client input
- Validate both type and content
- Sanitize before database operations
- Log validation failures for monitoring

Generated by {{.contexture.engine}} v{{.contexture.version}} on {{.date}}
```

### Example 2: Template-Driven Component Rule

```markdown
---
title: Component Generator
description: Generate consistent React components
tags: [react, components, templates]
trigger: manual
languages: [typescript, javascript]
frameworks: [react]
variables:
  componentType: functional
  includeTests: true
  includeStyles: true
---

# {{if .componentName}}{{pascalcase .componentName}} Component{{else}}Component Template{{end}}

## Generated Structure

{{if .componentName}}
### File: {{pascalcase .componentName}}.tsx

\`\`\`typescript
import React{{if eq .componentType "class"}}, { Component }{{end}} from 'react';
{{if .includeStyles}}import styles from './{{pascalcase .componentName}}.module.css';{{end}}

{{if .props}}
interface {{pascalcase .componentName}}Props {
{{range .props}}
  {{.name}}: {{.type}};
{{end}}
}
{{else}}
interface {{pascalcase .componentName}}Props {}
{{end}}

{{if eq .componentType "functional"}}
export const {{pascalcase .componentName}}: React.FC<{{pascalcase .componentName}}Props> = (props) => {
  return (
    <div{{if .includeStyles}} className={styles.container}{{end}}>
      {/* Component implementation */}
    </div>
  );
};
{{else}}
export class {{pascalcase .componentName}} extends Component<{{pascalcase .componentName}}Props> {
  render() {
    return (
      <div{{if .includeStyles}} className={styles.container}{{end}}>
        {/* Component implementation */}
      </div>
    );
  }
}
{{end}}
\`\`\`

{{if .includeTests}}
### File: {{pascalcase .componentName}}.test.tsx

\`\`\`typescript
import { render, screen } from '@testing-library/react';
import { {{pascalcase .componentName}} } from './{{pascalcase .componentName}}';

describe('{{pascalcase .componentName}}', () => {
  it('renders without crashing', () => {
    render(<{{pascalcase .componentName}} />);
    // Add assertions
  });
});
\`\`\`
{{end}}

{{if .includeStyles}}
### File: {{pascalcase .componentName}}.module.css

\`\`\`css
.container {
  /* Component styles */
}
\`\`\`
{{end}}
{{else}}
To use this template, provide a `componentName` variable.
{{end}}

## Usage

This component was generated with:
- Type: {{.componentType}}
- Tests: {{if .includeTests}}Included{{else}}Not included{{end}}
- Styles: {{if .includeStyles}}Included{{else}}Not included{{end}}
```

### Example 3: Multi-Language Rule with Conditions

```markdown
---
title: Error Handling Best Practices
description: Consistent error handling across languages
tags: [error-handling, best-practices, reliability]
trigger:
  type: model
  description: When implementing error handling or exception management
languages: [typescript, go, python, java]
---

# Error Handling Standards

## Language-Specific Guidelines

{{if .language}}
{{if eq .language "typescript"}}
### TypeScript Error Handling

\`\`\`typescript
// Use custom error classes
class ApplicationError extends Error {
  constructor(
    message: string,
    public code: string,
    public statusCode: number = 500
  ) {
    super(message);
    this.name = 'ApplicationError';
  }
}

// Use Result type for expected errors
type Result<T, E = Error> = 
  | { ok: true; value: T }
  | { ok: false; error: E };

// Handle async errors properly
async function fetchData(): Promise<Result<Data>> {
  try {
    const data = await api.get('/data');
    return { ok: true, value: data };
  } catch (error) {
    return { ok: false, error };
  }
}
\`\`\`

{{else if eq .language "go"}}
### Go Error Handling

\`\`\`go
// Define custom error types
type ValidationError struct {
    Field   string
    Message string
}

func (e ValidationError) Error() string {
    return fmt.Sprintf("validation failed for %s: %s", e.Field, e.Message)
}

// Always wrap errors with context
func processData(id string) error {
    data, err := fetchData(id)
    if err != nil {
        return fmt.Errorf("failed to process data for ID %s: %w", id, err)
    }
    return nil
}

// Use error variables for sentinel errors
var (
    ErrNotFound = errors.New("resource not found")
    ErrUnauthorized = errors.New("unauthorized access")
)
\`\`\`

{{else if eq .language "python"}}
### Python Error Handling

\`\`\`python
# Define custom exceptions
class ApplicationError(Exception):
    """Base exception for application errors"""
    def __init__(self, message: str, code: str = None):
        self.message = message
        self.code = code
        super().__init__(self.message)

class ValidationError(ApplicationError):
    """Raised when validation fails"""
    pass

# Use context managers for resource handling
from contextlib import contextmanager

@contextmanager
def managed_resource():
    resource = acquire_resource()
    try:
        yield resource
    finally:
        release_resource(resource)

# Handle specific exceptions
try:
    result = process_data()
except ValidationError as e:
    logger.error(f"Validation failed: {e.message}")
    raise
except Exception as e:
    logger.error(f"Unexpected error: {e}")
    raise ApplicationError("Processing failed") from e
\`\`\`

{{else}}
### {{titlecase .language}} Error Handling

Implement consistent error handling following language best practices.
{{end}}
{{else}}
### General Error Handling Principles

1. **Be Specific**: Use custom error types for different failure scenarios
2. **Add Context**: Always wrap errors with additional context
3. **Log Appropriately**: Log errors at the right level with context
4. **Fail Fast**: Detect and handle errors as early as possible
5. **Clean Up**: Always release resources in finally/defer blocks
{{end}}

## Common Patterns

- **Error Wrapping**: Add context without losing original error
- **Error Recovery**: Implement retry logic for transient failures
- **Error Reporting**: Send errors to monitoring services
- **Error Documentation**: Document possible errors in API specs

---
Generated on {{.datetime}} | Contexture {{.contexture.version}}
```

### Example 4: Rule with All Features

```markdown
---
title: Complete Rule Example
description: Demonstrates all available rule features
tags: [example, documentation, reference, templates, advanced]
trigger:
  type: glob
  globs: ["**/*.ts", "**/*.tsx", "**/*.js", "**/*.jsx"]
languages: [typescript, javascript]
frameworks: [react, vue, angular, svelte]
variables:
  projectName: MyProject
  author: Development Team
  features:
    - authentication
    - authorization
    - logging
    - monitoring
---

# {{.title}}

> {{.description}}

## Project: {{.projectName}}
**Author**: {{.author}}
**Date**: {{.date}}
**Time**: {{.time}}

## Configuration

This rule applies to:
- **Languages**: {{join_and .languages}}
- **Frameworks**: {{join_and .frameworks}}
- **File Patterns**: {{join_and .trigger.globs}}

## Features

{{range .features}}
### {{titlecase .}}

Implementation for {{.}} feature:

\`\`\`typescript
// {{.}} implementation
export const {{camelcase .}}Service = {
  name: '{{kebabcase .}}',
  initialize() {
    console.log('Initializing {{.}}...');
  }
};
\`\`\`
{{end}}

## Variable Transformations

| Original | Slugified | CamelCase | PascalCase | SnakeCase | KebabCase |
|----------|-----------|-----------|------------|-----------|-----------|
{{range .features}}
| {{.}} | {{slugify .}} | {{camelcase .}} | {{pascalcase .}} | {{snakecase .}} | {{kebabcase .}} |
{{end}}

## Conditional Rendering

{{if .projectName}}
Project name is set to: **{{.projectName}}**
{{else}}
No project name provided.
{{end}}

{{with .author}}
### Author Information
This rule was created by: {{.}}
{{end}}

## Array Operations

**All Tags**: {{join .tags ", "}}
**Tags with "and"**: {{join_and .tags}}
**Tag Count**: {{len .tags}}
**Unique Tags**: {{join_and (unique .tags)}}

## Built-in Variables

### Contexture Information
- Version: {{.contexture.version}}
- Engine: {{.contexture.engine}}
- Build Date: {{.contexture.build.date}}
- Build Commit: {{.contexture.build.commit}}
- Platform: {{.contexture.build.platform}}

### Date/Time Information
- Current Date: {{.date}}
- Current Time: {{.time}}
- Full DateTime: {{.datetime}}
- Unix Timestamp: {{.timestamp}}
- Year: {{.year}}

## Indentation Example

{{indent "This text\nwill be\nindented\nby 4 spaces" 4}}

## Default Values

- With value: {{default_if_empty .projectName "Unknown Project"}}
- Without value: {{default_if_empty .missingVar "Default Value"}}

## Text Transformations

- Uppercase: {{upper .projectName}}
- Lowercase: {{lower .projectName}}
- Trimmed: {{trim "  trimmed text  "}}
- Replaced: {{replace .projectName "My" "Your"}}

---

## Rule Metadata

{{if .rule}}
- **Rule ID**: {{.rule.id}}
- **Rule Title**: {{.rule.title}}
- **Rule Source**: {{.rule.source}}
- **Rule Ref**: {{.rule.ref}}
- **Trigger Type**: {{.rule.trigger.type}}
{{end}}

---
*Generated by {{.contexture.engine}} v{{.contexture.version}} on {{.datetime}}*
```

## Best Practices

### 1. Template Design

- Use meaningful variable names
- Provide default values for optional variables
- Include conditional blocks for flexibility
- Document expected variables in comments

### 2. Error Prevention

- Always check for variable existence before use
- Use `with` blocks for nested objects
- Provide fallback content for missing data
- Validate template syntax before deployment

### 3. Performance

- Minimize complex template logic
- Use built-in functions efficiently
- Cache processed templates when possible
- Avoid deeply nested conditions

### 4. Maintainability

- Keep templates focused and single-purpose
- Use consistent naming conventions
- Document template variables and usage
- Version control template changes

## Limitations

1. **Character Limits**:
   - Title: 80 characters maximum
   - Description: 200 characters maximum
   - Rule ID: 200 characters maximum

2. **Array Limits**:
   - Tags: 1-10 tags per rule
   - No limit on languages or frameworks arrays

3. **Template Constraints**:
   - Go template syntax only (not Liquid)
   - No HTML escaping (text/template, not html/template)
   - Variables must be valid Go identifiers when referenced

4. **Processing Constraints**:
   - Templates are processed once per rule
   - No recursive template inclusion
   - No external file includes

## Conclusion

This specification provides comprehensive documentation of all rule capabilities in the Contexture CLI. By following these guidelines and examples, LLMs and developers can create sophisticated, dynamic rules that leverage the full power of the template engine and validation system.

For updates and additional examples, refer to the Contexture rules repository and CLI documentation.