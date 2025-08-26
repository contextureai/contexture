package app

import (
	"strings"
	"testing"

	"github.com/contextureai/contexture/internal/domain"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewConfigManager(t *testing.T) {
	t.Parallel()
	fs := afero.NewMemMapFs()
	manager := NewConfigManager(fs)

	assert.NotNil(t, manager)
	assert.Equal(t, fs, manager.fs)
}

func TestConfigDisplayer_Load(t *testing.T) {
	fs := afero.NewMemMapFs()
	manager := NewConfigManager(fs)

	tests := []struct {
		name     string
		envVars  map[string]string
		expected *ConfigInfo
	}{
		{
			name:    "default values",
			envVars: map[string]string{},
			expected: &ConfigInfo{
				LogLevel:          "info",
				LogFormat:         "console",
				Verbose:           false,
				EnableDebug:       false,
				DefaultRepository: domain.DefaultRepository,
				DefaultBranch:     domain.DefaultBranch,
				DefaultFormats:    []string{"claude"},
				CacheEnabled:      true,
			},
		},
		{
			name: "custom values",
			envVars: map[string]string{
				"CONTEXTURE_LOG_LEVEL":          "debug",
				"CONTEXTURE_LOG_FORMAT":         "json",
				"CONTEXTURE_VERBOSE":            "true",
				"CONTEXTURE_DEBUG":              "true",
				"CONTEXTURE_DEFAULT_REPOSITORY": "https://example.com/rules.git",
				"CONTEXTURE_DEFAULT_BRANCH":     "develop",
				"CONTEXTURE_CACHE_ENABLED":      "false",
			},
			expected: &ConfigInfo{
				LogLevel:          "debug",
				LogFormat:         "json",
				Verbose:           true,
				EnableDebug:       true,
				DefaultRepository: "https://example.com/rules.git",
				DefaultBranch:     "develop",
				DefaultFormats:    []string{"claude"},
				CacheEnabled:      false,
			},
		},
		{
			name: "mixed boolean formats",
			envVars: map[string]string{
				"CONTEXTURE_VERBOSE":       "1",
				"CONTEXTURE_DEBUG":         "false",
				"CONTEXTURE_CACHE_ENABLED": "0",
			},
			expected: &ConfigInfo{
				LogLevel:          "info",
				LogFormat:         "console",
				Verbose:           true,
				EnableDebug:       false,
				DefaultRepository: domain.DefaultRepository,
				DefaultBranch:     domain.DefaultBranch,
				DefaultFormats:    []string{"claude"},
				CacheEnabled:      false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variables
			for key, value := range tt.envVars {
				t.Setenv(key, value)
			}

			config, err := manager.Load()
			require.NoError(t, err)
			assert.Equal(t, tt.expected, config)
		})
	}
}

func TestConfigDisplayer_GetEnvironmentDocumentation(t *testing.T) {
	t.Parallel()
	fs := afero.NewMemMapFs()
	manager := NewConfigManager(fs)

	doc := manager.GetEnvironmentDocumentation()

	// Check that documentation contains expected sections
	assert.Contains(t, doc, "LOGGING:")
	assert.Contains(t, doc, "REPOSITORY:")
	assert.Contains(t, doc, "CACHE:")
	assert.Contains(t, doc, "CONFIGURATION FILES:")
	assert.Contains(t, doc, "EXAMPLES:")

	// Check specific environment variables are documented
	assert.Contains(t, doc, "CONTEXTURE_LOG_LEVEL")
	assert.Contains(t, doc, "CONTEXTURE_LOG_FORMAT")
	assert.Contains(t, doc, "CONTEXTURE_VERBOSE")
	assert.Contains(t, doc, "CONTEXTURE_DEBUG")
	assert.Contains(t, doc, "CONTEXTURE_DEFAULT_REPOSITORY")
	assert.Contains(t, doc, "CONTEXTURE_DEFAULT_BRANCH")
	assert.Contains(t, doc, "CONTEXTURE_CACHE_ENABLED")

	// Check that default values are included
	assert.Contains(t, doc, domain.DefaultRepository)
	assert.Contains(t, doc, domain.DefaultBranch)

	// Check that documentation is well-formatted
	assert.Greater(t, len(doc), 100, "Documentation should be substantial")
	lines := strings.Split(doc, "\n")
	assert.Greater(t, len(lines), 10, "Documentation should be multi-line")
}

func TestGetEnvWithDefault(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue string
		envValue     string
		expected     string
	}{
		{
			name:         "environment variable set",
			key:          "TEST_KEY",
			defaultValue: "default",
			envValue:     "custom",
			expected:     "custom",
		},
		{
			name:         "environment variable empty",
			key:          "TEST_KEY",
			defaultValue: "default",
			envValue:     "",
			expected:     "default",
		},
		{
			name:         "environment variable not set",
			key:          "TEST_KEY",
			defaultValue: "default",
			envValue:     "", // Not setting environment variable
			expected:     "default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue != "" {
				t.Setenv(tt.key, tt.envValue)
			}

			result := getEnvWithDefault(tt.key, tt.defaultValue)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetBoolEnvWithDefault(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue bool
		envValue     string
		expected     bool
	}{
		{
			name:         "true string",
			key:          "TEST_BOOL",
			defaultValue: false,
			envValue:     "true",
			expected:     true,
		},
		{
			name:         "1 string",
			key:          "TEST_BOOL",
			defaultValue: false,
			envValue:     "1",
			expected:     true,
		},
		{
			name:         "false string",
			key:          "TEST_BOOL",
			defaultValue: true,
			envValue:     "false",
			expected:     false,
		},
		{
			name:         "0 string",
			key:          "TEST_BOOL",
			defaultValue: true,
			envValue:     "0",
			expected:     false,
		},
		{
			name:         "random string",
			key:          "TEST_BOOL",
			defaultValue: true,
			envValue:     "random",
			expected:     false,
		},
		{
			name:         "empty string uses default",
			key:          "TEST_BOOL",
			defaultValue: true,
			envValue:     "",
			expected:     true,
		},
		{
			name:         "not set uses default",
			key:          "TEST_BOOL",
			defaultValue: false,
			expected:     false,
		},
		{
			name:         "TRUE uppercase",
			key:          "TEST_BOOL",
			defaultValue: false,
			envValue:     "TRUE",
			expected:     false, // Only "true" and "1" are accepted
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue != "" {
				t.Setenv(tt.key, tt.envValue)
			}

			result := getBoolEnvWithDefault(tt.key, tt.defaultValue)
			assert.Equal(t, tt.expected, result)
		})
	}
}
