---
title: contexture rules add
description: Adds one or more rules to the `.contexture.yaml` configuration file.
---
Adds one or more rules to the `.contexture.yaml` configuration file.

## Synopsis

```bash
contexture rules add [rule-id...] [flags]
```

## Description

The `rules add` command adds new rules to the project. Rules must be specified by providing their rule IDs as arguments.

## Arguments

| Argument    | Description                                                                                             |
| :---------- | :------------------------------------------------------------------------------------------------------ |
| `[rule-id...]` | One or more rule reference strings. See [Rule References](../reference/rules/rule-references) for syntax details. |

## Flags

| Flag        | Description                                                                    |
| :---------- | :----------------------------------------------------------------------------- |
| `--global`, `-g` | Add rule to global configuration (`~/.contexture/.contexture.yaml`) instead of project configuration. |
| `--data`    | Provide rule variables as a JSON string.                                       |
| `--var`     | Set an individual variable (`key=value`) (can be used multiple times).           |
| `--source`, `--src` | Specify a custom Git repository URL to pull a rule from.                       |
| `--ref`     | Specify a Git branch, tag, or commit hash for a remote rule.                   |
| `--output`, `-o` | Choose the output format: `default` (terminal) or `json`.                  |

## Usage

### Adding Rules from Providers

The recommended way to add rules is using the `@provider/path` syntax:

```bash
# Add rules from the default @contexture provider
contexture rules add @contexture/code/clean-code @contexture/testing/unit-tests

# Add rules from a custom provider
contexture rules add @mycompany/security/auth @mycompany/testing/coverage

# Add multiple rules at once
contexture rules add @contexture/languages/go/code-organization @contexture/languages/go/error-handling
```

#### Alternative: Bracketed Syntax

You can also use the bracketed syntax:

```bash
# Bracketed format (alternative)
contexture rules add "[contexture:code/clean-code]" "[contexture:testing/unit-tests]"
```

### Adding a Rule with Variables

Variables can be provided using the `--var` flag (recommended):

```bash
# Using --var flag (recommended)
contexture rules add @contexture/testing/coverage --var threshold=90

# Multiple variables
contexture rules add @contexture/testing/coverage --var threshold=90 --var framework=jest
```

Alternatively, use the `--data` flag or inline JSON5:

```bash
# Using --data flag
contexture rules add @contexture/testing/coverage --data '{"threshold": 90}'

# Inline JSON5
contexture rules add '@contexture/testing/coverage {"threshold": 90, "framework": "jest"}'
```

### Using Specific Branches or Tags

To use a specific version of a rule, use the `--ref` flag:

```bash
# Use a specific branch
contexture rules add @contexture/experimental/new-feature --ref development

# Use a specific tag
contexture rules add @contexture/stable/patterns --ref v1.2.0

# Use a specific commit
contexture rules add @mycompany/security/auth --ref abc123def
```

### Adding Rules from Command-Line Sources

For one-time use, you can add rules from Git repositories that aren't configured as providers using the `--source` (or `--src`) and `--ref` flags:

```bash
# Using --source flag
contexture rules add security/auth \
  --source "https://github.com/company/rules.git" \
  --ref "main"

# Using --src shorthand
contexture rules add api/validation --src "git@github.com:team/rules.git"

# Multiple rules from the same source
contexture rules add api/validation api/rate-limiting \
  --src "https://github.com/team/api-rules.git" \
  --ref "v2.0"
```

**Note:** For frequently used repositories, consider adding them as providers using `contexture providers add` for cleaner syntax. See [Providers](./providers.md) for details.

### Adding Local Rules

You can add rules from local files by providing the file path:

```bash
# Add a local rule file
contexture rules add rules/project-specific-rule.md

# Add multiple local rules
contexture rules add rules/custom-1.md rules/custom-2.md
```

### Adding Global Rules

Global rules are stored in your user-level configuration and automatically included in all projects:

```bash
# Add a rule to global configuration
contexture rules add @contexture/languages/go/context --global

# Use shorthand flag
contexture rules add @contexture/testing/best-practices -g

# Add multiple global rules
contexture rules add @contexture/code/clean-code @contexture/security/input-validation --global
```

**Behavior:**
- Global rules are stored in `~/.contexture/.contexture.yaml`
- They are automatically included when running `contexture build` in any project
- Project-specific rules with matching IDs override global rules
- When adding a global rule from within a project directory, the project is automatically rebuilt to include the new global rule

After the rules are added, `contexture` immediately regenerates the enabled formats so the new guidance is written to `CLAUDE.md`, `.cursor/rules/`, and `.windsurf/rules/` without an additional `build` step.
