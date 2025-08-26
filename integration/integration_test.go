// Package integration provides integration tests for the Contexture CLI
// These tests focus on core functionality without interactive UI components
package integration

import (
	"path/filepath"
	"testing"

	"github.com/contextureai/contexture/internal/domain"
	"github.com/contextureai/contexture/internal/project"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestProjectManagerIntegration tests the project manager functionality end-to-end
func TestProjectManagerIntegration(t *testing.T) {
	fs := afero.NewMemMapFs()
	workingDir := "/integration-test"
	require.NoError(t, fs.MkdirAll(workingDir, 0o755))

	projectManager := project.NewManager(fs)

	t.Run("init_and_save_config", func(t *testing.T) {
		// Test initializing a configuration
		formatTypes := []domain.FormatType{
			domain.FormatClaude,
			domain.FormatCursor,
		}

		config, err := projectManager.InitConfig(workingDir, formatTypes, domain.ConfigLocationRoot)
		require.NoError(t, err)
		assert.NotNil(t, config)
		assert.Len(t, config.Formats, 2)

		// Verify config file was created
		configPath := filepath.Join(workingDir, ".contexture.yaml")
		exists, err := afero.Exists(fs, configPath)
		require.NoError(t, err)
		assert.True(t, exists)
	})

	t.Run("load_existing_config", func(t *testing.T) {
		// Load the config we just created
		configResult, err := projectManager.LoadConfig(workingDir)
		require.NoError(t, err)
		assert.NotNil(t, configResult.Config)
		assert.Equal(t, domain.ConfigLocationRoot, configResult.Location)
		assert.Len(t, configResult.Config.Formats, 2)
	})

	t.Run("add_rules_to_config", func(t *testing.T) {
		// Load current config
		configResult, err := projectManager.LoadConfig(workingDir)
		require.NoError(t, err)

		// Add a rule
		rule1 := domain.RuleRef{
			ID:     "[contexture:core/security/input-validation]",
			Source: "https://github.com/contextureai/rules.git",
			Ref:    "main",
		}

		err = projectManager.AddRule(configResult.Config, rule1)
		require.NoError(t, err)
		assert.Len(t, configResult.Config.Rules, 1)
		assert.Equal(t, rule1.ID, configResult.Config.Rules[0].ID)

		// Save the updated config
		err = projectManager.SaveConfig(configResult.Config, configResult.Location, workingDir)
		require.NoError(t, err)

		// Verify persistence
		reloadedConfig, err := projectManager.LoadConfig(workingDir)
		require.NoError(t, err)
		assert.Len(t, reloadedConfig.Config.Rules, 1)
		assert.Equal(t, rule1.ID, reloadedConfig.Config.Rules[0].ID)
	})

	t.Run("add_multiple_rules", func(t *testing.T) {
		// Load current config
		configResult, err := projectManager.LoadConfig(workingDir)
		require.NoError(t, err)

		// Add another rule
		rule2 := domain.RuleRef{
			ID:     "[contexture:typescript/strict-config]",
			Source: "https://github.com/contextureai/rules.git",
			Ref:    "main",
		}

		err = projectManager.AddRule(configResult.Config, rule2)
		require.NoError(t, err)
		assert.Len(t, configResult.Config.Rules, 2)

		// Save and verify
		err = projectManager.SaveConfig(configResult.Config, configResult.Location, workingDir)
		require.NoError(t, err)

		reloadedConfig, err := projectManager.LoadConfig(workingDir)
		require.NoError(t, err)
		assert.Len(t, reloadedConfig.Config.Rules, 2)
	})

	t.Run("check_if_rules_exist", func(t *testing.T) {
		configResult, err := projectManager.LoadConfig(workingDir)
		require.NoError(t, err)

		// Test HasRule functionality
		assert.True(t, projectManager.HasRule(
			configResult.Config,
			"[contexture:core/security/input-validation]",
		))
		assert.True(t, projectManager.HasRule(
			configResult.Config,
			"[contexture:typescript/strict-config]",
		))
		assert.False(t, projectManager.HasRule(
			configResult.Config,
			"[contexture:nonexistent/rule]",
		))
	})

	t.Run("remove_rules", func(t *testing.T) {
		configResult, err := projectManager.LoadConfig(workingDir)
		require.NoError(t, err)

		// Remove first rule
		err = projectManager.RemoveRule(
			configResult.Config,
			"[contexture:core/security/input-validation]",
		)
		require.NoError(t, err)
		assert.Len(t, configResult.Config.Rules, 1)
		assert.Equal(t, "[contexture:typescript/strict-config]", configResult.Config.Rules[0].ID)

		// Save and verify
		err = projectManager.SaveConfig(configResult.Config, configResult.Location, workingDir)
		require.NoError(t, err)

		reloadedConfig, err := projectManager.LoadConfig(workingDir)
		require.NoError(t, err)
		assert.Len(t, reloadedConfig.Config.Rules, 1)
		assert.Equal(t, "[contexture:typescript/strict-config]", reloadedConfig.Config.Rules[0].ID)
	})

	t.Run("config_location_preferences", func(t *testing.T) {
		// Test .contexture directory location
		contextureDir := "/contexture-test"
		require.NoError(t, fs.MkdirAll(contextureDir, 0o755))

		formatTypes := []domain.FormatType{domain.FormatWindsurf}
		config, err := projectManager.InitConfig(
			contextureDir,
			formatTypes,
			domain.ConfigLocationContexture,
		)
		require.NoError(t, err)
		assert.Len(t, config.Formats, 1)
		assert.Equal(t, domain.FormatWindsurf, config.Formats[0].Type)

		// Verify config file location
		contextureConfigPath := filepath.Join(contextureDir, ".contexture", ".contexture.yaml")
		exists, err := afero.Exists(fs, contextureConfigPath)
		require.NoError(t, err)
		assert.True(t, exists)

		// Load from contexture location
		reloadedConfig, err := projectManager.LoadConfig(contextureDir)
		require.NoError(t, err)
		assert.Equal(t, domain.ConfigLocationContexture, reloadedConfig.Location)
		assert.Len(t, reloadedConfig.Config.Formats, 1)
		assert.Equal(t, domain.FormatWindsurf, reloadedConfig.Config.Formats[0].Type)
	})
}

// TestFormatConfiguration tests format configuration functionality
func TestFormatConfiguration(t *testing.T) {
	fs := afero.NewMemMapFs()
	workingDir := "/format-test"
	require.NoError(t, fs.MkdirAll(workingDir, 0o755))

	projectManager := project.NewManager(fs)

	t.Run("create_config_with_all_formats", func(t *testing.T) {
		formatTypes := []domain.FormatType{
			domain.FormatClaude,
			domain.FormatCursor,
			domain.FormatWindsurf,
		}

		config, err := projectManager.InitConfig(workingDir, formatTypes, domain.ConfigLocationRoot)
		require.NoError(t, err)
		assert.Len(t, config.Formats, 3)

		// Verify all formats are enabled by default
		for _, format := range config.Formats {
			assert.True(t, format.Enabled, "Format %s should be enabled", format.Type)
		}
	})

	t.Run("get_enabled_formats", func(t *testing.T) {
		configResult, err := projectManager.LoadConfig(workingDir)
		require.NoError(t, err)

		enabledFormats := configResult.Config.GetEnabledFormats()
		assert.Len(t, enabledFormats, 3)
	})

	t.Run("get_format_by_type", func(t *testing.T) {
		configResult, err := projectManager.LoadConfig(workingDir)
		require.NoError(t, err)

		claudeFormat := configResult.Config.GetFormatByType(domain.FormatClaude)
		require.NotNil(t, claudeFormat)
		assert.Equal(t, domain.FormatClaude, claudeFormat.Type)
		assert.True(t, claudeFormat.Enabled)

		nonexistentFormat := configResult.Config.GetFormatByType("nonexistent")
		assert.Nil(t, nonexistentFormat)
	})

	t.Run("has_format", func(t *testing.T) {
		configResult, err := projectManager.LoadConfig(workingDir)
		require.NoError(t, err)

		assert.True(t, configResult.Config.HasFormat(domain.FormatClaude))
		assert.True(t, configResult.Config.HasFormat(domain.FormatCursor))
		assert.True(t, configResult.Config.HasFormat(domain.FormatWindsurf))
		assert.False(t, configResult.Config.HasFormat("nonexistent"))
	})
}

// TestErrorScenarios tests various error conditions
func TestErrorScenarios(t *testing.T) {
	fs := afero.NewMemMapFs()
	projectManager := project.NewManager(fs)

	t.Run("load_nonexistent_config", func(t *testing.T) {
		_, err := projectManager.LoadConfig("/nonexistent")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "no configuration file found")
	})

	t.Run("init_with_formats", func(t *testing.T) {
		workingDir := "/format-test"
		require.NoError(t, fs.MkdirAll(workingDir, 0o755))

		formatTypes := []domain.FormatType{
			domain.FormatClaude,
			domain.FormatCursor,
		}

		config, err := projectManager.InitConfig(workingDir, formatTypes, domain.ConfigLocationRoot)
		require.NoError(t, err)
		assert.Len(t, config.Formats, 2)
	})

	t.Run("add_duplicate_rule", func(t *testing.T) {
		workingDir := "/duplicate-test"
		require.NoError(t, fs.MkdirAll(workingDir, 0o755))

		// Create config with a rule
		config, err := projectManager.InitConfig(
			workingDir,
			[]domain.FormatType{domain.FormatClaude},
			domain.ConfigLocationRoot,
		)
		require.NoError(t, err)

		rule := domain.RuleRef{
			ID:     "[contexture:test/rule]",
			Source: "https://github.com/contextureai/rules.git",
			Ref:    "main",
		}

		// Add rule first time
		err = projectManager.AddRule(config, rule)
		require.NoError(t, err)
		assert.Len(t, config.Rules, 1)

		// Add same rule again - should update existing rule (not duplicate)
		modifiedRule := domain.RuleRef{
			ID:     "[contexture:test/rule]",
			Source: "https://github.com/contextureai/rules.git",
			Ref:    "develop", // Different branch
		}
		err = projectManager.AddRule(config, modifiedRule)
		require.NoError(t, err)
		assert.Len(t, config.Rules, 1, "Should not add duplicate rules")
		assert.Equal(t, "develop", config.Rules[0].Ref, "Should update existing rule")
	})

	t.Run("remove_nonexistent_rule", func(t *testing.T) {
		workingDir := "/remove-test"
		require.NoError(t, fs.MkdirAll(workingDir, 0o755))

		config, err := projectManager.InitConfig(
			workingDir,
			[]domain.FormatType{domain.FormatClaude},
			domain.ConfigLocationRoot,
		)
		require.NoError(t, err)

		// Try to remove a rule that doesn't exist - should return error
		err = projectManager.RemoveRule(config, "[contexture:nonexistent/rule]")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "rule not found")
		assert.Empty(t, config.Rules)
	})
}

