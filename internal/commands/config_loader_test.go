// Package commands provides CLI command implementations
package commands

import (
	"testing"

	"github.com/contextureai/contexture/internal/domain"
	"github.com/stretchr/testify/assert"
)

func TestConfigLoadResult(t *testing.T) {
	t.Parallel()
	t.Run("config_load_result_fields", func(t *testing.T) {
		projectConfig := &domain.Project{
			Rules:   []domain.RuleRef{{ID: "test"}},
			Formats: []domain.FormatConfig{{Type: domain.FormatClaude, Enabled: true}},
		}

		result := &ConfigLoadResult{
			Config:     projectConfig,
			ConfigPath: "/test/.contexture.yaml",
			CurrentDir: "/test",
		}

		assert.Equal(t, projectConfig, result.Config)
		assert.Equal(t, "/test/.contexture.yaml", result.ConfigPath)
		assert.Equal(t, "/test", result.CurrentDir)
	})
}
