---
title: contexture rules remove
description: Removes one or more rules from the `.contexture.yaml` configuration file.
---
Removes one or more rules from the `.contexture.yaml` configuration file.

## Synopsis

```bash
contexture rules remove [rule-id...] [flags]
```

## Aliases

-   `rm`

## Description

The `rules remove` command removes rule references from the project configuration and cleans any generated artifacts that referenced those rules. Rule IDs must be specified as arguments.

## Arguments

| Argument     | Description                                                                                             |
| :----------- | :------------------------------------------------------------------------------------------------------ |
| `[rule-id...]` | One or more rule reference strings to remove. See [Rule References](../reference/rules/rule-references) for syntax. |

## Flags

| Flag          | Description                                                |
| :------------ | :--------------------------------------------------------- |
| `--output`, `-o` | Choose the output format: `default` (terminal) or `json`. |

## Usage

### Removing Specific Rules

Remove one or more rules by providing their IDs.

```bash
contexture rules remove "[contexture:code/clean-code]" "rules/old-rule.md"
```

After the configuration is saved, `contexture` automatically regenerates the enabled formats so that the deleted rules are removed from `CLAUDE.md`, `.cursor/rules/`, and `.windsurf/rules/`.
