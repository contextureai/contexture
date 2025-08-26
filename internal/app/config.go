// Package app provides simplified configuration information for the contexture CLI
package app

import (
	"os"

	"github.com/contextureai/contexture/internal/domain"
	"github.com/spf13/afero"
)

// ConfigDisplayer provides configuration information display
type ConfigDisplayer struct {
	fs afero.Fs
}

// NewConfigManager creates a new configuration displayer
func NewConfigManager(fs afero.Fs) *ConfigDisplayer {
	return &ConfigDisplayer{
		fs: fs,
	}
}

// Load shows current configuration values from environment and defaults
func (c *ConfigDisplayer) Load() (*ConfigInfo, error) {
	return &ConfigInfo{
		LogLevel:    getEnvWithDefault("CONTEXTURE_LOG_LEVEL", "info"),
		LogFormat:   getEnvWithDefault("CONTEXTURE_LOG_FORMAT", "console"),
		Verbose:     getBoolEnvWithDefault("CONTEXTURE_VERBOSE", false),
		EnableDebug: getBoolEnvWithDefault("CONTEXTURE_DEBUG", false),
		DefaultRepository: getEnvWithDefault(
			"CONTEXTURE_DEFAULT_REPOSITORY",
			domain.DefaultRepository,
		),
		DefaultBranch: getEnvWithDefault("CONTEXTURE_DEFAULT_BRANCH", domain.DefaultBranch),
		DefaultFormats: []string{
			"claude",
		}, // Simplified - actual formats come from project config
		CacheEnabled: getBoolEnvWithDefault("CONTEXTURE_CACHE_ENABLED", true),
	}, nil
}

// ConfigInfo represents current configuration values
type ConfigInfo struct {
	LogLevel          string
	LogFormat         string
	Verbose           bool
	EnableDebug       bool
	DefaultRepository string
	DefaultBranch     string
	DefaultFormats    []string
	CacheEnabled      bool
}

// GetEnvironmentDocumentation returns documentation for environment variables
func (c *ConfigDisplayer) GetEnvironmentDocumentation() string {
	return `Contexture Environment Variables:

LOGGING:
  CONTEXTURE_LOG_LEVEL       Log level (debug, info, warn, error) [default: info]
  CONTEXTURE_LOG_FORMAT      Log format (console, json, text) [default: console]  
  CONTEXTURE_VERBOSE         Enable verbose output (true, false) [default: false]
  CONTEXTURE_DEBUG           Enable debug mode (true, false) [default: false]

REPOSITORY:
  CONTEXTURE_DEFAULT_REPOSITORY  Default rules repository URL [default: ` + domain.DefaultRepository + `]
  CONTEXTURE_DEFAULT_BRANCH      Default branch name [default: ` + domain.DefaultBranch + `]

CACHE:
  CONTEXTURE_CACHE_ENABLED       Enable caching (true, false) [default: true]

CONFIGURATION FILES:
  Project configuration files (.contexture.yaml):
  1. .contexture.yaml (current directory)
  2. .contexture/.contexture.yaml

EXAMPLES:
  export CONTEXTURE_LOG_LEVEL=debug
  export CONTEXTURE_VERBOSE=true
  export CONTEXTURE_DEFAULT_REPOSITORY=https://github.com/myorg/rules.git
`
}

// Helper functions
func getEnvWithDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getBoolEnvWithDefault(key string, defaultValue bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value == "true" || value == "1"
}
