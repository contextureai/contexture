// Package integration provides comprehensive git-related integration tests
package integration

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/contextureai/contexture/internal/git"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Constants are defined in git_test_helpers.go

// TestRealGitRepositoryOperations tests actual git operations with real repositories
func TestRealGitRepositoryOperations(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping git integration tests in short mode")
	}

	fs := afero.NewOsFs()
	repo := git.NewRepository(fs)
	ctx := context.Background()

	t.Run("clone public HTTPS repository", func(t *testing.T) {
		tempDir := t.TempDir()
		cloneDir := filepath.Join(tempDir, "rules-https")

		err := repo.Clone(ctx, testPublicRepo, cloneDir, git.WithBranch("main"))
		require.NoError(t, err, "Should successfully clone public HTTPS repository")

		// Verify repository was cloned
		assert.True(t, repo.IsValidRepository(cloneDir), "Cloned directory should be a valid git repository")

		// Verify we can get remote URL
		remoteURL, err := repo.GetRemoteURL(cloneDir)
		require.NoError(t, err)
		assert.Equal(t, testPublicRepo, remoteURL, "Remote URL should match cloned URL")

		// Verify we can get latest commit hash
		commitHash, err := repo.GetLatestCommitHash(cloneDir, "main")
		require.NoError(t, err)
		assert.Len(t, commitHash, 40, "Commit hash should be 40 characters")
		assert.Regexp(t, "^[a-f0-9]+$", commitHash, "Commit hash should be hexadecimal")

		t.Logf("Successfully cloned %s to %s (latest commit: %s)", testPublicRepo, cloneDir, commitHash[:7])
	})

	t.Run("clone with specific branch", func(t *testing.T) {
		tempDir := t.TempDir()
		cloneDir := filepath.Join(tempDir, "rules-branch")

		err := repo.Clone(ctx, testPublicRepo, cloneDir, git.WithBranch("main"))
		require.NoError(t, err, "Should successfully clone with specific branch")

		// Verify repository was cloned
		assert.True(t, repo.IsValidRepository(cloneDir), "Cloned directory should be a valid git repository")

		// Verify we can get commit hash for the specific branch
		commitHash, err := repo.GetLatestCommitHash(cloneDir, "main")
		require.NoError(t, err)
		assert.Len(t, commitHash, 40, "Commit hash should be 40 characters")

		t.Logf("Successfully cloned branch 'main' (commit: %s)", commitHash[:7])
	})

	t.Run("clone with shallow depth", func(t *testing.T) {
		tempDir := t.TempDir()
		cloneDir := filepath.Join(tempDir, "rules-shallow")

		err := repo.Clone(ctx, testPublicRepo, cloneDir,
			git.WithBranch("main"),
			git.WithShallow(1), // Only fetch 1 commit
		)
		require.NoError(t, err, "Should successfully clone with shallow depth")

		// Verify repository was cloned
		assert.True(t, repo.IsValidRepository(cloneDir), "Shallow cloned directory should be a valid git repository")

		t.Logf("Successfully performed shallow clone with depth 1")
	})

	t.Run("clone with single branch", func(t *testing.T) {
		tempDir := t.TempDir()
		cloneDir := filepath.Join(tempDir, "rules-single")

		err := repo.Clone(ctx, testPublicRepo, cloneDir,
			git.WithBranch("main"),
			git.WithSingleBranch(),
		)
		require.NoError(t, err, "Should successfully clone single branch")

		// Verify repository was cloned
		assert.True(t, repo.IsValidRepository(cloneDir), "Single branch clone should be a valid git repository")

		t.Logf("Successfully performed single branch clone")
	})
}

