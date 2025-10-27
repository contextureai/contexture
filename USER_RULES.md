# User Rules Implementation Design

## Problem Statement

The current global rules implementation injects user-level rules into project files (CLAUDE.md, .cursor/rules/, .windsurf/rules/), which creates conflicts in multi-developer environments where different developers have different global rules.

### Issues with Current Approach:
1. Developer A's global rules get injected into project files
2. Developer B doesn't have those rules, but sees them in the project
3. Unnecessary git diffs when developers have different global rules
4. Rules meant for personal use affect team members

## Solution: IDE-Native User Rules

### Key Insight
Some IDEs support user-level rules natively, so we should use those mechanisms instead of injecting into project files.

### IDE Support Matrix

| IDE | User Rules Support | File Location | Format |
|-----|-------------------|---------------|--------|
| **Windsurf** | ✅ Native | `~/.windsurf/global_rules.md` | Markdown + YAML frontmatter (12,000 chars/file) |
| **Claude Code** | ✅ Native | `~/.claude/CLAUDE.md` | Markdown + XML tags |
| **Cursor** | ❌ None | N/A - Project only | N/A |

## Design: Dual-Output Mode for User Rules

### Concept

**For IDEs with native user rules support:**
- Generate user rules in IDE's native user rules location
- Do NOT inject into project files
- Project files only contain project-level rules

**For IDEs without native user rules support:**
- Provide opt-in mechanism via project config
- Default: DO NOT inject user rules (safer for teams)
- If enabled: Inject user rules into project files (backward compatible)

### Implementation Strategy

#### 1. New Domain Concepts

```go
// UserRulesOutputMode defines how user rules are handled for a format
type UserRulesOutputMode string

const (
    // UserRulesNative - Output to IDE's native user rules location
    UserRulesNative UserRulesOutputMode = "native"

    // UserRulesProject - Inject into project files (opt-in)
    UserRulesProject UserRulesOutputMode = "project"

    // UserRulesDisabled - Don't output user rules at all
    UserRulesDisabled UserRulesOutputMode = "disabled"
)

// FormatCapabilities describes what a format supports
type FormatCapabilities struct {
    SupportsUserRules   bool   // IDE supports native user rules
    UserRulesPath       string // Where IDE expects user rules (e.g. "~/.windsurf/global_rules.md")
    DefaultOutputMode   UserRulesOutputMode
}
```

#### 2. Updated FormatConfig

```yaml
# .contexture.yaml
version: 1

formats:
  - type: windsurf
    enabled: true
    userRulesMode: native  # Output to ~/.windsurf/global_rules.md

  - type: claude
    enabled: true
    userRulesMode: native  # Output to ~/.claude/CLAUDE.md

  - type: cursor
    enabled: true
    # Default: "project" (inject user rules)
    # Set to "disabled" to exclude user rules (recommended for teams)
    # userRulesMode: disabled

rules:
  - id: "@contexture/languages/go/context"  # Project rule
```

#### 3. Build Behavior

```
contexture build
```

**For Windsurf:**
1. Generate `~/.windsurf/global_rules.md` with global rules
2. Generate `.windsurf/rules/` with ONLY project rules
3. No mixing of user/project rules in project files

**For Claude:**
1. Generate `~/.claude/CLAUDE.md` with global rules
2. Generate `CLAUDE.md` with ONLY project rules
3. No mixing of user/project rules in project files

**For Cursor:**
1. If `userRulesMode: project` (default):
   - Generate `.cursor/rules/` with both project AND global rules
   - Clearly mark which are global (e.g. comments/metadata)
2. If `userRulesMode: disabled` (opt-out):
   - Generate `.cursor/rules/` with ONLY project rules
   - Global rules are not used for Cursor

#### 4. List Command Behavior

```
contexture rules list
```

**Output should indicate where rules will be applied:**

```
Rules (5)

Project Rules:
  @contexture/languages/go/testing
    → CLAUDE.md, .windsurf/rules/, .cursor/rules/

  @mycompany/security/auth
    → CLAUDE.md, .windsurf/rules/, .cursor/rules/

Global Rules (User-level):
  @contexture/languages/go/context [global]
    → ~/.claude/CLAUDE.md (Claude Code)
    → ~/.windsurf/global_rules.md (Windsurf)
    → .cursor/rules/ (Cursor - userRulesMode: project)

  @contexture/testing/unit [global]
    → ~/.claude/CLAUDE.md (Claude Code)
    → ~/.windsurf/global_rules.md (Windsurf)
    → .cursor/rules/ (Cursor - userRulesMode: project)
```

#### 5. Configuration Migration

**Old behavior (pre-fix):**
- Global rules injected into all project files
- No distinction between user/project rules in output

**New behavior (post-fix):**
- Global rules → IDE user rules locations (Windsurf, Claude)
- Global rules → included by default for Cursor (can opt-out with disabled)
- Clear separation in outputs

**Migration path:**
- Users upgrading will see global rules move to native locations for Windsurf/Claude
- Cursor behavior unchanged (still includes user rules by default)
- Teams can opt-out Cursor user rules with `userRulesMode: disabled`
- No breaking changes to config file format

## Rule Ordering Fix

### Problem
Rule order in generated files is inconsistent, causing unnecessary git diffs.

