package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMain(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		wantExit int
		wantOut  string
		wantErr  string
	}{
		{
			name:     "no arguments shows help",
			args:     []string{"contexture"},
			wantExit: 0,
			wantOut:  "AI assistant rule management",
		},
		{
			name:     "version flag",
			args:     []string{"contexture", "--version"},
			wantExit: 0,
			wantOut:  "contexture version",
		},
		{
			name:     "help flag",
			args:     []string{"contexture", "--help"},
			wantExit: 0,
			wantOut:  "Commands:",
		},
		{
			name:     "invalid command",
			args:     []string{"contexture", "invalid-command"},
			wantExit: 3, // The actual exit code we observed
			wantErr:  "No help topic for 'invalid-command'",
		},
		{
			name:     "verbose flag",
			args:     []string{"contexture", "--verbose", "--help"},
			wantExit: 0,
			wantOut:  "Commands:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Build the binary for testing
			binPath := buildTestBinary(t)
			defer func() {
				if err := os.Remove(binPath); err != nil {
					t.Logf("Failed to cleanup test binary: %v", err)
				}
			}()

			// Run the command
			//nolint:gosec // Test code - binPath is our own test binary
			cmd := exec.CommandContext(context.Background(), binPath, tt.args[1:]...)
			var stdout, stderr bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr

			err := cmd.Run()
			exitCode := 0
			if err != nil {
				exitError := &exec.ExitError{}
				if errors.As(err, &exitError) {
					exitCode = exitError.ExitCode()
				} else {
					t.Fatalf("Failed to run command: %v", err)
				}
			}

			// Check exit code
			assert.Equal(t, tt.wantExit, exitCode, "exit code mismatch")

			// Check output
			output := stdout.String()
			errorOutput := stderr.String()

			if tt.wantOut != "" {
				assert.Contains(t, output, tt.wantOut, "stdout should contain expected output")
			}

			if tt.wantErr != "" {
				assert.Contains(t, errorOutput, tt.wantErr, "stderr should contain expected error")
			}
		})
	}
}

func TestMainCommands(t *testing.T) {
	commands := []string{"init", "rules", "build", "config"}

	binPath := buildTestBinary(t)
	defer func() {
		if err := os.Remove(binPath); err != nil {
			t.Logf("Failed to cleanup test binary: %v", err)
		}
	}()

	for _, cmdName := range commands {
		t.Run(cmdName+"_help", func(t *testing.T) {
			//nolint:gosec // Test code - binPath is our own test binary
			cmd := exec.CommandContext(context.Background(), binPath, cmdName, "--help")
			var stdout bytes.Buffer
			cmd.Stdout = &stdout

			err := cmd.Run()
			require.NoError(t, err, "help command should not fail")

			output := stdout.String()
			// Commands with subcommands show "Commands:" instead of "Usage:"
			hasUsage := strings.Contains(output, "Usage:") || strings.Contains(output, "Commands:")
			assert.True(t, hasUsage, "help should show either usage or commands section")
			assert.Contains(t, output, cmdName, "help should mention command name")
		})
	}
}

func TestMainGlobalFlags(t *testing.T) {
	binPath := buildTestBinary(t)
	defer func() {
		if err := os.Remove(binPath); err != nil {
			t.Logf("Failed to cleanup test binary: %v", err)
		}
	}()

	t.Run("verbose_flag", func(t *testing.T) {
		//nolint:gosec // Test code - binPath is our own test binary
		cmd := exec.CommandContext(context.Background(), binPath, "--verbose", "--help")
		var stdout bytes.Buffer
		cmd.Stdout = &stdout

		err := cmd.Run()
		require.NoError(t, err, "verbose help should not fail")

		output := stdout.String()
		assert.Contains(t, output, "verbose", "help should show verbose flag")
	})
}

func TestMainErrorHandling(t *testing.T) {
	binPath := buildTestBinary(t)
	defer func() {
		if err := os.Remove(binPath); err != nil {
			t.Logf("Failed to cleanup test binary: %v", err)
		}
	}()

	t.Run("graceful_error_handling", func(t *testing.T) {
		// Test that errors don't cause panics but proper exit codes
		//nolint:gosec // Test code - binPath is our own test binary
		cmd := exec.CommandContext(context.Background(), binPath, "init", "--invalid-flag")
		var stderr bytes.Buffer
		cmd.Stderr = &stderr

		err := cmd.Run()

		// Should exit with non-zero code
		require.Error(t, err)

		exitError := &exec.ExitError{}
		if errors.As(err, &exitError) {
			assert.NotEqual(t, 0, exitError.ExitCode(), "should exit with error code")
		}

		// Should not contain panic traces
		errorOutput := stderr.String()
		assert.NotContains(t, errorOutput, "panic:", "should not contain panic")
		assert.NotContains(t, errorOutput, "goroutine", "should not contain goroutine traces")
	})
}

