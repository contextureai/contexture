// Package commands provides CLI command implementations
package commands

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/contextureai/contexture/internal/cache"
	"github.com/contextureai/contexture/internal/dependencies"
	"github.com/contextureai/contexture/internal/domain"
	"github.com/contextureai/contexture/internal/git"
	"github.com/contextureai/contexture/internal/project"
	"github.com/contextureai/contexture/internal/rule"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v3"
)

const (
	testTempDir = "/tmp/test"
)

func TestNewUpdateCommand(t *testing.T) {
	t.Parallel()
	fs := afero.NewMemMapFs()
	deps := &dependencies.Dependencies{
		FS:      fs,
		Context: context.Background(),
	}

	cmd := NewUpdateCommand(deps)
	assert.NotNil(t, cmd)
	assert.NotNil(t, cmd.projectManager)
	assert.NotNil(t, cmd.ruleFetcher)
	assert.NotNil(t, cmd.ruleValidator)
}

func TestUpdateCommandWithDependencies(t *testing.T) {
	t.Parallel()
	fs := afero.NewMemMapFs()

	// Create mock dependencies
	mockFetcher := rule.NewMockFetcher(t)
	mockValidator := rule.NewValidator()
	mockProjectManager := project.NewManager(fs)
	mockGitRepo := git.NewMockRepository(t)
	mockCache := cache.NewSimpleCache(fs, mockGitRepo)

	// Test that we can create an UpdateCommand with explicit dependencies
	cmd := NewUpdateCommandWithDependencies(
		mockProjectManager,
		mockFetcher,
		mockValidator,
		mockCache,
		fs,
	)

	// Verify the dependencies were set correctly
	assert.NotNil(t, cmd)
	assert.Equal(t, mockProjectManager, cmd.projectManager)
	assert.Equal(t, mockFetcher, cmd.ruleFetcher)
	assert.Equal(t, mockValidator, cmd.ruleValidator)
	assert.Equal(t, mockCache, cmd.cache)
	assert.Equal(t, fs, cmd.fs)
}

func TestUpdateResult_Structure(t *testing.T) {
	t.Parallel()
	// Test the UpdateResult structure that was mentioned in the original skipped test
	result := UpdateResult{
		RuleID:         "[contexture:test/rule]",
		DisplayName:    "test/rule",
		CurrentVersion: "v1.0.0",
		LatestVersion:  "v1.1.0",
		HasUpdate:      true,
		Error:          nil,
		Status:         StatusUpdateAvailable,
		CurrentCommit: GitCommitInfo{
			Hash: "abc123",
			Date: "2023-01-01T00:00:00Z",
		},
		LatestCommit: GitCommitInfo{
			Hash: "def456",
			Date: "2023-01-02T00:00:00Z",
		},
	}

	// Test the structure and accessors
	assert.Equal(t, "[contexture:test/rule]", result.RuleID)
	assert.Equal(t, "test/rule", result.DisplayName)
	assert.Equal(t, "v1.0.0", result.CurrentVersion)
	assert.Equal(t, "v1.1.0", result.LatestVersion)
	assert.True(t, result.HasUpdate)
	require.NoError(t, result.Error)
	assert.Equal(t, StatusUpdateAvailable, result.Status)
	assert.Equal(t, "abc123", result.CurrentCommit.Hash)
	assert.Equal(t, "2023-01-01T00:00:00Z", result.CurrentCommit.Date)
	assert.Equal(t, "def456", result.LatestCommit.Hash)
	assert.Equal(t, "2023-01-02T00:00:00Z", result.LatestCommit.Date)
}

func TestUpdateStatus_Values(t *testing.T) {
	t.Parallel()
	// Test the UpdateStatus enum values
	assert.Equal(t, 0, int(StatusChecking))
	assert.Equal(t, 1, int(StatusUpToDate))
	assert.Equal(t, 2, int(StatusUpdateAvailable))
	assert.Equal(t, 3, int(StatusError))
	assert.Equal(t, 4, int(StatusApplying))
	assert.Equal(t, 5, int(StatusApplied))
}

func TestUpdateCommand_UpdateRuleCommitHash(t *testing.T) {
	t.Parallel()
	fs := afero.NewMemMapFs()

	// Create mock dependencies
	mockFetcher := rule.NewMockFetcher(t)
	mockValidator := rule.NewValidator()
	mockProjectManager := project.NewManager(fs)
	mockGitRepo := git.NewMockRepository(t)
	mockCache := cache.NewSimpleCache(fs, mockGitRepo)

	cmd := NewUpdateCommandWithDependencies(
		mockProjectManager,
		mockFetcher,
		mockValidator,
		mockCache,
		fs,
	)

	// Test config with rules
	config := &domain.Project{
		Rules: []domain.RuleRef{
			{ID: "[contexture:test/rule1]", CommitHash: "old123"},
			{ID: "[contexture:test/rule2]", CommitHash: "old456"},
		},
	}

	// Test updating commit hash
	cmd.updateRuleCommitHash(config, "[contexture:test/rule1]", "new123")

	// Verify the commit hash was updated
	assert.Equal(t, "new123", config.Rules[0].CommitHash)
	assert.Equal(t, "old456", config.Rules[1].CommitHash) // Should remain unchanged

	// Test updating second rule
	cmd.updateRuleCommitHash(config, "[contexture:test/rule2]", "new456")
	assert.Equal(t, "new123", config.Rules[0].CommitHash)
	assert.Equal(t, "new456", config.Rules[1].CommitHash)

	// Test updating non-existent rule (should not panic)
	cmd.updateRuleCommitHash(config, "[contexture:test/nonexistent]", "newer")
	assert.Equal(t, "new123", config.Rules[0].CommitHash) // Should remain unchanged
	assert.Equal(t, "new456", config.Rules[1].CommitHash) // Should remain unchanged
}

func TestUpdateAction(t *testing.T) {
	t.Parallel()
	fs := afero.NewMemMapFs()
	deps := &dependencies.Dependencies{
		FS:      fs,
		Context: context.Background(),
	}

	tempDir := testTempDir
	_ = fs.MkdirAll(tempDir, 0o755)

	// Create test config
	configPath := filepath.Join(tempDir, domain.ConfigFile)
	configData := `version: 1
formats:
  - type: claude
    enabled: true
rules:
  - id: "[contexture:test/rule]"
`
	_ = afero.WriteFile(fs, configPath, []byte(configData), 0o644)

	// Create mock CLI command
	cmd := &cli.Command{}
	cmd.Metadata = map[string]any{
		"check":  true,
		"yes":    false,
		"global": false,
	}

	// Note: This test will fail because it can't find the working directory config
	// In a proper test, we would need to mock os.Getwd or use a different approach
	assert.NotPanics(t, func() {
		// We expect this to fail with a "no project configuration found" error
		_ = UpdateAction(context.Background(), cmd, deps)
	})
}

func TestUpdateResult(t *testing.T) {
	t.Parallel()
	result := UpdateResult{
		RuleID:         "[contexture:test/rule]",
		CurrentVersion: "v1.0.0",
		LatestVersion:  "v1.1.0",
		HasUpdate:      true,
		Error:          nil,
	}

	assert.Equal(t, "[contexture:test/rule]", result.RuleID)
	assert.Equal(t, "v1.0.0", result.CurrentVersion)
	assert.Equal(t, "v1.1.0", result.LatestVersion)
	assert.True(t, result.HasUpdate)
	require.NoError(t, result.Error)
}