### Solution
**Deterministic Ordering:**
1. Sort rules by normalized ID (case-insensitive, alphabetical)
2. Within same provider, maintain insertion order from config
3. Consistent ordering across all formats

**Implementation:**
```go
func sortRulesForOutput(rules []*domain.Rule) []*domain.Rule {
    sorted := make([]*domain.Rule, len(rules))
    copy(sorted, rules)

    sort.SliceStable(sorted, func(i, j int) bool {
        // Normalize IDs for comparison
        idI := normalizeRuleIDForSort(sorted[i].ID)
        idJ := normalizeRuleIDForSort(sorted[j].ID)
        return idI < idJ
    })

    return sorted
}

func normalizeRuleIDForSort(id string) string {
    // Extract path from [contexture:path] or @provider/path
    // Convert to lowercase for case-insensitive sorting
    path := extractPath(id)
    return strings.ToLower(path)
}
```

## Configuration Schema

### Configuration Principles

**Omit Default Values:**
- When `userRulesMode` is set to the default value for a format, omit it from the config
- This keeps configs clean and makes non-default choices explicit
- Defaults:
  - Windsurf: `native`
  - Claude: `native`
  - Cursor: `project`

**Example - Minimal Config (using all defaults):**
```yaml
version: 1
formats:
  - type: windsurf
    enabled: true
  - type: claude
    enabled: true
  - type: cursor
    enabled: true
```

**Example - Explicit Non-Default:**
```yaml
version: 1
formats:
  - type: cursor
    enabled: true
    userRulesMode: disabled  # Explicitly opt-out (for teams)
```

### Project Config (.contexture.yaml)

```yaml
version: 1

formats:
  - type: windsurf
    enabled: true
    # userRulesMode defaults to "native" (output to ~/.windsurf/global_rules.md)
    # Omitted when using default value

  - type: claude
    enabled: true
    # userRulesMode defaults to "native" (output to ~/.claude/CLAUDE.md)
    # Omitted when using default value

  - type: cursor
    enabled: true
    # userRulesMode defaults to "project" (inject user rules into project files)
    # Omitted when using default value
    # To exclude user rules for teams, explicitly set:
    # userRulesMode: disabled

rules:
  # Project-level rules (apply to all developers)
  - id: "@contexture/languages/go/testing"
  - id: "@mycompany/security/auth"
```

### Global Config (~/.contexture/.contexture.yaml)

```yaml
version: 1

rules:
  # User-level rules (personal preferences)
  - id: "@contexture/languages/go/context"
  - id: "@contexture/testing/unit"
  - id: "[contexture(local):my-personal-rule]"

# Global config can also specify default format preferences
formats:
  - type: windsurf
    enabled: true
  - type: claude
    enabled: true
  - type: cursor
    enabled: true
```

## File Locations After Build

### Project Directory
```
project/
├── CLAUDE.md                      # Only project rules
├── .windsurf/
│   └── rules/                     # Only project rules
│       ├── go-testing.md
│       └── security-auth.md
└── .cursor/
    └── rules/                     # Only project rules (if userRulesMode: disabled)
        ├── go-testing.mdc
        └── security-auth.mdc
```

### User Home Directory
```
~/.windsurf/
└── global_rules.md                # User's global rules

~/.claude/
└── CLAUDE.md                      # User's global rules

~/.contexture/
└── .contexture.yaml               # User's global config
```

## Benefits

### ✅ Team Collaboration
- Project files only contain team-shared rules
- No git conflicts from different developers' personal rules
- Clear separation between team and personal preferences

### ✅ IDE Integration
- Leverages native IDE features
- Rules work as IDE intended
- Consistent with IDE user expectations

### ✅ Backward Compatibility
- Cursor behavior unchanged (default includes user rules)
- No breaking changes to config format
- Migration path is clear
- Teams can opt-out of Cursor user rules if needed

### ✅ Consistent Output
- Deterministic rule ordering eliminates unnecessary diffs
- Predictable file generation
- Easier code review

## Implementation Plan

### Phase 1: Core Infrastructure
1. Add `UserRulesOutputMode` to domain
2. Add `FormatCapabilities` interface
3. Update format registry with capabilities
4. Add `userRulesMode` to FormatConfig

### Phase 2: Format Implementations
1. Update Windsurf format to support native user rules
2. Update Claude format to support native user rules
3. Update Cursor format to support disabled/project modes
4. Implement rule sorting for consistent output

### Phase 3: Build Command
1. Separate project rules from global rules
2. Output project rules to project locations
3. Output global rules to user locations (native mode)
4. Respect userRulesMode configuration

### Phase 4: List Command
1. Show where each rule will be applied
2. Indicate user rules vs project rules
3. Show which IDEs will receive which rules

### Phase 5: Testing & Documentation
1. Unit tests for new behavior
2. Integration tests for file generation
3. E2E tests for complete workflows
4. Update documentation
5. Migration guide

## Success Criteria

- ✅ Project files contain only project rules
- ✅ User rules output to native IDE locations
- ✅ Cursor respects userRulesMode setting
- ✅ Rule ordering is consistent and deterministic
- ✅ No breaking changes for existing users
- ✅ Clear documentation and migration path
- ✅ All tests passing
