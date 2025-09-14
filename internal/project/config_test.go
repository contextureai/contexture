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

func TestNewManager(t *testing.T) {
	t.Parallel()
	fs := afero.NewMemMapFs()
	manager := NewManager(fs)
	assert.NotNil(t, manager)
	// Manager now uses interfaces, we can't directly access the fs field
	// but we can test that it works by performing an operation
}

func TestManager_InitConfig(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		formats    []domain.FormatType
		location   domain.ConfigLocation
		expectPath string
	}{
		{
			name:       "init with claude format in root",
			formats:    []domain.FormatType{domain.FormatClaude},
			location:   domain.ConfigLocationRoot,
			expectPath: ".contexture.yaml",
		},
		{
			name:       "init with multiple formats in contexture dir",
			formats:    []domain.FormatType{domain.FormatClaude, domain.FormatCursor},
			location:   domain.ConfigLocationContexture,
			expectPath: ".contexture/.contexture.yaml",
		},
		{
			name:       "init with windsurf format",
			formats:    []domain.FormatType{domain.FormatWindsurf},
			location:   domain.ConfigLocationRoot,
			expectPath: ".contexture.yaml",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := afero.NewMemMapFs()
			tempDir := "/tmp/test"
			_ = fs.MkdirAll(tempDir, 0o755)

			manager := NewManager(fs)
			config, err := manager.InitConfig(tempDir, tt.formats, tt.location)

			require.NoError(t, err)
			assert.NotNil(t, config)
			assert.Equal(t, 1, config.Version)
			assert.Len(t, config.Formats, len(tt.formats))
			assert.Empty(t, config.Rules)

			// Verify all formats are enabled
			for i, format := range config.Formats {
				assert.Equal(t, tt.formats[i], format.Type)
				assert.True(t, format.Enabled)
			}

			// Verify config file was created
			configPath := filepath.Join(tempDir, tt.expectPath)
			exists, err := afero.Exists(fs, configPath)
			require.NoError(t, err)
			assert.True(t, exists)
		})
	}
}

func TestManager_LoadConfig(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name           string
		setupConfig    func(afero.Fs, string)
		expectedLoc    domain.ConfigLocation
		expectedRules  int
		expectedFormat domain.FormatType
		expectError    bool
	}{
		{
			name: "load from root location",
			setupConfig: func(fs afero.Fs, dir string) {
				config := &domain.Project{
					Version: 1,
					Formats: []domain.FormatConfig{
						{Type: domain.FormatClaude, Enabled: true},
					},
					Rules: []domain.RuleRef{
						{ID: "[contexture:test/rule]"},
					},
				}
				manager := NewManager(fs)
				_ = manager.SaveConfig(config, domain.ConfigLocationRoot, dir)
			},
			expectedLoc:    domain.ConfigLocationRoot,
			expectedRules:  1,
			expectedFormat: domain.FormatClaude,
			expectError:    false,
		},
		{
			name: "load from contexture directory",
			setupConfig: func(fs afero.Fs, dir string) {
				config := &domain.Project{
					Version: 1,
					Formats: []domain.FormatConfig{
						{Type: domain.FormatCursor, Enabled: true},
					},
					Rules: []domain.RuleRef{},
				}
				manager := NewManager(fs)
				_ = manager.SaveConfig(config, domain.ConfigLocationContexture, dir)
			},
			expectedLoc:    domain.ConfigLocationContexture,
			expectedRules:  0,
			expectedFormat: domain.FormatCursor,
			expectError:    false,
		},
		{
			name: "no config file found",
			setupConfig: func(_ afero.Fs, _ string) {
				// Don't create any config
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := afero.NewMemMapFs()
			tempDir := "/tmp/test-load"
			_ = fs.MkdirAll(tempDir, 0o755)

			tt.setupConfig(fs, tempDir)

			manager := NewManager(fs)
			result, err := manager.LoadConfig(tempDir)

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.NotNil(t, result)
			assert.Equal(t, tt.expectedLoc, result.Location)
			assert.Equal(t, 1, result.Config.Version)
			assert.Len(t, result.Config.Rules, tt.expectedRules)
			assert.Len(t, result.Config.Formats, 1)
			assert.Equal(t, tt.expectedFormat, result.Config.Formats[0].Type)
		})
	}
}

