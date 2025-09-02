---
title: contexture init
description: Initialize a new `contexture` project in the current directory by creating a `.contexture.yaml` configuration file.
---
Initializes a new `contexture` project in the current directory by creating a `.contexture.yaml` configuration file.

## Synopsis

```bash
contexture init [flags]
```

## Description

The `init` command sets up a new project. By default, it runs in an interactive mode that prompts the user to select output formats and other features. For non-interactive environments, the `--no-interactive` flag can be used.

## Flags

| Flag               | Description                                         |
| :----------------- | :-------------------------------------------------- |
| `--force`, `-f`      | Overwrite an existing `.contexture.yaml` file.      |
| `--no-interactive` | Skip interactive prompts and use default settings. |

## Usage

### Interactive Initialization

Running `contexture init` without flags starts an interactive session to configure the project.

```bash
contexture init
```

### Non-Interactive Initialization

For automated setups, such as in CI/CD pipelines, use the `--no-interactive` flag.

```bash
contexture init --no-interactive
```

To re-initialize a project and overwrite an existing configuration, use `--force`.

```bash
contexture init --force --no-interactive
```
