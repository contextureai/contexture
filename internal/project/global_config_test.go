// Package project provides project configuration management
package project

import (
	"path/filepath"
	"testing"

	"github.com/contextureai/contexture/internal/domain"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testHomeDir     = "/home/testuser"
	testProjectDir  = "/project"
	testLocalSource = "local"
)

// mockHomeProvider is a test implementation of HomeDirectoryProvider
type mockHomeProvider struct {
	homeDir string
	err     error
}

func (m *mockHomeProvider) GetHomeDir() (string, error) {
	if m.err != nil {
		return "", m.err
	}
	return m.homeDir, nil
}

// newTestManagerWithHome creates a Manager with a mocked home directory for testing
//
//nolint:unparam // homeDir varies across test cases
func newTestManagerWithHome(fs afero.Fs, homeDir string) *Manager {
	matcher := &DefaultRuleMatcher{
		regex: domain.RuleIDParsePatternRegex,
	}

	return NewManagerForTesting(
		&DefaultConfigRepository{fs: fs},
		matcher,
		newDefaultConfigValidator(),
		&mockHomeProvider{homeDir: homeDir},
	)
}

// TestManager_LoadGlobalConfig tests loading global configuration
func TestManager_LoadGlobalConfig(t *testing.T) {
	t.Run("returns result with nil config when global config does not exist", func(t *testing.T) {
		fs := afero.NewMemMapFs()
		manager := newTestManagerWithHome(fs, testHomeDir)

		result, err := manager.LoadGlobalConfig()
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Nil(t, result.Config)
		assert.Equal(t, domain.ConfigLocationGlobal, result.Location)
	})

	t.Run("loads valid global config", func(t *testing.T) {
		fs := afero.NewMemMapFs()
		manager := newTestManagerWithHome(fs, testHomeDir)

		// Create global config directory
		globalDir := filepath.Join(testHomeDir, ".contexture")
		_ = fs.MkdirAll(globalDir, 0o755)

		// Create valid config
		configPath := filepath.Join(globalDir, ".contexture.yaml")
		configContent := `version: 1
formats:
  - type: claude
    enabled: true
rules:
  - id: "@contexture/test-rule"
providers:
  - name: mycompany
    url: https://github.com/mycompany/rules.git
`
		err := afero.WriteFile(fs, configPath, []byte(configContent), 0o644)
		require.NoError(t, err)

		result, err := manager.LoadGlobalConfig()
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, domain.ConfigLocationGlobal, result.Location)
		assert.Len(t, result.Config.Rules, 1)
		assert.Equal(t, "@contexture/test-rule", result.Config.Rules[0].ID)
		assert.Len(t, result.Config.Providers, 1)
		assert.Equal(t, "mycompany", result.Config.Providers[0].Name)
	})

	t.Run("returns error for invalid config", func(t *testing.T) {
		fs := afero.NewMemMapFs()
		manager := newTestManagerWithHome(fs, testHomeDir)

		// Create global config directory with invalid YAML
		globalDir := filepath.Join(testHomeDir, ".contexture")
		_ = fs.MkdirAll(globalDir, 0o755)

		configPath := filepath.Join(globalDir, ".contexture.yaml")
		invalidConfig := `version: 1
rules:
  - invalid yaml syntax here [[[
`
		err := afero.WriteFile(fs, configPath, []byte(invalidConfig), 0o644)
		require.NoError(t, err)

		result, err := manager.LoadGlobalConfig()
		require.Error(t, err)
		assert.Nil(t, result)
	})
}

