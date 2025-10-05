package rule

import (
	"context"
	"testing"

	"github.com/contextureai/contexture/internal/domain"
	"github.com/contextureai/contexture/internal/git"
	"github.com/contextureai/contexture/internal/provider"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestNewFetcher(t *testing.T) {
	t.Parallel()
	fs := afero.NewMemMapFs()
	mockRepo := git.NewMockRepository(t)

	config := FetcherConfig{
		DefaultURL: "https://github.com/test/repo.git",
	}

	fetcher := NewFetcher(fs, mockRepo, config, provider.NewRegistry())

	assert.NotNil(t, fetcher)
	// Verify it implements the interface (fetcher is already of type Fetcher)
	assert.Implements(t, (*Fetcher)(nil), fetcher)
}

func TestNewFetcherWithDefaults(t *testing.T) {
	t.Parallel()
	fs := afero.NewMemMapFs()
	mockRepo := git.NewMockRepository(t)

	fetcher := NewFetcher(fs, mockRepo, FetcherConfig{}, provider.NewRegistry())

	// Verify it implements the interface (fetcher is already of type Fetcher)
	assert.Implements(t, (*Fetcher)(nil), fetcher)
}

func TestGitFetcher_ParseRuleID(t *testing.T) {
	t.Parallel()
	fs := afero.NewMemMapFs()
	mockRepo := git.NewMockRepository(t)
	fetcher := NewFetcher(fs, mockRepo, FetcherConfig{
		DefaultURL: "https://github.com/contextureai/rules.git",
	}, provider.NewRegistry())

	tests := []struct {
		name          string
		ruleID        string
		wantErr       bool
		wantSource    string
		wantPath      string
		wantRef       string
		wantVariables map[string]any
	}{
		{
			name:       "basic rule ID",
			ruleID:     "[contexture:security/input-validation]",
			wantErr:    false,
			wantSource: "https://github.com/contextureai/rules.git",
			wantPath:   "security/input-validation",
			wantRef:    "main",
		},
		{
			name:       "rule ID with custom source",
			ruleID:     "[contexture(https://github.com/custom/repo.git):security/auth]",
			wantErr:    false,
			wantSource: "https://github.com/custom/repo.git",
			wantPath:   "security/auth",
			wantRef:    "main",
		},
		{
			name:       "rule ID with branch",
			ruleID:     "[contexture:typescript/strict,develop]",
			wantErr:    false,
			wantSource: "https://github.com/contextureai/rules.git",
			wantPath:   "typescript/strict",
			wantRef:    "develop",
		},
		{
			name:       "rule ID with custom source and branch",
			ruleID:     "[contexture(git@gitlab.com:company/rules.git):custom/rule,v1.2.3]",
			wantErr:    false,
			wantSource: "git@gitlab.com:company/rules.git",
			wantPath:   "custom/rule",
			wantRef:    "v1.2.3",
		},
		{
			name:    "invalid format",
			ruleID:  "invalid!@#$%^&*()rule",
			wantErr: true,
		},
		{
			name:    "empty rule ID",
			ruleID:  "",
			wantErr: true,
		},
		{
			name:       "rule ID with simple variables",
			ruleID:     `[contexture:database/config]{port: 5432, host: "localhost"}`,
			wantErr:    false,
			wantSource: "https://github.com/contextureai/rules.git",
			wantPath:   "database/config",
			wantRef:    "main",
			wantVariables: map[string]any{
				"port": float64(5432), // JSON numbers are parsed as float64
				"host": "localhost",
			},
		},
		{
			name:       "rule ID with complex variables",
			ruleID:     `[contexture:templates/readme]{project_name: "MyApp", features: ["auth", "logging"], config: {debug: true, level: "info"}}`,
			wantErr:    false,
			wantSource: "https://github.com/contextureai/rules.git",
			wantPath:   "templates/readme",
			wantRef:    "main",
			wantVariables: map[string]any{
				"project_name": "MyApp",
				"features":     []any{"auth", "logging"},
				"config": map[string]any{
					"debug": true,
					"level": "info",
				},
			},
		},
		{
			name:       "rule ID with variables and custom source",
			ruleID:     `[contexture(https://github.com/custom/repo.git):security/auth]{log_level: "debug", timeout: 30}`,
			wantErr:    false,
			wantSource: "https://github.com/custom/repo.git",
			wantPath:   "security/auth",
			wantRef:    "main",
			wantVariables: map[string]any{
				"log_level": "debug",
				"timeout":   float64(30),
			},
		},
		{
			name:       "rule ID with variables and branch",
			ruleID:     `[contexture:typescript/strict,v2.0.0]{target: "es2022", strict: true}`,
			wantErr:    false,
			wantSource: "https://github.com/contextureai/rules.git",
			wantPath:   "typescript/strict",
			wantRef:    "v2.0.0",
			wantVariables: map[string]any{
				"target": "es2022",
				"strict": true,
			},
		},
		{
			name:       "rule ID with everything",
			ruleID:     `[contexture(git@gitlab.com:company/rules.git):custom/rule,develop]{env: "staging", features: ["auth", "metrics"]}`,
			wantErr:    false,
			wantSource: "git@gitlab.com:company/rules.git",
			wantPath:   "custom/rule",
			wantRef:    "develop",
			wantVariables: map[string]any{
				"env":      "staging",
				"features": []any{"auth", "metrics"},
			},
		},
		{
			name:    "invalid JSON5 variables",
			ruleID:  `[contexture:test/rule]{invalid: json5 syntax}`,
			wantErr: true,
		},
		{
			name:          "empty variables object",
			ruleID:        `[contexture:test/rule]{}`,
			wantErr:       false,
			wantSource:    "https://github.com/contextureai/rules.git",
			wantPath:      "test/rule",
			wantRef:       "main",
			wantVariables: map[string]any{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsed, err := fetcher.ParseRuleID(tt.ruleID)

			if tt.wantErr {
				require.Error(t, err)
				assert.Nil(t, parsed)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, parsed)
				assert.Equal(t, tt.wantSource, parsed.Source)
				assert.Equal(t, tt.wantPath, parsed.RulePath)
				assert.Equal(t, tt.wantRef, parsed.Ref)
				if tt.wantVariables == nil {
					assert.Nil(t, parsed.Variables)
				} else {
					assert.Equal(t, tt.wantVariables, parsed.Variables)
				}
			}
		})
	}
}