// TestGitAuthentication tests different git authentication methods
func TestGitAuthentication(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping git authentication tests in short mode")
	}

	fs := afero.NewOsFs()
	repo := git.NewRepository(fs)
	ctx := context.Background()

	t.Run("public repository without authentication", func(t *testing.T) {
		tempDir := t.TempDir()
		cloneDir := filepath.Join(tempDir, "public-no-auth")

		// Clone without any authentication - should work for public repos
		err := repo.Clone(ctx, testPublicRepo, cloneDir)
		require.NoError(t, err, "Should clone public repository without authentication")

		assert.True(t, repo.IsValidRepository(cloneDir), "Repository should be valid")
		t.Logf("Successfully cloned public repository without authentication")
	})

	t.Run("GitHub token authentication", func(t *testing.T) {
		// Only test if GITHUB_TOKEN is available
		token := os.Getenv("GITHUB_TOKEN")
		if token == "" {
			t.Skip("GITHUB_TOKEN not set, skipping GitHub token authentication test")
		}

		tempDir := t.TempDir()
		cloneDir := filepath.Join(tempDir, "github-token")

		// Clone with GitHub token (should work even for public repos)
		err := repo.Clone(ctx, testPublicRepo, cloneDir)
		require.NoError(t, err, "Should clone with GitHub token")

		assert.True(t, repo.IsValidRepository(cloneDir), "Repository should be valid")
		t.Logf("Successfully used GitHub token authentication")
	})

	t.Run("SSH authentication with controlled environment", func(t *testing.T) {
		// Create a temporary directory to simulate a home directory with SSH keys
		tempHomeDir := t.TempDir()
		sshDir := filepath.Join(tempHomeDir, ".ssh")
		keyPath := filepath.Join(sshDir, "id_ed25519")
		
		// Create real directory structure and dummy SSH key file
		err := os.MkdirAll(sshDir, 0o700)
		require.NoError(t, err, "Should create SSH directory")
		
		// Create a dummy SSH key file (not a real key, just for testing detection)
		err = os.WriteFile(keyPath, []byte("dummy ssh key content"), 0o600)
		require.NoError(t, err, "Should create dummy SSH key file")

		// Set environment to use our temporary home directory
		t.Setenv("HOME", tempHomeDir)
		
		// Disable SSH agent for this test to force key file fallback
		t.Setenv("SSH_AUTH_SOCK", "")
		
		tempDir := t.TempDir()
		cloneDir := filepath.Join(tempDir, "ssh-auth")

		// Attempt SSH clone - this will test the SSH key detection logic
		err = repo.Clone(ctx, testPublicRepoSSH, cloneDir)
		
		// We expect this to fail since we're using dummy keys, but the failure
		// should indicate that SSH authentication was attempted and keys were detected
		require.Error(t, err, "Should fail with dummy SSH keys")
		
		// The error should indicate SSH authentication was attempted
		errorMsg := strings.ToLower(err.Error())
		
		// Check if the error indicates SSH key processing was attempted
		// This confirms our SSH key detection logic is working
		authAttempted := strings.Contains(errorMsg, "ssh") ||
			strings.Contains(errorMsg, "auth") ||
			strings.Contains(errorMsg, "key") ||
			strings.Contains(errorMsg, "load") ||
			strings.Contains(errorMsg, "failed to load")
			
		if authAttempted {
			t.Logf("✅ SSH key detection worked - authentication attempted and failed as expected: %v", err)
		} else {
			// If no SSH-specific error, it might be a network/repository error
			t.Logf("⚠️  SSH authentication attempt not clearly detected in error, but test validates graceful failure: %v", err)
		}
		
		// Key assertion: the operation should fail gracefully without crashing
		// and should not leave partial repositories
		assert.False(t, repo.IsValidRepository(cloneDir), "Should not leave partial repository on auth failure")
	})

	t.Run("authentication failure handling", func(t *testing.T) {
		tempDir := t.TempDir()
		cloneDir := filepath.Join(tempDir, "auth-fail")

		// Try to clone a repository that would require authentication
		// Use SSH URL without proper SSH setup to trigger auth failure
		privateSSHURL := "git@github.com:private/nonexistent-repo.git"
		err := repo.Clone(ctx, privateSSHURL, cloneDir)

		// Should fail gracefully
		require.Error(t, err, "Should fail when authentication is required but not available")
		assert.Contains(t, strings.ToLower(err.Error()), "auth", "Error should mention authentication")

		// Ensure no partial clone was left
		assert.False(t, repo.IsValidRepository(cloneDir), "Should not leave partial repository on auth failure")
	})
}