// TestManager_SaveGlobalConfig tests saving global configuration
func TestManager_SaveGlobalConfig(t *testing.T) {
	t.Run("creates directory and saves config", func(t *testing.T) {
		fs := afero.NewMemMapFs()
		manager := newTestManagerWithHome(fs, testHomeDir)

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

		// Verify directory was created
		globalDir := filepath.Join(testHomeDir, ".contexture")
		exists, _ := afero.DirExists(fs, globalDir)
		assert.True(t, exists)

		// Verify file was created
		configPath := filepath.Join(testHomeDir, ".contexture", ".contexture.yaml")
		exists, _ = afero.Exists(fs, configPath)
		assert.True(t, exists)

		// Verify content
		content, err := afero.ReadFile(fs, configPath)
		require.NoError(t, err)
		assert.Contains(t, string(content), "@contexture/test-rule")
		assert.Contains(t, string(content), "version: 1")
	})

	t.Run("returns error for nil config", func(t *testing.T) {
		fs := afero.NewMemMapFs()
		manager := newTestManagerWithHome(fs, testHomeDir)

		err := manager.SaveGlobalConfig(nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot be nil")
	})

	t.Run("overwrites existing config", func(t *testing.T) {
		fs := afero.NewMemMapFs()
		manager := newTestManagerWithHome(fs, testHomeDir)

		// Save first config
		config1 := &domain.Project{
			Version: 1,
			Rules: []domain.RuleRef{
				{ID: "@contexture/rule1"},
			},
			Formats: []domain.FormatConfig{
				{Type: domain.FormatClaude, Enabled: true},
			},
		}
		err := manager.SaveGlobalConfig(config1)
		require.NoError(t, err)

		// Save second config
		config2 := &domain.Project{
			Version: 1,
			Rules: []domain.RuleRef{
				{ID: "@contexture/rule2"},
			},
			Formats: []domain.FormatConfig{
				{Type: domain.FormatCursor, Enabled: true},
			},
		}
		err = manager.SaveGlobalConfig(config2)
		require.NoError(t, err)

		// Load and verify second config is there
		result, err := manager.LoadGlobalConfig()
		require.NoError(t, err)
		assert.Len(t, result.Config.Rules, 1)
		assert.Equal(t, "@contexture/rule2", result.Config.Rules[0].ID)
	})
}

// TestManager_InitializeGlobalConfig tests initializing global configuration
func TestManager_InitializeGlobalConfig(t *testing.T) {
	t.Run("creates directory structure", func(t *testing.T) {
		fs := afero.NewMemMapFs()
		manager := newTestManagerWithHome(fs, testHomeDir)

		err := manager.InitializeGlobalConfig()
		require.NoError(t, err)

		// Verify directory exists
		globalDir := filepath.Join(testHomeDir, ".contexture")
		exists, _ := afero.DirExists(fs, globalDir)
		assert.True(t, exists)

		// Verify config file exists with minimal content
		configPath := filepath.Join(globalDir, ".contexture.yaml")
		exists, _ = afero.Exists(fs, configPath)
		assert.True(t, exists)

		// Verify it's valid YAML
		result, err := manager.LoadGlobalConfig()
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, 1, result.Config.Version)
	})

	t.Run("does not overwrite existing config", func(t *testing.T) {
		fs := afero.NewMemMapFs()
		manager := newTestManagerWithHome(fs, testHomeDir)

		// Create existing config
		config := &domain.Project{
			Version: 1,
			Rules: []domain.RuleRef{
				{ID: "@contexture/existing-rule"},
			},
			Formats: []domain.FormatConfig{
				{Type: domain.FormatClaude, Enabled: true},
			},
		}
		err := manager.SaveGlobalConfig(config)
		require.NoError(t, err)

		// Initialize should not overwrite
		err = manager.InitializeGlobalConfig()
		require.NoError(t, err)

		// Verify original config is still there
		result, err := manager.LoadGlobalConfig()
		require.NoError(t, err)
		assert.Len(t, result.Config.Rules, 1)
		assert.Equal(t, "@contexture/existing-rule", result.Config.Rules[0].ID)
	})
}

