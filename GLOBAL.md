# Global User-Level Rules Implementation Plan

## Implementation Status

**Status: ✅ COMPLETE**
**Completion Date: 2025-10-26**

All planned features have been fully implemented, tested, documented, and meet code quality standards. This includes:
- ✅ All command-level `-g` flags (rules add/remove/update/new, providers add/remove, config)
- ✅ Automatic merging in list and build commands
- ✅ Global configuration lazy initialization
- ✅ Source indicators in output displays
- ✅ Full test coverage (unit, integration, E2E)
- ✅ Documentation updates (config, rules-list, providers-list)
- ✅ Code quality standards (all linter checks passing)
- ✅ Backward compatibility maintained

---

## Overview

This document outlines the implementation plan for adding global user-level rules support to Contexture. Global rules will be stored in `~/.contexture/` and automatically merged with project-specific rules, with project rules taking precedence when IDs conflict.

### Goals

1. Allow users to define rules once that apply across all their projects
2. Enable per-project overrides of global rules
3. Maintain backwards compatibility with existing projects
4. Keep implementation minimal and reuse existing infrastructure

### Non-Goals

- Complex inheritance mechanisms beyond simple override
- Global-specific command namespaces
- Automatic synchronization between projects
- Migration tools for moving rules between global/project

## User-Facing Behavior

### Directory Structure

Global configuration follows the same structure as project configuration:

```
~/.contexture/
├── .contexture.yaml    # Global configuration file
└── rules/              # Global rule files (optional)
    ├── my-rule.md
    └── team-standard.md
```

### Command Usage

All rule and provider commands support the `-g/--global` flag:

```bash
# Add a rule globally
contexture rules add -g @contexture/languages/go/context

# Remove a global rule
contexture rules remove -g @contexture/languages/go/context

# Create a new global rule
contexture rules new -g my-global-rule

# List all rules (global + project)
contexture rules list

# Update global rule variables
contexture rules update -g @contexture/languages/go/context --var key=value

# Manage global providers
contexture providers add -g teamrules https://github.com/company/rules.git

# View/edit global config
contexture config -g
```

### Build Behavior

The `contexture build` command automatically merges global and project rules:

```bash
# Generates output using global + project rules
contexture build
```

No flag is needed - global rules are always included. Project rules with matching IDs override global rules.

### List Output

The `list` command shows both global and project rules with source indicators:

```
Rules (5)

  @contexture/languages/go/context [global]
  @contexture/testing/unit [global]
  @mycompany/security/auth [project]
  languages/go/custom [project] [overrides global]
  security/custom [project]
```

Rules that override global rules are clearly marked.

## Configuration Structure

### Global Configuration File

`~/.contexture/.contexture.yaml` follows the same schema as project configuration:

```yaml
version: 1

providers:
  - name: teamrules
    url: https://github.com/company/contexture-rules.git
    defaultBranch: main

formats:
  - type: claude
    enabled: true
  - type: cursor
    enabled: true
  - type: windsurf
    enabled: true

rules:
  - id: "@contexture/languages/go/context"
  - id: "@teamrules/security/auth"
    variables:
      strict: true
  - id: "[contexture(local):my-global-rule]"

generation:
  parallelFetches: 5
  cacheEnabled: true
```

### Initialization

Global configuration is created lazily on first use of `-g` flag:

1. User runs `contexture rules add -g <rule-id>`
2. System checks if `~/.contexture/.contexture.yaml` exists
3. If not, creates directory and default config file
4. Adds rule to global config

No explicit initialization command is needed.

## Merge & Precedence Rules

### Configuration Merging

When loading configuration, the system:

1. Loads global config from `~/.contexture/.contexture.yaml` (if exists)
2. Loads project config from `.contexture/.contexture.yaml` or `.contexture.yaml`
3. Merges them with project taking precedence

### Rule Merging

Rules are merged by ID:

- **Unique rules**: All rules from both configs are included
- **Duplicate IDs**: Project rule completely replaces global rule
  - Project rule variables override global variables
  - No partial merging of variables
- **Order**: Global rules first, then project rules (maintaining list order)

Example:

```yaml
# Global config
rules:
  - id: "@contexture/languages/go/context"
    variables:
      style: "strict"
  - id: "@contexture/testing/unit"

# Project config
rules:
  - id: "@contexture/languages/go/context"
    variables:
      style: "relaxed"
  - id: "custom/project-rule"

# Merged result (what build sees)
rules:
  - id: "@contexture/testing/unit"           # From global (no conflict)
  - id: "@contexture/languages/go/context"   # From project (overrides global)
    variables:
      style: "relaxed"
  - id: "custom/project-rule"                # From project (unique)
```

### Rule ID Matching

Two rules are considered "the same" if their IDs match after normalization:

1. Extract the core path from both IDs (strip source, ref, variables)
2. Compare paths case-insensitively
3. Both `[contexture:path]` and `@contexture/path` match if paths are identical