func TestManager_AddRule(t *testing.T) {
	t.Parallel()
	fs := afero.NewMemMapFs()
	manager := NewManager(fs)

	config := &domain.Project{
		Version: 1,
		Formats: []domain.FormatConfig{
			{Type: domain.FormatClaude, Enabled: true},
		},
		Rules: []domain.RuleRef{},
	}

	// Add first rule
	rule1 := domain.RuleRef{
		ID:     "[contexture:test/rule1]",
		Source: "https://github.com/contextureai/rules.git",
		Ref:    "main",
	}

	err := manager.AddRule(config, rule1)
	require.NoError(t, err)
	assert.Len(t, config.Rules, 1)
	assert.Equal(t, rule1.ID, config.Rules[0].ID)

	// Add second rule
	rule2 := domain.RuleRef{
		ID:     "[contexture:test/rule2]",
		Source: "https://github.com/contextureai/rules.git",
		Ref:    "main",
	}

	err = manager.AddRule(config, rule2)
	require.NoError(t, err)
	assert.Len(t, config.Rules, 2)

	// Update existing rule (should replace, not add)
	rule1Updated := domain.RuleRef{
		ID:     "[contexture:test/rule1]",
		Source: "https://github.com/custom/rules.git",
		Ref:    "develop",
	}

	err = manager.AddRule(config, rule1Updated)
	require.NoError(t, err)
	assert.Len(t, config.Rules, 2) // Should still be 2

	// Find the updated rule
	var foundRule *domain.RuleRef
	for _, r := range config.Rules {
		if r.ID == rule1Updated.ID {
			foundRule = &r
			break
		}
	}
	require.NotNil(t, foundRule)
	assert.Equal(t, "https://github.com/custom/rules.git", foundRule.Source)
	assert.Equal(t, "develop", foundRule.Ref)
}

func TestManager_RemoveRule(t *testing.T) {
	t.Parallel()
	fs := afero.NewMemMapFs()
	manager := NewManager(fs)

	config := &domain.Project{
		Version: 1,
		Formats: []domain.FormatConfig{
			{Type: domain.FormatClaude, Enabled: true},
		},
		Rules: []domain.RuleRef{
			{ID: "[contexture:test/rule1]"},
			{ID: "[contexture:test/rule2]"},
			{ID: "[contexture:test/rule3]"},
		},
	}

	// Remove middle rule
	err := manager.RemoveRule(config, "[contexture:test/rule2]")
	require.NoError(t, err)
	assert.Len(t, config.Rules, 2)

	// Verify correct rules remain
	ruleIDs := make([]string, len(config.Rules))
	for i, rule := range config.Rules {
		ruleIDs[i] = rule.ID
	}
	assert.Contains(t, ruleIDs, "[contexture:test/rule1]")
	assert.Contains(t, ruleIDs, "[contexture:test/rule3]")
	assert.NotContains(t, ruleIDs, "[contexture:test/rule2]")

	// Remove non-existent rule
	err = manager.RemoveRule(config, "[contexture:test/nonexistent]")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "rule not found")
}

func TestManager_HasRule(t *testing.T) {
	t.Parallel()
	fs := afero.NewMemMapFs()
	manager := NewManager(fs)

	config := &domain.Project{
		Rules: []domain.RuleRef{
			{ID: "[contexture:test/rule1]"},
			{ID: "[contexture:test/rule2]"},
			{ID: "[contexture(git@github.com:user/custom-rules):custom/rule,branch]"},
		},
	}

	// Test exact matching
	assert.True(t, manager.HasRule(config, "[contexture:test/rule1]"))
	assert.True(t, manager.HasRule(config, "[contexture:test/rule2]"))
	assert.True(t, manager.HasRule(config, "[contexture(git@github.com:user/custom-rules):custom/rule,branch]"))
	assert.False(t, manager.HasRule(config, "[contexture:test/nonexistent]"))

	// Test short ID matching (the bug we fixed)
	assert.True(t, manager.HasRule(config, "test/rule1"))
	assert.True(t, manager.HasRule(config, "test/rule2"))
	assert.True(t, manager.HasRule(config, "custom/rule"))
	assert.False(t, manager.HasRule(config, "nonexistent/rule"))
}

