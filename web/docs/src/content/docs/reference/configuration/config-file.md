---
title: Configuration File
description: The `.contexture.yaml` file defines the project's rules, providers, output formats, and build settings.
---
The `.contexture.yaml` file defines the project's rules, providers, output formats, and build settings.

## Structure

```yaml
version: 1
providers: []
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

### `providers`

Defines custom named providers for rule sources. Providers enable `@provider/path` syntax for rule references.

-   **Type**: `list`
-   **Required**: `false`

**Provider Fields:**

| Field   | Type     | Required | Description                               |
| :------ | :------- | :------- | :---------------------------------------- |
| `name`    | `string`   | `true`     | Unique identifier for the provider.       |
| `url`     | `string`   | `true`     | Git repository URL (HTTPS or SSH).        |
| `defaultBranch`  | `string`   | `false`    | Default Git branch (defaults to `main`).  |
| `auth`    | `object`   | `false`    | Authentication configuration.             |

**Auth Fields:**

| Field | Type     | Required | Description                                     |
| :---- | :------- | :------- | :---------------------------------------------- |
| `type`  | `string`   | `true`     | Authentication type (`token` or `ssh`).         |
| `token` | `string`   | `false`    | A personal access token for HTTPS auth.         |

**Example:**
```yaml
providers:
  - name: mycompany
    url: https://github.com/mycompany/contexture-rules.git
    defaultBranch: main
    auth:
      type: token
      token: "ghp_..."
```

### `formats`

Defines the output formats to generate.

-   **Type**: `list`
-   **Required**: `true`

**Format Fields:**

| Field   | Type      | Required | Description                                     |
| :------ | :-------- | :------- | :---------------------------------------------- |
| `type`    | `string`    | `true`     | The format type (`claude`, `cursor`, `windsurf`). |
| `enabled` | `boolean`   | `false`    | Enable/disable the format (defaults to `true`).   |

**Example:**
```yaml
formats:
  - type: claude
    enabled: true
  - type: cursor
    enabled: false
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
  - id: "@mycompany/testing/coverage"
    variables:
      threshold: 90
  - id: "rules/local-project-rule.md"
```
