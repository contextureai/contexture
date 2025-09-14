---
title: Configuration File
description: The `.contexture.yaml` file defines the project's rules, sources, output formats, and build settings.
---
The `.contexture.yaml` file defines the project's rules, sources, output formats, and build settings.

## Structure

```yaml
version: 1
sources: []
formats: []
rules: []
generation: {}
```

## Top-Level Sections

### `version`

Specifies the configuration format version.

-   **Type**: `integer`
-   **Required**: `true`
-   **Current Value**: `1`

### `sources`

Defines custom Git repositories for rules. Each source is an object in a list.

-   **Type**: `list`
-   **Required**: `false`

**Source Fields:**

| Field   | Type     | Required | Description                               |
| :------ | :------- | :------- | :---------------------------------------- |
| `name`    | `string`   | `true`     | Unique identifier for the source.         |
| `type`    | `string`   | `true`     | Source type (currently only `git`).       |
| `url`     | `string`   | `true`     | Git repository URL (HTTPS or SSH).        |
| `branch`  | `string`   | `false`    | Default Git branch (defaults to `main`).    |
| `tag`     | `string`   | `false`    | Git tag to use (overrides `branch`).      |
| `enabled` | `boolean`  | `false`    | Enable/disable the source (defaults to `true`). |
| `auth`    | `object`   | `false`    | Authentication configuration.             |

**Auth Fields:**

| Field | Type     | Required | Description                                     |
| :---- | :------- | :------- | :---------------------------------------------- |
| `type`  | `string`   | `true`     | Authentication type (`token` or `ssh`).         |
| `token` | `string`   | `false`    | A personal access token for HTTPS auth.         |

**Example:**
```yaml
sources:
  - name: company
    type: git
    url: https://github.com/mycompany/contexture-rules.git
    branch: main
    auth:
      type: token
      token: "ghp_..."
```

### `formats`

Defines the output formats to generate.

-   **Type**: `list`
-   **Required**: `true`

**Format Fields:**

| Field      | Type      | Required | Description                                     |
| :--------- | :-------- | :------- | :---------------------------------------------- |
| `type`     | `string`    | `true`     | The format type (`claude`, `cursor`, `windsurf`). |
| `enabled`  | `boolean`   | `false`    | Enable/disable the format (defaults to `true`).   |
| `template` | `string`    | `false`    | Template file path (Claude format only).          |

**Example:**
```yaml
formats:
  - type: claude
    enabled: true
  - type: cursor
    enabled: false
  - type: claude
    enabled: true
    template: CLAUDE.template.md
```

**Template Field (Claude Format Only):**

The `template` field allows you to specify a custom template file for the Claude format. When specified, Contexture will use your template file instead of the default format.

- The template file should contain `{{.Rules}}` where you want the generated rules to be inserted
- If the template file is not found, Contexture falls back to the default format
- Template files support Go text/template syntax
- Only the `{{.Rules}}` variable is available in the template context

**Template Example:**
```markdown
# My Custom Instructions

## Project Overview
Custom project context here.

## AI Rules
{{.Rules}}

## Additional Notes  
Custom footer content.
```

### `rules`

Defines the rules to include in the project.

-   **Type**: `list`
-   **Required**: `false`

**Rule Reference Fields:**

| Field       | Type           | Required | Description                                                             |
| :---------- | :------------- | :------- | :---------------------------------------------------------------------- |
| `id`        | `string`         | `true`     | The rule reference string. See [Rule References](../reference/rules/rule-references). |
| `variables` | `map[string]any` | `false`    | Variables to apply to the rule.                                         |

**Example:**
```yaml
rules:
  - id: "[contexture:code/clean-code]"
  - id: "[contexture(company):testing/coverage]"
    variables:
      threshold: 90
  - id: "rules/local-project-rule.md"
```