func TestManager_GetConfigLocation(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name             string
		preferContexture bool
		createContexture bool
		expected         domain.ConfigLocation
	}{
		{
			name:             "prefer contexture when specified",
			preferContexture: true,
			createContexture: false,
			expected:         domain.ConfigLocationContexture,
		},
		{
			name:             "use contexture when directory exists",
			preferContexture: false,
			createContexture: true,
			expected:         domain.ConfigLocationContexture,
		},
		{
			name:             "default to root when no preference and no directory",
			preferContexture: false,
			createContexture: false,
			expected:         domain.ConfigLocationRoot,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := afero.NewMemMapFs()
			tempDir := "/tmp/test-location"
			_ = fs.MkdirAll(tempDir, 0o755)

			if tt.createContexture {
				contextureDir := filepath.Join(tempDir, domain.GetContextureDir())
				_ = fs.MkdirAll(contextureDir, 0o755)
			}

			manager := NewManager(fs)
			location := manager.GetConfigLocation(tempDir, tt.preferContexture)
			assert.Equal(t, tt.expected, location)
		})
	}
}

func TestManager_SaveConfig(t *testing.T) {
	t.Parallel()
	fs := afero.NewMemMapFs()
	tempDir := "/tmp/test-save"
	_ = fs.MkdirAll(tempDir, 0o755)

	manager := NewManager(fs)

	config := &domain.Project{
		Version: 1,
		Formats: []domain.FormatConfig{
			{Type: domain.FormatClaude, Enabled: true},
		},
		Rules: []domain.RuleRef{
			{ID: "[contexture:test/rule]"},
		},
	}

	// Test saving to root
	err := manager.SaveConfig(config, domain.ConfigLocationRoot, tempDir)
	require.NoError(t, err)

	rootPath := filepath.Join(tempDir, domain.ConfigFile)
	exists, err := afero.Exists(fs, rootPath)
	require.NoError(t, err)
	assert.True(t, exists)

	// Test saving to contexture directory
	err = manager.SaveConfig(config, domain.ConfigLocationContexture, tempDir)
	require.NoError(t, err)

	contexturePath := filepath.Join(tempDir, domain.GetContextureDir(), domain.ConfigFile)
	exists, err = afero.Exists(fs, contexturePath)
	require.NoError(t, err)
	assert.True(t, exists)
}

