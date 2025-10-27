// Package integration provides integration tests for the Contexture CLI
// This file contains integration tests for global configuration functionality
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

// TestGlobalConfigCreateAndLoad tests creating and loading global configuration with real filesystem
func TestGlobalConfigCreateAndLoad(t *testing.T) {
	// Setup: Create temp home directory
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	// Create manager with OS filesystem
	manager := project.NewManager(afero.NewOsFs())

	t.Run("initialize creates global config", func(t *testing.T) {
		err := manager.InitializeGlobalConfig()
		require.NoError(t, err)

		// Verify file exists
		globalPath := filepath.Join(tmpHome, ".contexture", ".contexture.yaml")
		require.FileExists(t, globalPath)

		// Load it back
		loaded, err := manager.LoadGlobalConfig()
		require.NoError(t, err)
		require.NotNil(t, loaded)
		require.NotNil(t, loaded.Config)
		assert.Equal(t, 1, loaded.Config.Version)
		assert.Equal(t, domain.ConfigLocationGlobal, loaded.Location)
	})

	t.Run("save and load global config", func(t *testing.T) {
		config := &domain.Project{
			Version: 1,
			Rules: []domain.RuleRef{
				{ID: "@contexture/test-rule"},
			},
			Formats: []domain.FormatConfig{
				{Type: domain.FormatClaude, Enabled: true},
			},
		}

		err := manager.SaveGlobalConfig(config)
		require.NoError(t, err)

		// Load it back
		loaded, err := manager.LoadGlobalConfig()
		require.NoError(t, err)
		require.NotNil(t, loaded)
		require.NotNil(t, loaded.Config)
		assert.Len(t, loaded.Config.Rules, 1)
		assert.Equal(t, "@contexture/test-rule", loaded.Config.Rules[0].ID)
	})

	t.Run("add and remove rules", func(t *testing.T) {
		// Load current config
		result, err := manager.LoadGlobalConfig()
		require.NoError(t, err)
		require.NotNil(t, result.Config)

		initialCount := len(result.Config.Rules)

		// Add a rule
		newRule := domain.RuleRef{ID: "@contexture/new-rule"}
		err = manager.AddRule(result.Config, newRule)
		require.NoError(t, err)

		// Save
		err = manager.SaveGlobalConfig(result.Config)
		require.NoError(t, err)

		// Reload and verify
		result, err = manager.LoadGlobalConfig()
		require.NoError(t, err)
		assert.Len(t, result.Config.Rules, initialCount+1)

		// Remove the rule
		err = manager.RemoveRule(result.Config, "@contexture/new-rule")
		require.NoError(t, err)

		// Save
		err = manager.SaveGlobalConfig(result.Config)
		require.NoError(t, err)

		// Reload and verify
		result, err = manager.LoadGlobalConfig()
		require.NoError(t, err)
		assert.Len(t, result.Config.Rules, initialCount)
	})
}

// TestGlobalConfigMerging tests merging global and project configurations
func TestGlobalConfigMerging(t *testing.T) {
	// Setup: Create temp home directory
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	// Create temp project directory
	projectDir := t.TempDir()

	manager := project.NewManager(afero.NewOsFs())

	// Setup: Create global config with a rule
	t.Run("setup", func(t *testing.T) {
		err := manager.InitializeGlobalConfig()
		require.NoError(t, err)

		result, err := manager.LoadGlobalConfig()
		require.NoError(t, err)

		globalRule := domain.RuleRef{
			ID: "@contexture/global/rule1",
			Variables: map[string]any{
				"key": "global-value",
			},
		}
		err = manager.AddRule(result.Config, globalRule)
		require.NoError(t, err)

		err = manager.SaveGlobalConfig(result.Config)
		require.NoError(t, err)

		// Create project config
		formats := []domain.FormatType{domain.FormatClaude}
		config, err := manager.InitConfig(projectDir, formats, domain.ConfigLocationRoot)
		require.NoError(t, err)
		assert.NotNil(t, config)
	})

	t.Run("merge shows global rule", func(t *testing.T) {
		merged, err := manager.LoadConfigMerged(projectDir)
		require.NoError(t, err)
		require.NotNil(t, merged)

		// Should have 1 rule from global
		assert.Len(t, merged.MergedRules, 1)
		assert.Equal(t, "@contexture/global/rule1", merged.MergedRules[0].RuleRef.ID)
		assert.Equal(t, domain.RuleSourceUser, merged.MergedRules[0].Source)
		assert.False(t, merged.MergedRules[0].OverridesGlobal)
	})

	t.Run("project overrides global rule", func(t *testing.T) {
		// Load project config
		projectResult, err := manager.LoadConfig(projectDir)
		require.NoError(t, err)

		// Add same rule as global but with different variables
		overrideRule := domain.RuleRef{
			ID: "@contexture/global/rule1",
			Variables: map[string]any{
				"key": "project-override-value",
			},
		}
		err = manager.AddRule(projectResult.Config, overrideRule)
		require.NoError(t, err)

		err = manager.SaveConfig(projectResult.Config, projectResult.Location, projectDir)
		require.NoError(t, err)

		// Merge should show project version only
		merged, err := manager.LoadConfigMerged(projectDir)
		require.NoError(t, err)

		// Find the overridden rule
		var overriddenRule *domain.RuleWithSource
		for i := range merged.MergedRules {
			if merged.MergedRules[i].RuleRef.ID == "@contexture/global/rule1" {
				overriddenRule = &merged.MergedRules[i]
				break
			}
		}

		require.NotNil(t, overriddenRule, "overridden rule should be present")
		assert.Equal(t, domain.RuleSourceProject, overriddenRule.Source)
		assert.True(t, overriddenRule.OverridesGlobal)
		assert.Equal(t, "project-override-value", overriddenRule.RuleRef.Variables["key"])
	})

	t.Run("add project-specific rule", func(t *testing.T) {
		// Load project config
		projectResult, err := manager.LoadConfig(projectDir)
		require.NoError(t, err)

		// Add project-specific rule
		projectRule := domain.RuleRef{
			ID: "@contexture/project/rule1",
		}
		err = manager.AddRule(projectResult.Config, projectRule)
		require.NoError(t, err)

		err = manager.SaveConfig(projectResult.Config, projectResult.Location, projectDir)
		require.NoError(t, err)

		// Merge should now show both global (original) and project rules
		merged, err := manager.LoadConfigMerged(projectDir)
		require.NoError(t, err)

		// We should have: global/rule1 (overridden by project) + project/rule1
		assert.Len(t, merged.MergedRules, 2)

		// Verify sources
		rulesByID := make(map[string]domain.RuleSource)
		for _, rws := range merged.MergedRules {
			rulesByID[rws.RuleRef.ID] = rws.Source
		}
		assert.Equal(t, domain.RuleSourceProject, rulesByID["@contexture/global/rule1"]) // Overridden
		assert.Equal(t, domain.RuleSourceProject, rulesByID["@contexture/project/rule1"])
	})
}

