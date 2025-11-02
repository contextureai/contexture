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
| `--global`, `-g` | Remove rule from global configuration (`~/.contexture/.contexture.yaml`) instead of project configuration. |
| `--output`, `-o` | Choose the output format: `default` (terminal) or `json`. |

## Usage

### Removing Specific Rules

Remove one or more rules by providing their IDs.

```bash
contexture rules remove "[contexture:code/clean-code]" "rules/old-rule.md"
```

### Removing Global Rules

Remove rules from your user-level global configuration:

```bash
# Remove a global rule
contexture rules remove @contexture/languages/go/context --global

# Use shorthand flag
contexture rules remove @contexture/testing/best-practices -g

# Remove multiple global rules
contexture rules remove @contexture/code/clean-code @contexture/security/input-validation --global
```

After the configuration is saved, `contexture` automatically regenerates the enabled formats so that the deleted rules are removed from `CLAUDE.md`, `.cursor/rules/`, and `.windsurf/rules/`.
