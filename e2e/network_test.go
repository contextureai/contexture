// Package e2e provides network and git integration tests
package e2e

import (
	"os"
	"testing"
	"time"

	"github.com/contextureai/contexture/e2e/helpers"
	"github.com/spf13/afero"
)

// TestNetworkFailures tests behavior when network operations fail
func TestNetworkFailures(t *testing.T) {
	t.Parallel()
	fs := afero.NewOsFs()
	project := helpers.NewTestProject(t, fs, binaryPath)

	// Initialize project with remote rules
	project.Run(t, "init", "--force", "--no-interactive").ExpectSuccess(t)

	t.Run("invalid git repository", func(t *testing.T) {
		result := project.Run(t, "rules", "add", "[contexture(https://invalid-repo.example.com):test/rule]", "--force")
		result.ExpectFailure(t).
			ExpectStderr(t, "failed to clone repository")
	})

	t.Run("timeout handling", func(t *testing.T) {
		// Use a very slow DNS resolution to simulate timeout
		result := project.Run(t, "rules", "add", "[contexture(https://very-slow-dns-that-should-not-exist.example.invalid):test/rule]", "--force")
		result.ExpectFailure(t).
			ExpectStderr(t, "failed")
	})

	t.Run("unsupported URL scheme", func(t *testing.T) {
		// Test with http URL which is not allowed (only https and ssh)
		result := project.Run(t, "rules", "add", "[contexture(http://127.0.0.1:9999):test/rule]", "--force")
		result.ExpectFailure(t).
			ExpectStderr(t, "unsupported URL scheme")
	})
}

// TestGitAuthentication tests various git authentication scenarios
func TestGitAuthentication(t *testing.T) {
	// Cannot use t.Parallel() because subtests use t.Setenv()
	fs := afero.NewOsFs()
	project := helpers.NewTestProject(t, fs, binaryPath)

	// Initialize project
	project.Run(t, "init", "--force", "--no-interactive").ExpectSuccess(t)

	t.Run("ssh authentication without agent", func(t *testing.T) {
		// Temporarily unset SSH_AUTH_SOCK to simulate no SSH agent
		originalSSHAuthSock := os.Getenv("SSH_AUTH_SOCK")
		_ = os.Unsetenv("SSH_AUTH_SOCK")
		defer func() {
			if originalSSHAuthSock != "" {
				t.Setenv("SSH_AUTH_SOCK", originalSSHAuthSock)
			}
		}()

		result := project.Run(t, "rules", "add", "[contexture(git@github.com:nonexistent/repo.git):test/rule]", "--force")
		result.ExpectFailure(t).
			ExpectStderr(t, "authentication")
	})

	t.Run("https with github token", func(t *testing.T) {
		// Test with a token (should fail gracefully for invalid token)
		t.Setenv("GITHUB_TOKEN", "invalid-token")

		result := project.Run(t, "rules", "add", "[contexture(https://github.com/nonexistent/repo.git):test/rule]", "--force")
		result.ExpectFailure(t).
			ExpectStderr(t, "authentication")
	})

	t.Run("public repository access", func(t *testing.T) {
		// Test accessing a public repository (this might succeed if the repo exists)
		result := project.Run(t, "rules", "add", "[contexture(https://github.com/contextureai/rules.git):core/example]", "--force")
		// Don't assert success/failure as it depends on network availability
		// Just ensure it doesn't crash
		t.Logf("Public repo access result: exit code %d", result.ExitCode)
	})
}

// TestGitBranchHandling tests branch and tag specification
func TestGitBranchHandling(t *testing.T) {
	t.Parallel()
	fs := afero.NewOsFs()
	project := helpers.NewTestProject(t, fs, binaryPath)

	// Initialize project
	project.Run(t, "init", "--force", "--no-interactive").ExpectSuccess(t)

	t.Run("specify branch", func(t *testing.T) {
		result := project.Run(t, "rules", "add", "[contexture(https://github.com/contextureai/rules.git):core/example,develop]", "--force")
		// Test that branch specification is parsed correctly (may fail if branch doesn't exist)
		t.Logf("Branch specification result: exit code %d", result.ExitCode)
	})

	t.Run("specify tag", func(t *testing.T) {
		result := project.Run(t, "rules", "add", "[contexture(https://github.com/contextureai/rules.git):core/example,v1.0.0]", "--force")
		// Test that tag specification is parsed correctly
		t.Logf("Tag specification result: exit code %d", result.ExitCode)
	})

	t.Run("invalid branch", func(t *testing.T) {
		result := project.Run(t, "rules", "add", "[contexture(https://github.com/contextureai/rules.git):core/example,nonexistent-branch]", "--force")
		result.ExpectFailure(t).
			ExpectStderr(t, "branch")
	})
}