Examples of matching IDs:
- `@contexture/languages/go/context` matches `[contexture:languages/go/context]`
- `@teamrules/security/auth` matches `[contexture(teamrules-git-url):security/auth]`
- Case-insensitive: `Security/Auth` matches `security/auth`

### Provider Merging

Providers are merged by name:

- **Unique providers**: All providers from both configs are included
- **Duplicate names**: Project provider completely replaces global provider
- Project can override global provider URL or authentication

### Format Configuration

Project format configuration takes complete precedence:

- If project defines formats, use project's format list
- If project doesn't define formats, fall back to global formats
- No per-format merging (all-or-nothing)

### Generation Settings

Generation settings use project values with global fallbacks:

- `parallelFetches`: Project value or global value or default (5)
- `defaultBranch`: Project value or global value or default ("main")
- `cacheEnabled`: Project value or global value or default (true)
- `cacheTTL`: Project value or global value or default

## Command Modifications

### `contexture rules add`

Add `-g/--global` flag:

```go
&cli.BoolFlag{
    Name:    "global",
    Aliases: []string{"g"},
    Usage:   "Add rule to global configuration (~/.contexture)",
}
```

**Behavior:**
- `-g` flag: Add to `~/.contexture/.contexture.yaml`
- No flag: Add to project config (existing behavior)
- Create global config directory/file if doesn't exist
- Validate rule before adding (same validation as project rules)
- Auto-run build after adding (respects merged config)

**Implementation Notes:**
- Modify `AddCommand.ExecuteWithDeps()` in `internal/commands/add.go`
- Add logic to detect `-g` flag and route to appropriate config
- Use `projectManager.LoadGlobalConfig()` (new method)
- Use `projectManager.SaveGlobalConfig()` (new method)

### `contexture rules remove`

Add `-g/--global` flag:

```go
&cli.BoolFlag{
    Name:    "global",
    Aliases: []string{"g"},
    Usage:   "Remove rule from global configuration",
}
```

**Behavior:**
- `-g` flag: Remove from global config only
- No flag: Remove from project config only (existing behavior)
- Show warning if rule doesn't exist in target config
- If rule exists in both configs and no `-g` flag, only remove from project (global remains)

**Implementation Notes:**
- Modify `RemoveCommand.Execute()` in `internal/commands/remove.go`
- Add logic to detect `-g` flag
- Handle case where global config doesn't exist (error or warning)

### `contexture rules list`

No new flags - automatically shows both global and project rules:

**Behavior:**
- Load both global and project configs
- Merge rules (as per build behavior)
- Display with source indicators:
  - `[global]` - Rule from global config only
  - `[project]` - Rule from project config only
  - `[project overrides global]` - Rule exists in both, showing project version
- Support existing `--pattern` flag (filters merged list)
- Support existing `--output` formats (json, default)

**JSON Output:**
```json
{
  "rules": [
    {
      "id": "@contexture/languages/go/context",
      "source": "global",
      "overridden": false,
      ...
    },
    {
      "id": "@mycompany/security/auth",
      "source": "project",
      "overridesGlobal": true,
      ...
    }
  ],
  "metadata": {
    "totalRules": 5,
    "globalRules": 2,
    "projectRules": 3,
    "overriddenGlobal": 1
  }
}
```

**Implementation Notes:**
- Modify `ListCommand.Execute()` in `internal/commands/list.go`
- Add new `listInstalledRulesWithGlobal()` method
- Track rule sources during merge
- Update display formatters to show source indicators
- Extend `output.ListMetadata` with source statistics

### `contexture rules update`

Add `-g/--global` flag:

```go
&cli.BoolFlag{
    Name:    "global",
    Aliases: []string{"g"},
    Usage:   "Update global rule configuration",
}
```

**Behavior:**
- `-g` flag: Update rule in global config
- No flag: Update rule in project config (existing behavior)
- Can update variables, pin/unpin rules
- Cannot move rules between global and project (must remove and re-add)

**Implementation Notes:**
- Modify `UpdateCommand.Execute()` in `internal/commands/update.go`
- Similar routing logic to add/remove commands

### `contexture rules new`

Add `-g/--global` flag:

```go
&cli.BoolFlag{
    Name:    "global",
    Aliases: []string{"g"},
    Usage:   "Create new rule in global rules directory",
}
```

**Behavior:**
- `-g` flag: Create rule file in `~/.contexture/rules/`
- No flag: Create in `.contexture/rules/` (existing behavior)
- After creation, automatically add rule reference to appropriate config
- Use same interactive prompts for metadata

**Implementation Notes:**
- Modify `NewCommand.Execute()` in `internal/commands/new.go`
- Detect `-g` flag and change target directory
- Create global rules directory if doesn't exist

### `contexture providers add`

Add `-g/--global` flag:

```go
&cli.BoolFlag{
    Name:    "global",
    Aliases: []string{"g"},
    Usage:   "Add provider to global configuration",
}
```

