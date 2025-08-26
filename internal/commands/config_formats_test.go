// Package commands provides CLI command implementations
package commands

import (
	"context"
	"testing"

	"github.com/contextureai/contexture/internal/dependencies"
	"github.com/contextureai/contexture/internal/domain"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

func TestNewFormatManager(t *testing.T) {
	deps := &dependencies.Dependencies{
		FS:      afero.NewMemMapFs(),
		Context: context.Background(),
	}

	fm := NewFormatManager(deps)
	assert.NotNil(t, fm)
	assert.NotNil(t, fm.projectManager)
	assert.NotNil(t, fm.registry)
	assert.NotNil(t, fm.fs)
}

func TestFormatManager_HelperMethods(t *testing.T) {
	fs := afero.NewMemMapFs()
	deps := &dependencies.Dependencies{
		FS:      fs,
		Context: context.Background(),
	}

	fm := NewFormatManager(deps)

	t.Run("getFormatDisplayName", func(t *testing.T) {
		// Test with valid format
		displayName := fm.getFormatDisplayName(domain.FormatClaude)
		assert.NotEmpty(t, displayName)

		// Test with unknown format (should fallback to string representation)
		unknownFormat := domain.FormatType("unknown")
		displayName = fm.getFormatDisplayName(unknownFormat)
		assert.Equal(t, "unknown", displayName)
	})

	t.Run("getFormatOutputPath", func(t *testing.T) {
		tests := []struct {
			format       domain.FormatType
			expectedPath string
		}{
			{domain.FormatClaude, domain.ClaudeOutputFile},
			{domain.FormatCursor, domain.CursorOutputDir + "/"},
			{domain.FormatWindsurf, domain.WindsurfOutputDir + "/"},
			{domain.FormatType("unknown"), "unknown"},
		}

		for _, tt := range tests {
			path := fm.getFormatOutputPath(tt.format)
			assert.Equal(t, tt.expectedPath, path)
		}
	})

	t.Run("registry_integration", func(t *testing.T) {
		// Test that format manager properly integrates with registry
		assert.NotNil(t, fm.registry)

		// Verify registry has handlers for all known format types
		knownFormats := []domain.FormatType{
			domain.FormatClaude,
			domain.FormatCursor,
			domain.FormatWindsurf,
		}

		for _, formatType := range knownFormats {
			handler, exists := fm.registry.GetHandler(formatType)
			assert.True(t, exists, "Registry should have handler for format %s", formatType)
			assert.NotNil(t, handler)
			assert.NotEmpty(t, handler.GetDisplayName())
			assert.NotEmpty(t, handler.GetDescription())
		}
	})
}
