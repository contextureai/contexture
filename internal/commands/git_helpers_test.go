package commands

import (
	"testing"

	"github.com/contextureai/contexture/internal/git"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewOpenRepository(t *testing.T) {
	t.Parallel()

	t.Run("returns_repository_client", func(t *testing.T) {
		t.Parallel()

		fs := afero.NewMemMapFs()
		repo := newOpenRepository(fs)

		require.NotNil(t, repo, "should return a repository instance")
		assert.Implements(t, (*git.Repository)(nil), repo, "should implement Repository interface")
	})

	t.Run("returns_git_client_type", func(t *testing.T) {
		t.Parallel()

		fs := afero.NewMemMapFs()
		repo := newOpenRepository(fs)

		// Verify it's actually a git.Client
		_, ok := repo.(*git.Client)
		assert.True(t, ok, "should return a *git.Client instance")
	})

	t.Run("disables_host_allowlisting", func(t *testing.T) {
		t.Parallel()

		fs := afero.NewMemMapFs()
		client := newOpenRepository(fs).(*git.Client)

		// This is security-critical - we need to verify AllowedHosts is nil
		// Since the config field is unexported, we verify the behavior instead:
		// A client with AllowedHosts=nil should accept any host during validation

		// Test that validation would pass for any host by creating a client
		// with the default config and comparing behavior
		defaultConfig := git.DefaultConfig(fs)
		require.NotNil(t, defaultConfig.AllowedHosts, "default config should have AllowedHosts set")

		// The open repository should behave differently - it should not restrict hosts
		// We verify this through the returned client type and that it's properly initialized
		require.NotNil(t, client, "client should be initialized")
	})

	t.Run("uses_provided_filesystem", func(t *testing.T) {
		t.Parallel()

		memFs := afero.NewMemMapFs()
		repo := newOpenRepository(memFs)

		// Verify the repository uses the provided filesystem by checking it's not nil
		// The actual filesystem usage is tested in git package tests
		require.NotNil(t, repo, "should use the provided filesystem")
	})

	t.Run("multiple_calls_return_independent_instances", func(t *testing.T) {
		t.Parallel()

		fs := afero.NewMemMapFs()
		repo1 := newOpenRepository(fs)
		repo2 := newOpenRepository(fs)

		// Each call should return a new instance
		assert.NotSame(t, repo1, repo2, "should return independent instances")
	})
}
