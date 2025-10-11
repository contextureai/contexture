---
title: contexture providers
description: Manage rule providers for your Contexture project.
---
Manage rule providers for your Contexture project.

## Synopsis

```bash
contexture providers [subcommand]
```

## Description

Providers are named references to rule repositories that enable clean, readable rule references using the `@provider/path` syntax. The `providers` command allows you to list, add, remove, and view details about configured providers.

Every Contexture project has access to the default `@contexture` provider, which points to the community-maintained rules repository. You can add custom providers for your team's private rules or other third-party rule collections.

## Subcommands

| Subcommand | Description                                  |
| :--------- | :------------------------------------------- |
| `list`     | List all available providers                 |
| `add`      | Add a custom provider                        |
| `remove`   | Remove a custom provider                     |
| `show`     | Show details for a specific provider         |

## Usage

### List Providers

Display all available providers, including the default @contexture provider and any custom providers.

```bash
contexture providers list
```

### Add a Provider

Add a custom provider to your project configuration.

```bash
# Add a provider with HTTPS URL
contexture providers add mycompany https://github.com/mycompany/rules.git

# Add a provider with SSH URL
contexture providers add team-security git@github.com:team/security-rules.git
```

Once added, you can reference rules from this provider using the `@provider/path` syntax:

```bash
contexture rules add @mycompany/security/auth
```

### Remove a Provider

Remove a custom provider from your project configuration.

```bash
contexture providers remove mycompany
```

**Note:** You cannot remove the default `@contexture` provider.

### Show Provider Details

Display detailed information about a specific provider, including its URL and default branch.

```bash
# Show default provider
contexture providers show contexture

# Show custom provider (with or without @ prefix)
contexture providers show @mycompany
contexture providers show mycompany
```

## Provider Configuration

When you add a provider, it is stored in your `.contexture.yaml` file:

```yaml
providers:
  - name: mycompany
    url: https://github.com/mycompany/rules.git
    defaultBranch: main
```

## Default Provider

The `@contexture` provider is always available and points to the community-maintained rules repository:
- **Name**: `contexture`
- **URL**: `https://github.com/contextureai/rules.git`
- **Default Branch**: `main`

## Related Commands

- [`contexture rules add`](./rules-add.md) - Add rules using provider syntax
- [`contexture query`](./query.md) - Search for rules across all providers
- [Configuration File Reference](../configuration/config-file.md) - Provider configuration details
