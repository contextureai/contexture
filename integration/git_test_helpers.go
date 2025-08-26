// Package integration provides test helpers for git integration tests
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

// Test constants shared across git integration tests
const (
	// Public repositories for testing (these should always be available)
	testPublicRepo    = "https://github.com/contextureai/rules.git"
	testPublicRepoSSH = "git@github.com:contextureai/rules.git"
	testSmallRepo     = "https://github.com/octocat/Hello-World.git"
	testInvalidRepo   = "https://github.com/nonexistent/repository-that-does-not-exist.git"
	testInvalidURL    = "not-a-valid-url"
	testMaliciousHost = "https://malicious.example.com/repo.git"
	testTimeout       = 30 * time.Second
	testShortTimeout  = 5 * time.Second
)

// TestProgressHandler implements git.ProgressHandler for testing
type TestProgressHandler struct {
	Messages  []string
	Completed bool
	Failed    bool
	LastError error
}

// OnProgress records progress messages
func (h *TestProgressHandler) OnProgress(message string, _, _ int64) {
	h.Messages = append(h.Messages, message)
}

// OnComplete marks the operation as completed
func (h *TestProgressHandler) OnComplete() {
	h.Completed = true
}

// OnError records the error and marks as failed
func (h *TestProgressHandler) OnError(err error) {
	h.Failed = true
	h.LastError = err
}

// Reset clears the handler state for reuse
func (h *TestProgressHandler) Reset() {
	h.Messages = nil
	h.Completed = false
	h.Failed = false
	h.LastError = nil
}

// GitTestSuite provides common git testing utilities
type GitTestSuite struct {
	t        *testing.T
	fs       afero.Fs
	repo     git.Repository
	ctx      context.Context
	tempDir  string
	progress *TestProgressHandler
}

// NewGitTestSuite creates a new test suite for git operations
func NewGitTestSuite(t *testing.T) *GitTestSuite {
	t.Helper()

	fs := afero.NewOsFs()
	repo := git.NewRepository(fs)
	ctx := context.Background()
	tempDir := t.TempDir()
	progress := &TestProgressHandler{}

	return &GitTestSuite{
		t:        t,
		fs:       fs,
		repo:     repo,
		ctx:      ctx,
		tempDir:  tempDir,
		progress: progress,
	}
}

// NewGitTestSuiteWithConfig creates a test suite with custom git configuration
func NewGitTestSuiteWithConfig(t *testing.T, config git.Config) *GitTestSuite {
	t.Helper()

	fs := afero.NewOsFs()
	client := git.NewClient(fs, config)
	ctx := context.Background()
	tempDir := t.TempDir()
	progress := &TestProgressHandler{}

	return &GitTestSuite{
		t:        t,
		fs:       fs,
		repo:     client,
		ctx:      ctx,
		tempDir:  tempDir,
		progress: progress,
	}
}

// CloneDir returns a unique clone directory path for the test
func (s *GitTestSuite) CloneDir(name string) string {
	s.t.Helper()
	return filepath.Join(s.tempDir, name)
}

// ClonePublicRepo clones the test public repository and returns the local path
func (s *GitTestSuite) ClonePublicRepo(name string) string {
	s.t.Helper()
	cloneDir := s.CloneDir(name)

	err := s.repo.Clone(s.ctx, testPublicRepo, cloneDir)
	require.NoError(s.t, err, "Should clone public repository")

	return cloneDir
}

// ClonePublicRepoWithBranch clones the test public repository with a specific branch
func (s *GitTestSuite) ClonePublicRepoWithBranch(name, branch string) string {
	s.t.Helper()
	cloneDir := s.CloneDir(name)

	err := s.repo.Clone(s.ctx, testPublicRepo, cloneDir, git.WithBranch(branch))
	require.NoError(s.t, err, "Should clone public repository with branch %s", branch)

	return cloneDir
}

// CloneSmallRepo clones the small test repository for faster tests
func (s *GitTestSuite) CloneSmallRepo(name string) string {
	s.t.Helper()
	cloneDir := s.CloneDir(name)

	err := s.repo.Clone(s.ctx, testSmallRepo, cloneDir)
	require.NoError(s.t, err, "Should clone small repository")

	return cloneDir
}

// ExpectValidRepository asserts that the path contains a valid git repository
func (s *GitTestSuite) ExpectValidRepository(path string) {
	s.t.Helper()
	assert.True(s.t, s.repo.IsValidRepository(path), "Path %s should be a valid git repository", path)
}

// ExpectInvalidRepository asserts that the path does not contain a valid git repository
func (s *GitTestSuite) ExpectInvalidRepository(path string) {
	s.t.Helper()
	assert.False(s.t, s.repo.IsValidRepository(path), "Path %s should not be a valid git repository", path)
}

