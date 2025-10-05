package provider

import (
	"testing"

	"github.com/contextureai/contexture/internal/domain"
)

func TestResolveRuleRef(t *testing.T) {
	registry := NewRegistry()
	_ = registry.Register(&domain.Provider{
		Name: "mycompany",
		URL:  "https://github.com/mycompany/rules.git",
	})

	resolver := NewResolver(registry)

	tests := []struct {
		name        string
		ruleRef     *domain.RuleRef
		expectedURL string
		expectError bool
	}{
		{
			name: "provider reference with @",
			ruleRef: &domain.RuleRef{
				ID:     "@contexture/path",
				Source: "@contexture",
			},
			expectedURL: domain.DefaultProviderURL,
			expectError: false,
		},
		{
			name: "custom provider with @",
			ruleRef: &domain.RuleRef{
				ID:     "@mycompany/path",
				Source: "@mycompany",
			},
			expectedURL: "https://github.com/mycompany/rules.git",
			expectError: false,
		},
		{
			name: "provider name without @",
			ruleRef: &domain.RuleRef{
				ID:     "path",
				Source: "contexture",
			},
			expectedURL: domain.DefaultProviderURL,
			expectError: false,
		},
		{
			name: "direct HTTPS URL",
			ruleRef: &domain.RuleRef{
				ID:     "[contexture(https://github.com/user/repo.git):path]",
				Source: "https://github.com/user/repo.git",
			},
			expectedURL: "https://github.com/user/repo.git",
			expectError: false,
		},
		{
			name: "direct git@ URL",
			ruleRef: &domain.RuleRef{
				ID:     "[contexture(git@github.com:user/repo.git):path]",
				Source: "git@github.com:user/repo.git",
			},
			expectedURL: "git@github.com:user/repo.git",
			expectError: false,
		},
		{
			name: "unknown provider",
			ruleRef: &domain.RuleRef{
				ID:     "@unknown/path",
				Source: "@unknown",
			},
			expectError: true,
		},
		{
			name:        "nil rule ref",
			ruleRef:     nil,
			expectError: true,
		},
		{
			name: "unresolvable source - not a provider or URL",
			ruleRef: &domain.RuleRef{
				ID:     "invalid-source-format",
				Source: "invalid-source",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url, err := resolver.ResolveRuleRef(tt.ruleRef)
			if tt.expectError {
				if err == nil {
					t.Error("expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if url != tt.expectedURL {
					t.Errorf("expected URL %s, got %s", tt.expectedURL, url)
				}
			}
		})
	}
}

func TestResolveProviderName(t *testing.T) {
	registry := NewRegistry()
	_ = registry.Register(&domain.Provider{
		Name: "mycompany",
		URL:  "https://github.com/mycompany/rules.git",
	})

	resolver := NewResolver(registry)

	tests := []struct {
		name         string
		providerName string
		expectedURL  string
		expectError  bool
	}{
		{
			name:         "with @ prefix",
			providerName: "@contexture",
			expectedURL:  domain.DefaultProviderURL,
			expectError:  false,
		},
		{
			name:         "without @ prefix",
			providerName: "contexture",
			expectedURL:  domain.DefaultProviderURL,
			expectError:  false,
		},
		{
			name:         "custom provider with @",
			providerName: "@mycompany",
			expectedURL:  "https://github.com/mycompany/rules.git",
			expectError:  false,
		},
		{
			name:         "custom provider without @",
			providerName: "mycompany",
			expectedURL:  "https://github.com/mycompany/rules.git",
			expectError:  false,
		},
		{
			name:         "unknown provider",
			providerName: "@unknown",
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url, err := resolver.ResolveProviderName(tt.providerName)
			if tt.expectError {
				if err == nil {
					t.Error("expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if url != tt.expectedURL {
					t.Errorf("expected URL %s, got %s", tt.expectedURL, url)
				}
			}
		})
	}
}
