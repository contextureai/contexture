---
title: References
description: Overview of rule references in contexture.
---
A rule reference is a string that identifies a rule to be included in a project. It supports referencing rules from Git repositories and local file paths.

## Reference Syntax

The syntax for a rule reference is:

`[contexture(source):path/to/rule,ref] {"key": "value"}`

| Component        | Description                                                                                             | Required |
| :--------------- | :------------------------------------------------------------------------------------------------------ | :------- |
| `[...]`          | Brackets enclosing the rule identifier.                                                                 | Yes      |
| `contexture`     | Namespace prefix.                                                                                       | Yes      |
| `(source)`       | The name of a source defined in the `sources` section of `.contexture.yaml`. Defaults to `contexture`. | No       |
| `:`              | Separator.                                                                                              | Yes      |
| `path/to/rule`   | The path to the rule file within the repository, without the `.md` extension.                               | Yes      |
| `,ref`           | A Git branch, tag, or commit hash. Defaults to the source's configured branch or `main`.                 | No       |
| `{...}`          | A JSON5 object of variables to apply to the rule.                                                       | No       |

## Remote Rule References

Remote rules are fetched from Git repositories.

### Default Repository

If the `source` is omitted, the default `contexture` community repository is used (`https://github.com/contextureai/rules.git`).

**Example:**
`[contexture:code/clean-code]`

This references the rule at `code/clean-code.md` in the default repository on the `main` branch.

### Branch and Tag References

To use a specific version of a rule, specify a branch, tag, or commit hash.

**Examples:**
- **Branch**: `[contexture:experimental/new-feature,development]`
- **Tag**: `[contexture:stable/patterns,v1.2.0]`

## Local Rule References

Local rules are referenced by their file path relative to the project root. The `rules/` directory is the conventional location for local rules.

**Example:**
`rules/project-specific-rule.md`

## Variables in References

Variables can be passed to a rule directly in the reference string as a JSON5 object.

**Example:**
`[contexture:testing/coverage] {"threshold": 90, "framework": "jest"}`

These variables have the highest precedence and will override any default variables defined in the rule's frontmatter.

## Command-Line Sources

In addition to configured sources in `.contexture.yaml`, you can specify custom sources directly via command-line flags when adding rules.

### Using `--source` or `--src` Flags

The `--source` (or its shorthand `--src`) and `--ref` flags allow you to add rules from repositories that aren't configured in your project.

```bash
# Interactive browsing from custom source
contexture rules add --src https://github.com/company/rules.git

# Adding specific rules from custom source  
contexture rules add "security/auth" --src "git@github.com:company/rules.git" --ref "v2.0"

# Multiple rules from same custom source
contexture rules add "api/validation" "api/errors" --src "https://github.com/team/api-rules.git"
```

### Differences from Configured Sources

| Aspect | Configured Sources | Command-Line Sources |
|:-------|:------------------|:-------------------|
| **Definition** | Defined in `.contexture.yaml` `sources` section | Specified via `--source`/`--src` flag |
| **Reusability** | Reusable across multiple rules | One-time use per command |
| **Authentication** | Configured with auth tokens/SSH keys | Uses system Git credentials |
| **Branch/Tag** | Default branch configured in source | Specified via `--ref` or uses `main` |

When using command-line sources, the resulting rule references in `.contexture.yaml` will include the full source URL and reference:

```yaml
rules:
  - id: "[contexture(https://github.com/company/rules.git):security/auth,v2.0]"
    source: "https://github.com/company/rules.git"
    ref: "v2.0"
```