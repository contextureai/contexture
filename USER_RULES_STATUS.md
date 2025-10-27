# User Rules Implementation Status

## Overview
Implementation of IDE-native user rules to prevent multi-developer conflicts and ensure consistent rule ordering.

## Completed ✅

### Phase 1: Core Infrastructure
- ✅ Added `UserRulesOutputMode` enum (native/project/disabled)
- ✅ Added `UserRulesMode` field to `FormatConfig`
- ✅ Added `FormatCapabilities` struct with user rules metadata
- ✅ Added `GetEffectiveUserRulesMode()` helper to FormatConfig
- ✅ Added `GetCapabilities()` to Handler interface
- ✅ Implemented GetCapabilities() in all format handlers:
  - Windsurf: `SupportsUserRules=true`, path=`~/.windsurf/global_rules.md`, default=`native`, max=12000 chars
  - Claude: `SupportsUserRules=true`, path=`~/.claude/CLAUDE.md`, default=`native`, max=unlimited
  - Cursor: `SupportsUserRules=false`, default=`project`
- ✅ Added `GetCapabilities(formatType)` helper to Registry

### Phase 2: Utilities
- ✅ Implemented `SortRulesDeterministically()` for consistent rule ordering
- ✅ Uses rule ID parser to normalize paths for case-insensitive sorting
- ✅ Stable sort preserves order for rules with same normalized ID

## Remaining Work 🔨

### Phase 3: Build Command Updates
**Status:** Not started
**Location:** `internal/commands/build.go`
**Tasks:**
1. Separate user rules from project rules in merged config
2. For each enabled format:
   - Get format capabilities
   - Determine effective user rules mode
   - If `native`: generate user rules to IDE's native location
   - If `project`: include user rules in project files
   - If `disabled`: exclude user rules
3. Sort rules before passing to format generators
4. Ensure user rules directories are created (e.g., `~/.windsurf/`, `~/.claude/`)

**Key Logic:**
```go
// Pseudocode for build command
for each format in config.formats {
    caps := registry.GetCapabilities(format.Type)
    mode := format.GetEffectiveUserRulesMode()

    // Separate rules
    projectRules := filter(mergedRules, source == "project")
    userRules := filter(mergedRules, source == "global")

    // Sort for consistency
    projectRules = SortRulesDeterministically(projectRules)
    userRules = SortRulesDeterministically(userRules)

    if mode == "native" && caps.SupportsUserRules {
        // Generate user rules to native location
        generateToPath(userRules, caps.UserRulesPath)
        // Generate only project rules to project
        generateToProject(projectRules)
    } else if mode == "project" {
        // Generate both to project
        allRules = append(projectRules, userRules)
        generateToProject(allRules)
    } else if mode == "disabled" {
        // Generate only project rules
        generateToProject(projectRules)
    }
}
```

### Phase 4: Format Implementations
**Status:** Not started
**Locations:** `internal/format/{windsurf,claude,cursor}/format.go`

**Windsurf:**
- Update `Format.Generate()` to accept separate user/project rule lists
- Add `GenerateUserRules(rules, outputPath)` method
- Ensure max 12,000 chars per file limit

**Claude:**
- Update to generate `~/.claude/CLAUDE.md` for user rules
- Keep project `CLAUDE.md` separate

**Cursor:**
- Add metadata/comments to distinguish user vs project rules when both included
- Handle disabled mode (project rules only)

### Phase 5: List Command Updates
**Status:** Not started
**Location:** `internal/commands/list.go`
**Tasks:**
1. Show where each rule will be applied (which IDE, which location)
2. Indicate format capabilities in output
3. Example output:
```
Project Rules:
  @contexture/languages/go/testing
    → CLAUDE.md, .windsurf/rules/, .cursor/rules/

Global Rules (User-level):
  @contexture/languages/go/context [global]
    → ~/.claude/CLAUDE.md (Claude)
    → ~/.windsurf/global_rules.md (Windsurf)
    → .cursor/rules/ (Cursor - userRulesMode: project)
```

### Phase 6: Config Serialization
**Status:** Not started
**Tasks:**
1. Ensure `userRulesMode` is omitted from YAML when set to default
2. Add custom YAML marshaller if needed
3. Test round-trip config loading/saving

