# End-to-End Test Suite

This package provides comprehensive end-to-end tests for the Contexture CLI application, testing the complete user experience from command invocation to final output generation.

## Purpose

The e2e package validates the entire application stack by:
- Testing the actual compiled binary in realistic scenarios
- Verifying complete user workflows from start to finish
- Ensuring all components work together correctly in real environments
- Catching integration issues that unit tests might miss

## Test Coverage

### Core CLI Functionality
- **Basic Commands**: Help, version, and command structure validation
- **Project Lifecycle**: Initialization, configuration management, and project operations
- **Rule Management**: Adding, removing, listing, and updating rules
- **Build Process**: Complete build workflows with multiple output formats
- **Interactive Features**: Terminal UI components and user interaction flows

### Network and External Dependencies
- **Git Repository Access**: Remote rule fetching and repository operations
- **Network Connectivity**: Testing with various network conditions
- **Cache Behavior**: Repository caching and cache invalidation scenarios

### User Experience Validation
- **Command-Line Interface**: Argument parsing, flag handling, and help output
- **Error Handling**: User-friendly error messages and recovery guidance
- **Terminal UI**: Interactive components like rule selectors and prompts
- **Output Formatting**: Generated files and console output validation

## Test Structure

### Test Files
- `cli_test.go`: Basic CLI functionality and command structure tests
- `workflow_test.go`: Complete user workflow scenarios
- `tui_test.go`: Terminal UI component testing
- `network_test.go`: Network-dependent functionality tests
- `integration_test.go`: Cross-component integration scenarios

### Support Infrastructure
- `helpers/cli.go`: CLI runner utilities for executing commands and validating results
- `fixtures/rules.go`: Test data and rule definitions for consistent testing

## Test Execution

### Prerequisites
- Compiled `contexture` binary must be available in `./bin/contexture`
- Tests require network access for Git repository operations
- Terminal capabilities for UI component testing

### Running Tests
```bash
# Run all e2e tests
go test ./e2e/...

# Skip e2e tests in short mode
go test -short ./...

# Run specific test suites
go test ./e2e/ -run TestCLI
go test ./e2e/ -run TestWorkflow
```

### Test Environment
- Uses temporary directories for isolated test execution
- Configures appropriate timeouts for external operations
- Manages test fixtures and cleanup automatically
- Supports both interactive and headless execution modes

## Helper Utilities

### CLI Runner
- `CLIRunner`: Executes CLI commands with configurable environments
- Result validation with fluent assertion interface
- Timeout management for long-running operations
- Working directory and environment variable management

### Test Fixtures
- Predefined rule sets for consistent test scenarios
- Sample project configurations for various use cases
- Network-accessible test repositories for Git functionality

## Relationship to Other Test Packages

- **Unit Tests**: Individual package tests focus on isolated component behavior
- **Integration Tests**: The `integration/` package tests component interactions without the full CLI
- **E2E Tests**: This package tests the complete user experience with the actual binary

## Continuous Integration

These tests serve as the final validation in CI/CD pipelines:
- Ensure the built binary works correctly in production-like environments
- Validate that all features work together as expected
- Catch regressions in user-facing functionality
- Verify cross-platform compatibility when run on multiple OS targets