**Behavior:**
- `-g` flag: Add to global providers list
- No flag: Add to project providers (existing behavior)

**Implementation Notes:**
- Modify provider command in `internal/commands/providers.go`
- Add routing logic for global vs project config

### `contexture providers remove`

Add `-g/--global` flag with same behavior as add.

### `contexture providers list`

No flag needed - automatically show both:

**Output:**
```
Providers (3)

  contexture [built-in]
    https://github.com/contextureai/rules.git

  teamrules [global]
    https://github.com/company/rules.git

  localdev [project]
    file:///home/user/custom-rules
```

**Implementation Notes:**
- Merge global and project providers
- Show source indicators
- Mark provider if it's overridden

### `contexture providers show`

Show provider details from either global or project:

**Behavior:**
- Search project providers first
- If not found, search global providers
- Show which config it's from

### `contexture build`

No flags added - automatically merges global and project:

**Behavior:**
- Load global config (if exists)
- Load project config
- Merge configurations (as per precedence rules)
- Generate output files for merged config
- No change to user-facing command

**Implementation Notes:**
- Modify `BuildCommand.Execute()` in `internal/commands/build.go`
- Use new `LoadConfigMerged()` method that handles global + project
- Ensure rule fetcher can find global rules

### `contexture config`

Add `-g/--global` flag:

```go
&cli.BoolFlag{
    Name:    "global",
    Aliases: []string{"g"},
    Usage:   "View or modify global configuration",
}
```

**Behavior:**
- `-g` flag: Display `~/.contexture/.contexture.yaml`
- No flag: Display project config (existing behavior)
- Support existing subcommands (show, edit, etc.) with global targeting

**Implementation Notes:**
- Modify config command in `internal/commands/config.go`
- Add routing based on `-g` flag

### `contexture init`

No changes - remains project-only:

**Rationale:**
- `init` is explicitly for project setup
- Global config is created lazily on first use
- No need for explicit global init

## Implementation Details

### Domain Models (`internal/domain/`)

#### New Types

Add source tracking for rules:

```go
// RuleSource indicates where a rule comes from
type RuleSource string

const (
    RuleSourceGlobal  RuleSource = "global"
    RuleSourceProject RuleSource = "project"
)

// RuleWithSource wraps a RuleRef with its source information
type RuleWithSource struct {
    RuleRef          RuleRef
    Source           RuleSource
    OverridesGlobal  bool
}

// ConfigSource identifies which configuration a value comes from
type ConfigSource string

const (
    ConfigSourceGlobal  ConfigSource = "global"
    ConfigSourceProject ConfigSource = "project"
)

// MergedConfig represents the result of merging global and project configs
type MergedConfig struct {
    Project      *Project
    GlobalConfig *Project
    MergedRules  []RuleWithSource
}
```

#### Extend Existing Types

Add location type for global:

```go
const (
    ConfigLocationRoot       ConfigLocation = "root"
    ConfigLocationContexture ConfigLocation = "contexture"
    ConfigLocationGlobal     ConfigLocation = "global"  // NEW
)

// GetGlobalConfigDir returns the global contexture directory
func GetGlobalConfigDir() (string, error) {
    homeDir, err := os.UserHomeDir()
    if err != nil {
        return "", err
    }
    return filepath.Join(homeDir, ".contexture"), nil
}

// GetGlobalConfigPath returns the global configuration file path
func GetGlobalConfigPath() (string, error) {
    dir, err := GetGlobalConfigDir()
    if err != nil {
        return "", err
    }
    return filepath.Join(dir, GetConfigFileName()), nil
}
```

### Project Manager (`internal/project/config.go`)

#### New Methods

Add global configuration management:

```go
// LoadGlobalConfig loads the global configuration from ~/.contexture
func (m *Manager) LoadGlobalConfig() (*domain.ConfigResult, error) {
    globalPath, err := domain.GetGlobalConfigPath()
    if err != nil {
        return nil, contextureerrors.Wrap(err, "get global config path")
    }

    // Check if global config exists
    exists, err := m.repo.Exists(globalPath)
    if err != nil {
        return nil, &ConfigError{
            Operation: "check existence",
            Path:      globalPath,
            Err:       err,
        }
    }

    if !exists {
        // Return nil without error - global config is optional
        return nil, nil
    }

    // Load and validate
    config, err := m.repo.Load(globalPath)
    if err != nil {
        return nil, &ConfigError{
            Operation: "load",
            Path:      globalPath,
            Err:       err,
        }
    }

    if err := m.validator.ValidateProject(config); err != nil {
        return nil, &ConfigError{
            Operation: "validate",
            Path:      globalPath,
            Err:       err,
        }
    }

    return &domain.ConfigResult{
        Config:   config,
        Location: domain.ConfigLocationGlobal,
        Path:     globalPath,
    }, nil
}

// SaveGlobalConfig saves the global configuration
func (m *Manager) SaveGlobalConfig(config *domain.Project) error {
    if config == nil {
        return contextureerrors.ValidationErrorf("config", "cannot be nil")
    }

    // Validate first
    if err := m.validator.ValidateProject(config); err != nil {
        return &ConfigError{
            Operation: "validate",
            Path:      "global",
            Err:       err,
        }
    }

    // Ensure global directory exists
    globalDir, err := domain.GetGlobalConfigDir()
    if err != nil {
        return contextureerrors.Wrap(err, "get global config dir")
    }

    if err := m.repo.GetFilesystem().MkdirAll(globalDir, configDirPermissions); err != nil {
        return contextureerrors.Wrap(err, "create global config directory")
    }

    // Get global config path
    globalPath, err := domain.GetGlobalConfigPath()
    if err != nil {
        return contextureerrors.Wrap(err, "get global config path")
    }

    // Save
    if err := m.repo.Save(config, globalPath); err != nil {
        return &ConfigError{
            Operation: "save",
            Path:      globalPath,
            Err:       err,
        }
    }

    return nil
}

// LoadConfigMerged loads both global and project configs and merges them
func (m *Manager) LoadConfigMerged(basePath string) (*domain.MergedConfig, error) {
    // Load global config (optional)
    globalResult, err := m.LoadGlobalConfig()
    if err != nil {
        return nil, contextureerrors.Wrap(err, "load global config")
    }

    // Load project config (required)
    projectResult, err := m.LoadConfig(basePath)
    if err != nil {
        return nil, contextureerrors.Wrap(err, "load project config")
    }

    // Merge configurations
    merged := m.MergeConfigs(globalResult, projectResult)

    return merged, nil
}

// MergeConfigs merges global and project configurations
func (m *Manager) MergeConfigs(global, project *domain.ConfigResult) *domain.MergedConfig {
    result := &domain.MergedConfig{
        Project:     project.Config,
        MergedRules: []domain.RuleWithSource{},
    }

    if global != nil {
        result.GlobalConfig = global.Config
    }

    // If no global config, just use project rules
    if global == nil || global.Config == nil {
        for _, rule := range project.Config.Rules {
            result.MergedRules = append(result.MergedRules, domain.RuleWithSource{
                RuleRef:         rule,
                Source:          domain.RuleSourceProject,
                OverridesGlobal: false,
            })
        }
        return result
    }

    // Build map of project rules for quick lookup
    projectRules := make(map[string]domain.RuleRef)
    for _, rule := range project.Config.Rules {
        normalizedID := m.normalizeRuleID(rule.ID)
        projectRules[normalizedID] = rule
    }

    // Add global rules first (checking for overrides)
    for _, globalRule := range global.Config.Rules {
        normalizedID := m.normalizeRuleID(globalRule.ID)
        if _, overridden := projectRules[normalizedID]; !overridden {
            // Not overridden - add global rule
            result.MergedRules = append(result.MergedRules, domain.RuleWithSource{
                RuleRef:         globalRule,
                Source:          domain.RuleSourceGlobal,
                OverridesGlobal: false,
            })
        }
    }

    // Add project rules (some may override global)
    for _, projectRule := range project.Config.Rules {
        normalizedID := m.normalizeRuleID(projectRule.ID)

        // Check if this overrides a global rule
        overridesGlobal := false
        for _, globalRule := range global.Config.Rules {
            if m.normalizeRuleID(globalRule.ID) == normalizedID {
                overridesGlobal = true
                break
            }
        }

        result.MergedRules = append(result.MergedRules, domain.RuleWithSource{
            RuleRef:         projectRule,
            Source:          domain.RuleSourceProject,
            OverridesGlobal: overridesGlobal,
        })
    }

    return result
}

// normalizeRuleID extracts and normalizes a rule ID for comparison
func (m *Manager) normalizeRuleID(ruleID string) string {
    // Use existing RuleMatcher logic to extract path
    path, err := m.matcher.ExtractPath(ruleID)
    if err != nil {
        // Fallback to the ID itself
        return strings.ToLower(ruleID)
    }
    return strings.ToLower(path)
}

// InitializeGlobalConfig creates a default global configuration
func (m *Manager) InitializeGlobalConfig() error {
    // Check if it already exists
    globalResult, err := m.LoadGlobalConfig()
    if err != nil {
        return err
    }

    if globalResult != nil {
        // Already exists, nothing to do
        return nil
    }

    // Create default global config
    defaultConfig := &domain.Project{
        Version: 1,
        Formats: []domain.FormatConfig{
            {Type: domain.FormatTypeClaude, Enabled: true},
            {Type: domain.FormatTypeCursor, Enabled: true},
            {Type: domain.FormatTypeWindsurf, Enabled: true},
        },
        Rules: []domain.RuleRef{},
    }

    return m.SaveGlobalConfig(defaultConfig)
}
```

#### Extend Existing Methods

Modify `LoadConfigWithLocalRules()` to optionally include global config:

