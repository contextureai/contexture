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
| `--source`  | Specify a custom Git repository URL to pull a rule from.                       |
| `--ref`     | Specify a Git branch, tag, or commit hash for a remote rule.                   |

## Usage

### Interactive Mode

To browse and select rules interactively, run the command without arguments.

```bash
contexture rules add
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

To add a rule from a Git repository that is not configured in `sources`, use the `--source` and `--ref` flags.

```bash
contexture rules add "my/custom-rule" \
  --source "https://github.com/my-org/rules.git" \
  --ref "main"
```
