// Package integration provides examples of how to use the git test helpers
package integration

import (
	"testing"
	"time"

	"github.com/contextureai/contexture/internal/git"
)

// ExampleGitTestSuite shows how to use the GitTestSuite for comprehensive testing
func ExampleGitTestSuite() {
	// This is an example function showing how to use the GitTestSuite
	// It would typically be used within actual test functions

	t := &testing.T{} // In real tests, this comes from the test function parameter

	// Create a test suite
	suite := NewGitTestSuite(t)

	// Clone a repository easily
	repoPath := suite.ClonePublicRepo("my-test-repo")

	// Verify it's a valid repository
	suite.ExpectValidRepository(repoPath)

	// Test error scenarios
	suite.ExpectCloneError("https://invalid-url", "/tmp/test", "unsupported")

	// Test concurrent operations
	suite.TestConcurrentClones(3, testSmallRepo)

	// Measure performance
	duration := suite.MeasureCloneTime(testSmallRepo, "/tmp/perf-test")
	_ = duration // Use the duration for assertions
}

// ExampleGitTestSuiteWithCustomConfig shows how to test with custom git configuration
func ExampleGitTestSuiteWithCustomConfig() {
	t := &testing.T{} // In real tests, this comes from the test function parameter

	// Create custom configuration
	config := git.DefaultConfig()
	config.CloneTimeout = 10 * time.Second
	config.AllowedHosts = []string{"github.com"}

	// Create test suite with custom config
	suite := NewGitTestSuiteWithConfig(t, config)

	// Test with the custom configuration
	repoPath := suite.ClonePublicRepo("custom-config-test")
	suite.ExpectValidRepository(repoPath)
}

// ExampleProgressHandling shows how to test progress handling
func ExampleProgressHandling() {
	t := &testing.T{} // In real tests, this comes from the test function parameter

	suite := NewGitTestSuite(t)

	// Clone with progress tracking
	cloneDir := suite.CloneDir("progress-test")
	repo := suite.repo

	err := repo.Clone(suite.ctx, testSmallRepo, cloneDir, suite.WithProgress())
	if err != nil {
		return // Handle error
	}

	// Verify progress was called
	suite.ExpectProgressCalled()
}

// ExampleTestRepositoryLoop shows how to test against multiple repositories
func ExampleTestRepositoryLoop() {
	t := &testing.T{} // In real tests, this comes from the test function parameter

	suite := NewGitTestSuite(t)

	// Test against all available test repositories
	testRepos := GetTestRepositories()
	for _, testRepo := range testRepos {
		if !testRepo.IsPublic {
			continue // Skip private repos in this example
		}

		// Clone each repository
		repoPath := suite.CloneDir(testRepo.Name)
		err := suite.repo.Clone(suite.ctx, testRepo.URL, repoPath)
		if err != nil {
			continue // Skip if clone fails
		}

		// Verify expected files exist
		for _, expectedFile := range testRepo.HasFiles {
			if !suite.TestFileExists(repoPath, expectedFile) {
				t.Errorf("Expected file %s not found in %s", expectedFile, testRepo.Name)
			}
		}
	}
}

// ExampleErrorScenarios shows how to test error scenarios systematically
func ExampleErrorScenarios() {
	t := &testing.T{} // In real tests, this comes from the test function parameter

	suite := NewGitTestSuite(t)

	// Test all error scenarios
	errorScenarios := GetErrorScenarios()
	for _, scenario := range errorScenarios {
		cloneDir := suite.CloneDir(scenario.Name)

		// Test the error scenario
		err := suite.repo.Clone(suite.ctx, scenario.URL, cloneDir)
		if err == nil {
			t.Errorf("Expected error for scenario %s but got none", scenario.Name)
			continue
		}

		// Verify error contains expected message
		suite.ExpectCloneError(scenario.URL, cloneDir, scenario.ExpectedError)

		// Verify cleanup if expected
		if scenario.ShouldCleanup {
			suite.ExpectInvalidRepository(cloneDir)
		}
	}
}

// ExampleNetworkSkipping shows how to skip tests when network is unavailable
func ExampleNetworkSkipping() {
	t := &testing.T{} // In real tests, this comes from the test function parameter

	suite := NewGitTestSuite(t)

	// Skip if network is not available
	suite.SkipIfNoNetwork()

	// Skip if SSH is not available
	suite.SkipIfNoSSH()

	// Skip if GitHub token is not available
	suite.SkipIfNoGitHubToken()

	// Proceed with network-dependent tests
	repoPath := suite.ClonePublicRepo("network-test")
	suite.ExpectValidRepository(repoPath)
}

// ExampleFileOperations shows how to test file operations in cloned repositories
func ExampleFileOperations() {
	t := &testing.T{} // In real tests, this comes from the test function parameter

	suite := NewGitTestSuite(t)

	// Clone repository
	repoPath := suite.ClonePublicRepo("file-ops-test")

	// Find a common file for testing
	commonFile, err := suite.GetCommonFile(repoPath)
	if err != nil {
		t.Skip("No common files found for testing")
	}

	// Test file operations
	if !suite.TestFileExists(repoPath, commonFile) {
		t.Errorf("Common file %s should exist", commonFile)
	}

	// Test with repository operations
	commitInfo, err := suite.repo.GetFileCommitInfo(repoPath, commonFile, "main")
	if err != nil {
		t.Errorf("Failed to get commit info for %s: %v", commonFile, err)
		return
	}

	// Verify commit hash format
	suite.ExpectCommitHash(commitInfo.Hash)
}

// ExampleCustomDirectories shows how to test with existing directories and non-repositories
func ExampleCustomDirectories() {
	t := &testing.T{} // In real tests, this comes from the test function parameter

	suite := NewGitTestSuite(t)

	// Create a non-repository directory
	nonRepo := suite.CreateNonRepository("not-a-repo")
	suite.ExpectInvalidRepository(nonRepo)

	// Create an existing directory
	existingDir := suite.CreateExistingDirectory("existing-content")

	// Try to clone into existing directory
	err := suite.repo.Clone(suite.ctx, testSmallRepo, existingDir)
	if err != nil {
		suite.LogGitOperation("clone-into-existing", testSmallRepo, existingDir, 0, err)
	} else {
		// If it succeeded, it should now be a valid repository
		suite.ExpectValidRepository(existingDir)
	}
}
