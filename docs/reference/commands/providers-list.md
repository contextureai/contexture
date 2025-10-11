---
title: contexture providers list
description: List all available providers including the default and custom providers.
---
List all available providers including the default @contexture provider and any custom providers configured in your project.

## Synopsis

```bash
contexture providers list
```

## Aliases

- `ls`

## Description

The `providers list` command displays all providers available to your project. This includes:
- The default `@contexture` provider (always available)
- Any custom providers you've added to your project

For each provider, the command shows:
- Provider name (with @ prefix)
- Git repository URL
- Whether it's the default provider

## Usage

### List All Providers

```bash
contexture providers list

# Using alias
contexture providers ls
```

### Example Output

```
Providers

Available providers:
  @contexture (default)
    https://github.com/contextureai/rules.git
  @mycompany
    https://github.com/mycompany/rules.git
  @team-rules
    git@github.com:team/security-rules.git
```

## Related Commands

- [`contexture providers add`](./providers-add.md) - Add a custom provider
- [`contexture providers show`](./providers-show.md) - View details for a specific provider
- [`contexture query`](./query.md) - Search rules across all providers
