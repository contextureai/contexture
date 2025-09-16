package commands

import (
	"github.com/contextureai/contexture/internal/git"
	"github.com/spf13/afero"
)

// newOpenRepository returns a git repository client with host allowlisting disabled.
func newOpenRepository(fs afero.Fs) git.Repository {
	config := git.DefaultConfig(fs)
	config.AllowedHosts = nil
	return git.NewClient(fs, config)
}