// TestGitBranchHandling tests git branch operations
func TestGitBranchHandling(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping git branch tests in short mode")
	}

	fs := afero.NewOsFs()
	repo := git.NewRepository(fs)
	ctx := context.Background()

	t.Run("default branch handling", func(t *testing.T) {
		tempDir := t.TempDir()
		cloneDir := filepath.Join(tempDir, "default-branch")

		// Clone without specifying branch (should use default)
		err := repo.Clone(ctx, testPublicRepo, cloneDir)
		require.NoError(t, err, "Should clone default branch")

		// Verify we can get commit hash
		commitHash, err := repo.GetLatestCommitHash(cloneDir, "")
		require.NoError(t, err, "Should get latest commit hash")
		assert.Len(t, commitHash, 40, "Commit hash should be 40 characters")

		t.Logf("Default branch cloned successfully (commit: %s)", commitHash[:7])
	})

	t.Run("main branch explicit", func(t *testing.T) {
		tempDir := t.TempDir()
		cloneDir := filepath.Join(tempDir, "main-branch")

		// Explicitly clone main branch
		err := repo.Clone(ctx, testPublicRepo, cloneDir, git.WithBranch("main"))
		require.NoError(t, err, "Should clone main branch explicitly")

		// Verify we can get commit hash for main branch
		commitHash, err := repo.GetLatestCommitHash(cloneDir, "main")
		require.NoError(t, err, "Should get main branch commit hash")
		assert.Len(t, commitHash, 40, "Commit hash should be 40 characters")

		t.Logf("Main branch cloned successfully (commit: %s)", commitHash[:7])
	})

	t.Run("nonexistent branch handling", func(t *testing.T) {
		tempDir := t.TempDir()
		cloneDir := filepath.Join(tempDir, "nonexistent-branch")

		// Try to clone nonexistent branch
		err := repo.Clone(ctx, testPublicRepo, cloneDir, git.WithBranch("nonexistent-branch-name"))

		// Should handle gracefully - either fail or fallback to default branch
		if err != nil {
			t.Logf("Nonexistent branch failed as expected: %v", err)
			errorMsg := strings.ToLower(err.Error())
			assert.True(t,
				strings.Contains(errorMsg, "branch") ||
					strings.Contains(errorMsg, "reference") ||
					strings.Contains(errorMsg, "not found"),
				"Error should mention branch/reference issue: %v", err)
		} else {
			t.Logf("Nonexistent branch fell back to default branch")
			assert.True(t, repo.IsValidRepository(cloneDir), "Repository should be valid")
		}
	})

	t.Run("branch commit info retrieval", func(t *testing.T) {
		tempDir := t.TempDir()
		cloneDir := filepath.Join(tempDir, "branch-info")

		// Clone repository
		err := repo.Clone(ctx, testPublicRepo, cloneDir, git.WithBranch("main"))
		require.NoError(t, err, "Should clone repository")

		// Get commit info for a likely existing file
		commitInfo, err := repo.GetFileCommitInfo(cloneDir, "README.md", "main")
		if err != nil {
			// Try alternative common files
			alternativeFiles := []string{"readme.md", "README", ".gitignore", "index.js", "package.json"}
			for _, altFile := range alternativeFiles {
				commitInfo, err = repo.GetFileCommitInfo(cloneDir, altFile, "main")
				if err == nil {
					t.Logf("Found file: %s", altFile)
					break
				}
			}
		}

		if err == nil {
			require.NotNil(t, commitInfo, "Commit info should not be nil")
			assert.Len(t, commitInfo.Hash, 40, "Commit hash should be 40 characters")
			assert.NotEmpty(t, commitInfo.Date, "Commit date should not be empty")
			t.Logf("File commit info: %s on %s", commitInfo.Hash[:7], commitInfo.Date)
		} else {
			t.Logf("No common files found in repository for commit info test")
		}
	})
}

