# Contexture

Contexture is a Go CLI tool for managing AI assistant rules across multiple platforms (Claude, Cursor, Windsurf). It fetches rules from local/remote sources, processes templates with variables, and generates platform-specific output files.

## Commands

### Essential Development Commands
- `make build` - Build binary to `./bin/contexture`
- `make test` - Run unit tests with coverage
- `make lint` - Run golangci-lint
- `make fmt` - Format code with goimports and gofumpt
- `make deps` - Download and tidy dependencies
- `make generate` - Generate mocks with mockery

### Working Locally
- Run `go test ./...`, then `make fmt`, and finally `make lint` before returning changes.
- Prefer the in-memory filesystem helpers (`internal/commands/test_helpers.go:12`, `internal/dependencies/deps.go:31`) for new tests to avoid touching the host filesystem.

### Application Commands
- `contexture init` - Initialize project with `.contexture.yaml`
- `contexture rules add` - Add rules by ID
- `contexture rules list/remove/update` - Manage rules
- `contexture config` - View/modify configuration
- `contexture build` - Generate platform-specific files

## Project Structure

```
cmd/contexture/        # CLI entry point
internal/
├── app/               # Main application and command orchestration
├── cli/               # CLI help templates and formatting
├── commands/          # Command implementations (init, build, rules, config)
├── domain/            # Core business models and interfaces
├── format/            # Output format handlers (claude, cursor, windsurf)
├── rule/              # Rule processing, fetching, validation
├── template/          # Go template engine for variable substitution
├── git/               # Git repository operations
├── tui/               # Terminal UI components (Bubble Tea)
├── ui/                # UI components and styling
├── project/           # Project configuration management
├── cache/             # Simple caching implementation
├── errors/            # Error handling and display
└── version/           # Version information

e2e/                   # End-to-end tests
integration/           # Git integration tests
testproject/           # Test fixtures and examples
docs/                  # Documentation
web/                   # Web assets
```

## Core Architecture

**Rule Flow**: Sources → Fetcher → Template Engine → Format Transformer → Output Files

### Key Modules

- **app**: CLI app setup, dependency injection, error handling
- **commands**: Command implementations using dependency injection
- **domain**: Core types (`Project`, `Rule`, `Format`, `Config`)
- **rule**: Rule fetching (git/local), parsing, validation, template processing
- **format**: Platform-specific transformations (claude.md, .cursor/rules/, .windsurf/rules/)
- **git**: Git operations for remote rule repositories
- **tui**: Interactive rule selection and file browsing

### Extension Hooks

- Register new CLI commands in `internal/app/app.go:95` so they are wired into the binary and help output.
- Register new output formats through `internal/format/registry.go:48` to expose them in prompts and builds.

## Domain Concepts

### Rules
Markdown files with YAML frontmatter containing AI instructions. Support template variables via Go's `text/template`.

### Formats
- `claude`: Single `CLAUDE.md` file (supports custom templates via `CLAUDE.template.md`)
- `cursor`: Individual files in `.cursor/rules/`
- `windsurf`: Individual files in `.windsurf/rules/`

### Sources
- Default community repository
- Custom Git repositories
- Local project files

### Project Configuration (`.contexture.yaml`)
```yaml
version: 1
formats: 
  - type: claude
    enabled: true
    template: CLAUDE.template.md  # Optional custom template
  - type: cursor
    enabled: true
  - type: windsurf
    enabled: true
rules: [rule references with optional variables]
sources: [custom git repositories]
generation: [build settings]
```

## Testing

- **Unit**: `make test` - All `*_test.go` files
- **Integration**: Git operations with real repositories
- **E2E**: Full CLI workflow testing with fixtures
- **Coverage**: Generates `coverage.out` for analysis
- **Tips**: Reuse the test dependency helpers (`internal/commands/test_helpers.go:12`, `internal/dependencies/deps.go:31`) to keep tests hermetic.

## Dependencies

- **CLI**: urfave/cli/v3
- **TUI**: Charm ecosystem (bubbletea, lipgloss, huh)
- **Git**: go-git/go-git/v5
- **Validation**: go-playground/validator/v10
- **Rendering**: charmbracelet/glamour (markdown)

## Error Handling

Custom error types in `internal/errors` with display formatting and exit codes.

## Build System

Go 1.25+ required. Binary name: `contexture`. CI runs on Ubuntu with matrix builds for linux/windows/darwin on amd64/arm64.

# Rules

# Preferred Tools

Preferred tools for the project

**Applies:** Always active


**Tags:** tools
You run in an environment where several high performance, productivity enhancing tools are available to help you with your tasks. You should always use these tools rather than the default system tools wherever possible.

Whenever a search requires syntax-aware or structural matching, default to ast-grep run --lang go --pattern "<pattern>" (or set --lang appropriately) and avoid falling back to text-only tools like rg or grep unless I explicitly request a plain-text search.

If you need to find specific files, use 'fd'

If you need to find specific text/strings, use 'rg' (ripgrep)

If you need to select from multiple results, pipe to 'fzf'

Do you need to interact with JSON? use 'jq' -- or 'yq' for YAML

<!-- id: [contexture(local):tools] -->

---

# Thought Process

Thought process when working on the project

**Applies:** Always active


**Tags:** process
Think carefully and only action the specific task that I have given you, with the most concise and elegant solution that changes as little code as possible.

<!-- id: [contexture(local):think] -->