func TestGitFetcher_FetchRule(t *testing.T) {
	t.Parallel()
	fs := afero.NewMemMapFs()
	mockRepo := git.NewMockRepository(t)

	fetcher := NewFetcher(fs, mockRepo, FetcherConfig{
		DefaultURL: "https://github.com/contextureai/rules.git",
	}, provider.NewRegistry())

	// Mock the Clone method to create test data in temporary directory
	mockRepo.On("Clone", mock.Anything, "https://github.com/contextureai/rules.git", mock.AnythingOfType("string"), mock.AnythingOfType("[]git.CloneOption")).
		Run(func(args mock.Arguments) {
			tempPath := args.Get(2).(string)
			// Create test data structure in the cloned repo
			_ = fs.MkdirAll(tempPath+"/core/security", 0o755)
			_ = afero.WriteFile(
				fs,
				tempPath+"/core/security/input-validation.md",
				[]byte("---\ntitle: Test Rule\ndescription: This is a test rule description\n"+
					"tags:\n  - test\n  - example\n---\n\n# Test Rule\nThis is a test rule."),
				0o644,
			)
		}).
		Return(nil)

	ctx := context.Background()
	ruleID := "[contexture:core/security/input-validation]"

	rule, err := fetcher.FetchRule(ctx, ruleID)

	require.NoError(t, err)
	assert.NotNil(t, rule)
	assert.Equal(t, ruleID, rule.ID)
	assert.Contains(t, rule.Content, "# Test Rule")
	assert.Equal(t, "https://github.com/contextureai/rules.git", rule.Source)
	assert.Equal(t, "main", rule.Ref)
}

func TestGitFetcher_FetchRule_NotFound(t *testing.T) {
	t.Parallel()
	fs := afero.NewMemMapFs()
	mockRepo := git.NewMockRepository(t)

	fetcher := NewFetcher(fs, mockRepo, FetcherConfig{
		DefaultURL: "https://github.com/contextureai/rules.git",
	}, provider.NewRegistry())

	// Mock the Clone method to create empty directory structure (no rule file)
	mockRepo.On("Clone", mock.Anything, "https://github.com/contextureai/rules.git", mock.AnythingOfType("string"), mock.AnythingOfType("[]git.CloneOption")).
		Run(func(args mock.Arguments) {
			tempPath := args.Get(2).(string)
			// Create empty core directory structure
			_ = fs.MkdirAll(tempPath+"/core", 0o755)
		}).
		Return(nil)

	ctx := context.Background()
	ruleID := "[contexture:nonexistent/rule]"

	rule, err := fetcher.FetchRule(ctx, ruleID)

	require.Error(t, err)
	assert.Nil(t, rule)
	assert.Contains(t, err.Error(), "rule not found")
}