// TestManager_LoadConfigMerged tests merging global and project configurations
func TestManager_LoadConfigMerged(t *testing.T) {
	t.Run("merges global and project rules", func(t *testing.T) {
		fs := afero.NewMemMapFs()
		manager := newTestManagerWithHome(fs, testHomeDir)
		projectDir := testProjectDir

		// Create global config
		globalConfig := &domain.Project{
			Version: 1,
			Rules: []domain.RuleRef{
				{ID: "@contexture/global-rule"},
			},
			Formats: []domain.FormatConfig{
				{Type: domain.FormatClaude, Enabled: true},
			},
		}
		err := manager.SaveGlobalConfig(globalConfig)
		require.NoError(t, err)

		// Create project config
		_ = fs.MkdirAll(projectDir, 0o755)
		projectConfigPath := filepath.Join(projectDir, ".contexture.yaml")
		projectConfigContent := `version: 1
formats:
  - type: cursor
    enabled: true
rules:
  - id: "@contexture/project-rule"
`
		err = afero.WriteFile(fs, projectConfigPath, []byte(projectConfigContent), 0o644)
		require.NoError(t, err)

		// Load merged config
		merged, err := manager.LoadConfigMerged(projectDir)
		require.NoError(t, err)
		require.NotNil(t, merged)

		// Should have 2 rules total
		assert.Len(t, merged.MergedRules, 2)

		// Verify sources
		ruleIDs := make(map[string]domain.RuleSource)
		for _, rws := range merged.MergedRules {
			ruleIDs[rws.RuleRef.ID] = rws.Source
		}

		assert.Equal(t, domain.RuleSourceGlobal, ruleIDs["@contexture/global-rule"])
		assert.Equal(t, domain.RuleSourceProject, ruleIDs["@contexture/project-rule"])
	})

	t.Run("project rule overrides global rule", func(t *testing.T) {
		fs := afero.NewMemMapFs()
		manager := newTestManagerWithHome(fs, testHomeDir)
		projectDir := testProjectDir

		// Create global config with variables
		globalConfig := &domain.Project{
			Version: 1,
			Rules: []domain.RuleRef{
				{
					ID:        "@contexture/shared-rule",
					Variables: map[string]any{"key": "global-value"},
				},
			},
			Formats: []domain.FormatConfig{
				{Type: domain.FormatClaude, Enabled: true},
			},
		}
		err := manager.SaveGlobalConfig(globalConfig)
		require.NoError(t, err)

		// Create project config with same rule but different variables
		_ = fs.MkdirAll(projectDir, 0o755)
		projectConfigPath := filepath.Join(projectDir, ".contexture.yaml")
		projectConfigContent := `version: 1
formats:
  - type: claude
    enabled: true
rules:
  - id: "@contexture/shared-rule"
    variables:
      key: "project-value"
`
		err = afero.WriteFile(fs, projectConfigPath, []byte(projectConfigContent), 0o644)
		require.NoError(t, err)

		// Load merged config
		merged, err := manager.LoadConfigMerged(projectDir)
		require.NoError(t, err)
		require.NotNil(t, merged)

		// Should have only 1 rule (project overrides global)
		assert.Len(t, merged.MergedRules, 1)

		rule := merged.MergedRules[0]
		assert.Equal(t, "@contexture/shared-rule", rule.RuleRef.ID)
		assert.Equal(t, domain.RuleSourceProject, rule.Source)
		assert.True(t, rule.OverridesGlobal)
		assert.Equal(t, "project-value", rule.RuleRef.Variables["key"])
	})

	t.Run("works with only project config", func(t *testing.T) {
		fs := afero.NewMemMapFs()
		manager := newTestManagerWithHome(fs, testHomeDir)
		projectDir := testProjectDir

		// No global config, only project
		_ = fs.MkdirAll(projectDir, 0o755)
		projectConfigPath := filepath.Join(projectDir, ".contexture.yaml")
		projectConfigContent := `version: 1
formats:
  - type: claude
    enabled: true
rules:
  - id: "@contexture/project-only-rule"
`
		err := afero.WriteFile(fs, projectConfigPath, []byte(projectConfigContent), 0o644)
		require.NoError(t, err)

		// Load merged config
		merged, err := manager.LoadConfigMerged(projectDir)
		require.NoError(t, err)
		require.NotNil(t, merged)

		assert.Len(t, merged.MergedRules, 1)
		assert.Equal(t, "@contexture/project-only-rule", merged.MergedRules[0].RuleRef.ID)
		assert.Equal(t, domain.RuleSourceProject, merged.MergedRules[0].Source)
		assert.False(t, merged.MergedRules[0].OverridesGlobal)
	})

	t.Run("works with only global config", func(t *testing.T) {
		fs := afero.NewMemMapFs()
		manager := newTestManagerWithHome(fs, testHomeDir)
		projectDir := testProjectDir

		// Create global config
		globalConfig := &domain.Project{
			Version: 1,
			Rules: []domain.RuleRef{
				{ID: "@contexture/global-only-rule"},
			},
			Formats: []domain.FormatConfig{
				{Type: domain.FormatClaude, Enabled: true},
			},
		}
		err := manager.SaveGlobalConfig(globalConfig)
		require.NoError(t, err)

		// Create empty project config
		_ = fs.MkdirAll(projectDir, 0o755)
		projectConfigPath := filepath.Join(projectDir, ".contexture.yaml")
		projectConfigContent := `version: 1
formats:
  - type: claude
    enabled: true
`
		err = afero.WriteFile(fs, projectConfigPath, []byte(projectConfigContent), 0o644)
		require.NoError(t, err)

		// Load merged config
		merged, err := manager.LoadConfigMerged(projectDir)
		require.NoError(t, err)
		require.NotNil(t, merged)

		assert.Len(t, merged.MergedRules, 1)
		assert.Equal(t, "@contexture/global-only-rule", merged.MergedRules[0].RuleRef.ID)
		assert.Equal(t, domain.RuleSourceGlobal, merged.MergedRules[0].Source)
		assert.False(t, merged.MergedRules[0].OverridesGlobal)
	})
}

