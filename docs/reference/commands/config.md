---
title: contexture config
description: Display and manage project configuration from the `.contexture.yaml` file.
---
Displays and manages project configuration from the `.contexture.yaml` file.

## Synopsis

```bash
contexture config [subcommand]
```

## Description

The `config` command is used to view the project's configuration and manage output formats. When run without a subcommand, it defaults to the `show` action.

By default, it displays the project configuration (`.contexture.yaml`). Use the `--global` flag to view or manage your user-level global configuration (`~/.contexture/.contexture.yaml`).

## Flags

| Flag          | Shorthand | Description                                                              |
| :------------ | :-------- | :----------------------------------------------------------------------- |
| `--global`    | `-g`      | View or manage global configuration (`~/.contexture/.contexture.yaml`) instead of project configuration. |

## Subcommands

### `show`

Displays a summary of the current project configuration. This is the default action.

**Synopsis**

```bash
contexture config
contexture config show
```

**Aliases**

-   `s`

**Examples**

```bash
# View project configuration
contexture config
contexture config show

# View global configuration
contexture config --global
contexture config -g
```

### `formats`

Provides tools for managing the output formats defined in the configuration file.

**Synopsis**

```bash
contexture config formats [subcommand]
```

**Subcommands for `formats`:**

| Subcommand | Description                                          |
| :--------- | :--------------------------------------------------- |
| `list`     | List all available and configured output formats.    |
| `add`      | Add one or more formats to the configuration.        |
| `remove`   | Remove one or more formats from the configuration.   |
| `enable`   | Enable one or more configured formats.               |
| `disable`  | Disable one or more configured formats.              |

**Usage Examples for `formats`**

```bash
# List all formats
contexture config formats list

# Add and enable the 'claude' and 'cursor' formats
contexture config formats add claude cursor

# Disable the 'windsurf' format
contexture config formats disable windsurf
```