### Phase 7: Testing
**Status:** Not started
**Tests Needed:**
1. **Unit Tests:**
   - FormatConfig.GetEffectiveUserRulesMode()
   - SortRulesDeterministically()
   - Format capabilities

2. **Integration Tests:**
   - Build command with user rules native mode
   - Build command with user rules project mode
   - Build command with user rules disabled
   - File generation to correct locations

3. **E2E Tests:**
   - Complete workflow with Windsurf native user rules
   - Complete workflow with Claude native user rules
   - Complete workflow with Cursor project mode
   - Mixed configuration (different modes per format)
   - Rule ordering consistency

### Phase 8: Documentation
**Status:** Not started
**Files to Update:**
1. `docs/reference/commands/build.md` - document user rules behavior
2. `docs/reference/configuration/config-file.md` - document userRulesMode
3. `docs/core-concepts/rules.md` - explain user vs project rules
4. Add migration guide for existing users
5. Update GLOBAL.md to reference user rules implementation

## Design Decisions

### Defaults
- **Windsurf:** `native` (use `~/.windsurf/global_rules.md`)
- **Claude:** `native` (use `~/.claude/CLAUDE.md`)
- **Cursor:** `project` (inject into `.cursor/rules/` - backward compatible)

### Config Philosophy
- Omit `userRulesMode` from YAML when using default value
- Makes configs clean and non-default choices explicit

### Backward Compatibility
- ✅ No breaking changes
- ✅ Existing Cursor behavior unchanged (still includes user rules)
- ✅ Teams can opt-out Cursor user rules with `userRulesMode: disabled`
- ✅ Windsurf/Claude automatically use native locations (better for teams)

## Next Steps

1. **Implement Build Command Logic** (Phase 3)
   - This is the core functionality
   - Most complex part of the implementation
   - Requires careful handling of file paths, directories, and format-specific logic

2. **Update Format Implementations** (Phase 4)
   - Modify each format to support separate user/project rule generation
   - Handle edge cases (file size limits, directory creation, etc.)

3. **Update List Command** (Phase 5)
   - Show comprehensive rule destination information
   - Help users understand where their rules will be applied

4. **Write Comprehensive Tests** (Phase 7)
   - Ensure all modes work correctly
   - Verify file generation
   - Test edge cases

5. **Update Documentation** (Phase 8)
   - Clear explanation of user rules concept
   - Migration guide
   - Examples for each mode

## Estimated Remaining Effort

- Phase 3 (Build Command): **4-6 hours** (complex logic, error handling, file I/O)
- Phase 4 (Format Implementations): **3-4 hours** (3 formats to update)
- Phase 5 (List Command): **1-2 hours** (display logic)
- Phase 6 (Config Serialization): **1 hour** (if custom marshalling needed)
- Phase 7 (Testing): **4-5 hours** (comprehensive test coverage)
- Phase 8 (Documentation): **2-3 hours** (multiple files, examples)

**Total:** ~15-21 hours of focused development time

## Files Modified So Far

1. `internal/domain/format.go` - Added UserRulesOutputMode, FormatCapabilities, FormatConfig.UserRulesMode
2. `internal/format/registry.go` - Added GetCapabilities() to Handler interface
3. `internal/format/windsurf/ui.go` - Implemented GetCapabilities()
4. `internal/format/claude/ui.go` - Implemented GetCapabilities()
5. `internal/format/cursor/ui.go` - Implemented GetCapabilities()
6. `internal/rule/utils.go` - Added SortRulesDeterministically()
7. `USER_RULES.md` - Comprehensive design document
8. `USER_RULES_STATUS.md` - This file

## Testing Strategy

Before proceeding with full implementation:
1. ✅ Ensure current code compiles
2. ✅ Run existing tests to verify no regressions
3. ✅ Run linter to ensure code quality

After completing each phase:
1. Write unit tests for new functionality
2. Run full test suite
3. Manual testing with real projects

## Success Criteria

- ✅ All code compiles without errors
- ✅ All existing tests pass
- ✅ Linter shows 0 issues
- ⬜ New unit tests pass
- ⬜ New integration tests pass
- ⬜ New E2E tests pass
- ⬜ Manual testing successful
- ⬜ Documentation complete and accurate
