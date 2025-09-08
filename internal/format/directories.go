// Package format provides format utilities
package format

import (
	"fmt"

	"github.com/contextureai/contexture/internal/domain"
	"github.com/spf13/afero"
)

// DirectoryManager handles creation of format-specific directories
type DirectoryManager struct {
	fs      afero.Fs
	builder *Builder
}

// NewDirectoryManager creates a new directory manager
func NewDirectoryManager(fs afero.Fs) *DirectoryManager {
	return &DirectoryManager{
		fs:      fs,
		builder: NewBuilder(),
	}
}

// CreateFormatDirectories creates necessary directories for specific formats
func (dm *DirectoryManager) CreateFormatDirectories(formatType domain.FormatType) error {
	// Create a format instance to delegate directory creation
	format, err := dm.builder.Build(formatType, dm.fs, nil)
	if err != nil {
		return fmt.Errorf("failed to create format instance: %w", err)
	}

	// Use the format's own directory creation method
	config := &domain.FormatConfig{Type: formatType}
	return format.CreateDirectories(config)
}
