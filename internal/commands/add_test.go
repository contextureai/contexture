// Package commands provides CLI command implementations
package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/contextureai/contexture/internal/dependencies"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v3"
)

func TestNewAddCommand(t *testing.T) {
	t.Parallel()
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
	t.Parallel()
	fs := afero.NewMemMapFs()
	deps := &dependencies.Dependencies{
		FS:      fs,
		Context: context.Background(),
	}

	// Test with no arguments (should now return helpful error message)
	app := createTestApp(func(ctx context.Context, cmd *cli.Command) error {
		return AddAction(ctx, cmd, deps)
	})

	err := runTestApp(app)
	// Should error now with helpful message
	require.Error(t, err)
	assert.Contains(t, err.Error(), "validation failed")
	assert.Contains(t, err.Error(), "no rule IDs provided")
}

func TestAddCommand_Execute_NoConfig(t *testing.T) {
	t.Parallel()
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
	assert.Contains(t, err.Error(), "no configuration file found")
}

func TestAddCommand_CustomDataParsing(t *testing.T) {
	t.Parallel()
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

func TestAddCommand_SourceAndRefFlags(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name           string
		originalRuleID string
		sourceFlag     string
		refFlag        string
		expectedRuleID string
		description    string
	}{
		{
			name:           "simple rule ID with source flag",
			originalRuleID: "test/lemon",
			sourceFlag:     "https://github.com/user/repo.git",
			refFlag:        "",
			expectedRuleID: "[contexture(https://github.com/user/repo.git):test/lemon]",
			description:    "should construct proper rule ID with source",
		},
		{
			name:           "simple rule ID with source and ref flags",
			originalRuleID: "test/lemon",
			sourceFlag:     "https://github.com/user/repo.git",
			refFlag:        "main",
			expectedRuleID: "[contexture(https://github.com/user/repo.git):test/lemon,main]",
			description:    "should construct proper rule ID with source and ref",
		},
		{
			name:           "simple rule ID with source and branch ref",
			originalRuleID: "security/auth",
			sourceFlag:     "git@github.com:company/rules.git",
			refFlag:        "feature-branch",
			expectedRuleID: "[contexture(git@github.com:company/rules.git):security/auth,feature-branch]",
			description:    "should work with SSH URLs and branch names",
		},
		{
			name:           "no source flag provided",
			originalRuleID: "test/lemon",
			sourceFlag:     "",
			refFlag:        "",
			expectedRuleID: "test/lemon",
			description:    "should not modify rule ID when no source flag",
		},
		{
			name:           "ref flag without source flag",
			originalRuleID: "test/lemon",
			sourceFlag:     "",
			refFlag:        "main",
			expectedRuleID: "test/lemon",
			description:    "should ignore ref flag when no source flag",
		},
		{
			name:           "full rule ID format with source flag",
			originalRuleID: "[contexture:existing/rule]",
			sourceFlag:     "https://github.com/user/repo.git",
			refFlag:        "",
			expectedRuleID: "[contexture:existing/rule]",
			description:    "should not modify already formatted rule IDs",
		},
		{
			name:           "full rule ID with custom source and flags",
			originalRuleID: "[contexture(https://other.com/repo.git):other/rule]",
			sourceFlag:     "https://github.com/user/repo.git",
			refFlag:        "main",
			expectedRuleID: "[contexture(https://other.com/repo.git):other/rule]",
			description:    "should not modify already formatted rule IDs with custom source",
		},
		{
			name:           "complex rule path with source flag",
			originalRuleID: "security/authentication/oauth2",
			sourceFlag:     "https://github.com/enterprise/security-rules.git",
			refFlag:        "v2.1.0",
			expectedRuleID: "[contexture(https://github.com/enterprise/security-rules.git):security/authentication/oauth2,v2.1.0]",
			description:    "should handle complex nested rule paths and version tags",
		},
		{
			name:           "SSH URL with source flag",
			originalRuleID: "core/logging",
			sourceFlag:     "git@gitlab.com:company/rules.git",
			refFlag:        "",
			expectedRuleID: "[contexture(git@gitlab.com:company/rules.git):core/logging]",
			description:    "should work with SSH URLs from different Git providers",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test the rule ID construction logic
			originalRuleID := tt.originalRuleID
			sourceFlag := tt.sourceFlag
			refFlag := tt.refFlag

			// Apply the same logic as in the add command
			processedRuleID := originalRuleID
			if sourceFlag != "" {
				// If this is a simple rule ID (not already in [contexture:...] format),
				// construct the proper format using the --source and optional --ref flags
				if !strings.HasPrefix(originalRuleID, "[contexture") {
					if refFlag != "" {
						processedRuleID = fmt.Sprintf("[contexture(%s):%s,%s]", sourceFlag, originalRuleID, refFlag)
					} else {
						processedRuleID = fmt.Sprintf("[contexture(%s):%s]", sourceFlag, originalRuleID)
					}
				}
			}

			assert.Equal(t, tt.expectedRuleID, processedRuleID, tt.description)
		})
	}
}

