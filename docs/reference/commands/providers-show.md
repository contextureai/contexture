---
title: contexture providers show
description: Display detailed information about a specific provider.
---
Display detailed information about a specific provider.

## Synopsis

```bash
contexture providers show <name>
```

## Arguments

| Argument | Description                                        |
| :------- | :------------------------------------------------- |
| `name`   | Name of the provider (with or without @ prefix)    |

## Description

The `providers show` command displays detailed information about a specific provider, including:
- Provider name
- Git repository URL
- Default branch
- Whether it's the default provider

You can specify the provider name with or without the `@` prefix.

## Usage

### Show Default Provider

```bash
contexture providers show contexture
contexture providers show @contexture
```

### Show Custom Provider

```bash
contexture providers show mycompany
contexture providers show @mycompany
```

### Example Output

```
Provider: @mycompany

Details:
  Name: mycompany
  URL: https://github.com/mycompany/rules.git
  Default Branch: main
  Type: Custom Provider

Rules from this provider can be referenced with:
  contexture rules add @mycompany/path/to/rule
```

## Related Commands

- [`contexture providers list`](./providers-list.md) - View all configured providers
- [`contexture providers add`](./providers-add.md) - Add a provider
- [`contexture query --provider`](./query.md) - Search rules from a specific provider
