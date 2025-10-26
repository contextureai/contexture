---
title: Configuration File
description: The `.contexture.yaml` file defines the project's rules, providers, output formats, and build settings.
---
The `.contexture.yaml` file defines the project's rules, providers, output formats, and build settings.

## Configuration Locations

Contexture supports two configuration locations:

### Project Configuration
Located in the project root directory:
- `.contexture.yaml` (root location)
- `.contexture/.contexture.yaml` (contexture directory)

### Global Configuration
Located in your home directory at `~/.contexture/.contexture.yaml`

Global configuration:
- Applies to all projects automatically
- Can define rules, providers, and formats
- Project-specific rules override global rules with matching IDs
- Modified using the `--global` or `-g` flag with rule and provider commands

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

| Field      | Type      | Required | Description                                     |
| :--------- | :-------- | :------- | :---------------------------------------------- |
| `type`     | `string`    | `true`     | The format type (`claude`, `cursor`, `windsurf`). |
| `enabled`  | `boolean`   | `false`    | Enable/disable the format (defaults to `true`; generated configs include the explicit value for clarity). |
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

| Field        | Type             | Required | Description                                                             |
| :----------- | :--------------- | :------- | :---------------------------------------------------------------------- |
| `id`         | `string`         | `true`     | The rule reference string. See [Rule References](../reference/rules/rule-references). |
| `variables`  | `map[string]any` | `false`    | Variables to apply to the rule.                                         |
| `source`     | `string`         | `false`    | The resolved source identifier or repository URL. Populated automatically. |
| `ref`        | `string`         | `false`    | The resolved branch, tag, or commit hash. Defaults to `main`.            |
| `commitHash` | `string`         | `false`    | The exact commit that was fetched. Used by `contexture rules update`.     |
| `pinned`     | `boolean`        | `false`    | Marks the rule as pinned to the recorded commit.                         |

**Example:**
```yaml
rules:
  - id: "[contexture:code/clean-code]"
  - id: "@mycompany/testing/coverage"
    variables:
      threshold: 90
  - id: "rules/local-project-rule.md"
```

`contexture` manages the `source`, `ref`, `commitHash`, and `pinned` fields automatically when you add, update, or pin rules. In most cases you only need to edit the `id` and `variables` entries.