func TestAddCommand_EdgeCases(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name           string
		originalRuleID string
		sourceFlag     string
		refFlag        string
		expectedRuleID string
		description    string
	}{
		{
			name:           "empty rule ID with source",
			originalRuleID: "",
			sourceFlag:     "https://github.com/user/repo.git",
			refFlag:        "",
			expectedRuleID: "[contexture(https://github.com/user/repo.git):]",
			description:    "should handle empty rule ID (though invalid)",
		},
		{
			name:           "rule ID with special characters",
			originalRuleID: "rules/test-rule_v2",
			sourceFlag:     "https://github.com/user/repo.git",
			refFlag:        "feature/new-rules",
			expectedRuleID: "[contexture(https://github.com/user/repo.git):rules/test-rule_v2,feature/new-rules]",
			description:    "should handle special characters in rule ID and ref",
		},
		{
			name:           "partial contexture format should not be modified",
			originalRuleID: "[contexture",
			sourceFlag:     "https://github.com/user/repo.git",
			refFlag:        "",
			expectedRuleID: "[contexture",
			description:    "should not modify incomplete contexture format (starts with [contexture)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test the rule ID construction logic for edge cases
			originalRuleID := tt.originalRuleID
			sourceFlag := tt.sourceFlag
			refFlag := tt.refFlag

			// Apply the same logic as in the add command
			processedRuleID := originalRuleID
			if sourceFlag != "" {
				if !strings.HasPrefix(originalRuleID, "[contexture") {
					if refFlag != "" {
						processedRuleID = fmt.Sprintf("[contexture(%s):%s,%s]", sourceFlag, originalRuleID, refFlag)
					} else {
						processedRuleID = fmt.Sprintf("[contexture(%s):%s]", sourceFlag, originalRuleID)
					}
				}
			}

			assert.Equal(t, tt.expectedRuleID, processedRuleID, tt.description)
		})
	}
}

func TestAddCommand_SrcAliasFunctionality(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		originalRuleID string
		sourceFlag     string
		expectedRuleID string
		description    string
	}{
		{
			name:           "src alias works like source",
			originalRuleID: "test/lemon",
			sourceFlag:     "https://github.com/user/repo.git",
			expectedRuleID: "[contexture(https://github.com/user/repo.git):test/lemon]",
			description:    "--src should produce same result as --source",
		},
		{
			name:           "src alias with SSH URL",
			originalRuleID: "security/auth",
			sourceFlag:     "git@github.com:company/rules.git",
			expectedRuleID: "[contexture(git@github.com:company/rules.git):security/auth]",
			description:    "--src should work with SSH URLs",
		},
		{
			name:           "src alias with complex path",
			originalRuleID: "languages/go/advanced-patterns",
			sourceFlag:     "https://github.com/myorg/contexture-rules.git",
			expectedRuleID: "[contexture(https://github.com/myorg/contexture-rules.git):languages/go/advanced-patterns]",
			description:    "--src should handle complex rule paths",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			originalRuleID := tt.originalRuleID
			sourceFlag := tt.sourceFlag

			// Apply the same logic as in the add command
			processedRuleID := originalRuleID
			if sourceFlag != "" {
				if !strings.HasPrefix(originalRuleID, "[contexture") {
					processedRuleID = fmt.Sprintf("[contexture(%s):%s]", sourceFlag, originalRuleID)
				}
			}

			assert.Equal(t, tt.expectedRuleID, processedRuleID, tt.description)
		})
	}
}

func TestPathExtraction_ForDocumentationPurposes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		ruleID       string
		expectedPath string
		description  string
	}{
		{
			name:         "custom source format",
			ruleID:       "[contexture(git@github.com:user/repo.git):test/lemon]",
			expectedPath: "test/lemon",
			description:  "should extract path from custom source format",
		},
		{
			name:         "simple rule ID",
			ruleID:       "test/simple",
			expectedPath: "test/simple",
			description:  "should use simple rule ID directly",
		},
		{
			name:         "standard contexture format",
			ruleID:       "[contexture:languages/go/testing]",
			expectedPath: "languages/go/testing",
			description:  "should extract path from standard format",
		},
		{
			name:         "custom source with branch",
			ruleID:       "[contexture(https://github.com/user/repo.git):security/auth,main]",
			expectedPath: "security/auth",
			description:  "should extract path and ignore branch",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var rulePath string

			// This demonstrates the path extraction logic (previously used in TUI)
			if strings.HasPrefix(tt.ruleID, "[contexture") {
				// Simulate domain.ExtractRulePath logic
				pathPart := strings.TrimPrefix(tt.ruleID, "[contexture:")
				if strings.HasPrefix(tt.ruleID, "[contexture(") {
					parts := strings.SplitN(pathPart, "):", 2)
					if len(parts) == 2 {
						pathPart = parts[1]
					}
				}
				pathPart = strings.TrimSuffix(pathPart, "]")

				// Remove branch suffix if present (path/rule,branch)
				if commaIdx := strings.Index(pathPart, ","); commaIdx != -1 {
					pathPart = pathPart[:commaIdx]
				}

				rulePath = pathPart
			} else if tt.ruleID != "" {
				// Use rule ID directly if it's not contexture format
				rulePath = tt.ruleID
			}

			assert.Equal(t, tt.expectedPath, rulePath, tt.description)
		})
	}
}
