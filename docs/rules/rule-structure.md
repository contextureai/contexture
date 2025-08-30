# Rule Structure

A `contexture` rule is a markdown document composed of two main sections: frontmatter (YAML metadata) and content (Markdown body). This document details the specification for these sections.

## Anatomy

A rule file has the following structure:

```markdown
---
# 1. Frontmatter (YAML metadata)
title: <string>
description: <string>
tags: [<string>]
languages: [<string>]
frameworks: [<string>]
trigger: <string | object>
variables: {<key>: <value>}
---

# 2. Content (Markdown body)
The content for the AI assistant, which can use {{.variableName}} syntax.
```

## Frontmatter Specification

The frontmatter section contains structured metadata about the rule, enclosed in `---` markers.

### Required Fields

| Field         | Type     | Description                                     |
| :------------ | :------- | :---------------------------------------------- |
| `title`       | `string`   | A concise, descriptive name for the rule.       |
| `description` | `string`   | A brief explanation of the rule's purpose.      |
| `tags`        | `[]string` | An array of categorization tags.                |

### Optional Fields

| Field        | Type           | Description                                                        |
| :----------- | :------------- | :----------------------------------------------------------------- |
| `languages`  | `[]string`       | A list of applicable programming languages.                        |
| `frameworks` | `[]string`       | A list of applicable frameworks or libraries.                        |
| `trigger`    | `string|object` | Defines when the rule is applied. See [Rule Triggers](#rule-triggers). |
| `variables`  | `map[string]any` | Default values for template variables.                             |

### Rule Triggers

The `trigger` field controls when a rule is applied.

| Type     | Description                                         | Use Case                      |
| :------- | :-------------------------------------------------- | :---------------------------- |
| `manual` | (Default) The rule is applied only when explicitly requested. | Specialized or optional rules |
| `always` | The rule is always included in generated outputs.   | Core standards and conventions  |
| `model`  | The rule is applied based on AI model context.      | Advanced conditional logic    |
| `glob`   | The rule is applied when file glob patterns match.  | Context-specific rules        |

A `glob` trigger is configured as an object:

```yaml
trigger:
  type: glob
  globs:
    - "*.test.js"
    - "**/__tests__/**"
```

## Content Specification

The content section contains the AI assistant instructions in standard markdown format. It supports template variables using Go's `text/template` syntax.

### Template Variables

Variables allow for dynamic content within rules.

**Syntax:**
- **Substitution**: `{{.variableName}}`
- **Conditional Logic**: `{{if .variableName}}...{{end}}`
- **Iteration**: `{{range .arrayName}}...{{end}}`

### Variable Resolution

Variables are resolved with the following order of precedence (highest to lowest):

1.  **Rule Reference Variables**: Defined in `.contexture.yaml` for a specific rule reference.
2.  **Rule Frontmatter Defaults**: Defined in the `variables` section of the rule's frontmatter.
3.  **Global Variables**: System-provided variables.

See the [Variables and Templates](../core-concepts/variables.md) documentation for more details.
