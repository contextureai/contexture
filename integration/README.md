# Integration Test Suite

This package provides integration tests for the Contexture CLI application, focusing on testing component interactions and core functionality without the full end-to-end binary execution overhead.

## Purpose

The integration package serves as a middle layer between unit tests and end-to-end tests by:
- Testing interactions between multiple internal packages
- Validating core business logic workflows without CLI overhead
- Focusing on component integration rather than user interface
- Providing faster feedback than full e2e tests while maintaining broader coverage than unit tests

## Test Focus Areas

### Core Component Integration
- **Project Management**: Configuration loading, validation, and persistence workflows
- **Rule Processing**: Integration between fetchers, parsers, processors, and validators
- **Git Operations**: Repository access, caching, and rule fetching from remote sources
- **Format Generation**: End-to-end format processing without CLI interface

### Business Logic Validation
- **Configuration Management**: Project initialization and configuration manipulation
- **Rule Lifecycle**: Complete rule processing from source to processed output
- **Validation Workflows**: Cross-package validation and error handling
- **Template Processing**: Rule content processing with variable substitution

## Test Structure

### Core Integration Tests
- `integration_test.go`: Primary integration test scenarios for core functionality
- `git_integration_test.go`: Git repository operations and caching integration

### Git Testing Infrastructure
- `git_test_helpers.go`: Utilities for Git repository testing and setup
- `git_test_examples.go`: Sample Git repository structures and test data

## Testing Approach

### In-Memory Testing
- Uses `afero.MemMapFs` for filesystem operations to avoid side effects
- Creates isolated test environments for each scenario
- Enables fast, repeatable test execution without external dependencies

### Component Composition
- Tests real component implementations rather than mocks
- Validates actual integration points between packages
- Ensures components work correctly when composed together

### Realistic Scenarios
- Tests common user workflows at the component level
- Validates error handling and edge cases in integrated scenarios
- Ensures business logic consistency across package boundaries

## Test Categories

### Project Management Integration
```go
func TestProjectManagerIntegration(t *testing.T)
```
- Project configuration initialization and loading
- Configuration validation with complex rule references
- Configuration persistence and atomic updates

### Git Operations Integration  
```go
func TestGitIntegration(t *testing.T)
```
- Repository cloning and caching behavior
- Rule fetching from various Git sources
- Branch and commit-specific rule access

### Rule Processing Integration
```go  
func TestRuleProcessingIntegration(t *testing.T)
```
- End-to-end rule processing pipeline
- Template processing with variable resolution
- Validation integration across rule lifecycle

## Relationship to Other Test Packages

### Compared to Unit Tests
- **Broader Scope**: Tests multiple packages working together
- **Integration Focus**: Validates component interactions rather than isolated behavior
- **Realistic Data Flow**: Uses real data flows between components

### Compared to E2E Tests
- **No CLI Overhead**: Tests components directly without binary execution
- **Faster Execution**: Runs significantly faster than full e2e tests
- **Component Focus**: Tests business logic rather than user interface
- **Isolated Environment**: Uses in-memory implementations for consistency

## Test Execution

### Running Integration Tests
```bash
# Run all integration tests
go test ./integration/...

# Run specific integration test suites
go test ./integration/ -run TestProjectManager
go test ./integration/ -run TestGit
```

### Test Requirements
- No external dependencies required (uses in-memory implementations)
- No network access needed (mocks external Git repositories where appropriate)
- Fast execution suitable for frequent development feedback

## Helper Utilities

### Git Test Infrastructure
- Repository creation and management utilities
- Sample repository structures for testing various scenarios
- Mock Git server functionality for network operation testing

### Test Data Management
- Consistent test fixtures for repeatable scenarios
- Sample project configurations and rule definitions
- Validation test cases covering edge conditions

## Continuous Integration Role

Integration tests serve as:
- **Fast Feedback**: Quicker validation than e2e tests during development
- **Component Validation**: Ensures internal packages work together correctly
- **Regression Prevention**: Catches integration issues before they reach e2e testing
- **Development Confidence**: Provides assurance that refactoring preserves functionality