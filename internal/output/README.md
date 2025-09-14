# Output Package

The `output` package provides extensible output formatting for CLI commands, supporting multiple output formats including terminal display and JSON.

## Architecture

### Core Components

- **`Writer` Interface**: Defines methods for writing different types of command output
- **`Manager`**: Handles format selection and delegates to appropriate writers
- **Format Types**: `FormatDefault` (terminal) and `FormatJSON` 
- **Metadata Structs**: Contextual information for each command type

### Writers

- **`JSONWriter`**: Outputs structured JSON for programmatic consumption
- **`TerminalWriter`**: Delegates to existing terminal display logic

## Usage

### Basic Usage

```go
import "github.com/contextureai/contexture/internal/output"

// Create manager for desired format
manager, err := output.NewManager(output.FormatJSON)
if err != nil {
    return err
}

// Prepare metadata
metadata := output.ListMetadata{
    Command:       "rules list",
    Pattern:       "testing",
    TotalRules:    10,
    FilteredRules: 3,
    Timestamp:     time.Now(),
}

// Write output
err = manager.WriteRulesList(rules, metadata)
```

### Adding New Output Formats

1. Define new format constant:
```go
const FormatYAML Format = "yaml"
```

2. Implement Writer interface:
```go
type YAMLWriter struct{}

func (w *YAMLWriter) WriteRulesList(rules []*domain.Rule, metadata ListMetadata) error {
    // Implementation
}
```

3. Update Manager factory:
```go
case FormatYAML:
    writer = NewYAMLWriter()
```

### Adding New Command Types

1. Define metadata struct:
```go
type ConfigMetadata struct {
    Command   string    `json:"command"`
    Version   string    `json:"version"`
    Path      string    `json:"path"`
    Timestamp time.Time `json:"timestamp"`
}
```

2. Add method to Writer interface:
```go
type Writer interface {
    WriteRulesList(rules []*domain.Rule, metadata ListMetadata) error
    WriteConfig(config *domain.Config, metadata ConfigMetadata) error
}
```

3. Implement in all writers

## JSON Schema

### Rules List Output

```json
{
  "command": "rules list",
  "version": "1.0",
  "metadata": {
    "command": "rules list",
    "version": "1.0",
    "pattern": "optional-filter-pattern",
    "totalRules": 10,
    "filteredRules": 3, 
    "timestamp": "2025-09-14T19:30:45Z"
  },
  "rules": [
    {
      "id": "rule-identifier",
      "title": "Rule Title",
      "description": "Rule description",
      "tags": ["tag1", "tag2"],
      "languages": ["go", "python"],
      "frameworks": ["gin", "fastapi"],
      "trigger": {
        "type": "glob",
        "globs": ["**/*.go"]
      },
      "content": "Rule content...",
      "variables": {},
      "defaultVariables": {},
      "filePath": "path/to/rule.md",
      "source": "https://github.com/repo/rules.git",
      "ref": "main",
      "createdAt": "2025-01-01T00:00:00Z",
      "updatedAt": "2025-01-01T00:00:00Z"
    }
  ]
}
```

## Error Handling

The package provides typed errors for better error handling:

```go
var unsupportedErr *output.UnsupportedFormatError
if errors.As(err, &unsupportedErr) {
    fmt.Printf("Format '%s' not supported\n", unsupportedErr.Format)
}
```

## Testing

The package includes comprehensive tests:

- **Unit Tests**: Individual writer and manager functionality
- **Integration Tests**: End-to-end output generation  
- **E2E Tests**: Full CLI workflow testing

Run tests with:
```bash
go test ./internal/output
```

## Design Principles

1. **Extensibility**: Easy to add new formats and command types
2. **Consistency**: Uniform JSON schema across commands
3. **Separation of Concerns**: Clear boundaries between formatting logic
4. **Backward Compatibility**: Terminal output unchanged
5. **Testability**: Comprehensive test coverage with clear interfaces