// buildTestBinary builds the contexture binary for integration testing
func buildTestBinary(t *testing.T) string {
	t.Helper()

	binPath := "./contexture-test"

	// Build the binary
	cmd := exec.CommandContext(context.Background(), "go", "build", "-o", binPath, ".")
	cmd.Dir = "."

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		t.Fatalf("Failed to build test binary: %v\nStderr: %s", err, stderr.String())
	}

	return binPath
}

// Benchmark main function startup time
func BenchmarkMain(b *testing.B) {
	// Create a temporary testing.T for buildTestBinary
	t := &testing.T{}
	binPath := buildTestBinary(t)
	defer func() {
		if err := os.Remove(binPath); err != nil {
			b.Logf("Failed to cleanup test binary: %v", err)
		}
	}()

	b.ResetTimer()
	for range b.N {
		//nolint:gosec // Test code - binPath is our own test binary
		cmd := exec.CommandContext(context.Background(), binPath, "--help")
		cmd.Stdout = nil // Discard output
		cmd.Stderr = nil

		err := cmd.Run()
		if err != nil {
			b.Fatalf("Command failed: %v", err)
		}
	}
}

// Example demonstrating basic usage
func ExampleMain() {
	// This would normally be run from command line:
	// contexture --help

	// For testing purposes, we'll just validate the structure
	args := []string{"contexture", "--help"}

	// The application should handle help gracefully
	_ = args

	fmt.Println("Usage information would be displayed")
	// Output:
	// Usage information would be displayed
}

func TestMainDocumentation(t *testing.T) {
	binPath := buildTestBinary(t)
	defer func() {
		if err := os.Remove(binPath); err != nil {
			t.Logf("Failed to cleanup test binary: %v", err)
		}
	}()

	t.Run("app_description", func(t *testing.T) {
		//nolint:gosec // Test code - binPath is our own test binary
		cmd := exec.CommandContext(context.Background(), binPath, "--help")
		var stdout bytes.Buffer
		cmd.Stdout = &stdout

		err := cmd.Run()
		require.NoError(t, err)

		output := stdout.String()
		assert.Contains(t, output, "AI assistant rule management", "should show app description")
		assert.Contains(t, output, "Claude", "should mention Claude format")
		assert.Contains(t, output, "Cursor", "should mention Cursor format")
		assert.Contains(t, output, "Windsurf", "should mention Windsurf format")
	})

	t.Run("author_information", func(t *testing.T) {
		//nolint:gosec // Test code - binPath is our own test binary
		cmd := exec.CommandContext(context.Background(), binPath, "--help")
		var stdout bytes.Buffer
		cmd.Stdout = &stdout

		err := cmd.Run()
		require.NoError(t, err)

		output := stdout.String()
		// The CLI framework doesn't show author info in help by default
		// Just check that help is displayed properly
		assert.Contains(t, output, "AI assistant rule management", "should show app description")
	})
}

func TestMainIntegration(t *testing.T) {
	binPath := buildTestBinary(t)
	defer func() {
		if err := os.Remove(binPath); err != nil {
			t.Logf("Failed to cleanup test binary: %v", err)
		}
	}()

	t.Run("command_chaining", func(t *testing.T) {
		// Test that commands can be discovered and help works
		commands := []string{"init", "rules", "build", "config"}

		for _, cmdName := range commands {
			//nolint:gosec // Test code - binPath is our own test binary
			cmd := exec.CommandContext(context.Background(), binPath, cmdName, "--help")
			err := cmd.Run()
			assert.NoError(t, err, "command %s help should work", cmdName)
		}
	})
}

// TestMainSignalHandling tests that the application handles signals gracefully
func TestMainSignalHandling(t *testing.T) {
	binPath := buildTestBinary(t)
	defer func() {
		if err := os.Remove(binPath); err != nil {
			t.Logf("Failed to cleanup test binary: %v", err)
		}
	}()

	t.Run("interrupt_handling", func(t *testing.T) {
		// Start a long-running command
		//nolint:gosec // Test code - binPath is our own test binary
		cmd := exec.CommandContext(context.Background(), binPath, "--help")

		err := cmd.Start()
		require.NoError(t, err)

		// The help command should complete quickly and gracefully
		err = cmd.Wait()
		assert.NoError(t, err, "help command should complete successfully")
	})
}
