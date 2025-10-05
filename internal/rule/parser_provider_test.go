package rule

import (
	"testing"

	"github.com/contextureai/contexture/internal/domain"
	"github.com/contextureai/contexture/internal/provider"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRuleIDParser_ProviderSyntax(t *testing.T) {
	t.Parallel()

	// Create registry with custom providers
	registry := provider.NewRegistry()
	err := registry.Register(&domain.Provider{
		Name: "mycompany",
		URL:  "https://github.com/mycompany/rules.git",
	})
	require.NoError(t, err)

	parser := NewRuleIDParser(domain.DefaultRepository, registry)

	tests := []struct {
		name        string
		ruleID      string
		wantSource  string
		wantPath    string
		wantRef     string
		wantErr     bool
		errContains string
	}{
		{
			name:       "@contexture provider syntax",
			ruleID:     "@contexture/typescript/naming",
			wantSource: domain.DefaultProviderURL,
			wantPath:   "typescript/naming",
			wantRef:    "main",
			wantErr:    false,
		},
		{
			name:       "custom @provider syntax",
			ruleID:     "@mycompany/security/auth",
			wantSource: "https://github.com/mycompany/rules.git",
			wantPath:   "security/auth",
			wantRef:    "main",
			wantErr:    false,
		},
		{
			name:       "@provider with nested path",
			ruleID:     "@contexture/languages/go/testing",
			wantSource: domain.DefaultProviderURL,
			wantPath:   "languages/go/testing",
			wantRef:    "main",
			wantErr:    false,
		},
		{
			name:        "unknown provider",
			ruleID:      "@unknown/security/rule",
			wantErr:     true,
			errContains: "unknown provider",
		},
		{
			name:       "full format with @provider",
			ruleID:     "[contexture(@contexture):typescript/naming]",
			wantSource: domain.DefaultProviderURL,
			wantPath:   "typescript/naming",
			wantRef:    "main",
			wantErr:    false,
		},
		{
			name:       "full format with custom @provider",
			ruleID:     "[contexture(@mycompany):security/auth]",
			wantSource: "https://github.com/mycompany/rules.git",
			wantPath:   "security/auth",
			wantRef:    "main",
			wantErr:    false,
		},
		{
			name:       "full format with @provider and ref",
			ruleID:     "[contexture(@contexture):go/testing,v1.2.0]",
			wantSource: domain.DefaultProviderURL,
			wantPath:   "go/testing",
			wantRef:    "v1.2.0",
			wantErr:    false,
		},
		{
			name:       "simple format defaults to @contexture",
			ruleID:     "typescript/naming",
			wantSource: domain.DefaultRepository,
			wantPath:   "typescript/naming",
			wantRef:    "main",
			wantErr:    false,
		},
		{
			name:       "direct URL still works",
			ruleID:     "https://github.com/user/repo.git#path/to/rule",
			wantSource: "https://github.com/user/repo.git",
			wantPath:   "path/to/rule",
			wantRef:    "main",
			wantErr:    false,
		},
		{
			name:       "full format with direct URL",
			ruleID:     "[contexture(https://github.com/user/repo.git):path/to/rule]",
			wantSource: "https://github.com/user/repo.git",
			wantPath:   "path/to/rule",
			wantRef:    "main",
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsed, err := parser.ParseRuleID(tt.ruleID)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				return
			}

			require.NoError(t, err)
			assert.NotNil(t, parsed)
			assert.Equal(t, tt.wantSource, parsed.Source)
			assert.Equal(t, tt.wantPath, parsed.RulePath)
			assert.Equal(t, tt.wantRef, parsed.Ref)
		})
	}
}

func TestRuleIDParser_ProviderWithVariables(t *testing.T) {
	t.Parallel()

	registry := provider.NewRegistry()
	parser := NewRuleIDParser(domain.DefaultRepository, registry)

	tests := []struct {
		name          string
		ruleID        string
		wantVariables map[string]any
		wantErr       bool
	}{
		{
			name:   "full format @provider with variables",
			ruleID: "[contexture(@contexture):go/testing] {coverage: 80}",
			wantVariables: map[string]any{
				"coverage": float64(80),
			},
			wantErr: false,
		},
		{
			name:   "full format custom provider with variables",
			ruleID: "[contexture:typescript/naming] {style: \"strict\"}",
			wantVariables: map[string]any{
				"style": "strict",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsed, err := parser.ParseRuleID(tt.ruleID)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.NotNil(t, parsed)
			if tt.wantVariables != nil {
				assert.Equal(t, tt.wantVariables, parsed.Variables)
			}
		})
	}
}

func TestRuleIDParser_WithoutRegistry(t *testing.T) {
	t.Parallel()

	// Parser without registry should fall back gracefully
	parser := NewRuleIDParser(domain.DefaultRepository, nil)

	parsed, err := parser.ParseRuleID("@contexture/typescript/naming")
	require.NoError(t, err)
	assert.NotNil(t, parsed)
	// Should fallback to default URL when registry is nil
	assert.Equal(t, domain.DefaultRepository, parsed.Source)
	assert.Equal(t, "typescript/naming", parsed.RulePath)
}
