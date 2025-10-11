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
    cursor (.cursor/rules/)
    windsurf (.windsurf/rules/)
```

Claude is preselected; use the space bar to enable Cursor and Windsurf before pressing enter if you want them generated.

When all formats are enabled, the resulting `.contexture.yaml` file looks like:

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

## Step 2: Discover Rules (Optional)

Before adding rules, you can search for available rules across all providers using the `query` command:

```bash
# Search for testing-related rules
contexture query testing

# Search for Go language rules
contexture query go

# Use advanced expressions for complex searches
contexture query --expr 'Language == "go" and any(Tags, # in ["testing", "best-practices"])'

# Limit results for easier browsing
contexture query testing --limit 10

# Get JSON output for scripting
contexture query testing -o json
```

The query command helps you discover rules before adding them to your project. See [`contexture query`](../reference/commands/query) for more details.

## Step 3: Add Rules

Add rules by specifying their IDs as arguments:

```bash
# Add code quality rules
contexture rules add code/clean-code code/error-handling

# Add documentation rules  
contexture rules add docs/readme-best-practices
```

#### Using Custom Sources

You can also add rules from your own repositories:

```bash
# Add rules from a custom repository
contexture rules add my/custom-rule --src https://github.com/mycompany/contexture-rules.git

# Add specific rules from a custom source
contexture rules add "security/auth" --src "git@github.com:mycompany/rules.git"
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

#### Creating Custom Rules (Optional)

You can create your own project-specific rules using the `rules new` command:

```bash
# Create a custom rule with metadata
contexture rules new project-conventions \
  --name "Project Conventions" \
  --description "Our team's coding standards" \
  --tags "conventions,team,standards"

# Edit the generated file to add your content
vim rules/project-conventions.md

# Add the custom rule to your project
contexture rules add rules/project-conventions.md
```

This creates a new rule file in your project's `rules/` directory with YAML frontmatter. See [`contexture rules new`](../reference/commands/rules-new) for more details.

## Step 4: Generate Output

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

## Step 5: Customize with Templates (Optional)

You can customize the structure of your `CLAUDE.md` file using a custom template:

1. **Create a template file** in your project root:

```bash
# Create a custom template
cat > CLAUDE.template.md << 'EOF'
# My Team's AI Assistant Instructions

## Project Overview
This project follows our established development practices and coding standards.

## Team Guidelines  
- All code must be reviewed before merging
- Tests are required for new functionality
- Follow the style guide in our documentation

## Generated Rules
{{.Rules}}

## Additional Resources
- Check our internal documentation wiki
- Refer to the project README for setup instructions
- Contact the team leads for architectural decisions
EOF
```

2. **Update your configuration** to use the template:

```yaml
# In .contexture.yaml
formats:
  - type: claude
    enabled: true
    template: CLAUDE.template.md
  - type: cursor
    enabled: true
  - type: windsurf
    enabled: true
```

3. **Rebuild to apply the template**:

```bash
contexture build
```

Now your `CLAUDE.md` file will use your custom structure while still including all the generated rules.

## Step 6: Verify the Setup

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
contexture init                                      # Initialize a new project
contexture query testing                            # Search for rules across all providers
contexture query --expr 'Language == "go"'         # Advanced expression search
contexture rules add code/clean-code              # Add specific rules
contexture rules add security/auth --src https://github.com/...  # Add rules from a custom source
contexture rules new my-rule --name "My Rule"     # Create a new custom rule
contexture rules list                               # Show configured rules
contexture rules list -o json                      # Show rules as JSON
contexture build                                    # Generate output files
contexture config                                   # View the project configuration
contexture --help                                  # Show help

# Template customization (Claude format only)
# 1. Create CLAUDE.template.md with {{.Rules}} placeholder
# 2. Add 'template: CLAUDE.template.md' to claude format config
# 3. Run 'contexture build' to apply template
```
