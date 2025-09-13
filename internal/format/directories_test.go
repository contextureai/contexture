package format

import (
	"testing"

	"github.com/contextureai/contexture/internal/domain"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDirectoryManager(t *testing.T) {
	t.Parallel()
	fs := afero.NewMemMapFs()
	dm := NewDirectoryManager(fs)

	assert.NotNil(t, dm)
	assert.Equal(t, fs, dm.fs)
}

func TestDirectoryManager_CreateFormatDirectories(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name            string
		formatType      domain.FormatType
		expectedDir     string
		shouldCreateDir bool
	}{
		{
			name:            "cursor format creates directory",
			formatType:      domain.FormatCursor,
			expectedDir:     domain.CursorOutputDir,
			shouldCreateDir: true,
		},
		{
			name:            "windsurf format creates directory",
			formatType:      domain.FormatWindsurf,
			expectedDir:     domain.WindsurfOutputDir,
			shouldCreateDir: true,
		},
		{
			name:            "claude format does not create directory",
			formatType:      domain.FormatClaude,
			expectedDir:     "",
			shouldCreateDir: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := afero.NewMemMapFs()
			dm := NewDirectoryManager(fs)

			err := dm.CreateFormatDirectories(tt.formatType)
			require.NoError(t, err)

			if tt.shouldCreateDir {
				// Check that directory was created
				exists, err := afero.DirExists(fs, tt.expectedDir)
				require.NoError(t, err)
				assert.True(t, exists, "Expected directory %s to be created", tt.expectedDir)

				// Check directory permissions
				info, err := fs.Stat(tt.expectedDir)
				require.NoError(t, err)
				assert.True(t, info.IsDir())
				assert.Equal(t, "drwxr-x---", info.Mode().String())
			}
		})
	}
}

func TestDirectoryManager_CreateFormatDirectories_UnknownFormat(t *testing.T) {
	t.Parallel()
	fs := afero.NewMemMapFs()
	dm := NewDirectoryManager(fs)

	err := dm.CreateFormatDirectories(domain.FormatType("unknown"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported format type")
}

func TestDirectoryManager_CreateFormatDirectories_ErrorHandling(t *testing.T) {
	t.Parallel()
	// Test with read-only filesystem to trigger errors
	fs := afero.NewReadOnlyFs(afero.NewMemMapFs())
	dm := NewDirectoryManager(fs)

	err := dm.CreateFormatDirectories(domain.FormatCursor)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "operation not permitted")

	err = dm.CreateFormatDirectories(domain.FormatWindsurf)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "operation not permitted")

	// Claude should not error since it doesn't create directories
	err = dm.CreateFormatDirectories(domain.FormatClaude)
	require.NoError(t, err)
}

func TestDirectoryManager_CreateFormatDirectories_Idempotent(t *testing.T) {
	t.Parallel()
	fs := afero.NewMemMapFs()
	dm := NewDirectoryManager(fs)

	// Create directory first time
	err := dm.CreateFormatDirectories(domain.FormatCursor)
	require.NoError(t, err)

	// Verify directory exists
	exists, err := afero.DirExists(fs, domain.CursorOutputDir)
	require.NoError(t, err)
	assert.True(t, exists)

	// Create directory second time - should not error
	err = dm.CreateFormatDirectories(domain.FormatCursor)
	require.NoError(t, err)

	// Directory should still exist
	exists, err = afero.DirExists(fs, domain.CursorOutputDir)
	require.NoError(t, err)
	assert.True(t, exists)
}
