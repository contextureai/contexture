// Package commands provides CLI command implementations
package commands

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/contextureai/contexture/internal/dependencies"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v3"
)

func TestNewAddCommand(t *testing.T) {
	fs := afero.NewMemMapFs()
	deps := &dependencies.Dependencies{
		FS:      fs,
		Context: context.Background(),
	}

	cmd := NewAddCommand(deps)
	assert.NotNil(t, cmd)
	assert.NotNil(t, cmd.projectManager)
	assert.NotNil(t, cmd.ruleFetcher)
	assert.NotNil(t, cmd.ruleValidator)
}

func TestAddAction(t *testing.T) {
	fs := afero.NewMemMapFs()
	deps := &dependencies.Dependencies{
		FS:      fs,
		Context: context.Background(),
	}

	// Create a context with empty arguments to simulate CLI with no args
	ctx := context.Background()

	// Test with no arguments (should now show available rules instead of failing)
	// Create a command that will have no arguments
	app := &cli.Command{
		Name: "test",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			return AddAction(ctx, cmd, deps)
		},
	}

	err := app.Run(ctx, []string{"test"})
	// Should not error now - it should show available rules
	require.NoError(t, err)
}

func TestAddCommand_Execute_NoConfig(t *testing.T) {
	fs := afero.NewMemMapFs()
	tempDir := "/tmp/test-add"
	_ = fs.MkdirAll(tempDir, 0o755)

	deps := &dependencies.Dependencies{
		FS:      fs,
		Context: context.Background(),
	}

	cmd := NewAddCommand(deps)

	// Create mock CLI command
	cliCmd := &cli.Command{}

	// Test with no project configuration (should fail)
	err := cmd.Execute(context.Background(), cliCmd, []string{"[contexture:test/rule]"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no Contexture project found")
}

func TestAddCommand_CustomDataParsing(t *testing.T) {
	tests := []struct {
		name         string
		dataInput    string
		expectError  bool
		expectedData map[string]any
	}{
		{
			name:        "valid JSON data",
			dataInput:   `{"name": "test", "version": "1.0", "enabled": true}`,
			expectError: false,
			expectedData: map[string]any{
				"name":    "test",
				"version": "1.0",
				"enabled": true,
			},
		},
		{
			name:        "nested JSON data",
			dataInput:   `{"config": {"timeout": 30, "retries": 3}, "tags": ["production", "api"]}`,
			expectError: false,
			expectedData: map[string]any{
				"config": map[string]any{
					"timeout": float64(30),
					"retries": float64(3),
				},
				"tags": []any{"production", "api"},
			},
		},
		{
			name:         "empty data",
			dataInput:    "",
			expectError:  false,
			expectedData: nil,
		},
		{
			name:        "invalid JSON",
			dataInput:   `{"name": "test", "invalid":}`,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test the JSON parsing logic directly
			var customData map[string]any
			var err error

			if tt.dataInput != "" {
				err = json.Unmarshal([]byte(tt.dataInput), &customData)
			}

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.expectedData != nil {
					assert.Equal(t, tt.expectedData, customData)
				} else {
					assert.Nil(t, customData)
				}
			}
		})
	}
}
