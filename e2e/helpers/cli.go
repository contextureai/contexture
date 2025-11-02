// Package helpers provides utilities for end-to-end testing of the Contexture CLI
package helpers

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

// CLIRunner provides utilities for running CLI commands in tests
type CLIRunner struct {
	BinaryPath     string
	WorkDir        string
	Env            []string
	Timeout        time.Duration
	SuppressOutput bool // Suppress CLI output logging unless test fails
}

// NewCLIRunner creates a new CLI runner with default settings
func NewCLIRunner(binaryPath string) *CLIRunner {
	return &CLIRunner{
		BinaryPath:     binaryPath,
		Timeout:        30 * time.Second,
		Env:            os.Environ(),
		SuppressOutput: false,
	}
}

// WithWorkDir sets the working directory for CLI commands
func (r *CLIRunner) WithWorkDir(dir string) *CLIRunner {
	r.WorkDir = dir
	return r
}

// WithTimeout sets the timeout for CLI commands
func (r *CLIRunner) WithTimeout(timeout time.Duration) *CLIRunner {
	r.Timeout = timeout
	return r
}

// WithEnv adds environment variables for CLI commands
func (r *CLIRunner) WithEnv(key, value string) *CLIRunner {
	r.Env = append(r.Env, fmt.Sprintf("%s=%s", key, value))
	return r
}

// WithSuppressedOutput suppresses CLI output logging unless test fails
func (r *CLIRunner) WithSuppressedOutput() *CLIRunner {
	r.SuppressOutput = true
	return r
}

// CLIResult contains the result of a CLI command execution
type CLIResult struct {
	ExitCode int
	Stdout   string
	Stderr   string
	Error    error
}

// Run executes a CLI command and returns the result
func (r *CLIRunner) Run(t *testing.T, args ...string) *CLIResult {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), r.Timeout)
	defer cancel()

	// #nosec G204 - BinaryPath is controlled in test environment
	cmd := exec.CommandContext(ctx, r.BinaryPath, args...)
	cmd.Dir = r.WorkDir
	cmd.Env = r.Env

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Always log the command being run (this is useful debug info)
	t.Logf("Running: %s %s", r.BinaryPath, strings.Join(args, " "))
	if r.WorkDir != "" {
		t.Logf("Working directory: %s", r.WorkDir)
	}

	err := cmd.Run()
	exitCode := 0
	if err != nil {
		var exitError *exec.ExitError
		if errors.As(err, &exitError) {
			exitCode = exitError.ExitCode()
		} else {
			// Command failed to start or was killed
			exitCode = -1
		}
	}

	result := &CLIResult{
		ExitCode: exitCode,
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		Error:    err,
	}

	// Log output based on suppression settings and command success
	shouldLogOutput := !r.SuppressOutput || exitCode != 0

	t.Logf("Exit code: %d", result.ExitCode)
	if shouldLogOutput && result.Stdout != "" {
		t.Logf("Stdout: %s", result.Stdout)
	} else if r.SuppressOutput && result.Stdout != "" && exitCode == 0 {
		t.Logf("Stdout: [suppressed - %d characters]", len(result.Stdout))
	}

	if shouldLogOutput && result.Stderr != "" {
		t.Logf("Stderr: %s", result.Stderr)
	} else if r.SuppressOutput && result.Stderr != "" && exitCode == 0 {
		t.Logf("Stderr: [suppressed - %d characters]", len(result.Stderr))
	}

	return result
}

// ExpectSuccess asserts that the command succeeded (exit code 0)
func (r *CLIResult) ExpectSuccess(t *testing.T) *CLIResult {
	t.Helper()
	require.Equal(t, 0, r.ExitCode, "Command should succeed. Stderr: %s", r.Stderr)
	return r
}

// ExpectFailure asserts that the command failed (non-zero exit code)
func (r *CLIResult) ExpectFailure(t *testing.T) *CLIResult {
	t.Helper()
	require.NotEqual(t, 0, r.ExitCode, "Command should fail")
	return r
}

// ExpectExitCode asserts the specific exit code
func (r *CLIResult) ExpectExitCode(t *testing.T, code int) *CLIResult {
	t.Helper()
	require.Equal(t, code, r.ExitCode, "Unexpected exit code. Stderr: %s", r.Stderr)
	return r
}

// ExpectStdout asserts that stdout contains the given text
func (r *CLIResult) ExpectStdout(t *testing.T, text string) *CLIResult {
	t.Helper()
	require.Contains(t, r.Stdout, text, "Stdout should contain text")
	return r
}

// ExpectStderr asserts that stderr contains the given text
func (r *CLIResult) ExpectStderr(t *testing.T, text string) *CLIResult {
	t.Helper()
	require.Contains(t, r.Stderr, text, "Stderr should contain text")
	return r
}

