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

The `rules add` command adds new rules to the project. When run without arguments, it opens an interactive browser to select rules from available sources. Rules can also be added directly by providing their rule IDs as arguments.

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

### Interactive Mode

To browse and select rules interactively, run the command without arguments.

```bash
contexture rules add
```

#### Interactive Mode with Custom Sources

You can also browse rules from custom repositories interactively by specifying a custom source:

```bash
# Browse rules from a custom repository
contexture rules add --src https://github.com/mycompany/contexture-rules.git

# Browse rules from a custom repository at a specific branch
contexture rules add --source git@github.com:mycompany/rules.git --ref develop
```

### Adding Specific Rules

Add one or more rules by providing their IDs.

```bash
contexture rules add "[contexture:code/clean-code]" "[contexture:testing/unit-tests]"
```

### Adding a Rule with Variables

Variables can be provided inline as a JSON5 string.

```bash
contexture rules add '[contexture:testing/coverage] {"threshold": 90}'
```

Alternatively, use the `--data` or `--var` flags.

```bash
contexture rules add "[contexture:testing/coverage]" --data '{"threshold": 90}'
contexture rules add "[contexture:testing/coverage]" --var threshold=90
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
