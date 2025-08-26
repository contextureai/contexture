// Package format provides format utilities
package format

import (
	"fmt"

	"github.com/contextureai/contexture/internal/domain"
	"github.com/spf13/afero"
)

// DirectoryManager handles creation of format-specific directories
type DirectoryManager struct {
	fs afero.Fs
}

// NewDirectoryManager creates a new directory manager
func NewDirectoryManager(fs afero.Fs) *DirectoryManager {
	return &DirectoryManager{fs: fs}
}

// CreateFormatDirectories creates necessary directories for specific formats
func (dm *DirectoryManager) CreateFormatDirectories(formatType domain.FormatType) error {
	switch formatType {
	case domain.FormatCursor:
		if err := dm.fs.MkdirAll(domain.CursorOutputDir, 0o755); err != nil {
			return fmt.Errorf("failed to create cursor directory: %w", err)
		}
	case domain.FormatWindsurf:
		if err := dm.fs.MkdirAll(domain.WindsurfOutputDir, 0o755); err != nil {
			return fmt.Errorf("failed to create windsurf directory: %w", err)
		}
	case domain.FormatClaude:
		// Claude format doesn't need a directory, just a file
	}
	return nil
}
