---
title: contexture rules list
description: Displays the rules configured in the `.contexture.yaml` file.
---
Displays the rules configured in the `.contexture.yaml` file.

## Synopsis

```bash
contexture rules list [flags]
```

## Aliases

-   `ls`

## Description

The `rules list` command displays all rules that have been added to the project in a clean, terminal-friendly format. Each rule shows its path, title, and source information. The command supports pattern-based filtering to help you find specific rules quickly.

## Flags

| Flag          | Description                                                                  |
| :------------ | :--------------------------------------------------------------------------- |
| `--pattern`, `-p` | Filter rules using a regex pattern (matches ID, title, description, tags, frameworks, languages, source) |

## Usage

### List All Rules

Displays all configured rules with their paths, titles, and source information.

```bash
contexture rules list
```

### Filter by Pattern

Use regex patterns to filter rules across multiple fields. Patterns are case-insensitive by default.

```bash
# Find rules related to Go
contexture rules list --pattern "go"

# Find testing-related rules
contexture rules list -p "testing"

# Use regex patterns
contexture rules list --pattern "(python|javascript)"

# Find security rules
contexture rules list --pattern "security.*validation"
```

## Output Format

The command displays rules in a compact format:
- **Rule Path**: The rule's identifier path (e.g., `languages/go/testing`)
- **Title**: A descriptive title (e.g., `Go Testing Best Practices`)
- **Source**: Where the rule comes from (only shown for non-default sources)

When using a pattern filter, the active pattern is shown in the header for clarity.
