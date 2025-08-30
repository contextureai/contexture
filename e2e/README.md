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

## E2E Test Architecture

### Test Execution Infrastructure

```mermaid
graph TB
    subgraph "Test Environment"
        TEST[Test Suite]
        HELPER[Helper Utilities]
        FIXTURE[Test Fixtures]
        CLEANUP[Cleanup Manager]
    end
    
    subgraph "Isolated Test Project"
        TESTDIR[Temporary Test Directory]
        CONFIG[".contexture.yaml"]
        RULES["rules/" Directory]
        OUTPUT["Output Files"]
    end
    
    subgraph "Binary Execution"
        BINARY["./bin/contexture"]
        PROCESS[OS Process]
        STDOUT[Standard Output]
        STDERR[Standard Error]
        EXITCODE[Exit Code]
    end
    
    subgraph "Filesystem Operations"
        FS[afero.Fs]
        OSFS[afero.OsFs]
        MEMFS[afero.MemMapFs]
        FILEOPS[File Operations]
    end
    
    subgraph "Validation Layer"
        EXPECT[Expectation Matchers]
        ASSERT[Assertions]
        RESULTS[Test Results]
    end
    
    TEST --> HELPER
    HELPER --> FIXTURE
    HELPER --> TESTDIR
    
    TESTDIR --> CONFIG
    TESTDIR --> RULES
    TESTDIR --> OUTPUT
    
    HELPER --> BINARY
    BINARY --> PROCESS
    PROCESS --> STDOUT
    PROCESS --> STDERR
    PROCESS --> EXITCODE
    
    HELPER --> FS
    FS --> OSFS
    FS --> MEMFS
    FS --> FILEOPS
    
    FILEOPS --> CONFIG
    FILEOPS --> RULES
    FILEOPS --> OUTPUT
    
    STDOUT --> EXPECT
    STDERR --> EXPECT
    EXITCODE --> EXPECT
    OUTPUT --> EXPECT
    
    EXPECT --> ASSERT
    ASSERT --> RESULTS
    
    CLEANUP --> TESTDIR
    CLEANUP --> OUTPUT
    
    style TEST fill:#e1f5fe
    style BINARY fill:#f3e5f5
    style FS fill:#e8f5e8
    style EXPECT fill:#fff3e0
```

### Complete Workflow Test Flow

```mermaid
sequenceDiagram
    participant Test as Test Case
    participant Project as TestProject
    participant CLI as CLIRunner
    participant Binary as contexture
    participant FS as Filesystem
    participant Validate as Validator
    
    Test->>Project: NewTestProject(fs, binaryPath)
    Project->>FS: Create temp directory
    Project->>Project: Setup CLI runner
    
    Note over Test: Test Setup Phase
    Test->>Project: WithConfig(yaml)
    Project->>FS: Write .contexture.yaml
    Test->>Project: WithLocalRule(path, content)
    Project->>FS: Write rules/*.md
    
    Note over Test: Command Execution Phase
    Test->>Project: Run("init", "--force")
    Project->>CLI: Execute command
    CLI->>Binary: Spawn process with args
    Binary->>FS: Read/Write config files
    Binary-->>CLI: stdout, stderr, exit code
    CLI-->>Project: CLIResult
    Project-->>Test: CLIResult
    
    Note over Test: Validation Phase
    Test->>Validate: ExpectSuccess()
    Test->>Validate: ExpectStdout("text")
    Test->>Project: AssertFileExists(path)
    Project->>FS: Check file existence
    FS-->>Project: File status
    Project-->>Test: Assertion result
    
    Test->>Project: AssertFileContains(path, content)
    Project->>FS: Read file content
    FS-->>Project: File content
    Project->>Validate: Content validation
    Validate-->>Test: Assertion result
    
    Note over Test: Multi-Command Workflows
    Test->>Project: Run("build")
    Project->>CLI: Execute build
    CLI->>Binary: Spawn build process
    Binary->>FS: Read rules, generate output
    Binary-->>CLI: Build results
    CLI-->>Test: Build completion
    
    Test->>Project: GetFileContent("CLAUDE.md")
    Project->>FS: Read generated file
    FS-->>Project: File content
    Project-->>Test: Content for validation
    
    Note over Test: Cleanup Phase
    Test->>Project: Cleanup (automatic)
    Project->>FS: Remove temp directory
    Project->>FS: Clean up test files
```