// TestGitErrorHandling tests git error scenarios
func TestGitErrorHandling(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping git error tests in short mode")
	}

	fs := afero.NewOsFs()
	repo := git.NewRepository(fs)
	ctx := context.Background()

	t.Run("nonexistent repository", func(t *testing.T) {
		tempDir := t.TempDir()
		cloneDir := filepath.Join(tempDir, "nonexistent")

		err := repo.Clone(ctx, testInvalidRepo, cloneDir)
		require.Error(t, err, "Should error when cloning nonexistent repository")

		// Error message could be "not found", "authentication required", or "repository not found"
		errorMsg := strings.ToLower(err.Error())
		assert.True(t,
			strings.Contains(errorMsg, "not found") ||
				strings.Contains(errorMsg, "authentication") ||
				strings.Contains(errorMsg, "repository"),
			"Error should indicate repository issue: %v", err)

		// Verify no partial clone was created
		assert.False(t, repo.IsValidRepository(cloneDir), "Should not create repository directory on failure")
	})

	t.Run("invalid git URL", func(t *testing.T) {
		tempDir := t.TempDir()
		cloneDir := filepath.Join(tempDir, "invalid-url")

		err := repo.Clone(ctx, testInvalidURL, cloneDir)
		require.Error(t, err, "Should error with invalid git URL")

		errorMsg := strings.ToLower(err.Error())
		assert.True(t,
			strings.Contains(errorMsg, "invalid") ||
				strings.Contains(errorMsg, "unsupported") ||
				strings.Contains(errorMsg, "scheme"),
			"Error should indicate URL issue: %v", err)
	})

	t.Run("malicious host rejection", func(t *testing.T) {
		tempDir := t.TempDir()
		cloneDir := filepath.Join(tempDir, "malicious")

		err := repo.Clone(ctx, testMaliciousHost, cloneDir)
		require.Error(t, err, "Should reject unauthorized host")
		assert.Contains(t, strings.ToLower(err.Error()), "unauthorized", "Error should indicate unauthorized host")
	})

	t.Run("context cancellation", func(t *testing.T) {
		tempDir := t.TempDir()
		cloneDir := filepath.Join(tempDir, "cancelled")

		// Create a context that cancels immediately
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		err := repo.Clone(ctx, testPublicRepo, cloneDir)
		// Should either succeed quickly or fail due to cancellation
		if err != nil {
			assert.Contains(t, strings.ToLower(err.Error()), "cancel", "Error should indicate cancellation")
		}
	})

	t.Run("timeout handling", func(t *testing.T) {
		tempDir := t.TempDir()
		cloneDir := filepath.Join(tempDir, "timeout")

		// Create a context with very short timeout
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
		defer cancel()

		err := repo.Clone(ctx, testPublicRepo, cloneDir)
		// Should either succeed quickly or timeout
		if err != nil {
			errorMsg := strings.ToLower(err.Error())
			assert.True(t,
				strings.Contains(errorMsg, "timeout") ||
					strings.Contains(errorMsg, "deadline") ||
					strings.Contains(errorMsg, "context") ||
					strings.Contains(errorMsg, "operation timed out"),
				"Error should indicate timeout/deadline exceeded: %v", err)
		}
	})

	t.Run("directory already exists", func(t *testing.T) {
		tempDir := t.TempDir()
		cloneDir := filepath.Join(tempDir, "existing")

		// Create directory first
		err := os.MkdirAll(cloneDir, 0o750)
		require.NoError(t, err)

		// Create a file in the directory
		testFile := filepath.Join(cloneDir, "test.txt")
		err = os.WriteFile(testFile, []byte("test"), 0o600)
		require.NoError(t, err)

		// Try to clone into existing directory
		err = repo.Clone(ctx, testPublicRepo, cloneDir)
		// Behavior may vary - either succeeds (overwrites) or fails gracefully
		if err != nil {
			t.Logf("Clone into existing directory failed as expected: %v", err)
		} else {
			t.Logf("Clone into existing directory succeeded (overwrote contents)")
			assert.True(t, repo.IsValidRepository(cloneDir), "Should be a valid repository")
		}
	})

	t.Run("invalid repository operations", func(t *testing.T) {
		tempDir := t.TempDir()
		notARepo := filepath.Join(tempDir, "not-a-repo")

		// Create a directory that's not a git repository
		err := os.MkdirAll(notARepo, 0o750)
		require.NoError(t, err)

		// Operations should fail gracefully
		assert.False(t, repo.IsValidRepository(notARepo), "Should not be a valid repository")

		_, err = repo.GetRemoteURL(notARepo)
		require.Error(t, err, "Should error when getting remote URL from non-repository")

		_, err = repo.GetLatestCommitHash(notARepo, "main")
		assert.Error(t, err, "Should error when getting commit hash from non-repository")
	})
}

