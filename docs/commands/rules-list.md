# `contexture rules list`

Displays the rules configured in the `.contexture.yaml` file.

## Synopsis

```bash
contexture rules list [flags]
```

## Aliases

-   `ls`

## Description

The `rules list` command prints a list of all rules that have been added to the project. The output includes the rule's ID, title, and description. Using the `--verbose` flag provides a more detailed view.

## Flags

| Flag          | Description                                                                  |
| :------------ | :--------------------------------------------------------------------------- |
| `--verbose`, `-v` | Show detailed information for each rule, including metadata and source info. |
| `--formats`   | Filter the list to show only rules compatible with the specified format(s).  |

## Usage

### Standard List

Displays a summary of each configured rule.

```bash
contexture rules list
```

### Verbose List

Displays detailed information for each rule.

```bash
contexture rules list --verbose
```

### Filtering by Format

To see which rules apply to a specific output format, use the `--formats` flag. The flag can be used multiple times.

```bash
# Show rules for the 'claude' format
contexture rules list --formats claude

# Show rules for both 'cursor' and 'windsurf' formats
contexture rules list --formats cursor --formats windsurf
```
