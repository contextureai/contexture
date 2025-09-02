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