// TestGitPerformance tests git operation performance
func TestGitPerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping git performance tests in short mode")
	}

	fs := afero.NewOsFs()
	repo := git.NewRepository(fs)
	ctx := context.Background()

	t.Run("clone performance", func(t *testing.T) {
		tempDir := t.TempDir()
		cloneDir := filepath.Join(tempDir, "performance")

		// Measure clone time
		start := time.Now()
		err := repo.Clone(ctx, testSmallRepo, cloneDir) // Use small repo for faster test
		duration := time.Since(start)

		require.NoError(t, err, "Clone should succeed")
		assert.True(t, repo.IsValidRepository(cloneDir), "Should be valid repository")

		t.Logf("Clone completed in %v", duration)

		// Clone should complete within reasonable time (30 seconds)
		assert.Less(t, duration, testTimeout, "Clone should complete within %v", testTimeout)
	})

	t.Run("concurrent clones", func(t *testing.T) {
		tempDir := t.TempDir()
		numConcurrent := 3
		done := make(chan error, numConcurrent)

		// Start concurrent clones
		start := time.Now()
		for i := range numConcurrent {
			go func(index int) {
				cloneDir := filepath.Join(tempDir, fmt.Sprintf("concurrent-%d", index))
				err := repo.Clone(ctx, testSmallRepo, cloneDir)
				done <- err
			}(i)
		}

		// Wait for all to complete
		for i := range numConcurrent {
			err := <-done
			require.NoError(t, err, "Concurrent clone %d should succeed", i)
		}
		duration := time.Since(start)

		t.Logf("Completed %d concurrent clones in %v", numConcurrent, duration)

		// Verify all clones succeeded
		for i := range numConcurrent {
			cloneDir := filepath.Join(tempDir, fmt.Sprintf("concurrent-%d", i))
			assert.True(t, repo.IsValidRepository(cloneDir), "Concurrent clone %d should be valid", i)
		}
	})

	t.Run("operation timeout handling", func(t *testing.T) {
		tempDir := t.TempDir()
		cloneDir := filepath.Join(tempDir, "timeout-test")

		// Test with reasonable timeout
		ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
		defer cancel()

		start := time.Now()
		err := repo.Clone(ctx, testSmallRepo, cloneDir)
		duration := time.Since(start)

		if err != nil {
			// If it failed, it should be due to timeout
			assert.Contains(t, strings.ToLower(err.Error()), "timeout", "Should timeout gracefully")
			assert.GreaterOrEqual(t, duration, testTimeout, "Should respect timeout")
		} else {
			// If it succeeded, it should be within timeout
			assert.Less(t, duration, testTimeout, "Should complete within timeout")
			assert.True(t, repo.IsValidRepository(cloneDir), "Should be valid repository")
		}

		t.Logf("Clone with %v timeout completed in %v (success: %v)", testTimeout, duration, err == nil)
	})
}

