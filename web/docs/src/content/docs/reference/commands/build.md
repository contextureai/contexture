---
title: contexture build
description: Generate output files for all enabled formats based on the rules in `.contexture.yaml`.
---
Generates output files for all enabled formats based on the rules in `.contexture.yaml`.

## Synopsis

```bash
contexture build [flags]
```

## Description

The `build` command is the primary command for generating AI assistant rule files. It executes the entire process of fetching rule content, resolving variables, processing templates, and writing the final output to the format-specific directories (e.g., `CLAUDE.md`, `.cursor/rules/`).

This command should be run whenever rules are added, removed, or updated in the configuration.

## Flags

| Flag          | Description                                                              |
| :------------ | :----------------------------------------------------------------------- |
| `--verbose`, `-v` | Show detailed logs during the build process.                             |
| `--formats`   | Build only for the specified output formats (can be used multiple times). |

## Usage

### Standard Build

Generates output for all enabled formats.

```bash
contexture build
```

### Verbose Build

To see detailed step-by-step logging of the build process, use the `--verbose` flag.

```bash
contexture build --verbose
```

### Building Specific Formats

To generate output for only a subset of the enabled formats, use the `--formats` flag.

```bash
# Build only the 'claude' format
contexture build --formats claude

# Build both 'cursor' and 'windsurf' formats
contexture build --formats cursor --formats windsurf
```
