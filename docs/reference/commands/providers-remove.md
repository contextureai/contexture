---
title: contexture providers remove
description: Remove a custom provider from your project configuration.
---
Remove a custom provider from your project configuration.

## Synopsis

```bash
contexture providers remove <name>
```

## Aliases

- `rm`

## Arguments

| Argument | Description                                 |
| :------- | :------------------------------------------ |
| `name`   | Name of the provider to remove (without @ prefix) |

## Description

The `providers remove` command removes a custom provider from your `.contexture.yaml` configuration.

**Important:** You cannot remove the default `@contexture` provider.

## Usage

### Remove a Provider

```bash
contexture providers remove mycompany

# Using alias
contexture providers rm team-security
```

## Behavior

When you remove a provider:
1. The provider is removed from your `.contexture.yaml` file
2. **Rules from that provider are NOT automatically removed**
3. If you have rules from the removed provider, you'll need to manually remove them with `contexture rules remove`

## Warning

If you remove a provider that is still referenced by rules in your configuration, you'll encounter errors during the build process. Make sure to either:
1. Remove all rules from that provider first, or
2. Update the rules to use a different provider or direct repository URLs

## Related Commands

- [`contexture providers list`](./providers-list.md) - View all configured providers
- [`contexture providers add`](./providers-add.md) - Add a provider
- [`contexture rules remove`](./rules-remove.md) - Remove rules from the project