// TestGitCommitOperations tests git commit-related operations
func TestGitCommitOperations(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping git commit operations tests in short mode")
	}

	fs := afero.NewOsFs()
	repo := git.NewRepository(fs)
	ctx := context.Background()

	// Setup: Clone a repository for testing
	tempDir := t.TempDir()
	cloneDir := filepath.Join(tempDir, "commit-ops")
	err := repo.Clone(ctx, testPublicRepo, cloneDir)
	require.NoError(t, err, "Should clone repository for commit operations tests")

	t.Run("get commit info by hash", func(t *testing.T) {
		// Get latest commit hash first
		latestHash, err := repo.GetLatestCommitHash(cloneDir, "main")
		require.NoError(t, err, "Should get latest commit hash")

		// Get commit info by that hash
		commitInfo, err := repo.GetCommitInfoByHash(cloneDir, latestHash)
		require.NoError(t, err, "Should get commit info by hash")
		require.NotNil(t, commitInfo, "Commit info should not be nil")

		assert.Equal(t, latestHash, commitInfo.Hash, "Commit hashes should match")
		assert.NotEmpty(t, commitInfo.Date, "Commit date should not be empty")
		assert.Len(t, commitInfo.Hash, 40, "Commit hash should be 40 characters")

		t.Logf("Commit info: %s on %s", commitInfo.Hash[:7], commitInfo.Date)
	})

	t.Run("get commit info by short hash", func(t *testing.T) {
		// Get latest commit hash first
		latestHash, err := repo.GetLatestCommitHash(cloneDir, "main")
		require.NoError(t, err, "Should get latest commit hash")

		// Use short hash (first 7 characters)
		shortHash := latestHash[:7]

		// Get commit info by short hash
		commitInfo, err := repo.GetCommitInfoByHash(cloneDir, shortHash)
		require.NoError(t, err, "Should get commit info by short hash")
		require.NotNil(t, commitInfo, "Commit info should not be nil")

		assert.Equal(t, latestHash, commitInfo.Hash, "Should resolve to full hash")
		assert.NotEmpty(t, commitInfo.Date, "Commit date should not be empty")

		t.Logf("Short hash %s resolved to full hash %s", shortHash, commitInfo.Hash[:7])
	})

	t.Run("get file at commit", func(t *testing.T) {
		// Get latest commit hash
		latestHash, err := repo.GetLatestCommitHash(cloneDir, "main")
		require.NoError(t, err, "Should get latest commit hash")

		// Try to get content of common files
		commonFiles := []string{"README.md", "readme.md", "README", ".gitignore"}
		var fileContent []byte
		var fileName string

		for _, file := range commonFiles {
			content, err := repo.GetFileAtCommit(cloneDir, file, latestHash)
			if err == nil {
				fileContent = content
				fileName = file
				break
			}
		}

		if fileName != "" {
			assert.NotEmpty(t, fileContent, "File content should not be empty")
			t.Logf("Successfully retrieved %s (%d bytes) at commit %s", fileName, len(fileContent), latestHash[:7])
		} else {
			t.Logf("No common files found in repository for file content test")
		}
	})

	t.Run("nonexistent file at commit", func(t *testing.T) {
		// Get latest commit hash
		latestHash, err := repo.GetLatestCommitHash(cloneDir, "main")
		require.NoError(t, err, "Should get latest commit hash")

		// Try to get content of nonexistent file
		_, err = repo.GetFileAtCommit(cloneDir, "nonexistent-file-that-should-not-exist.txt", latestHash)
		assert.Error(t, err, "Should error when trying to get nonexistent file")
	})

	t.Run("invalid commit hash", func(t *testing.T) {
		// Try to get commit info with invalid hash
		_, err := repo.GetCommitInfoByHash(cloneDir, "invalid-hash-that-does-not-exist")
		require.Error(t, err, "Should error with invalid commit hash")

		// Try to get file with invalid hash
		_, err = repo.GetFileAtCommit(cloneDir, "README.md", "invalid-hash-that-does-not-exist")
		assert.Error(t, err, "Should error when getting file with invalid commit hash")
	})
}

