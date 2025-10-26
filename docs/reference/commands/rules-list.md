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

The command automatically merges and displays rules from:
1. **Global configuration** (`~/.contexture/.contexture.yaml`) - Rules available across all projects
2. **Project configuration** (`.contexture.yaml`) - Project-specific rules
3. **Local rules** (`rules/` directory) - Rules stored in the project

When rules have matching IDs, project-specific rules take precedence over global rules.

## Flags

| Flag          | Description                                                                  |
| :------------ | :--------------------------------------------------------------------------- |
| `--pattern`, `-p` | Filter rules using a regex pattern (matches ID, title, description, tags, frameworks, languages, source) |
| `--output`, `-o` | Output format: `default` for terminal display, `json` for JSON output |

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

### JSON Output

Use JSON output for programmatic processing or integration with other tools.

```bash
# Output rules as JSON
contexture rules list --output json

# Use short flag
contexture rules list -o json

# Combine with pattern filtering
contexture rules list --pattern "go" --output json
```

**JSON Structure:**
```json
{
  "metadata": {
    "pattern": "go",
    "totalRules": 5,
    "filteredRules": 2
  },
  "rules": [
    {
      "id": "[contexture:languages/go/testing]",
      "title": "Go Testing Best Practices",
      "description": "Write idiomatic table-driven tests...",
      "tags": ["go", "testing", "best-practices"],
      "trigger": {
        "type": "glob",
        "globs": ["**/*_test.go"]
      },
      "languages": ["go"],
      "frameworks": [],
      "content": "Rule content...",
      "variables": {},
      "defaultVariables": {},
      "filePath": "languages/go/testing",
      "source": "https://github.com/contextureai/rules.git",
      "ref": "main"
    }
  ]
}
```

## Output Format

### Terminal Output Format

The default terminal output displays rules in a compact format:
- **Rule Path**: The rule's identifier path (e.g., `languages/go/testing`)
- **Title**: A descriptive title (e.g., `Go Testing Best Practices`)
- **Source**: Where the rule comes from (only shown for non-default sources)

When using a pattern filter, the active pattern is shown in the header for clarity.

### JSON Output Format  

JSON output provides structured data suitable for programmatic processing:
- **Metadata**: Pattern filter (if used) and rule counts
- **Rules Array**: Complete rule objects with IDs, metadata, variables, and content
- **Consistent Schema**: Stable field names that match the CLI structs

The JSON format is ideal for:
- Integration with CI/CD pipelines  
- Custom tooling and automation
- Data analysis and reporting
- API responses and web interfaces