### Error Handling and Recovery Testing

```mermaid
flowchart TD
    START([Start Error Test]) --> SETUP[Setup Test Environment]
    
    SETUP --> CORRUPT{Introduce Error Condition}
    
    CORRUPT -->|Config Error| BADCONFIG[Create Invalid Config]
    CORRUPT -->|Rule Error| BADRULE[Create Invalid Rule]
    CORRUPT -->|Permission Error| PERMISSION[Simulate Permission Issue]
    CORRUPT -->|Network Error| NETWORK[Simulate Network Failure]
    
    BADCONFIG --> TESTFAIL[Execute Command - Expect Failure]
    BADRULE --> TESTFAIL
    PERMISSION --> TESTFAIL
    NETWORK --> TESTFAIL
    
    TESTFAIL --> VALIDATE[Validate Error Response]
    VALIDATE --> RECOVERY[Test Recovery Action]
    
    RECOVERY -->|Re-init| REINIT[Run init --force]
    RECOVERY -->|Fix Config| FIXCONFIG[Correct Configuration]
    RECOVERY -->|Fix Rule| FIXRULE[Correct Rule Content]
    RECOVERY -->|Retry| RETRY[Retry Operation]
    
    REINIT --> TESTOK[Execute Command - Expect Success]
    FIXCONFIG --> TESTOK
    FIXRULE --> TESTOK
    RETRY --> TESTOK
    
    TESTOK --> FINAL[Validate Recovery Success]
    FINAL --> END([Test Complete])
    
    subgraph "Error Categories"
        direction TB
        E1["Configuration Errors<br/>• Malformed YAML<br/>• Invalid structure<br/>• Missing required fields"]
        E2["Rule Validation Errors<br/>• Missing frontmatter<br/>• Invalid metadata<br/>• Malformed content"]
        E3["File System Errors<br/>• Permission denied<br/>• Missing directories<br/>• Disk space issues"]
        E4["Network Errors<br/>• Git clone failures<br/>• Repository access<br/>• Timeout issues"]
    end
    
    subgraph "Recovery Strategies"
        direction TB
        R1["Graceful Degradation<br/>• Continue with partial success<br/>• Skip problematic rules<br/>• Generate available formats"]
        R2["User Guidance<br/>• Clear error messages<br/>• Suggested fixes<br/>• Recovery instructions"]
        R3["State Recovery<br/>• Config regeneration<br/>• Cache invalidation<br/>• Directory recreation"]
    end
    
    style START fill:#e1f5fe
    style TESTFAIL fill:#ffebee
    style TESTOK fill:#e8f5e8
    style END fill:#f3e5f5
```

### Test Organization and Coverage

```mermaid
mindmap
  root((E2E Tests))
    CLI Basics
      Help Commands
      Version Info
      Invalid Args
      Error Handling
    
    Project Lifecycle
      Init Command
        Default Config
        Force Override
        Format Selection
        Directory Structure
      
      Config Management
        Show Config
        Format Operations
          Add Format
          Enable/Disable
          Remove Format
        
        Rule Operations
          List Rules
          Add Rules
          Remove Rules
          Update Rules
    
    Complete Workflows
      Basic Workflow
        init → config → build
        Local Rules
        Output Validation
      
      Mixed Rules
        Local + Remote
        Variable Substitution
        Template Processing
      
      Error Recovery
        Config Corruption
        Rule Validation
        Network Failures
        Permission Issues
      
      Large Projects
        Many Rules
        Performance Testing
        Memory Usage
        Build Time
    
    Cross-Platform
      Windows
      macOS
      Linux
      
    Network Operations
      Git Repository
        Clone Operations
        Authentication
        Branch Selection
        Cache Management
      
      Offline Mode
        Cache Usage
        Local Fallback
        Error Handling
```

## Continuous Integration

These tests serve as the final validation in CI/CD pipelines:
- Ensure the built binary works correctly in production-like environments
- Validate that all features work together as expected
- Catch regressions in user-facing functionality
- Verify cross-platform compatibility when run on multiple OS targets