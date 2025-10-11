---
title: contexture providers add
description: Add a custom provider to your project configuration.
---
Add a custom provider to your project configuration.

## Synopsis

```bash
contexture providers add <name> <url>
```

## Arguments

| Argument | Description                                   |
| :------- | :-------------------------------------------- |
| `name`   | Unique name for the provider (without @ prefix) |
| `url`    | Git repository URL (HTTPS or SSH)              |

## Description

The `providers add` command adds a new custom provider to your `.contexture.yaml` configuration. Once added, the provider can be used with the `@provider/path` syntax when adding rules.

Provider names should be unique and cannot conflict with the default `contexture` provider name.

## Usage

### Add a Provider with HTTPS URL

```bash
contexture providers add mycompany https://github.com/mycompany/rules.git
```

### Add a Provider with SSH URL

```bash
contexture providers add team-security git@github.com:team/security-rules.git
```

### Use the Provider

After adding a provider, you can reference its rules using the `@provider/path` syntax:

```bash
# Discover rules from the new provider
contexture query --provider mycompany

# Add rules from the new provider
contexture rules add @mycompany/security/auth
contexture rules add @mycompany/testing/coverage
```

## Configuration

When you add a provider, it's stored in your `.contexture.yaml` file:

```yaml
providers:
  - name: mycompany
    url: https://github.com/mycompany/rules.git
    defaultBranch: main  # defaults to 'main' if not specified
```

## Tips

- **Naming**: Choose descriptive provider names (e.g., `mycompany`, `team-security`, `internal-rules`)
- **SSH vs HTTPS**: Use SSH URLs if you have SSH keys configured; use HTTPS for public repositories or with access tokens
- **Private Repositories**: Ensure your Git credentials are configured for the repository URL

## Related Commands

- [`contexture providers list`](./providers-list.md) - View all configured providers
- [`contexture providers remove`](./providers-remove.md) - Remove a provider
- [`contexture rules add`](./rules-add.md) - Add rules using the provider