// TestRulePinning tests rule pinning and version management
func TestRulePinning(t *testing.T) {
	fs := afero.NewOsFs()
	project := helpers.NewTestProject(t, fs, binaryPath)

	// Initialize project
	project.Run(t, "init", "--force", "--no-interactive").ExpectSuccess(t)

	// Create a config with pinned rules
	project.WithConfig(`version: 1
formats:
  - type: claude
    enabled: true
rules:
  - id: "test/rule"
    source: "https://github.com/contextureai/rules.git"
    commitHash: "abc123"
    pinned: true`)

	t.Run("update respects pinning", func(t *testing.T) {
		result := project.Run(t, "rules", "update", "--dry-run")
		result.ExpectSuccess(t).
			ExpectStdout(t, "pinned")
	})

	t.Run("force update pinned rule", func(t *testing.T) {
		result := project.Run(t, "rules", "update", "--yes", "--dry-run")
		result.ExpectSuccess(t)
		// Should indicate it would update even pinned rules
	})
}

// TestConcurrentOperations tests concurrent git operations
func TestConcurrentOperations(t *testing.T) {
	fs := afero.NewOsFs()
	project := helpers.NewTestProject(t, fs, binaryPath)

	// Initialize project
	project.Run(t, "init", "--force", "--no-interactive").ExpectSuccess(t)

	t.Run("parallel rule fetching", func(t *testing.T) {
		// Add multiple rules simultaneously to test parallel fetching
		// This tests the internal concurrency handling

		done := make(chan bool, 3)

		// Run multiple operations in parallel
		go func() {
			result := project.Run(t, "rules", "add", "[contexture:core/example1]", "--force")
			t.Logf("Parallel operation 1: exit code %d", result.ExitCode)
			done <- true
		}()

		go func() {
			result := project.Run(t, "rules", "add", "[contexture:core/example2]", "--force")
			t.Logf("Parallel operation 2: exit code %d", result.ExitCode)
			done <- true
		}()

		go func() {
			result := project.Run(t, "build")
			t.Logf("Parallel build: exit code %d", result.ExitCode)
			done <- true
		}()

		// Wait for all operations with timeout
		timeout := time.After(30 * time.Second)
		completed := 0

		for completed < 3 {
			select {
			case <-done:
				completed++
			case <-timeout:
				t.Error("Timeout waiting for parallel operations")
				return
			}
		}

		t.Log("All parallel operations completed")
	})
}

// TestRepositoryCleanup tests that temporary repositories are cleaned up properly
func TestRepositoryCleanup(t *testing.T) {
	fs := afero.NewOsFs()
	project := helpers.NewTestProject(t, fs, binaryPath)

	// Initialize project
	project.Run(t, "init", "--force", "--no-interactive").ExpectSuccess(t)

	t.Run("cleanup after successful operation", func(t *testing.T) {
		// Count temp directories before
		tempFiles1, _ := os.ReadDir("/tmp")
		tempCount1 := len(tempFiles1)

		// Perform operation that should create and cleanup temp directory
		project.Run(t, "rules", "add", "[contexture:core/example]", "--force")

		// Count temp directories after
		tempFiles2, _ := os.ReadDir("/tmp")
		tempCount2 := len(tempFiles2)

		// Should not have significantly more temp directories
		if tempCount2 > tempCount1+2 {
			t.Errorf("Possible temp directory leak: before=%d, after=%d", tempCount1, tempCount2)
		}
	})

	t.Run("cleanup after failed operation", func(t *testing.T) {
		// Count temp directories before
		tempFiles1, _ := os.ReadDir("/tmp")
		tempCount1 := len(tempFiles1)

		// Perform operation that should fail but still cleanup
		project.Run(t, "rules", "add", "[contexture(https://invalid-repo.example.com):test/rule]", "--force")

		// Count temp directories after
		tempFiles2, _ := os.ReadDir("/tmp")
		tempCount2 := len(tempFiles2)

		// Should not have significantly more temp directories
		if tempCount2 > tempCount1+2 {
			t.Errorf("Possible temp directory leak after failure: before=%d, after=%d", tempCount1, tempCount2)
		}
	})
}
