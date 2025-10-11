<div align="center">
  <a href="https://contexture.sh">
    <picture>
      <source srcset="web/assets/full-logo-dark.svg" media="(prefers-color-scheme: dark)">
      <source srcset="web/assets/full-logo-light.svg" media="(prefers-color-scheme: light)">
      <img src="web/assets/full-logo-light.svg" alt="Contexture logo">
    </picture>
  </a>
</div>
<p align="center">Manage agent context across all of your favorite AI platforms</p>
<p align="center">
  <a href="https://github.com/contextureai/contexture/actions/workflows/release.yml"><img alt="Build Status" src="https://img.shields.io/github/actions/workflow/status/contextureai/contexture/release.yml?style=flat-square&branch=main" /></a>
  <a href="https://github.com/contextureai/contexture/releases"><img alt="Latest Release" src="https://img.shields.io/github/v/release/contextureai/contexture?sort=semver&display_name=release&style=flat-square" /></a>
</p>

## What is Contexture?

Contexture is a CLI tool for managing AI assistant rules across multiple platforms (Claude, Cursor, Windsurf). It fetches rules from sources, processes templates with variables, and generates platform-specific output files.

## Installation

```bash
# via Go Install:
go install github.com/contextureai/contexture/cmd/contexture@latest
```

## Quick Start

1. Initialize a project
   ```bash
   contexture init
   ```
2. Discover available rules (optional)
   ```bash
   contexture query "go"
   ```
3. Add a rule
   ```bash
   contexture rules add @contexture/go/thought-process
   ```
4. Create your own custom rules (optional)
   ```bash
   contexture rules new my-custom-rule --name "My Rule" --tags "custom"
   ```

## Rules

Contexture comes with a set of official, maintained rules that you can find here:

[https://github.com/contextureai/rules](https://github.com/contextureai/rules)
