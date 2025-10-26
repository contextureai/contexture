---
title: contexture rules new
description: Creates a new rule file with metadata and boilerplate content.
---
Creates a new rule file with metadata and boilerplate content.

## Synopsis

```bash
contexture rules new <path> [flags]
```

## Description

The `rules new` command creates a new rule file with YAML frontmatter and boilerplate content. The command supports two modes of operation:

1. **Inside a Contexture Project**: When run within a directory containing a `.contexture.yaml` configuration file, the command creates the rule in the project's `rules/` directory.
2. **Outside a Contexture Project**: When run outside a Contexture project, the command creates the rule at the literal path specified.

This dual-mode behavior makes it convenient to create rules both in individual projects (for local, project-specific rules) and in rule repositories (which typically don't have a `.contexture.yaml` file).

## Arguments

| Argument | Description                                                                 |
| :------- | :-------------------------------------------------------------------------- |
| `<path>` | The path where the rule should be created. The `.md` extension is optional and will be added automatically if not provided. |

## Flags

| Flag               | Shorthand | Description                                                                    |
| :----------------- | :-------- | :----------------------------------------------------------------------------- |
| `--global`         | `-g`      | Create rule in global rules directory (`~/.contexture/rules/`) instead of project rules directory. |
| `--name`           | `-n`      | Set the title of the rule (displayed in frontmatter and as the main heading). |
| `--description`    | `-d`      | Set the description of the rule (displayed in frontmatter).                    |
| `--tags`           | `-t`      | Set comma-separated tags for the rule (e.g., `security,auth,critical`).       |

## Usage

### Creating a Basic Rule

Create a simple rule without metadata:

```bash
# Creates rules/my-custom-rule.md in a Contexture project
contexture rules new my-custom-rule
```

This generates a minimal file with empty title and description:

```markdown
---
description: ""
title: ""
trigger: manual
---

```

Note: Title and description fields are always present in the frontmatter but will be empty strings if not specified. Tags are only included when explicitly provided via the `--tags` flag.

### Creating a Rule with Custom Metadata

Specify custom title, description, and tags:

```bash
# Using long flags
contexture rules new security-check \
  --name "Security Validation" \
  --description "Validates security best practices" \
  --tags "security,validation,critical"

# Using short flags
contexture rules new api-design \
  -n "API Design Guidelines" \
  -d "REST API design best practices" \
  -t "api,rest,design"
```

This generates:

```markdown
---
description: Validates security best practices
tags:
  - security
  - validation
  - critical
title: Security Validation
trigger: manual
---

# Security Validation

Validates security best practices
```

### Creating Nested Rules

Create rules in subdirectories using path separators:

```bash
# Creates rules/languages/go/error-handling.md
contexture rules new languages/go/error-handling \
  --name "Go Error Handling" \
  --description "Best practices for error handling in Go" \
  --tags "go,errors,best-practices"

# Creates rules/security/auth/jwt.md
contexture rules new security/auth/jwt \
  --name "JWT Authentication" \
  --description "Guidelines for implementing JWT authentication" \
  --tags "security,auth,jwt"
```

Parent directories are created automatically if they don't exist.

### Creating Rules Outside a Project

When run outside a Contexture project, the path is interpreted literally:

```bash
# Creates ./custom-rule.md in the current directory
contexture rules new custom-rule

# Creates ./team-rules/coding-standards.md
contexture rules new team-rules/coding-standards \
  --name "Team Coding Standards" \
  --tags "standards,team"

# Creates an absolute path
contexture rules new /tmp/test-rule \
  --name "Test Rule" \
  --tags "testing"
```

This is particularly useful when working in rule repositories that don't have a `.contexture.yaml` configuration file.

## Behavior Details

### Path Normalization

The command automatically handles file extensions:

```bash
# Both commands create the same file: rules/my-rule.md
contexture rules new my-rule
contexture rules new my-rule.md
```

### Project Detection

The command searches for `.contexture.yaml` in the current directory and parent directories. When found:

- Rules are created in the `rules/` directory relative to the project root
- The path argument is interpreted relative to the `rules/` directory
- Parent directories are created automatically

When not found:

- The path is interpreted literally (relative to current directory or absolute)
- Parent directories are created automatically

### File Existence Check

If a rule file already exists at the target path, the command will fail with an error:

```bash
contexture rules new existing-rule
# Error: rule file already exists: /path/to/rules/existing-rule.md
```

## Examples

### Workflow: Creating a Project-Specific Rule

```bash
# 1. Initialize a new project
cd /path/to/project
contexture init

# 2. Create a custom rule for the project
contexture rules new project-conventions \
  --name "Project Conventions" \
  --description "Project-specific coding conventions" \
  --tags "conventions,local,team"

# 3. Edit the rule file
vim rules/project-conventions.md

# 4. Add the rule to your project
contexture rules add rules/project-conventions.md

# 5. The rule is now included in generated output files
contexture build
```

### Workflow: Creating Rules in a Rule Repository

```bash
# Clone or create a rule repository (no .contexture.yaml needed)
git clone https://github.com/mycompany/rules.git
cd rules

# Create new rules
contexture rules new languages/python/type-hints \
  --name "Python Type Hints" \
  --description "Use type hints for better code clarity" \
  --tags "python,types,best-practices"

contexture rules new security/input-validation \
  --name "Input Validation" \
  --description "Validate all user inputs" \
  --tags "security,validation,critical"

# Commit and push
git add .
git commit -m "Add new rules"
git push
```

### Multiple Rules at Once

```bash
# Create several related rules
contexture rules new go/context -n "Context Usage" -t "go,context"
contexture rules new go/errors -n "Error Handling" -t "go,errors"
contexture rules new go/testing -n "Testing Practices" -t "go,testing"
contexture rules new go/concurrency -n "Concurrency Patterns" -t "go,concurrency"
```

## Default Values

If flags are not provided, the following behavior applies:

| Field           | Default Value                                      |
| :-------------- | :------------------------------------------------- |
| `--name`        | Empty string (`""`) - field present in frontmatter |
| `--description` | Empty string (`""`) - field present in frontmatter |
| `--tags`        | Not included in frontmatter when not specified     |
| `trigger`       | `"manual"` (always present)                        |

## Next Steps

After creating a rule:

1. Edit the generated file to add your rule content
2. If in a project, add the rule with `contexture rules add rules/your-rule.md`
3. If in a repository, commit and push the rule for others to use
4. See [Rule Structure](../rules/rule-structure) for details on rule file format
5. See [Rule References](../rules/rule-references) for how to reference rules in projects

## See Also

- [`contexture rules add`](./rules-add) - Add rules to your project
- [`contexture rules list`](./rules-list) - List configured rules
- [Rule Structure](../rules/rule-structure) - Learn about rule file format
- [Quick Start](../../getting-started/quick-start) - Get started with Contexture
