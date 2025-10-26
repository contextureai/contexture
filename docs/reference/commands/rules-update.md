---
title: contexture rules update
description: Updates remote rules to their latest versions from their respective Git repositories.
---
Updates remote rules to their latest versions from their respective Git repositories.

## Synopsis

```bash
contexture rules update [flags]
```

## Description

The `rules update` command checks all configured remote rule sources for new commits or tags. It compares the latest available version with the locally cached version and applies updates if available. This command does not affect local rules.

## Flags

| Flag        | Description                                               |
| :---------- | :-------------------------------------------------------- |
| `--global`, `-g` | Update rules in global configuration instead of project configuration. |
| `--dry-run` | Show available updates without applying them.             |
| `--yes`, `-y` | Skip the confirmation prompt and apply all updates.       |
| `--output`, `-o` | Choose the output format: `default` (terminal) or `json`. |

## Usage

### Checking for Updates

To see a summary of available updates without making any changes, use the `--dry-run` flag.

```bash
contexture rules update --dry-run
```

### Applying Updates

To apply all available updates, run the command without flags. It will present a summary and prompt for confirmation before proceeding.

```bash
contexture rules update
```

For automated environments, use the `--yes` flag to bypass the confirmation prompt.

```bash
contexture rules update --yes
```

When updates are applied successfully, `contexture` automatically regenerates all enabled formats so the freshly fetched rule content is reflected in your `CLAUDE.md`, `.cursor/rules/`, and `.windsurf/rules/` directories.
