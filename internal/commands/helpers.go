package commands

import (
	"os"

	"github.com/contextureai/contexture/internal/domain"
	contextureerrors "github.com/contextureai/contexture/internal/errors"
	"github.com/contextureai/contexture/internal/project"
)

// loadConfigByScope loads either global or project configuration based on the isGlobal flag
// Returns the config, config path, and any error encountered
func loadConfigByScope(projectManager *project.Manager, isGlobal bool) (*domain.Project, string, error) {
	if isGlobal {
		// Initialize global config if needed (only for add/update commands)
		err := projectManager.InitializeGlobalConfig()
		if err != nil {
			return nil, "", contextureerrors.Wrap(err, "initialize global config")
		}

		// Load global config
		globalResult, err := projectManager.LoadGlobalConfig()
		if err != nil {
			return nil, "", contextureerrors.Wrap(err, "load global config")
		}
		if globalResult == nil || globalResult.Config == nil {
			return nil, "", contextureerrors.ValidationErrorf("global config", "no global configuration found")
		}
		return globalResult.Config, globalResult.Path, nil
	}

	// Get current directory and load project configuration
	currentDir, err := os.Getwd()
	if err != nil {
		return nil, "", contextureerrors.Wrap(err, "get current directory")
	}

	configResult, err := projectManager.LoadConfigWithLocalRules(currentDir)
	if err != nil {
		return nil, "", contextureerrors.Wrap(err, "load config")
	}
	return configResult.Config, configResult.Path, nil
}
