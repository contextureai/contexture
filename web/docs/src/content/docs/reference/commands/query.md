---
title: contexture query
description: Search for rules across all configured providers using text or expression-based queries.
---
Search for rules across all configured providers using text or expression-based queries.

## Synopsis

```bash
contexture query [flags] <search-text>
```

## Description

The `query` command provides a powerful way to search for rules across all configured providers (both default and custom). It supports two query modes:

1. **Simple text search** (default): Searches for text in rule titles and IDs using AND logic
2. **Advanced expression search** (`--expr` flag): Uses the expr-lang expression language for complex queries with access to all rule fields

Unlike `rules list` which only shows rules already added to your project, `query` searches across all available rules from all providers, making it ideal for discovering new rules to add to your project.

## Flags

| Flag          | Description                                                                  |
| :------------ | :--------------------------------------------------------------------------- |
| `--expr` | Use advanced expression syntax for complex queries (see https://expr-lang.org) |
| `--output`, `-o` | Output format: `default` for terminal display, `json` for JSON output |
| `--provider` | Search specific providers only (can be specified multiple times) |
| `--limit`, `-n` | Maximum number of results to display (default: 50) |

## Usage

### Simple Text Search

By default, the query command performs a simple text search across rule titles and IDs. All search terms must match (AND logic).

```bash
# Find rules about testing
contexture query testing

# Find rules about Go testing (both "go" AND "testing" must match)
contexture query "go testing"

# Find authentication and security rules
contexture query auth security
```

### Advanced Expression Search

Use the `--expr` flag to unlock powerful expression-based queries using the expr-lang syntax.

```bash
# Find rules with specific tags
contexture query --expr 'Tag contains "testing"'

# Find Go language rules
contexture query --expr 'Language == "go"'

# Find rules with variables
contexture query --expr 'HasVars == true'

# Complex queries with AND/OR logic
contexture query --expr 'Language == "go" and any(Tags, # in ["security", "auth"])'

# Find rules by content
contexture query --expr 'Content contains "best practices"'

# Find rules with specific trigger types
contexture query --expr 'TriggerType == "file_change"'
```

#### Available Fields for Expressions

**Direct fields:**
- `ID` (string) - Rule identifier
- `Title` (string) - Rule title
- `Description` (string) - Rule description
- `Tags` ([]string) - Rule tags
- `Languages` ([]string) - Supported languages
- `Frameworks` ([]string) - Supported frameworks
- `Content` (string) - Rule content/body
- `Variables` (map) - Rule variables
- `Source` (string) - Source provider
- `FilePath` (string) - File path in repository

**Computed fields:**
- `Tag` (string) - Space-joined tags
- `Language` (string) - Space-joined languages
- `Framework` (string) - Space-joined frameworks
- `Provider` (string) - Provider name
- `Path` (string) - Extracted path from ID
- `HasVars` (bool) - Has variables
- `VarCount` (int) - Number of variables
- `TriggerType` (string) - Trigger type

For complete expression syntax documentation, see: https://expr-lang.org/docs/language-definition

### Filter by Provider

Search only specific providers using the `--provider` flag.

```bash
# Search only the default contexture provider
contexture query --provider contexture testing

# Search multiple providers
contexture query --provider mycompany --provider team-rules security

# Combine with expression queries
contexture query --provider contexture --expr 'Language == "go"'
```

### Limit Results

Control the number of results displayed.

```bash
# Show only first 10 results
contexture query --limit 10 testing

# Use short flag
contexture query -n 5 security

# Increase limit for comprehensive search
contexture query --limit 100 "error handling"
```

### JSON Output

Use JSON output for programmatic processing or integration with other tools.

```bash
# Output as JSON
contexture query --output json testing

# Use short flag
contexture query -o json "go test"

# Combine with expressions
contexture query --expr 'HasVars == true' -o json
```

**JSON Structure:**
```json
{
  "metadata": {
    "query": "testing",
    "queryType": "text",
    "totalResults": 15
  },
  "rules": [
    {
      "id": "@contexture/languages/go/testing",
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

The default terminal output displays matching rules in a compact format:
- **Rule Path**: The rule's full identifier including provider (e.g., `@contexture/languages/go/testing`)
- **Title**: A descriptive title (e.g., `Go Testing Best Practices`)
- **Header**: Shows total count of matching rules

Results are sorted by path for consistent output and limited to 50 by default (configurable with `--limit`).

### JSON Output Format

JSON output provides structured data suitable for programmatic processing:
- **Metadata**: Query string, query type (text/expr), and total results count
- **Rules Array**: Complete rule objects with IDs, metadata, variables, and content
- **Consistent Schema**: Stable field names that match the CLI structs

The JSON format is ideal for:
- Discovery workflows and rule browsing tools
- Integration with search interfaces
- Automated rule selection and installation
- Analytics and reporting on available rules

## Examples

### Discovering New Rules

```bash
# Find all Go-related rules across all providers
contexture query go

# Find security rules you might want to add
contexture query --expr 'any(Tags, # in ["security", "auth"])'

# Browse testing rules
contexture query testing --limit 20
```

### Complex Searches

```bash
# Find Go rules with variables from a specific provider
contexture query --provider mycompany --expr 'Language == "go" and HasVars == true'

# Find all file-triggered rules
contexture query --expr 'TriggerType == "file_change"'

# Find rules with specific content
contexture query --expr 'Content contains "performance" and Language == "go"'
```

### Programmatic Usage

```bash
# Get rule IDs for automation
contexture query -o json testing | jq -r '.rules[].id'

# Find rules to add to your project
contexture query --expr 'Language == "typescript"' -o json | jq '.metadata.totalResults'

# Filter and process with jq
contexture query go -o json | jq '.rules[] | select(.tags[] | contains("testing"))'
```

## Tips

- **Start simple**: Use text search first, then switch to expressions for complex queries
- **Use expressions for precision**: When you need exact matches or complex logic
- **Explore with limit**: Use `--limit 10` when exploring to avoid overwhelming output
- **Check available providers**: Run `contexture providers list` to see what providers you can search
- **JSON for scripting**: Use JSON output with `jq` for powerful automation workflows

## Related Commands

- [`contexture rules add`](./rules-add.md) - Add discovered rules to your project
- [`contexture rules list`](./rules-list.md) - List rules already in your project
- [`contexture providers list`](../providers/list.md) - View available providers