// TestManager_LoadConfigMergedWithLocalRules tests merging including local rules
func TestManager_LoadConfigMergedWithLocalRules(t *testing.T) {
	t.Run("includes local rules in merge", func(t *testing.T) {
		fs := afero.NewMemMapFs()
		manager := newTestManagerWithHome(fs, testHomeDir)
		projectDir := testProjectDir

		// Create global config
		globalConfig := &domain.Project{
			Version: 1,
			Rules: []domain.RuleRef{
				{ID: "@contexture/global-rule"},
			},
			Formats: []domain.FormatConfig{
				{Type: domain.FormatClaude, Enabled: true},
			},
		}
		err := manager.SaveGlobalConfig(globalConfig)
		require.NoError(t, err)

		// Create project config
		_ = fs.MkdirAll(projectDir, 0o755)
		projectConfigPath := filepath.Join(projectDir, ".contexture.yaml")
		projectConfigContent := `version: 1
formats:
  - type: claude
    enabled: true
rules:
  - id: "@contexture/project-rule"
`
		err = afero.WriteFile(fs, projectConfigPath, []byte(projectConfigContent), 0o644)
		require.NoError(t, err)

		// Create local rule
		rulesDir := filepath.Join(projectDir, "rules")
		_ = fs.MkdirAll(rulesDir, 0o755)
		localRulePath := filepath.Join(rulesDir, "local-rule.md")
		localRuleContent := `---
title: Local Rule
---
# Local Rule Content
`
		err = afero.WriteFile(fs, localRulePath, []byte(localRuleContent), 0o644)
		require.NoError(t, err)

		// Load merged config with local rules
		merged, err := manager.LoadConfigMergedWithLocalRules(projectDir)
		require.NoError(t, err)
		require.NotNil(t, merged)

		// Should have 3 rules: global, project, and local
		assert.Len(t, merged.MergedRules, 3)

		// Verify all sources are present
		sources := make(map[string]bool)
		for _, rws := range merged.MergedRules {
			if rws.RuleRef.Source == testLocalSource {
				sources[testLocalSource] = true
			} else {
				sources[string(rws.Source)] = true
			}
		}

		assert.True(t, sources["global"])
		assert.True(t, sources["project"])
		assert.True(t, sources[testLocalSource])
	})
}