// ExpectCommitHash verifies that a commit hash has the expected format
func (s *GitTestSuite) ExpectCommitHash(hash string) {
	s.t.Helper()
	assert.Len(s.t, hash, 40, "Commit hash should be 40 characters")
	assert.Regexp(s.t, "^[a-f0-9]+$", hash, "Commit hash should be hexadecimal")
}

// ExpectCloneError attempts to clone and expects it to fail with specific error content
func (s *GitTestSuite) ExpectCloneError(repoURL, cloneDir string, expectedError string) {
	s.t.Helper()

	err := s.repo.Clone(s.ctx, repoURL, cloneDir)
	require.Error(s.t, err, "Should error when cloning %s", repoURL)
	assert.Contains(s.t, strings.ToLower(err.Error()), strings.ToLower(expectedError),
		"Error should contain %s", expectedError)
}

// MeasureCloneTime measures how long a clone operation takes
func (s *GitTestSuite) MeasureCloneTime(repoURL, cloneDir string, opts ...git.CloneOption) time.Duration {
	s.t.Helper()

	start := time.Now()
	err := s.repo.Clone(s.ctx, repoURL, cloneDir, opts...)
	duration := time.Since(start)

	require.NoError(s.t, err, "Clone should succeed")
	s.t.Logf("Clone of %s completed in %v", repoURL, duration)

	return duration
}

// TestConcurrentClones tests cloning multiple repositories concurrently
func (s *GitTestSuite) TestConcurrentClones(numClones int, repoURL string) {
	s.t.Helper()

	done := make(chan error, numClones)

	start := time.Now()
	for i := range numClones {
		go func(index int) {
			cloneDir := s.CloneDir(fmt.Sprintf("concurrent-%d", index))
			err := s.repo.Clone(s.ctx, repoURL, cloneDir)
			done <- err
		}(i)
	}

	// Wait for all to complete
	for i := range numClones {
		err := <-done
		require.NoError(s.t, err, "Concurrent clone %d should succeed", i)
	}
	duration := time.Since(start)

	s.t.Logf("Completed %d concurrent clones in %v", numClones, duration)

	// Verify all clones succeeded
	for i := range numClones {
		cloneDir := s.CloneDir(fmt.Sprintf("concurrent-%d", i))
		s.ExpectValidRepository(cloneDir)
	}
}

// CreateNonRepository creates a directory that is not a git repository
func (s *GitTestSuite) CreateNonRepository(name string) string {
	s.t.Helper()

	dir := s.CloneDir(name)
	err := os.MkdirAll(dir, 0o750)
	require.NoError(s.t, err, "Should create directory")

	// Add a non-git file to make it clear it's not a repository
	testFile := filepath.Join(dir, "not-a-repo.txt")
	err = os.WriteFile(testFile, []byte("This is not a git repository"), 0o600)
	require.NoError(s.t, err, "Should create test file")

	return dir
}

// CreateExistingDirectory creates a directory with some content
func (s *GitTestSuite) CreateExistingDirectory(name string) string {
	s.t.Helper()

	dir := s.CloneDir(name)
	err := os.MkdirAll(dir, 0o750)
	require.NoError(s.t, err, "Should create directory")

	// Add some files to the directory
	testFile := filepath.Join(dir, "existing-file.txt")
	err = os.WriteFile(testFile, []byte("existing content"), 0o600)
	require.NoError(s.t, err, "Should create existing file")

	return dir
}

// WithProgress adds progress handler to clone options
func (s *GitTestSuite) WithProgress() git.CloneOption {
	s.progress.Reset()
	return git.WithProgress(s.progress)
}

// ExpectProgressCalled asserts that progress callbacks were invoked
func (s *GitTestSuite) ExpectProgressCalled() {
	s.t.Helper()
	assert.NotEmpty(s.t, s.progress.Messages, "Progress should have been called")
}

// ExpectProgressCompleted asserts that progress completion was called
func (s *GitTestSuite) ExpectProgressCompleted() {
	s.t.Helper()
	assert.True(s.t, s.progress.Completed, "Progress should have completed")
}

// ExpectProgressFailed asserts that progress error was called
func (s *GitTestSuite) ExpectProgressFailed() {
	s.t.Helper()
	assert.True(s.t, s.progress.Failed, "Progress should have failed")
	assert.Error(s.t, s.progress.LastError, "Progress should have recorded error")
}