func TestGitFetcher_FetchRules(t *testing.T) {
	t.Parallel()
	fs := afero.NewMemMapFs()
	mockRepo := git.NewMockRepository(t)

	fetcher := NewFetcher(fs, mockRepo, FetcherConfig{
		DefaultURL: "https://github.com/contextureai/rules.git",
	}, provider.NewRegistry())

	// Mock the Clone method to create test data for each rule fetch (called multiple times)
	mockRepo.On("Clone", mock.Anything, "https://github.com/contextureai/rules.git", mock.AnythingOfType("string"), mock.AnythingOfType("[]git.CloneOption")).
		Run(func(args mock.Arguments) {
			tempPath := args.Get(2).(string)
			// Create test data structure for multiple rules
			_ = fs.MkdirAll(tempPath+"/core/security", 0o755)
			_ = fs.MkdirAll(tempPath+"/core/typescript", 0o755)
			_ = afero.WriteFile(
				fs,
				tempPath+"/core/security/input-validation.md",
				[]byte("---\ntitle: Security Rule\ndescription: Security rule description\n"+
					"tags:\n  - security\n---\n\n# Security Rule"),
				0o644,
			)
			_ = afero.WriteFile(
				fs,
				tempPath+"/core/typescript/strict-mode.md",
				[]byte("---\ntitle: TypeScript Rule\ndescription: TypeScript rule description\n"+
					"tags:\n  - typescript\n---\n\n# TypeScript Rule"),
				0o644,
			)
		}).
		Return(nil)

	ctx := context.Background()
	ruleIDs := []string{
		"[contexture:core/security/input-validation]",
		"[contexture:core/typescript/strict-mode]",
	}

	rules, err := fetcher.FetchRules(ctx, ruleIDs)

	require.NoError(t, err)
	assert.Len(t, rules, 2)

	// Verify both rules were fetched
	ruleMap := make(map[string]*domain.Rule)
	for _, rule := range rules {
		ruleMap[rule.ID] = rule
	}

	assert.Contains(t, ruleMap, "[contexture:core/security/input-validation]")
	assert.Contains(t, ruleMap, "[contexture:core/typescript/strict-mode]")
}

func TestGitFetcher_ListAvailableRules(t *testing.T) {
	t.Parallel()
	fs := afero.NewMemMapFs()
	mockRepo := git.NewMockRepository(t)

	fetcher := NewFetcher(fs, mockRepo, FetcherConfig{
		DefaultURL: "https://github.com/contextureai/rules.git",
	}, provider.NewRegistry())

	// Mock the Clone method to create test repository structure
	mockRepo.On("Clone", mock.Anything, "https://github.com/contextureai/rules.git", mock.AnythingOfType("string"), mock.AnythingOfType("[]git.CloneOption")).
		Run(func(args mock.Arguments) {
			tempPath := args.Get(2).(string)
			// Create test rule structure
			_ = fs.MkdirAll(tempPath+"/core/security", 0o755)
			_ = fs.MkdirAll(tempPath+"/core/typescript", 0o755)
			_ = fs.MkdirAll(tempPath+"/core/go/best-practices", 0o755)

			// Create rule files
			_ = afero.WriteFile(
				fs,
				tempPath+"/core/security/input-validation.md",
				[]byte("content"),
				0o644,
			)
			_ = afero.WriteFile(
				fs,
				tempPath+"/core/security/authentication.md",
				[]byte("content"),
				0o644,
			)
			_ = afero.WriteFile(
				fs,
				tempPath+"/core/typescript/strict-mode.md",
				[]byte("content"),
				0o644,
			)
			_ = afero.WriteFile(
				fs,
				tempPath+"/core/go/best-practices/error-handling.md",
				[]byte("content"),
				0o644,
			)

			// Create non-rule files (should be ignored)
			_ = afero.WriteFile(fs, tempPath+"/core/README.md", []byte("readme"), 0o644)
			_ = afero.WriteFile(fs, tempPath+"/core/security/notes.txt", []byte("notes"), 0o644)
		}).
		Return(nil)

	ctx := context.Background()
	rules, err := fetcher.ListAvailableRules(ctx, "", "")

	require.NoError(t, err)
	assert.Len(t, rules, 4) // README.md is now excluded

	expected := []string{
		"core/security/input-validation",
		"core/security/authentication",
		"core/typescript/strict-mode",
		"core/go/best-practices/error-handling",
	}

	for _, expectedRule := range expected {
		assert.Contains(t, rules, expectedRule)
	}
}

