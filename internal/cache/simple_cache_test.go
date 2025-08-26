package cache

import (
	"context"
	"fmt"
	"testing"

	"github.com/contextureai/contexture/internal/git"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const (
	testMainBranch = "main"
)

func TestSimpleCache_generateCacheKey(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		repoURL  string
		gitRef   string
		expected string
	}{
		{
			name:     "GitHub HTTPS URL",
			repoURL:  "https://github.com/contextureai/rules.git",
			gitRef:   testMainBranch,
			expected: "github.com_contextureai_rules-main",
		},
		{
			name:     "GitHub SSH URL",
			repoURL:  "git@github.com:user/repo.git",
			gitRef:   "develop",
			expected: "github.com_user_repo-develop",
		},
		{
			name:     "GitLab with tag",
			repoURL:  "https://gitlab.com/company/project.git",
			gitRef:   "v1.2.3",
			expected: "gitlab.com_company_project-v1.2.3",
		},
		{
			name:     "GitHub HTTPS without .git suffix",
			repoURL:  "https://github.com/user/repo",
			gitRef:   "feature-branch",
			expected: "github.com_user_repo-feature-branch",
		},
		{
			name:     "Bitbucket SSH URL",
			repoURL:  "git@bitbucket.org:team/project.git",
			gitRef:   "release/1.0",
			expected: "bitbucket.org_team_project-release/1.0",
		},
		{
			name:     "Complex URL fallback",
			repoURL:  "invalid://complex@url:with:colons",
			gitRef:   "main",
			expected: "invalid___complex_url_with_colons-main",
		},
		{
			name:     "GitLab SSH with nested path",
			repoURL:  "git@gitlab.company.com:group/subgroup/project.git",
			gitRef:   "develop",
			expected: "gitlab.company.com_group_subgroup_project-develop",
		},
	}

	cache := &SimpleCache{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cache.generateCacheKey(tt.repoURL, tt.gitRef)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSimpleCache_isValidRepository(t *testing.T) {
	t.Parallel()
	fs := afero.NewMemMapFs()

	tests := []struct {
		name     string
		setup    func() string // Returns path to test
		expected bool
	}{
		{
			name: "Non-existent directory",
			setup: func() string {
				return "/tmp/non-existent-dir"
			},
			expected: false,
		},
		{
			name: "Directory exists with .git subdirectory",
			setup: func() string {
				path := "/tmp/valid-repo"
				_ = fs.MkdirAll(path+"/.git", 0o755)
				return path
			},
			expected: true,
		},
		{
			name: "Directory exists without .git subdirectory",
			setup: func() string {
				path := "/tmp/invalid-repo"
				_ = fs.MkdirAll(path, 0o755)
				return path
			},
			expected: false,
		},
	}

	cache := &SimpleCache{fs: fs}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := tt.setup()
			result := cache.isValidRepository(path)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSimpleCache_GetRepository(t *testing.T) {
	t.Parallel()
	fs := afero.NewMemMapFs()
	mockRepo := git.NewMockRepository(t)
	cache := NewSimpleCache(fs, mockRepo)

	t.Run("clone repository when not cached", func(t *testing.T) {
		repoURL := "https://github.com/test/repo.git"
		gitRef := testMainBranch
		expectedPath := "/tmp/contexture/github.com_test_repo-main"

		// Mock successful clone
		mockRepo.On("Clone", mock.Anything, repoURL, expectedPath, mock.Anything).Return(nil)

		path, err := cache.GetRepository(context.Background(), repoURL, gitRef)

		require.NoError(t, err)
		assert.Equal(t, expectedPath, path)
		mockRepo.AssertExpectations(t)
	})

	t.Run("use cached repository when available", func(t *testing.T) {
		repoURL := "https://github.com/test/cached.git"
		gitRef := testMainBranch
		cachedPath := "/tmp/contexture/github.com_test_cached-main"

		// Set up cached repository
		_ = fs.MkdirAll(cachedPath+"/.git", 0o755)

		path, err := cache.GetRepository(context.Background(), repoURL, gitRef)

		require.NoError(t, err)
		assert.Equal(t, cachedPath, path)
		// Should not call Clone since repository is cached
	})
}

func TestSimpleCache_GetRepositoryWithUpdate(t *testing.T) {
	t.Parallel()
	fs := afero.NewMemMapFs()
	mockRepo := git.NewMockRepository(t)
	cache := NewSimpleCache(fs, mockRepo)

	t.Run("clone repository when not cached", func(t *testing.T) {
		repoURL := "https://github.com/test/update-repo.git"
		gitRef := "develop"
		expectedPath := "/tmp/contexture/github.com_test_update-repo-develop"

		// Mock successful clone
		mockRepo.On("Clone", mock.Anything, repoURL, expectedPath, mock.Anything).Return(nil)

		path, err := cache.GetRepositoryWithUpdate(context.Background(), repoURL, gitRef)

		require.NoError(t, err)
		assert.Equal(t, expectedPath, path)
		mockRepo.AssertExpectations(t)
	})

	t.Run("pull updates when repository is cached", func(t *testing.T) {
		repoURL := "https://github.com/test/update-cached.git"
		gitRef := testMainBranch
		cachedPath := "/tmp/contexture/github.com_test_update-cached-main"

		// Set up cached repository
		_ = fs.MkdirAll(cachedPath+"/.git", 0o755)

		// Mock successful pull
		mockRepo.On("Pull", mock.Anything, cachedPath, mock.Anything).Return(nil)

		path, err := cache.GetRepositoryWithUpdate(context.Background(), repoURL, gitRef)

		require.NoError(t, err)
		assert.Equal(t, cachedPath, path)
		mockRepo.AssertExpectations(t)
	})

	t.Run("continue with cached version when pull fails", func(t *testing.T) {
		repoURL := "https://github.com/test/pull-fail.git"
		gitRef := testMainBranch
		cachedPath := "/tmp/contexture/github.com_test_pull-fail-main"

		// Set up cached repository
		_ = fs.MkdirAll(cachedPath+"/.git", 0o755)

		// Mock failed pull
		mockRepo.On("Pull", mock.Anything, cachedPath, mock.Anything).
			Return(fmt.Errorf("network error"))

		path, err := cache.GetRepositoryWithUpdate(context.Background(), repoURL, gitRef)

		// Should succeed despite pull failure
		require.NoError(t, err)
		assert.Equal(t, cachedPath, path)
		mockRepo.AssertExpectations(t)
	})
}