```go
// Add a new method that includes global
func (m *Manager) LoadConfigWithLocalRulesAndGlobal(basePath string) (*domain.MergedConfig, error) {
    merged, err := m.LoadConfigMerged(basePath)
    if err != nil {
        return nil, err
    }

    // Discover local rules and add to project config
    // (existing logic from LoadConfigWithLocalRules)

    return merged, nil
}
```

### Rule Fetcher (`internal/rule/fetcher.go`)

#### Modifications

Update local rule discovery to check global rules directory:

```go
// LocalFetcher needs to check both project and global directories
func (f *LocalFetcher) FetchRule(ctx context.Context, ruleID string) (*domain.Rule, error) {
    // Try project-local first (.contexture/rules/, rules/)
    rule, err := f.fetchFromProject(ctx, ruleID)
    if err == nil {
        return rule, nil
    }

    // Try global (~/.contexture/rules/)
    rule, err = f.fetchFromGlobal(ctx, ruleID)
    if err == nil {
        return rule, nil
    }

    return nil, contextureerrors.NotFoundErrorf("rule", "rule not found: %s", ruleID)
}

func (f *LocalFetcher) fetchFromGlobal(ctx context.Context, ruleID string) (*domain.Rule, error) {
    globalDir, err := domain.GetGlobalConfigDir()
    if err != nil {
        return nil, err
    }

    rulesDir := filepath.Join(globalDir, "rules")

    // Use existing discovery logic but with global directory
    // ...
}
```

### Command Layer (`internal/commands/`)

#### Shared Flag Definition

Create shared flag constant:

```go
// In internal/commands/flags.go (new file)

var GlobalFlag = &cli.BoolFlag{
    Name:    "global",
    Aliases: []string{"g"},
    Usage:   "Operate on global configuration (~/.contexture)",
}
```

#### Add Command (`internal/commands/add.go`)

Modify `ExecuteWithDeps()`:

```go
func (c *AddCommand) ExecuteWithDeps(ctx context.Context, cmd *cli.Command, ruleIDs []string, deps *dependencies.Dependencies) error {
    // ... existing setup ...

    isGlobal := cmd.Bool("global")

    var config *domain.Project
    var configPath string

    if isGlobal {
        // Initialize global config if needed
        if err := c.projectManager.InitializeGlobalConfig(); err != nil {
            return contextureerrors.Wrap(err, "initialize global config")
        }

        // Load global config
        globalResult, err := c.projectManager.LoadGlobalConfig()
        if err != nil {
            return contextureerrors.Wrap(err, "load global config")
        }
        config = globalResult.Config
        configPath = globalResult.Path
    } else {
        // Existing project config loading
        currentDir, err := os.Getwd()
        if err != nil {
            return contextureerrors.Wrap(err, "get current directory")
        }

        configResult, err := c.projectManager.LoadConfigWithLocalRules(currentDir)
        if err != nil {
            return contextureerrors.Wrap(err, "load config")
        }
        config = configResult.Config
        configPath = configResult.Path
    }

    // ... rest of existing logic ...

    // Save to appropriate location
    if isGlobal {
        err = c.projectManager.SaveGlobalConfig(config)
    } else {
        location := c.projectManager.GetConfigLocation(currentDir, false)
        err = c.projectManager.SaveConfig(config, location, currentDir)
    }

    // ... rest of existing logic ...
}
```

#### List Command (`internal/commands/list.go`)

Modify `Execute()`:

```go
func (c *ListCommand) Execute(ctx context.Context, cmd *cli.Command) error {
    currentDir, err := os.Getwd()
    if err != nil {
        return contextureerrors.Wrap(err, "get current directory")
    }

    // Load merged config (global + project)
    merged, err := c.projectManager.LoadConfigMerged(currentDir)
    if err != nil {
        // If project config fails, try just global
        if errors.Is(err, ErrConfigNotFound) {
            globalResult, err := c.projectManager.LoadGlobalConfig()
            if err != nil || globalResult == nil {
                return contextureerrors.Wrap(err, "load configuration")
            }
            // Show just global rules
            return c.showRulesFromGlobalOnly(ctx, cmd, globalResult.Config)
        }
        return err
    }

    // Load providers
    if merged.GlobalConfig != nil {
        if err := c.providerRegistry.LoadFromProject(merged.GlobalConfig); err != nil {
            return contextureerrors.Wrap(err, "load global providers")
        }
    }
    if err := c.providerRegistry.LoadFromProject(merged.Project); err != nil {
        return contextureerrors.Wrap(err, "load project providers")
    }

    // Fetch rules with source tracking
    return c.showMergedRuleList(ctx, cmd, merged)
}

func (c *ListCommand) showMergedRuleList(ctx context.Context, cmd *cli.Command, merged *domain.MergedConfig) error {
    // Fetch actual rule content for each RuleWithSource
    rulesWithSource := make([]RuleWithSourceContent, 0, len(merged.MergedRules))

    for _, rws := range merged.MergedRules {
        rule, err := c.ruleFetcher.FetchRule(ctx, rws.RuleRef.ID)
        if err != nil {
            fmt.Printf("Warning: Failed to fetch rule %s: %v\n", rws.RuleRef.ID, err)
            continue
        }

        rulesWithSource = append(rulesWithSource, RuleWithSourceContent{
            Rule:            rule,
            Source:          rws.Source,
            OverridesGlobal: rws.OverridesGlobal,
        })
    }

    // Display with source indicators
    return c.displayRulesWithSource(rulesWithSource, cmd)
}
```

