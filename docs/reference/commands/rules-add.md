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
| `--force`, `-f` | Update a rule's configuration if it already exists in `.contexture.yaml`.      |
| `--formats` | Specify which output formats a rule should apply to (can be used multiple times). |
| `--data`    | Provide rule variables as a JSON string.                                       |
| `--var`     | Set an individual variable (`key=value`) (can be used multiple times).           |
| `--source`, `--src` | Specify a custom Git repository URL to pull a rule from.                       |
| `--ref`     | Specify a Git branch, tag, or commit hash for a remote rule.                   |

## Usage

### Adding Rules

Add one or more rules by providing their IDs.

```bash
# Simple format (recommended)
contexture rules add languages/go/code-organization testing/unit-tests

# Full format
contexture rules add "[contexture:code/clean-code]" "[contexture:testing/unit-tests]"
```

### Adding a Rule with Variables

Variables can be provided inline as a JSON5 string.

```bash
contexture rules add 'testing/coverage {"threshold": 90}'
```

Alternatively, use the `--data` or `--var` flags.

```bash
contexture rules add testing/coverage --data '{"threshold": 90}'
contexture rules add testing/coverage --var threshold=90
```

### Adding a Rule from a Custom Source

To add a rule from a Git repository that is not configured in `sources`, use the `--source` (or `--src`) and `--ref` flags.

```bash
# Using --source flag
contexture rules add "my/custom-rule" \
  --source "https://github.com/my-org/rules.git" \
  --ref "main"

# Using --src shorthand
contexture rules add "security/auth" --src "git@github.com:company/rules.git"

# Multiple rules from the same custom source
contexture rules add "api/validation" "api/rate-limiting" \
  --src "https://github.com/myteam/api-rules.git" \
  --ref "v2.0"
```
