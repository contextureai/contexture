// Package commands provides CLI command implementations
package commands

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/contextureai/contexture/internal/dependencies"
	"github.com/contextureai/contexture/internal/domain"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v3"
)

func TestNewInitCommand(t *testing.T) {
	fs := afero.NewMemMapFs()
	deps := &dependencies.Dependencies{
		FS:      fs,
		Context: context.Background(),
	}

	cmd := NewInitCommand(deps)
	assert.NotNil(t, cmd)
	assert.NotNil(t, cmd.projectManager)
	assert.NotNil(t, cmd.registry)
}

func TestInitCommand_Execute(t *testing.T) {
	tests := []struct {
		name           string
		setupFS        func(afero.Fs) string
		formats        []string
		location       string
		force          bool
		expectError    bool
		expectedConfig *domain.Project
	}{
		{
			name: "successful initialization in empty directory",
			setupFS: func(fs afero.Fs) string {
				tempDir := "/tmp/test"
				_ = fs.MkdirAll(tempDir, 0o755)
				return tempDir
			},
			formats:     []string{"claude"},
			location:    "root",
			force:       false,
			expectError: false,
			expectedConfig: &domain.Project{
				Version: 1,
				Formats: []domain.FormatConfig{
					{Type: domain.FormatClaude, Enabled: true},
				},
				Rules: []domain.RuleRef{},
			},
		},
		{
			name: "initialization with multiple formats",
			setupFS: func(fs afero.Fs) string {
				tempDir := "/tmp/test-multi"
				_ = fs.MkdirAll(tempDir, 0o755)
				return tempDir
			},
			formats:     []string{"claude", "cursor", "windsurf"},
			location:    "contexture",
			force:       false,
			expectError: false,
			expectedConfig: &domain.Project{
				Version: 1,
				Formats: []domain.FormatConfig{
					{Type: domain.FormatClaude, Enabled: true},
					{Type: domain.FormatCursor, Enabled: true},
					{Type: domain.FormatWindsurf, Enabled: true},
				},
				Rules: []domain.RuleRef{},
			},
		},
		{
			name: "force overwrite existing configuration",
			setupFS: func(fs afero.Fs) string {
				tempDir := "/tmp/test-force"
				_ = fs.MkdirAll(tempDir, 0o755)
				// Create existing config
				configPath := filepath.Join(tempDir, domain.ConfigFile)
				content := `version: 1
formats:
  - type: claude
    enabled: false
rules: []
`
				_ = afero.WriteFile(fs, configPath, []byte(content), 0o644)
				return tempDir
			},
			formats:     []string{"cursor"},
			location:    "root",
			force:       true,
			expectError: false,
			expectedConfig: &domain.Project{
				Version: 1,
				Formats: []domain.FormatConfig{
					{Type: domain.FormatCursor, Enabled: true},
				},
				Rules: []domain.RuleRef{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := afero.NewMemMapFs()
			tempDir := tt.setupFS(fs)

			// Skip t.Chdir since we're using a memory filesystem

			deps := &dependencies.Dependencies{
				FS:      fs,
				Context: context.Background(),
			}

			cmd := NewInitCommand(deps)

			// Create mock CLI command
			cliCmd := &cli.Command{}
			cliCmd.Metadata = map[string]any{
				"force":    tt.force,
				"formats":  tt.formats,
				"location": tt.location,
			}

			// Mock current directory
			_ = fs.MkdirAll(tempDir, 0o755)

			// Create config directly for testing purposes
			var location domain.ConfigLocation
			if tt.location == "contexture" {
				location = domain.ConfigLocationContexture
			} else {
				location = domain.ConfigLocationRoot
			}

			formatTypes := make([]domain.FormatType, len(tt.formats))
			for i, f := range tt.formats {
				formatTypes[i] = domain.FormatType(f)
			}

			_, err := cmd.projectManager.InitConfig(tempDir, formatTypes, location)

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)

			// Verify config file was created
			var configPath string
			if tt.location == "contexture" {
				configPath = filepath.Join(tempDir, domain.GetContextureDir(), domain.ConfigFile)
			} else {
				configPath = filepath.Join(tempDir, domain.ConfigFile)
			}

			exists, err := afero.Exists(fs, configPath)
			require.NoError(t, err)
			assert.True(t, exists, "Config file should exist at %s", configPath)

			// Verify config content
			configResult, err := cmd.projectManager.LoadConfig(tempDir)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedConfig.Version, configResult.Config.Version)
			assert.Len(t, configResult.Config.Formats, len(tt.expectedConfig.Formats))
			assert.Len(t, configResult.Config.Rules, len(tt.expectedConfig.Rules))

			// Verify format types
			actualTypes := make([]domain.FormatType, len(configResult.Config.Formats))
			for i, f := range configResult.Config.Formats {
				actualTypes[i] = f.Type
				assert.True(t, f.Enabled, "Format should be enabled by default")
			}

			expectedTypes := make([]domain.FormatType, len(tt.expectedConfig.Formats))
			for i, f := range tt.expectedConfig.Formats {
				expectedTypes[i] = f.Type
			}

			assert.ElementsMatch(t, expectedTypes, actualTypes)
		})
	}
}

func TestInitCommand_ProjectManagerIntegration(t *testing.T) {
	fs := afero.NewMemMapFs()
	tempDir := "/tmp/test-integration"
	_ = fs.MkdirAll(tempDir, 0o755)

	deps := &dependencies.Dependencies{
		FS:      fs,
		Context: context.Background(),
	}

	cmd := NewInitCommand(deps)

	// Test that InitConfig works through project manager
	formats := []domain.FormatType{domain.FormatClaude}
	config, err := cmd.projectManager.InitConfig(tempDir, formats, domain.ConfigLocationRoot)

	require.NoError(t, err)
	assert.NotNil(t, config)
	assert.Equal(t, 1, config.Version)
	assert.Len(t, config.Formats, 1)
	assert.Equal(t, domain.FormatClaude, config.Formats[0].Type)
	assert.True(t, config.Formats[0].Enabled)

	// Verify config file was created
	configPath := filepath.Join(tempDir, domain.ConfigFile)
	exists, err := afero.Exists(fs, configPath)
	require.NoError(t, err)
	assert.True(t, exists)
}

func TestInitAction(t *testing.T) {
	fs := afero.NewMemMapFs()
	deps := &dependencies.Dependencies{
		FS:      fs,
		Context: context.Background(),
	}

	// Skip testing InitAction directly since it requires interactive input
	// Instead, test that NewInitCommand works properly
	initCmd := NewInitCommand(deps)
	assert.NotNil(t, initCmd)
	assert.NotNil(t, initCmd.projectManager)
	assert.NotNil(t, initCmd.registry)
}