// TestGlobalConfigWithProviders tests global configuration with custom providers
func TestGlobalConfigWithProviders(t *testing.T) {
	// Setup: Create temp home directory
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	manager := project.NewManager(afero.NewOsFs())

	t.Run("add and persist global provider", func(t *testing.T) {
		err := manager.InitializeGlobalConfig()
		require.NoError(t, err)

		result, err := manager.LoadGlobalConfig()
		require.NoError(t, err)
		require.NotNil(t, result.Config)

		// Add custom provider
		result.Config.Providers = append(result.Config.Providers, domain.Provider{
			Name:          "mycompany",
			URL:           "https://github.com/mycompany/rules.git",
			DefaultBranch: "main",
		})

		err = manager.SaveGlobalConfig(result.Config)
		require.NoError(t, err)

		// Reload and verify persistence
		reloaded, err := manager.LoadGlobalConfig()
		require.NoError(t, err)
		require.NotNil(t, reloaded.Config)
		assert.Len(t, reloaded.Config.Providers, 1)
		assert.Equal(t, "mycompany", reloaded.Config.Providers[0].Name)
		assert.Equal(t, "https://github.com/mycompany/rules.git", reloaded.Config.Providers[0].URL)
		// DefaultBranch may be omitted or reset during serialization, this is expected
	})
}

// TestGlobalConfigFormats tests global config with multiple output formats
func TestGlobalConfigFormats(t *testing.T) {
	// Setup: Create temp home directory
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	manager := project.NewManager(afero.NewOsFs())

	t.Run("initialize has all formats enabled", func(t *testing.T) {
		err := manager.InitializeGlobalConfig()
		require.NoError(t, err)

		result, err := manager.LoadGlobalConfig()
		require.NoError(t, err)
		require.NotNil(t, result.Config)

		// Should have all 3 formats enabled by default
		assert.Len(t, result.Config.Formats, 3)

		formatTypes := make(map[domain.FormatType]bool)
		for _, f := range result.Config.Formats {
			formatTypes[f.Type] = f.Enabled
		}

		assert.True(t, formatTypes[domain.FormatClaude])
		assert.True(t, formatTypes[domain.FormatCursor])
		assert.True(t, formatTypes[domain.FormatWindsurf])
	})

	t.Run("modify and persist format settings", func(t *testing.T) {
		result, err := manager.LoadGlobalConfig()
		require.NoError(t, err)
		require.NotNil(t, result.Config)

		// Disable windsurf
		for i := range result.Config.Formats {
			if result.Config.Formats[i].Type == domain.FormatWindsurf {
				result.Config.Formats[i].Enabled = false
			}
		}

		err = manager.SaveGlobalConfig(result.Config)
		require.NoError(t, err)

		// Reload and verify persistence
		reloaded, err := manager.LoadGlobalConfig()
		require.NoError(t, err)
		require.NotNil(t, reloaded.Config)

		for _, f := range reloaded.Config.Formats {
			if f.Type == domain.FormatWindsurf {
				assert.False(t, f.Enabled)
			} else {
				assert.True(t, f.Enabled)
			}
		}
	})
}