// SkipIfNoNetwork skips the test if network access is not available
func (s *GitTestSuite) SkipIfNoNetwork() {
	s.t.Helper()

	// Try a quick connectivity test
	ctx, cancel := context.WithTimeout(s.ctx, 5*time.Second)
	defer cancel()

	testDir := s.CloneDir("network-test")
	err := s.repo.Clone(ctx, testSmallRepo, testDir)
	if err != nil {
		s.t.Skipf("Skipping test due to network connectivity issues: %v", err)
	}

	// Clean up test clone
	_ = os.RemoveAll(testDir)
}

// SkipIfNoSSH skips the test if SSH agent is not available
func (s *GitTestSuite) SkipIfNoSSH() {
	s.t.Helper()

	if os.Getenv("SSH_AUTH_SOCK") == "" {
		s.t.Skip("SSH_AUTH_SOCK not set, skipping SSH test")
	}
}

// SkipIfNoGitHubToken skips the test if GITHUB_TOKEN is not available
func (s *GitTestSuite) SkipIfNoGitHubToken() {
	s.t.Helper()

	if os.Getenv("GITHUB_TOKEN") == "" {
		s.t.Skip("GITHUB_TOKEN not set, skipping GitHub token test")
	}
}

// TestFileExists checks if a file exists in the cloned repository
func (s *GitTestSuite) TestFileExists(repoPath, fileName string) bool {
	s.t.Helper()

	filePath := filepath.Join(repoPath, fileName)
	_, err := os.Stat(filePath)
	return err == nil
}

// GetCommonFile tries to find a common file in the repository for testing
func (s *GitTestSuite) GetCommonFile(repoPath string) (string, error) {
	s.t.Helper()

	commonFiles := []string{
		"README.md", "readme.md", "README", "readme",
		".gitignore", "LICENSE", "license", "package.json",
		"go.mod", "Cargo.toml", "pyproject.toml", "index.js",
	}

	for _, file := range commonFiles {
		if s.TestFileExists(repoPath, file) {
			return file, nil
		}
	}

	return "", fmt.Errorf("no common files found in repository")
}

// TestRepository provides configuration for testing different repository scenarios
type TestRepository struct {
	Name        string
	URL         string
	SSHURL      string
	Branch      string
	IsSmall     bool
	IsPublic    bool
	HasFiles    []string
	Description string
}

// GetTestRepositories returns a list of test repositories for comprehensive testing
func GetTestRepositories() []TestRepository {
	return []TestRepository{
		{
			Name:        "contexture-rules",
			URL:         testPublicRepo,
			SSHURL:      testPublicRepoSSH,
			Branch:      "main",
			IsSmall:     false,
			IsPublic:    true,
			HasFiles:    []string{"README.md"},
			Description: "Main Contexture rules repository",
		},
		{
			Name:        "hello-world",
			URL:         testSmallRepo,
			SSHURL:      "git@github.com:octocat/Hello-World.git",
			Branch:      "master",
			IsSmall:     true,
			IsPublic:    true,
			HasFiles:    []string{"README"},
			Description: "Small Hello World repository for fast testing",
		},
	}
}

// TestErrorScenario represents an error scenario to test
type TestErrorScenario struct {
	Name          string
	URL           string
	ExpectedError string
	Description   string
	ShouldCleanup bool
}

// GetErrorScenarios returns common error scenarios for testing
func GetErrorScenarios() []TestErrorScenario {
	return []TestErrorScenario{
		{
			Name:          "nonexistent-repository",
			URL:           testInvalidRepo,
			ExpectedError: "not found",
			Description:   "Repository that does not exist",
			ShouldCleanup: true,
		},
		{
			Name:          "invalid-url",
			URL:           testInvalidURL,
			ExpectedError: "invalid",
			Description:   "Malformed URL",
			ShouldCleanup: true,
		},
		{
			Name:          "unauthorized-host",
			URL:           testMaliciousHost,
			ExpectedError: "unauthorized",
			Description:   "Host not in allowed list",
			ShouldCleanup: true,
		},
	}
}

// LogGitOperation logs details about a git operation for debugging
func (s *GitTestSuite) LogGitOperation(operation, repoURL, path string, duration time.Duration, err error) {
	s.t.Helper()

	if err != nil {
		s.t.Logf("%s failed: %s -> %s (took %v): %v", operation, repoURL, path, duration, err)
	} else {
		s.t.Logf("%s succeeded: %s -> %s (took %v)", operation, repoURL, path, duration)
	}
}

// Cleanup performs cleanup operations for the test suite
func (s *GitTestSuite) Cleanup() {
	s.t.Helper()

	// The tempDir is automatically cleaned up by t.TempDir()
	// This method is provided for any additional cleanup that might be needed
	s.progress.Reset()
}
