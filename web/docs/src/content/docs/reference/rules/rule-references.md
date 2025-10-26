---
title: References
description: Overview of rule references in contexture.
---
A rule reference is a string that identifies a rule to be included in a project. It supports referencing rules from Git repositories and local file paths.

## Reference Syntax

Contexture supports two syntaxes for referencing rules:

### Recommended: Provider Syntax (Preferred)

The modern, clean syntax using the `@provider/path` format:

```
@provider/path/to/rule
```

| Component        | Description                                                                                    | Required |
| :--------------- | :--------------------------------------------------------------------------------------------- | :------- |
| `@`              | Provider prefix                                                                                | Yes      |
| `provider`       | Provider name (`contexture` for default, or custom provider name)                              | Yes      |
| `/`              | Separator                                                                                      | Yes      |
| `path/to/rule`   | Path to the rule file within the repository (without `.md` extension)                          | Yes      |

**Examples:**
- `@contexture/code/clean-code` - Rule from default provider
- `@mycompany/security/auth` - Rule from custom provider

### Alternative: Bracketed Syntax (Legacy)

The legacy syntax for advanced use cases:

```
[contexture(source):path/to/rule,ref] {"key": "value"}
```

| Component        | Description                                                                                             | Required |
| :--------------- | :------------------------------------------------------------------------------------------------------ | :------- |
| `[...]`          | Brackets enclosing the rule identifier                                                                  | Yes      |
| `contexture`     | Namespace prefix                                                                                        | Yes      |
| `(source)`       | Provider name or direct URL. Defaults to `contexture`                                                   | No       |
| `:`              | Separator                                                                                               | Yes      |
| `path/to/rule`   | Path to the rule file within the repository (without `.md` extension)                                   | Yes      |
| `,ref`           | Git branch, tag, or commit hash. Defaults to provider's default branch or `main`                        | No       |
| `{...}`          | JSON5 object of variables to apply to the rule                                                          | No       |

## Remote Rule References

Remote rules are fetched from Git repositories using providers.

### Default Provider

The `@contexture` provider is always available and points to the community-maintained rules repository (`https://github.com/contextureai/rules.git`).

**Examples:**
```bash
# Provider syntax (recommended)
contexture rules add @contexture/code/clean-code

# Bracketed syntax (alternative)
contexture rules add "[contexture:code/clean-code]"
```

This references the rule at `code/clean-code.md` in the default repository on the `main` branch.

### Branch and Tag References

To use a specific version of a rule, use the `--ref` flag when adding rules:

**Examples:**
```bash
# Use a specific branch
contexture rules add @contexture/experimental/new-feature --ref development

# Use a specific tag
contexture rules add @contexture/stable/patterns --ref v1.2.0

# Bracketed syntax alternative
contexture rules add "[contexture:experimental/new-feature,development]"
contexture rules add "[contexture:stable/patterns,v1.2.0]"
```

### Custom Providers

To use rules from a custom provider, first add the provider, then reference its rules:

```bash
# Add a custom provider
contexture providers add mycompany https://github.com/mycompany/rules.git

# Use rules from the custom provider
contexture rules add @mycompany/security/auth
contexture rules add @mycompany/testing/coverage
```

## Local Rule References

Local rules are referenced by their file path relative to the project root. The `rules/` directory is the conventional location for local rules.

**Example:**
```bash
contexture rules add rules/project-specific-rule.md
```

## Variables in References

Variables can be passed to a rule using the `--var` flag or inline as a JSON5 object.

**Examples:**
```bash
# Using --var flag (recommended)
contexture rules add @contexture/testing/coverage --var threshold=90

# Inline JSON5 (alternative)
contexture rules add '@contexture/testing/coverage {"threshold": 90, "framework": "jest"}'

# Bracketed syntax with inline JSON5
contexture rules add '[contexture:testing/coverage] {"threshold": 90, "framework": "jest"}'
```

These variables have the highest precedence and will override any default variables defined in the rule's frontmatter.

## Command-Line Sources

In addition to configured providers in `.contexture.yaml`, you can specify custom sources directly via command-line flags when adding rules.

### Using `--source` or `--src` Flags

The `--source` (or its shorthand `--src`) and `--ref` flags allow you to add rules from repositories that aren't configured as providers.

**Examples:**
```bash
# Adding a rule from a custom source
contexture rules add security/auth --src git@github.com:company/rules.git --ref v2.0

# Multiple rules from the same custom source
contexture rules add api/validation api/errors --src https://github.com/team/api-rules.git
```

**Note:** For frequently used repositories, consider adding them as providers using `contexture providers add` for cleaner syntax.

### Configured Providers vs. Command-Line Sources

| Aspect | Configured Providers | Command-Line Sources |
|:-------|:------------------|:-------------------|
| **Definition** | Defined in `.contexture.yaml` `providers` section | Specified via `--source`/`--src` flag |
| **Syntax** | Clean `@provider/path` syntax | Path only, source in flag |
| **Reusability** | Reusable across multiple rules with `@provider` | One-time use per command |
| **Authentication** | Configured with auth tokens/SSH keys | Uses system Git credentials |
| **Branch/Tag** | Default branch configured in provider | Specified via `--ref` or uses `main` |

When using command-line sources, the resulting rule references in `.contexture.yaml` include the full source URL:

```yaml
rules:
  - id: "[contexture(https://github.com/company/rules.git):security/auth,v2.0]"
    source: "https://github.com/company/rules.git"
    ref: "v2.0"
```

## Summary of Syntax Options

```bash
# Recommended: Provider syntax (clean and reusable)
contexture rules add @contexture/code/clean-code
contexture rules add @mycompany/security/auth

# With branch/tag
contexture rules add @contexture/testing/patterns --ref v1.2.0

# With variables
contexture rules add @contexture/testing/coverage --var threshold=90

# Command-line source (one-time use)
contexture rules add security/auth --src git@github.com:company/rules.git

# Local files
contexture rules add rules/project-specific-rule.md

# Bracketed syntax (legacy/alternative)
contexture rules add "[contexture:code/clean-code]"
```