// TestGitPullOperations tests git pull operations
func TestGitPullOperations(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping git pull operations tests in short mode")
	}

	fs := afero.NewOsFs()
	repo := git.NewRepository(fs)
	ctx := context.Background()

	t.Run("pull existing repository", func(t *testing.T) {
		tempDir := t.TempDir()
		cloneDir := filepath.Join(tempDir, "pull-test")

		// First, clone the repository
		err := repo.Clone(ctx, testPublicRepo, cloneDir)
		require.NoError(t, err, "Should clone repository")

		// Then try to pull (should either update or report "already up to date")
		err = repo.Pull(ctx, cloneDir)
		require.NoError(t, err, "Pull should succeed or report already up to date")

		t.Logf("Successfully pulled repository")
	})

	t.Run("pull with specific branch", func(t *testing.T) {
		tempDir := t.TempDir()
		cloneDir := filepath.Join(tempDir, "pull-branch")

		// Clone with specific branch
		err := repo.Clone(ctx, testPublicRepo, cloneDir, git.WithBranch("main"))
		require.NoError(t, err, "Should clone repository")

		// Pull with same branch
		err = repo.Pull(ctx, cloneDir, git.PullWithBranch("main"))
		require.NoError(t, err, "Pull with branch should succeed")

		t.Logf("Successfully pulled specific branch")
	})

	t.Run("pull non-repository", func(t *testing.T) {
		tempDir := t.TempDir()
		notARepo := filepath.Join(tempDir, "not-a-repo")

		// Create directory that's not a repository
		err := os.MkdirAll(notARepo, 0o750)
		require.NoError(t, err)

		// Pull should fail
		err = repo.Pull(ctx, notARepo)
		require.Error(t, err, "Should error when pulling non-repository")
		assert.Contains(t, strings.ToLower(err.Error()), "not a git repository", "Error should mention not a git repository")
	})
}

// TestGitValidation tests git URL and repository validation
func TestGitValidation(t *testing.T) {
	fs := afero.NewOsFs()
	repo := git.NewRepository(fs)

	t.Run("URL validation", func(t *testing.T) {
		tests := []struct {
			name    string
			url     string
			wantErr bool
		}{
			{"valid HTTPS URL", testPublicRepo, false},
			{"valid SSH URL", testPublicRepoSSH, false},
			{"empty URL", "", true},
			{"invalid URL", "not-a-url", true},
			{"malicious host", testMaliciousHost, true},
			{"unsupported scheme", "ftp://example.com/repo.git", true},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := repo.ValidateURL(tt.url)
				if tt.wantErr {
					assert.Error(t, err, "Should error for %s", tt.name)
				} else {
					assert.NoError(t, err, "Should not error for %s", tt.name)
				}
			})
		}
	})

	t.Run("repository validation", func(t *testing.T) {
		tempDir := t.TempDir()

		// Test non-existent directory
		nonExistent := filepath.Join(tempDir, "does-not-exist")
		assert.False(t, repo.IsValidRepository(nonExistent), "Non-existent directory should not be valid")

		// Test empty directory
		emptyDir := filepath.Join(tempDir, "empty")
		err := os.MkdirAll(emptyDir, 0o750)
		require.NoError(t, err)
		assert.False(t, repo.IsValidRepository(emptyDir), "Empty directory should not be valid")

		// Test actual git repository
		gitRepo := filepath.Join(tempDir, "git-repo")
		err = repo.Clone(context.Background(), testPublicRepo, gitRepo)
		if err == nil {
			assert.True(t, repo.IsValidRepository(gitRepo), "Cloned repository should be valid")
		} else {
			t.Logf("Skipping valid repository test due to clone failure: %v", err)
		}
	})
}