func TestLocalFetcher(t *testing.T) {
	t.Parallel()
	fs := afero.NewMemMapFs()
	baseDir := "/local/rules"

	// Create test rules with proper frontmatter
	_ = fs.MkdirAll(baseDir+"/security", 0o755)
	testRuleContent := `---
title: Local Test Rule
description: A test rule for local fetcher
tags: [test, local]
---

# Local Test Rule

This is a test rule content.`
	_ = afero.WriteFile(fs, baseDir+"/security/test-rule.md", []byte(testRuleContent), 0o644)

	fetcher := NewLocalFetcher(fs, baseDir)

	t.Run("fetch local rule", func(t *testing.T) {
		ctx := context.Background()
		rule, err := fetcher.FetchRule(ctx, "security/test-rule")

		require.NoError(t, err)
		assert.NotNil(t, rule)
		assert.Equal(t, "[contexture(local):security/test-rule]", rule.ID)
		assert.Equal(t, "Local Test Rule", rule.Title)
		assert.Equal(t, "A test rule for local fetcher", rule.Description)
		assert.Equal(t, []string{"test", "local"}, rule.Tags)
		assert.Contains(t, rule.Content, "This is a test rule content.")
		assert.Equal(t, "local", rule.Source)
	})

	t.Run("list local rules", func(t *testing.T) {
		ctx := context.Background()
		rules, err := fetcher.ListAvailableRules(ctx, "", "")

		require.NoError(t, err)
		assert.Len(t, rules, 1)
		assert.Equal(t, "security/test-rule", rules[0])
	})
}

func TestExtractRuleIDsFromContent(t *testing.T) {
	t.Parallel()
	content := `
# Template Content

This template includes several rules:

[contexture:security/input-validation]

Some more content...

[contexture(https://github.com/custom/repo.git):typescript/strict-mode,develop]

And another rule:

[contexture:go/error-handling]

# Duplicate rule (should only appear once)
[contexture:security/input-validation]
`

	ruleIDs := ExtractRuleIDsFromContent(content)

	assert.Len(t, ruleIDs, 3)

	expected := []string{
		"[contexture:security/input-validation]",
		"[contexture(https://github.com/custom/repo.git):typescript/strict-mode,develop]",
		"[contexture:go/error-handling]",
	}

	for _, expectedID := range expected {
		assert.Contains(t, ruleIDs, expectedID)
	}
}

func TestLocalFetcher_ParseRuleID(t *testing.T) {
	t.Parallel()
	fs := afero.NewMemMapFs()
	fetcher := NewLocalFetcher(fs, "/local/rules")

	tests := []struct {
		name          string
		ruleID        string
		wantErr       bool
		wantPath      string
		wantVariables map[string]any
	}{
		{
			name:     "simple path",
			ruleID:   "security/input-validation",
			wantErr:  false,
			wantPath: "security/input-validation",
		},
		{
			name:     "path with variables",
			ruleID:   `security/config{port: 8080, host: "localhost"}`,
			wantErr:  false,
			wantPath: "security/config",
			wantVariables: map[string]any{
				"port": float64(8080),
				"host": "localhost",
			},
		},
		{
			name:     "path with complex variables",
			ruleID:   `templates/docker{services: ["web", "db"], config: {memory: "1g", cpu: 2}}`,
			wantErr:  false,
			wantPath: "templates/docker",
			wantVariables: map[string]any{
				"services": []any{"web", "db"},
				"config": map[string]any{
					"memory": "1g",
					"cpu":    float64(2),
				},
			},
		},
		{
			name:          "empty variables object",
			ruleID:        "test/rule{}",
			wantErr:       false,
			wantPath:      "test/rule",
			wantVariables: map[string]any{},
		},
		{
			name:    "invalid JSON5 variables",
			ruleID:  "test/rule{invalid: syntax",
			wantErr: true,
		},
		{
			name:    "empty rule ID",
			ruleID:  "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsed, err := fetcher.ParseRuleID(tt.ruleID)

			if tt.wantErr {
				require.Error(t, err)
				assert.Nil(t, parsed)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, parsed)
				assert.Equal(t, tt.wantPath, parsed.RulePath)
				assert.Equal(t, "local", parsed.Source)
				assert.Empty(t, parsed.Ref)
				if tt.wantVariables == nil {
					assert.Nil(t, parsed.Variables)
				} else {
					assert.Equal(t, tt.wantVariables, parsed.Variables)
				}
			}
		})
	}
}
