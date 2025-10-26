---
title: Variables and Templates
description: Overview of variables and templates in contexture.
---
`contexture` uses Go's `text/template` engine to allow rule customization with variables.

## Template System

Variables can be used in rules to:
-   Customize behavior for different frameworks or languages.
-   Inject project-specific values like naming conventions.
-   Conditionally render content.

## Variable Sources

Variables are resolved from multiple sources in the following order of precedence:

1.  **Rule reference variables**: Variables defined alongside a rule in `.contexture.yaml`.
2.  **Rule frontmatter**: Default values in the rule's YAML header.
3.  **Global context**: System-provided variables.

## Template Syntax

### Basic Usage

```markdown
# Simple variable substitution
Maximum line length: {{.maxLineLength}} characters.

# With a default value
Test framework: {{default_if_empty .framework "jest"}}
```

### Nested Variables

```markdown
# Object properties
Database host: {{.database.host}}

# Array elements
Primary language: {{index .languages 0}}
```

## Built-in Functions

### String Functions

```markdown
# Case conversion
{{.name | upper}}   # UPPERCASE
{{.name | lower}}   # lowercase
{{.name | titlecase}} # Title Case

# String manipulation
{{.text | trim}}  # Remove whitespace
```

### Conditional Logic

```markdown
{{if .useTypeScript}}
- Add proper type annotations.
- Enable strict mode.
{{else}}
- Use JSDoc for type hints.
{{end}}

# Equality checks
{{if eq .environment "production"}}
Production-specific rules apply.
{{end}}
```

### Array Functions

```markdown
# Iteration
{{range .languages}}
- {{.}} support enabled.
{{end}}

# Joining
Languages: {{join .languages ", "}}
```

## Advanced Patterns

### Dynamic Content

```markdown
{{range .languages}}
{{if eq . "python"}}
## Python Guidelines
- Use type hints for function parameters.
- Follow PEP 8 style guidelines.
{{end}}
{{end}}

{{range .frameworks}}
{{if eq . "django"}}
### Django Specifics
- Use Django's built-in User model.
{{end}}
{{end}}
```

### Defining Variables

Variables can be defined in the rule's frontmatter or in the rule reference in `.contexture.yaml`.

#### In Rule Frontmatter

```yaml
---
title: Code Style Guidelines
variables:
  maxLineLength: 100
  database:
    type: "postgresql"
    port: 5432
  supportedLanguages: ["javascript", "typescript"]
---
```

#### In `.contexture.yaml`

```yaml
rules:
  - id: '[contexture:style/formatting]'
    variables:
      maxLineLength: 120
      useSpaces: true
  - id: '[contexture:database/config]'
    variables:
      database:
        type: "postgresql"
        host: "localhost"
```

## Global Variables

`contexture` provides several global variables that are always available:

-   `{{.now}}`: The current time.
-   `{{.date}}`: The current date (YYYY-MM-DD).
-   `{{.time}}`: The current time (HH:MM:SS).
-   `{{.datetime}}`: The current date and time (YYYY-MM-DD HH:MM:SS).
-   `{{.timestamp}}`: The current Unix timestamp.
-   `{{.contexture}}`: A map containing build information:
    -   `{{.contexture.version}}`: The version of `contexture`.
    -   `{{.contexture.build.commit}}`: The Git commit hash.
    -   `{{.contexture.build.date}}`: The build date.

## Next Steps

-   **[Rules Documentation](../reference/rules/rule-references)**: Creating rules with variables.
-   **[Commands Reference](../reference/commands/init)**: The command reference.
