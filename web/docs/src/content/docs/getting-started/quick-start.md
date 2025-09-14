---
title: Quick Start
description: A quick start guide to getting started with contexture.
---
This tutorial explains how to set up a new `contexture` project and generate AI assistant rules.

## Prerequisites

- `contexture` is [installed](./installation).
- Git is available in your `PATH`.

## Step 1: Create a Project

To create a new project, navigate to a directory and run `contexture init`:

```bash
# Navigate to your project directory
cd /path/to/your/project

# Initialize a new project
contexture init
```

The `init` command opens an interactive prompt to select output formats:

```
? Select output formats to enable:
  ✓ claude (CLAUDE.md)
  ✓ cursor (.cursor/rules/)
  ✓ windsurf (.windsurf/rules/)
```

This creates a `.contexture.yaml` file with the initial configuration:

```yaml
version: 1
formats:
  - type: claude
    enabled: true
  - type: cursor
    enabled: true
  - type: windsurf
    enabled: true
rules: []
```

## Step 2: Add Rules

Add rules by specifying their IDs as arguments:

```bash
# Add code quality rules
contexture rules add code/clean-code code/error-handling

# Add documentation rules  
contexture rules add docs/readme-best-practices
```

Your `.contexture.yaml` will be updated to include these rules:

```yaml
version: 1
formats:
  - type: claude
    enabled: true
  - type: cursor 
    enabled: true
  - type: windsurf
    enabled: true
rules:
  - id: "[contexture:code/clean-code]"
  - id: "[contexture:code/error-handling]"
  - id: "[contexture:docs/readme-best-practices]"
```

## Step 3: Generate Output

To generate the output files for all enabled formats, run the `build` command:

```bash
contexture build
```

The `build` command:
1.  Fetches rules from the remote repository.
2.  Processes templates and variables.
3.  Generates format-specific output files.

The `build` command generates the following directory structure:

```
your-project/
├── .contexture.yaml          # Your configuration
├── CLAUDE.md                # Claude AI assistant rules
├── .cursor/
│   └── rules/               # Cursor IDE rules
└── .windsurf/
    └── rules/               # Windsurf IDE rules
```

Note: `contexture` automatically runs a build when rules are added or updated.

## Step 4: Verify the Setup

To verify the setup:

```bash
# List configured rules
contexture rules list

# Show the project configuration
contexture config

# Check for generated files
ls -la CLAUDE.md .cursor/rules/ .windsurf/rules/
```

## Next Steps

- **[Core Concepts](../core-concepts/overview)**: Understand how `contexture` works.
- **[Rules Documentation](../reference/rules/rule-references)**: Learn about rule structure and customization.
- **[Commands Reference](../reference/commands/init)**: Explore all available commands.

## Quick Reference

```bash
# Common commands
contexture init                    # Initialize a new project
contexture rules add              # Browse and add rules
contexture rules list             # Show configured rules
contexture build                  # Generate output files
contexture config           # View the project configuration
contexture --help                # Show help
```
