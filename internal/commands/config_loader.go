// Package commands provides CLI command implementations
package commands

import (
	"os"

	"github.com/contextureai/contexture/internal/domain"
	contextureerrors "github.com/contextureai/contexture/internal/errors"
	"github.com/contextureai/contexture/internal/project"
)

// ConfigLoadResult represents the result of loading configuration
type ConfigLoadResult struct {
	Config       *domain.Project
	ConfigPath   string
	CurrentDir   string
	ConfigResult *domain.ConfigResult
}

// LoadProjectConfig loads project configuration
// This function encapsulates the common pattern used across multiple commands
func LoadProjectConfig(projectManager *project.Manager) (*ConfigLoadResult, error) {
	// Get current directory
	currentDir, err := os.Getwd()
	if err != nil {
		return nil, contextureerrors.Wrap(err, "get current directory")
	}

	// Load project configuration with local rules
	configResult, err := projectManager.LoadConfigWithLocalRules(currentDir)
	if err != nil {
		return nil, contextureerrors.Wrap(err, "load project configuration").
			WithSuggestions("Run 'contexture init' to create a project configuration")
	}

	result := &ConfigLoadResult{
		Config:       configResult.Config,
		ConfigPath:   configResult.Path,
		CurrentDir:   currentDir,
		ConfigResult: configResult,
	}

	return result, nil
}

// LoadProjectConfigOptional loads project configuration but doesn't fail if not found
// This is useful for commands that can work without existing configuration
func LoadProjectConfigOptional(projectManager *project.Manager) (*ConfigLoadResult, error) {
	// Get current directory
	currentDir, err := os.Getwd()
	if err != nil {
		return nil, contextureerrors.Wrap(err, "get current directory")
	}

	result := &ConfigLoadResult{
		CurrentDir: currentDir,
	}

	// Try to load project configuration with local rules, but don't fail if not found
	configResult, err := projectManager.LoadConfigWithLocalRules(currentDir)
	if err != nil {
		// Return nil config to indicate no configuration found
		result.Config = nil
		result.ConfigPath = ""
		result.ConfigResult = nil
	} else {
		result.Config = configResult.Config
		result.ConfigPath = configResult.Path
		result.ConfigResult = configResult
	}

	return result, nil
}

// SaveConfig saves the project configuration back to the appropriate location
func (r *ConfigLoadResult) SaveConfig(projectManager *project.Manager) error {
	// Save as project configuration
	if r.ConfigResult != nil {
		// Use the original location
		return projectManager.SaveConfig(r.Config, r.ConfigResult.Location, r.CurrentDir)
	}

	// Determine appropriate location for new config
	location := projectManager.GetConfigLocation(r.CurrentDir, false)
	return projectManager.SaveConfig(r.Config, location, r.CurrentDir)
}
