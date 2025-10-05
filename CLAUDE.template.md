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

### Application Commands
- `contexture init` - Initialize project with `.contexture.yaml`
- `contexture rules add` - Add rules by ID
- `contexture rules list/remove/update` - Manage rules
- `contexture providers list/add/remove/show` - Manage rule providers
- `contexture config` - View/modify configuration
- `contexture build` - Generate platform-specific files

## Project Structure

```
cmd/contexture/        # CLI entry point
internal/
├── app/               # Main application and command orchestration
├── cli/               # CLI help templates and formatting
├── commands/          # Command implementations (init, build, rules, config, providers)
├── domain/            # Core business models and interfaces
├── format/            # Output format handlers (claude, cursor, windsurf)
├── rule/              # Rule processing, fetching, validation
├── provider/          # Provider registry and resolution
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

## Domain Concepts

### Rules
Markdown files with YAML frontmatter containing AI instructions. Support template variables via Go's `text/template`.

### Formats
- `claude`: Single `CLAUDE.md` file (supports custom templates via `CLAUDE.template.md`)
- `cursor`: Individual files in `.cursor/rules/`
- `windsurf`: Individual files in `.windsurf/rules/`

### Providers
- Default `@contexture` provider (bundled community repository)
- Custom named providers defined in `.contexture.yaml`
- Direct Git repository URLs
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
providers: [custom named providers]
generation: [build settings]
```

## Testing

- **Unit**: `make test` - All `*_test.go` files
- **Integration**: Git operations with real repositories
- **E2E**: Full CLI workflow testing with fixtures
- **Coverage**: Generates `coverage.out` for analysis

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

{{ .Rules }}