// ExpectNotStdout asserts that stdout does not contain the given text
func (r *CLIResult) ExpectNotStdout(t *testing.T, text string) *CLIResult {
	t.Helper()
	require.NotContains(t, r.Stdout, text, "Stdout should not contain text")
	return r
}

// ExpectNotStderr asserts that stderr does not contain the given text
func (r *CLIResult) ExpectNotStderr(t *testing.T, text string) *CLIResult {
	t.Helper()
	require.NotContains(t, r.Stderr, text, "Stderr should not contain text")
	return r
}

// TestProject represents a test project setup
type TestProject struct {
	Dir    string
	FS     afero.Fs
	Runner *CLIRunner
}

// NewTestProject creates a new test project in a parallel-safe test directory
func NewTestProject(t *testing.T, fs afero.Fs, binaryPath string) *TestProject {
	t.Helper()

	// Use t.TempDir() for parallel-safe temporary directories
	testDir := t.TempDir()

	// Create a unique HOME directory for this test to avoid conflicts with global config
	tmpHome := t.TempDir()

	// Get the project root directory (where the binary should be)
	cwd, err := os.Getwd()
	require.NoError(t, err)

	// If we're in the e2e directory, go up one level to project root
	if filepath.Base(cwd) == "e2e" {
		cwd = filepath.Dir(cwd)
	}

	// Ensure binary path is absolute
	var absPath string
	if filepath.IsAbs(binaryPath) {
		absPath = binaryPath
	} else {
		absPath = filepath.Join(cwd, binaryPath)
	}

	// No need for explicit cleanup - t.TempDir() handles it automatically

	// Set up runner with custom HOME environment
	runner := NewCLIRunner(absPath).
		WithWorkDir(testDir).
		WithSuppressedOutput().
		WithEnv("HOME", tmpHome)

	return &TestProject{
		Dir:    testDir,
		FS:     fs,
		Runner: runner,
	}
}

// WithFile adds a file to the test project
func (p *TestProject) WithFile(path, content string) *TestProject {
	fullPath := filepath.Join(p.Dir, path)
	dir := filepath.Dir(fullPath)
	_ = p.FS.MkdirAll(dir, 0o755)
	_ = afero.WriteFile(p.FS, fullPath, []byte(content), 0o644)
	return p
}

// WithConfig creates a .contexture.yaml config in the project
func (p *TestProject) WithConfig(content string) *TestProject {
	return p.WithFile(".contexture.yaml", content)
}

// WithLocalRule adds a local rule file to the rules directory
func (p *TestProject) WithLocalRule(rulePath, content string) *TestProject {
	return p.WithFile(filepath.Join("rules", rulePath+".md"), content)
}

// AssertFileExists asserts that a file exists in the project
func (p *TestProject) AssertFileExists(t *testing.T, path string) {
	t.Helper()
	fullPath := filepath.Join(p.Dir, path)
	exists, err := afero.Exists(p.FS, fullPath)
	require.NoError(t, err)
	require.True(t, exists, "File should exist: %s", path)
}

// AssertFileContains asserts that a file contains the given content
func (p *TestProject) AssertFileContains(t *testing.T, path, content string) {
	t.Helper()
	fullPath := filepath.Join(p.Dir, path)
	data, err := afero.ReadFile(p.FS, fullPath)
	require.NoError(t, err)
	require.Contains(t, string(data), content, "File should contain content: %s", path)
}

// GetFileContent reads and returns file content
func (p *TestProject) GetFileContent(t *testing.T, path string) string {
	t.Helper()
	fullPath := filepath.Join(p.Dir, path)
	data, err := afero.ReadFile(p.FS, fullPath)
	require.NoError(t, err)
	return string(data)
}

// GetDirectoryContent reads and returns concatenated content from all files in a directory
func (p *TestProject) GetDirectoryContent(t *testing.T, path string) string {
	t.Helper()
	fullPath := filepath.Join(p.Dir, path)

	// Check if it's a directory
	isDir, err := afero.IsDir(p.FS, fullPath)
	require.NoError(t, err)
	require.True(t, isDir, "Path should be a directory: %s", path)

	var content strings.Builder
	err = afero.Walk(p.FS, fullPath, func(walkPath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories, only read files
		if info.IsDir() {
			return nil
		}

		data, err := afero.ReadFile(p.FS, walkPath)
		if err != nil {
			return err
		}

		content.Write(data)
		content.WriteString("\n")
		return nil
	})
	require.NoError(t, err)

	return content.String()
}

// Run executes a CLI command in the project directory
func (p *TestProject) Run(t *testing.T, args ...string) *CLIResult {
	t.Helper()
	return p.Runner.Run(t, args...)
}