#### Remove, Update, New Commands

Similar modifications to Add command - detect `-g` flag and route to global config operations.

#### Build Command (`internal/commands/build.go`)

Modify to use merged config:

```go
func (c *BuildCommand) Execute(ctx context.Context, cmd *cli.Command) error {
    currentDir, err := os.Getwd()
    if err != nil {
        return contextureerrors.Wrap(err, "get current directory")
    }

    // Load merged configuration (global + project)
    merged, err := c.projectManager.LoadConfigMerged(currentDir)
    if err != nil {
        return contextureerrors.Wrap(err, "load config")
    }

    // Load providers from both configs
    if merged.GlobalConfig != nil {
        if err := c.providerRegistry.LoadFromProject(merged.GlobalConfig); err != nil {
            return contextureerrors.Wrap(err, "load global providers")
        }
    }
    if err := c.providerRegistry.LoadFromProject(merged.Project); err != nil {
        return contextureerrors.Wrap(err, "load project providers")
    }

    // Create a synthetic Project with merged rules for generation
    mergedProject := *merged.Project
    mergedProject.Rules = make([]domain.RuleRef, len(merged.MergedRules))
    for i, rws := range merged.MergedRules {
        mergedProject.Rules[i] = rws.RuleRef
    }

    // Use existing build logic with merged project
    // ... rest of build command ...
}
```

### CLI Flag Registration (`cmd/contexture/main.go` or relevant CLI setup)

Add `-g` flag to appropriate commands:

```go
// In command definitions
{
    Name:  "add",
    Flags: []cli.Flag{
        GlobalFlag,  // Add to existing flags
        // ... other flags ...
    },
    // ...
}

// Similarly for: remove, update, new, providers add/remove, config
```

### Output Formatting (`internal/output/`)

#### Extend Metadata Types

```go
// In internal/output/metadata.go

type ListMetadata struct {
    Pattern         string `json:"pattern,omitempty"`
    TotalRules      int    `json:"totalRules"`
    FilteredRules   int    `json:"filteredRules"`
    GlobalRules     int    `json:"globalRules"`      // NEW
    ProjectRules    int    `json:"projectRules"`     // NEW
    OverriddenRules int    `json:"overriddenRules"`  // NEW
}

// Add source tracking to rule output
type RuleOutput struct {
    ID              string         `json:"id"`
    Title           string         `json:"title"`
    Description     string         `json:"description"`
    Source          string         `json:"source"`          // NEW: "global" or "project"
    OverridesGlobal bool           `json:"overridesGlobal"` // NEW
    // ... other fields ...
}
```

#### Update Writers

Modify default and JSON writers to show source information:

```go
// In internal/output/default_writer.go

func (w *DefaultWriter) WriteRulesList(rules []*domain.Rule, metadata ListMetadata) error {
    // ... existing code ...

    for _, rule := range rulesWithSource {
        fmt.Printf("  %s", rule.Title)

        // Show source indicator
        if rule.Source == "global" {
            fmt.Printf(" %s", grayStyle.Render("[global]"))
        } else if rule.OverridesGlobal {
            fmt.Printf(" %s", grayStyle.Render("[project overrides global]"))
        } else {
            fmt.Printf(" %s", grayStyle.Render("[project]"))
        }

        fmt.Println()
    }

    // ... rest of display ...
}
```

## Testing Strategy

### Unit Tests

#### Config Merging Tests (`internal/project/config_test.go`)

