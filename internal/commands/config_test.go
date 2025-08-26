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

func TestNewMainConfigCommand(t *testing.T) {
	deps := &dependencies.Dependencies{
		FS:      afero.NewMemMapFs(),
		Context: context.Background(),
	}

	cmd := NewMainConfigCommand(deps)
	assert.NotNil(t, cmd)
	assert.NotNil(t, cmd.projectManager)
	assert.NotNil(t, cmd.registry)
}

func TestMainConfigCommand_IntegrationWithRegistry(t *testing.T) {
	// Test that the config command properly integrates with the format registry
	fs := afero.NewMemMapFs()
	deps := &dependencies.Dependencies{
		FS:      fs,
		Context: context.Background(),
	}

	// Create command and verify it can access registry
	cmd := NewMainConfigCommand(deps)
	assert.NotNil(t, cmd.registry)

	// Verify registry has handlers for all format types
	supportedFormats := []string{"claude", "cursor", "windsurf"}
	for _, formatStr := range supportedFormats {
		formatType := getFormatTypeFromString(formatStr)
		if formatType != "" {
			handler, exists := cmd.registry.GetHandler(formatType)
			assert.True(t, exists, "Registry should have handler for format %s", formatType)
			assert.NotNil(t, handler)
			assert.NotEmpty(t, handler.GetDisplayName())
		}
	}
}

// Helper function to convert string to format type
func getFormatTypeFromString(formatStr string) domain.FormatType {
	switch formatStr {
	case "claude":
		return domain.FormatClaude
	case "cursor":
		return domain.FormatCursor
	case "windsurf":
		return domain.FormatWindsurf
	default:
		return ""
	}
}