func TestManager_DiscoverLocalRules(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name           string
		configLocation domain.ConfigLocation
		setupFiles     map[string]string // file path -> content
		expectedRules  []string          // expected rule IDs
	}{
		{
			name:           "no rules directory",
			configLocation: domain.ConfigLocationRoot,
			setupFiles:     map[string]string{},
			expectedRules:  nil,
		},
		{
			name:           "rules in project root structure",
			configLocation: domain.ConfigLocationRoot,
			setupFiles: map[string]string{
				"rules/security/auth.md":     "# Auth Rule\nTest auth rule",
				"rules/performance/cache.md": "# Cache Rule\nTest cache rule",
				"rules/typescript/strict.md": "# TypeScript Rule\nStrict mode rule",
			},
			expectedRules: []string{"security/auth", "performance/cache", "typescript/strict"},
		},
		{
			name:           "rules in contexture directory structure",
			configLocation: domain.ConfigLocationContexture,
			setupFiles: map[string]string{
				".contexture/rules/security/auth.md":     "# Auth Rule\nTest auth rule",
				".contexture/rules/performance/cache.md": "# Cache Rule\nTest cache rule",
				".contexture/rules/ui/components.md":     "# UI Rule\nComponent rule",
			},
			expectedRules: []string{"security/auth", "performance/cache", "ui/components"},
		},
		{
			name:           "mixed files in rules directory",
			configLocation: domain.ConfigLocationRoot,
			setupFiles: map[string]string{
				"rules/security/auth.md":      "# Auth Rule\nTest auth rule",
				"rules/security/readme.txt":   "This is not a rule",
				"rules/performance/cache.md":  "# Cache Rule\nTest cache rule",
				"rules/performance/notes.doc": "Notes",
				"rules/typescript/strict.md":  "# TypeScript Rule\nStrict mode rule",
			},
			expectedRules: []string{"security/auth", "performance/cache", "typescript/strict"},
		},
		{
			name:           "single level rules",
			configLocation: domain.ConfigLocationRoot,
			setupFiles: map[string]string{
				"rules/auth.md":   "# Auth Rule\nTest auth rule",
				"rules/cache.md":  "# Cache Rule\nTest cache rule",
				"rules/strict.md": "# TypeScript Rule\nStrict mode rule",
			},
			expectedRules: []string{"auth", "cache", "strict"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := afero.NewMemMapFs()
			manager := NewManager(fs)

			basePath := "/test/project"

			// Create config file
			config := &domain.Project{
				Version: 1,
				Formats: []domain.FormatConfig{{Type: domain.FormatClaude, Enabled: true}},
				Rules:   []domain.RuleRef{{ID: "[contexture:existing/rule]"}},
			}
			err := manager.SaveConfig(config, tt.configLocation, basePath)
			require.NoError(t, err)

			// Setup test files
			for filePath, content := range tt.setupFiles {
				fullPath := filepath.Join(basePath, filePath)
				err := fs.MkdirAll(filepath.Dir(fullPath), 0o755)
				require.NoError(t, err)
				err = afero.WriteFile(fs, fullPath, []byte(content), 0o644)
				require.NoError(t, err)
			}

			// Load config to get ConfigResult
			configResult, err := manager.LoadConfig(basePath)
			require.NoError(t, err)

			// Test DiscoverLocalRules
			localRules, err := manager.DiscoverLocalRules(configResult)
			require.NoError(t, err)

			if tt.expectedRules == nil {
				assert.Nil(t, localRules)
			} else {
				assert.Len(t, localRules, len(tt.expectedRules))

				// Check that all expected rules are found
				actualRuleIDs := make([]string, len(localRules))
				for i, rule := range localRules {
					actualRuleIDs[i] = rule.ID
					assert.Equal(t, "local", rule.Source, "Rule source should be 'local'")
				}

				for _, expectedRule := range tt.expectedRules {
					assert.Contains(t, actualRuleIDs, expectedRule, "Expected rule not found: %s", expectedRule)
				}
			}
		})
	}
}

func TestManager_LoadConfigWithLocalRules(t *testing.T) {
	t.Parallel()
	fs := afero.NewMemMapFs()
	manager := NewManager(fs)

	basePath := "/test/project"

	// Create config file with existing remote rules
	config := &domain.Project{
		Version: 1,
		Formats: []domain.FormatConfig{{Type: domain.FormatClaude, Enabled: true}},
		Rules:   []domain.RuleRef{{ID: "[contexture:remote/rule]"}},
	}
	err := manager.SaveConfig(config, domain.ConfigLocationRoot, basePath)
	require.NoError(t, err)

	// Create local rules
	localRulesFiles := map[string]string{
		"rules/security/auth.md":     "# Auth Rule\nTest auth rule",
		"rules/performance/cache.md": "# Cache Rule\nTest cache rule",
	}

	for filePath, content := range localRulesFiles {
		fullPath := filepath.Join(basePath, filePath)
		err := fs.MkdirAll(filepath.Dir(fullPath), 0o755)
		require.NoError(t, err)
		err = afero.WriteFile(fs, fullPath, []byte(content), 0o644)
		require.NoError(t, err)
	}

	// Test LoadConfigWithLocalRules
	configResult, err := manager.LoadConfigWithLocalRules(basePath)
	require.NoError(t, err)

	// Should have original rule plus local rules
	assert.Len(t, configResult.Config.Rules, 3)

	// Check that we have both remote and local rules
	ruleIDs := make([]string, len(configResult.Config.Rules))
	sources := make([]string, len(configResult.Config.Rules))
	for i, rule := range configResult.Config.Rules {
		ruleIDs[i] = rule.ID
		sources[i] = rule.Source
	}

	// Should contain the original remote rule
	assert.Contains(t, ruleIDs, "[contexture:remote/rule]")

	// Should contain the local rules
	assert.Contains(t, ruleIDs, "security/auth")
	assert.Contains(t, ruleIDs, "performance/cache")

	// Check sources
	localRuleCount := 0
	for _, source := range sources {
		if source == "local" {
			localRuleCount++
		}
	}
	assert.Equal(t, 2, localRuleCount, "Should have 2 local rules")
}
