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

The `rules remove` command removes rule references from the project configuration. Rule IDs must be specified as arguments.

Removing a rule only affects the configuration file. The generated output files are not automatically updated. Run `contexture build` after removing rules to apply the changes to the generated formats.

## Arguments

| Argument     | Description                                                                                             |
| :----------- | :------------------------------------------------------------------------------------------------------ |
| `[rule-id...]` | One or more rule reference strings to remove. See [Rule References](../reference/rules/rule-references) for syntax. |

## Flags

| Flag           | Description                                                                    |
| :------------- | :----------------------------------------------------------------------------- |
| `--keep-outputs` | Prevents the `build` command from removing the rule's content from generated files. |
| `--formats`    | Only remove the rule from the specified output formats.                        |

## Usage

### Removing Specific Rules

Remove one or more rules by providing their IDs.

```bash
contexture rules remove "[contexture:code/clean-code]" "rules/old-rule.md"
```

### Preserving Generated Content

To remove a rule from the configuration but leave its content in the generated output files, use the `--keep-outputs` flag. This content will be removed on the next build unless this flag is used.

```bash
contexture rules remove "[contexture:testing/unit-tests]" --keep-outputs
```