```go
func TestMergeConfigs(t *testing.T) {
    tests := []struct {
        name           string
        globalConfig   *domain.Project
        projectConfig  *domain.Project
        expectedRules  []domain.RuleWithSource
    }{
        {
            name: "no overlap - all rules included",
            globalConfig: &domain.Project{
                Rules: []domain.RuleRef{
                    {ID: "@contexture/global-rule"},
                },
            },
            projectConfig: &domain.Project{
                Rules: []domain.RuleRef{
                    {ID: "@contexture/project-rule"},
                },
            },
            expectedRules: []domain.RuleWithSource{
                {
                    RuleRef: domain.RuleRef{ID: "@contexture/global-rule"},
                    Source: domain.RuleSourceGlobal,
                    OverridesGlobal: false,
                },
                {
                    RuleRef: domain.RuleRef{ID: "@contexture/project-rule"},
                    Source: domain.RuleSourceProject,
                    OverridesGlobal: false,
                },
            },
        },
        {
            name: "project overrides global",
            globalConfig: &domain.Project{
                Rules: []domain.RuleRef{
                    {ID: "@contexture/rule", Variables: map[string]any{"key": "global"}},
                },
            },
            projectConfig: &domain.Project{
                Rules: []domain.RuleRef{
                    {ID: "@contexture/rule", Variables: map[string]any{"key": "project"}},
                },
            },
            expectedRules: []domain.RuleWithSource{
                {
                    RuleRef: domain.RuleRef{
                        ID: "@contexture/rule",
                        Variables: map[string]any{"key": "project"},
                    },
                    Source: domain.RuleSourceProject,
                    OverridesGlobal: true,
                },
            },
        },
        {
            name: "nil global config",
            globalConfig: nil,
            projectConfig: &domain.Project{
                Rules: []domain.RuleRef{
                    {ID: "@contexture/project-rule"},
                },
            },
            expectedRules: []domain.RuleWithSource{
                {
                    RuleRef: domain.RuleRef{ID: "@contexture/project-rule"},
                    Source: domain.RuleSourceProject,
                    OverridesGlobal: false,
                },
            },
        },
        // Add tests for:
        // - Different rule ID formats matching
        // - Case-insensitive matching
        // - Provider merging
        // - Format config precedence
        // - Generation settings fallback
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

#### Rule ID Normalization Tests

```go
func TestNormalizeRuleID(t *testing.T) {
    tests := []struct {
        name     string
        ruleID1  string
        ruleID2  string
        shouldMatch bool
    }{
        {
            name:     "provider syntax matches full format",
            ruleID1:  "@contexture/languages/go/context",
            ruleID2:  "[contexture:languages/go/context]",
            shouldMatch: true,
        },
        {
            name:     "case insensitive",
            ruleID1:  "@contexture/Security/Auth",
            ruleID2:  "[contexture:security/auth]",
            shouldMatch: true,
        },
        // More test cases
    }
}
```

### Integration Tests

#### Global Config I/O Tests (`integration/global_config_test.go`)

```go
func TestGlobalConfigCreateAndLoad(t *testing.T) {
    // Setup: Create temp home directory
    tmpHome := t.TempDir()
    t.Setenv("HOME", tmpHome)

    // Create global config
    manager := project.NewManager(afero.NewOsFs())

    config := &domain.Project{
        Version: 1,
        Rules: []domain.RuleRef{
            {ID: "@contexture/test-rule"},
        },
    }

    err := manager.SaveGlobalConfig(config)
    require.NoError(t, err)

    // Verify file exists
    globalPath := filepath.Join(tmpHome, ".contexture", ".contexture.yaml")
    require.FileExists(t, globalPath)

    // Load it back
    loaded, err := manager.LoadGlobalConfig()
    require.NoError(t, err)
    require.NotNil(t, loaded)
    require.Equal(t, 1, len(loaded.Config.Rules))
    require.Equal(t, "@contexture/test-rule", loaded.Config.Rules[0].ID)
}
```

### E2E Tests

#### Global Rule Workflow Tests (`e2e/global_rules_test.go`)

```go
func TestGlobalRuleWorkflow(t *testing.T) {
    // Setup: Create test environment with temp home and project
    tmpHome := t.TempDir()
    tmpProject := t.TempDir()
    t.Setenv("HOME", tmpHome)

    // Initialize project
    runCommand(t, tmpProject, "contexture", "init")

    // Add global rule
    out := runCommand(t, tmpProject, "contexture", "rules", "add", "-g", "@contexture/testing/echo")
    require.Contains(t, out, "added successfully")

    // Verify global config exists
    globalConfig := filepath.Join(tmpHome, ".contexture", ".contexture.yaml")
    require.FileExists(t, globalConfig)

    // List rules - should show global rule
    out = runCommand(t, tmpProject, "contexture", "rules", "list")
    require.Contains(t, out, "@contexture/testing/echo")
    require.Contains(t, out, "[global]")

    // Add project rule with same ID
    out = runCommand(t, tmpProject, "contexture", "rules", "add", "@contexture/testing/echo", "--var", "message=project")

    // List again - should show override
    out = runCommand(t, tmpProject, "contexture", "rules", "list")
    require.Contains(t, out, "[project overrides global]")

    // Build should use project version
    out = runCommand(t, tmpProject, "contexture", "build")
    require.NoError(t, err)

    // Verify generated file uses project variables
    claudeFile := filepath.Join(tmpProject, "CLAUDE.md")
    content, err := os.ReadFile(claudeFile)
    require.NoError(t, err)
    require.Contains(t, string(content), "project") // Project variable used

    // Remove project rule
    out = runCommand(t, tmpProject, "contexture", "rules", "remove", "@contexture/testing/echo")

    // List should show global rule again (not overridden)
    out = runCommand(t, tmpProject, "contexture", "rules", "list")
    require.Contains(t, out, "[global]")
    require.NotContains(t, out, "overrides")
}
```

### Test Coverage Goals

- Config merging logic: 100%
- Global config I/O: 100%
- Command flag handling: 100%
- Rule ID normalization: 100%
- E2E workflows: Major user paths covered

## Migration & Compatibility

### Backwards Compatibility

**No breaking changes:**
- Existing projects work without modification
- All commands work without `-g` flag (existing behavior)
- Global config is completely optional
- No changes to file formats or schemas

**Graceful degradation:**
- If global config doesn't exist, commands work normally
- If global config is malformed, show clear error and continue with project config
- List/build commands handle missing global config transparently

### User Migration Path

**For users who want global rules:**

1. Add first global rule: `contexture rules add -g @contexture/my-rule`
   - System automatically creates `~/.contexture/` directory and config

2. Migrate existing project rules to global (manual):
   - List project rules: `contexture rules list`
   - For each rule to migrate:
     - Add to global: `contexture rules add -g <rule-id>`
     - Remove from project: `contexture rules remove <rule-id>`

3. Verify with build: `contexture build`

**No automatic migration tool needed:**
- Users can gradually adopt global rules
- Can keep some rules project-specific, others global
- No pressure to migrate existing projects

### Documentation Updates Needed

1. **User Guide** - New section on global rules:
   - What are global rules
   - When to use global vs project rules
   - How to add/remove/manage global rules
   - How precedence works

2. **Command Reference** - Update each command:
   - Document `-g` flag
   - Show examples with global flag
   - Explain merged behavior for build/list

3. **Configuration Reference** - Document:
   - Global config file location
   - Same schema as project config
   - Merge behavior

4. **Migration Guide** - Best practices:
   - Which rules are good candidates for global
   - How to handle team vs personal rules
   - Troubleshooting common issues

## Implementation Phases

### Phase 1: Core Infrastructure (MVP)

1. Domain models for source tracking
2. Global config loading/saving in ProjectManager
3. Config merging logic with tests
4. Unit tests for all new functions

**Deliverable:** Library functions work, no CLI changes yet

### Phase 2: Command Integration

1. Add `-g` flag to all relevant commands
2. Route commands to global vs project config
3. Update command implementations (add, remove, new, update)
4. Integration tests for commands

**Deliverable:** All commands support `-g` flag

### Phase 3: Merged Operations

1. Update list command to show both sources
2. Update build command to use merged config
3. Update providers list to show both sources
4. Output formatting with source indicators

**Deliverable:** List and build work with merged configs

### Phase 4: Polish & Documentation

1. E2E tests for complete workflows
2. Error messages and user feedback
3. Documentation updates
4. Examples and tutorials

**Deliverable:** Production-ready feature

## Success Criteria

### Functional Requirements

✓ Users can add rules globally with `-g` flag
✓ Global rules apply to all projects automatically
✓ Project rules override global rules by ID
✓ List command shows source of each rule
✓ Build command merges global and project rules correctly
✓ No breaking changes to existing projects

### Non-Functional Requirements

✓ Config loading adds < 50ms overhead (for typical global config)
✓ Merge logic handles 1000+ rules efficiently
✓ Clear error messages for global config issues
✓ 90%+ test coverage for new code
✓ Documentation covers all use cases

### User Experience Goals

✓ Intuitive `-g` flag usage
✓ Clear visual indicators of rule source
✓ Helpful error messages
✓ No confusion about precedence rules
✓ Smooth migration path from project-only to global+project

## Open Questions & Decisions

### Resolved

1. **Directory location:** `~/.contexture/` (mirrors project structure)
2. **Flag name:** `-g/--global` (standard Unix convention)
3. **Merge strategy:** Project completely overrides global (no partial merge)
4. **Initialization:** Lazy (create on first use, no explicit init command)
5. **Build behavior:** Always merge (no flag needed)

### Future Considerations

1. **Global templates:** Should global config support custom Claude templates?
   - **Decision:** Yes, follow same logic as rules (project overrides global)

2. **Environment-specific globals:** Should we support `~/.contexture/environments/`?
   - **Decision:** Out of scope for MVP, can be added later

3. **Sync mechanism:** Should global config be shareable via git?
   - **Decision:** Out of scope - users can manually version `~/.contexture/` if desired

4. **Rule namespacing:** Should global rules have special namespace?
   - **Decision:** No - use same ID scheme as project rules

5. **Conflict resolution UI:** Should we prompt user when adding rule that exists globally?
   - **Decision:** No prompt - project always overrides (can show info message)

## Summary

This plan provides a comprehensive, backwards-compatible approach to adding global user-level rules support to Contexture. The implementation:

- **Minimal:** Reuses existing infrastructure (Project struct, validation, fetching)
- **Intuitive:** Standard `-g` flag pattern, clear source indicators
- **Safe:** No breaking changes, graceful degradation, extensive tests
- **Complete:** Covers all commands, edge cases, and user workflows

The phased approach allows for incremental development and testing, with each phase delivering working functionality.