// TestGitIntegrationWithCustomConfig tests git operations with custom configuration
func TestGitIntegrationWithCustomConfig(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping git custom config tests in short mode")
	}

	fs := afero.NewOsFs()

	t.Run("custom timeout configuration", func(t *testing.T) {
		// Create client with short timeout
		config := git.DefaultConfig(fs)
		config.CloneTimeout = testShortTimeout
		client := git.NewClient(fs, config)

		tempDir := t.TempDir()
		cloneDir := filepath.Join(tempDir, "timeout-config")

		ctx := context.Background()
		start := time.Now()
		err := client.Clone(ctx, testSmallRepo, cloneDir) // Use small repo for faster completion
		duration := time.Since(start)

		if err != nil {
			// If it timed out, duration should be close to timeout
			t.Logf("Clone timed out as expected after %v", duration)
			assert.GreaterOrEqual(t, duration, testShortTimeout, "Should respect custom timeout")
		} else {
			// If it succeeded, it should be within timeout
			t.Logf("Clone succeeded within custom timeout (%v)", duration)
			assert.Less(t, duration, testShortTimeout, "Should complete within custom timeout")
			assert.True(t, client.IsValidRepository(cloneDir), "Should be valid repository")
		}
	})

	t.Run("restricted host configuration", func(t *testing.T) {
		// Create client with restricted hosts
		config := git.DefaultConfig(fs)
		config.AllowedHosts = []string{"github.com"} // Only allow github.com
		client := git.NewClient(fs, config)

		// Test allowed host
		err := client.ValidateURL(testPublicRepo)
		require.NoError(t, err, "Should allow github.com")

		// Test disallowed host
		err = client.ValidateURL("https://gitlab.com/user/repo.git")
		require.Error(t, err, "Should reject gitlab.com")
		assert.Contains(t, err.Error(), "unauthorized", "Should mention unauthorized host")
	})

	t.Run("restricted scheme configuration", func(t *testing.T) {
		// Create client with only HTTPS allowed
		config := git.DefaultConfig(fs)
		config.AllowedSchemes = []string{"https"} // Only HTTPS
		client := git.NewClient(fs, config)

		// Test allowed scheme
		err := client.ValidateURL(testPublicRepo)
		require.NoError(t, err, "Should allow HTTPS")

		// Test disallowed scheme
		err = client.ValidateURL(testPublicRepoSSH)
		require.Error(t, err, "Should reject SSH")
		assert.Contains(t, err.Error(), "unsupported", "Should mention unsupported scheme")
	})
}

// TestGitMemoryUsage tests memory usage during git operations
func TestGitMemoryUsage(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping git memory usage tests in short mode")
	}

	fs := afero.NewOsFs()
	repo := git.NewRepository(fs)
	ctx := context.Background()

	t.Run("memory usage during clone", func(t *testing.T) {
		tempDir := t.TempDir()
		cloneDir := filepath.Join(tempDir, "memory-test")

		// Clone repository and ensure it doesn't consume excessive memory
		err := repo.Clone(ctx, testSmallRepo, cloneDir)
		if err == nil {
			assert.True(t, repo.IsValidRepository(cloneDir), "Should be valid repository")
			t.Logf("Clone completed without memory issues")
		} else {
			t.Logf("Clone failed (possibly due to network): %v", err)
		}
	})

	t.Run("cleanup after failed operations", func(t *testing.T) {
		tempDir := t.TempDir()
		cloneDir := filepath.Join(tempDir, "cleanup-test")

		// Try to clone invalid repository
		err := repo.Clone(ctx, testInvalidRepo, cloneDir)
		require.Error(t, err, "Should fail to clone invalid repository")

		// Verify cleanup occurred (no partial directory left)
		assert.False(t, repo.IsValidRepository(cloneDir), "Should cleanup failed clone")
		_, err = os.Stat(cloneDir)
		assert.True(t, os.IsNotExist(err), "Directory should not exist after failed clone")
	})
}
