# Projects and Configuration

A `contexture` project is a directory that contains a `.contexture.yaml` file. This file defines how AI assistant rules are managed and generated.

## Project Structure

A basic project layout includes the configuration file and any generated format outputs. Local rules can be stored in an optional `rules/` directory.

```
my-project/
├── .contexture.yaml
├── rules/
│   └── project-specific.md
├── CLAUDE.md
└── .cursor/
    └── rules/
```

## Configuration File

The `.contexture.yaml` file is the central configuration for a project.

### Example

```yaml
# .contexture.yaml
version: 1

# Output format configurations
formats:
  - type: claude
    enabled: true
  - type: cursor
    enabled: true

# Rule references with variables
rules:
  # Rule from the default repository with variables
  - id: "[contexture:code/clean-code]"
    variables:
      maxLineLength: 100

  # Rule from a custom source
  - id: "[contexture(company):standards/coding-conventions]"
```

### Top-Level Keys

-   `version`: The configuration format version. Currently `1`.
-   `formats`: A list of output format configurations.
-   `rules`: A list of rules to include in the project.
-   `generation`: Performance and behavior settings.

### `formats`

The `formats` key is a list of output formats to generate.

```yaml
formats:
  - type: claude         # 'claude', 'cursor', or 'windsurf'
    enabled: true        # Whether to generate this format
```

### `rules`

The `rules` key is a list of rules to include.

```yaml
rules:
  # Remote rule from the default repository
  - id: "[contexture:path/to/rule]"

  # Remote rule from a custom source and branch
  - id: "[contexture(company):path/to/rule,feature-branch]"

  # Rule with variables
  - id: "[contexture:path/to/rule]"
    variables:
      key: value
```

## Next Steps

-   **[Variables](variables.md)**: Template variables and customization.
-   **[Configuration Guide](../configuration/)**: Detailed configuration documentation.
-   **[Commands Reference](../commands/)**: All available commands.