// TestConfigurationPersistence tests that configurations persist correctly across operations
func TestConfigurationPersistence(t *testing.T) {
	fs := afero.NewMemMapFs()
	workingDir := "/persistence-test"
	require.NoError(t, fs.MkdirAll(workingDir, 0o755))

	projectManager := project.NewManager(fs)

	// Create initial configuration
	formatTypes := []domain.FormatType{domain.FormatClaude, domain.FormatCursor}
	config, err := projectManager.InitConfig(workingDir, formatTypes, domain.ConfigLocationRoot)
	require.NoError(t, err)

	// Add multiple rules
	rules := []domain.RuleRef{
		{
			ID:     "[contexture:core/security/input-validation]",
			Source: "https://github.com/contextureai/rules.git",
			Ref:    "main",
		},
		{
			ID:     "[contexture:typescript/strict-config]",
			Source: "https://github.com/contextureai/rules.git",
			Ref:    "main",
		},
		{
			ID:     "[contexture:react/component-naming]",
			Source: "https://github.com/contextureai/rules.git",
			Ref:    "develop",
		},
	}

	for _, rule := range rules {
		err = projectManager.AddRule(config, rule)
		require.NoError(t, err)
	}

	// Save configuration
	err = projectManager.SaveConfig(config, domain.ConfigLocationRoot, workingDir)
	require.NoError(t, err)

	// Reload and verify all data persisted
	reloadedConfigResult, err := projectManager.LoadConfig(workingDir)
	require.NoError(t, err)

	assert.Len(t, reloadedConfigResult.Config.Formats, 2)
	assert.Len(t, reloadedConfigResult.Config.Rules, 3)

	// Verify specific rule details
	for i, expectedRule := range rules {
		actualRule := reloadedConfigResult.Config.Rules[i]
		assert.Equal(t, expectedRule.ID, actualRule.ID)
		// For cleaned configs, default sources and branches are omitted
		// Verify functional equivalence rather than exact field values
		if expectedRule.Source == domain.DefaultRepository {
			// Default repository should result in empty source after cleaning
			assert.Empty(t, actualRule.Source,
				"Default repository source should be omitted after cleaning")
		} else {
			assert.Equal(t, expectedRule.Source, actualRule.Source)
		}

		if expectedRule.Ref == domain.DefaultBranch {
			// Default branch should result in empty branch after cleaning
			assert.Empty(t, actualRule.Ref, "Default branch should be omitted after cleaning")
		} else {
			assert.Equal(t, expectedRule.Ref, actualRule.Ref)
		}

		// Helper methods should still return correct effective values
		assert.Equal(t, expectedRule.GetRef(), actualRule.GetRef())
	}

	// Test operations after reload
	err = projectManager.RemoveRule(
		reloadedConfigResult.Config,
		"[contexture:typescript/strict-config]",
	)
	require.NoError(t, err)
	assert.Len(t, reloadedConfigResult.Config.Rules, 2)

	// Save and reload again
	err = projectManager.SaveConfig(
		reloadedConfigResult.Config,
		reloadedConfigResult.Location,
		workingDir,
	)
	require.NoError(t, err)

	finalConfigResult, err := projectManager.LoadConfig(workingDir)
	require.NoError(t, err)
	assert.Len(t, finalConfigResult.Config.Rules, 2)

	// Verify the correct rule was removed
	ruleIDs := make([]string, len(finalConfigResult.Config.Rules))
	for i, rule := range finalConfigResult.Config.Rules {
		ruleIDs[i] = rule.ID
	}
	assert.Contains(t, ruleIDs, "[contexture:core/security/input-validation]")
	assert.Contains(t, ruleIDs, "[contexture:react/component-naming]")
	assert.NotContains(t, ruleIDs, "[contexture:typescript/strict-config]")
}
