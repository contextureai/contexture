# TUI Package

This package contains simplified terminal user interface components for Contexture.

## Components

- **Prompts** (`prompt.go`): Simple inline prompts using huh library
  - `HandleFormError`: Error handling for huh forms
  - `Select`: Single selection prompts  
  - `MultiSelect`: Multiple selection prompts
  - `ErrUserCancelled`: User cancellation error

## Usage

The TUI components provide inline prompts for CLI commands:

- `init` - Uses prompts for project initialization and format selection
- `config formats` - Uses prompts for format management
- `update` - Uses prompts for confirmation dialogs

## Architecture

The TUI components use the huh library for simple inline prompts that:

- **Stay inline**: Don't take over the full terminal screen
- **Are scriptable**: Work well in CI/CD environments  
- **Follow CLI conventions**: Consistent with standard CLI tools
- **Are easy to test**: Simple prompt-based interactions

## Design Philosophy

This package follows the principle of **simple, inline interactions** rather than complex full-screen UIs:

- ✅ Inline prompts that stay in terminal flow
- ✅ Consistent with standard CLI tools
- ✅ Easy to test and maintain
- ❌ No full-screen terminal takeover
- ❌ No complex state management
- ❌ No bubble tea full-screen components

## Charmbracelet Integration

This package uses the charmbracelet ecosystem:
- **`huh`**: For simple inline form and prompt components

## API

The package exposes the following components:

- `Select(opts SelectOptions)`: Single selection prompt
- `MultiSelect(opts MultiSelectOptions)`: Multiple selection prompt  
- `HandleFormError(err)`: Convert huh errors to user-friendly messages
- `ErrUserCancelled`: Standard cancellation error

## Error Handling

- `HandleFormError(err)`: Converts `charmbracelet/huh` errors into user-friendly messages
- `ErrUserCancelled`: Returned when user cancels an operation (e.g., by pressing ESC)

## Testing

Components include focused tests for:
- Prompt functionality
- Error handling
- User cancellation scenarios