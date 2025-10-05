package provider

import (
	"testing"

	"github.com/contextureai/contexture/internal/domain"
)

func TestNewRegistry(t *testing.T) {
	registry := NewRegistry()

	if registry == nil {
		t.Fatal("expected non-nil registry")
	}

	// Should have default provider registered
	url, err := registry.Resolve("contexture")
	if err != nil {
		t.Errorf("expected default provider to be registered: %v", err)
	}

	expectedURL := domain.DefaultProviderURL
	if url != expectedURL {
		t.Errorf("expected URL %s, got %s", expectedURL, url)
	}
}

func TestRegister(t *testing.T) {
	registry := NewRegistry()

	tests := []struct {
		name        string
		provider    *domain.Provider
		expectError bool
	}{
		{
			name: "valid provider",
			provider: &domain.Provider{
				Name: "custom",
				URL:  "https://github.com/custom/rules.git",
			},
			expectError: false,
		},
		{
			name:        "nil provider",
			provider:    nil,
			expectError: true,
		},
		{
			name: "empty name",
			provider: &domain.Provider{
				Name: "",
				URL:  "https://github.com/custom/rules.git",
			},
			expectError: true,
		},
		{
			name: "empty URL",
			provider: &domain.Provider{
				Name: "custom",
				URL:  "",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := registry.Register(tt.provider)
			if tt.expectError && err == nil {
				t.Error("expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestResolve(t *testing.T) {
	registry := NewRegistry()

	// Register custom provider
	customProvider := &domain.Provider{
		Name: "mycompany",
		URL:  "https://github.com/mycompany/rules.git",
	}
	if err := registry.Register(customProvider); err != nil {
		t.Fatalf("failed to register provider: %v", err)
	}

	tests := []struct {
		name         string
		providerName string
		expectedURL  string
		expectError  bool
	}{
		{
			name:         "default provider",
			providerName: "contexture",
			expectedURL:  domain.DefaultProviderURL,
			expectError:  false,
		},
		{
			name:         "custom provider",
			providerName: "mycompany",
			expectedURL:  "https://github.com/mycompany/rules.git",
			expectError:  false,
		},
		{
			name:         "unknown provider",
			providerName: "unknown",
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url, err := registry.Resolve(tt.providerName)
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

func TestLoadFromProject(t *testing.T) {
	t.Run("valid project with providers", func(t *testing.T) {
		registry := NewRegistry()

		project := &domain.Project{
			Providers: []domain.Provider{
				{
					Name: "custom1",
					URL:  "https://github.com/custom1/rules.git",
				},
				{
					Name: "custom2",
					URL:  "https://github.com/custom2/rules.git",
				},
			},
		}

		err := registry.LoadFromProject(project)
		if err != nil {
			t.Fatalf("failed to load providers from project: %v", err)
		}

		// Verify providers were registered
		url1, err := registry.Resolve("custom1")
		if err != nil {
			t.Errorf("custom1 provider not found: %v", err)
		}
		if url1 != "https://github.com/custom1/rules.git" {
			t.Errorf("unexpected URL for custom1: %s", url1)
		}

		url2, err := registry.Resolve("custom2")
		if err != nil {
			t.Errorf("custom2 provider not found: %v", err)
		}
		if url2 != "https://github.com/custom2/rules.git" {
			t.Errorf("unexpected URL for custom2: %s", url2)
		}
	})

	t.Run("nil project", func(t *testing.T) {
		registry := NewRegistry()
		err := registry.LoadFromProject(nil)
		if err != nil {
			t.Errorf("expected no error for nil project, got: %v", err)
		}
	})

	t.Run("project with invalid provider - missing name", func(t *testing.T) {
		registry := NewRegistry()

		project := &domain.Project{
			Providers: []domain.Provider{
				{
					Name: "",
					URL:  "https://github.com/custom/rules.git",
				},
			},
		}

		err := registry.LoadFromProject(project)
		if err == nil {
			t.Error("expected error for provider with missing name")
		}
	})

	t.Run("project with invalid provider - missing URL", func(t *testing.T) {
		registry := NewRegistry()

		project := &domain.Project{
			Providers: []domain.Provider{
				{
					Name: "custom",
					URL:  "",
				},
			},
		}

		err := registry.LoadFromProject(project)
		if err == nil {
			t.Error("expected error for provider with missing URL")
		}
	})
}

func TestGet(t *testing.T) {
	registry := NewRegistry()

	customProvider := &domain.Provider{
		Name:          "mycompany",
		URL:           "https://github.com/mycompany/rules.git",
		DefaultBranch: "production",
	}
	if err := registry.Register(customProvider); err != nil {
		t.Fatalf("failed to register provider: %v", err)
	}

	provider, err := registry.Get("mycompany")
	if err != nil {
		t.Fatalf("failed to get provider: %v", err)
	}

	if provider.Name != "mycompany" {
		t.Errorf("expected name 'mycompany', got '%s'", provider.Name)
	}
	if provider.DefaultBranch != "production" {
		t.Errorf("expected branch 'production', got '%s'", provider.DefaultBranch)
	}

	_, err = registry.Get("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent provider")
	}
}

func TestList(t *testing.T) {
	registry := NewRegistry()

	// Should include default provider
	names := registry.List()
	if len(names) != 1 {
		t.Errorf("expected 1 provider, got %d", len(names))
	}

	// Add more providers
	_ = registry.Register(&domain.Provider{Name: "custom1", URL: "https://github.com/custom1/rules.git"})
	_ = registry.Register(&domain.Provider{Name: "custom2", URL: "https://github.com/custom2/rules.git"})

	names = registry.List()
	if len(names) != 3 {
		t.Errorf("expected 3 providers, got %d", len(names))
	}
}

func TestListProviders(t *testing.T) {
	registry := NewRegistry()

	// Should include default provider with all details
	providers := registry.ListProviders()
	if len(providers) != 1 {
		t.Fatalf("expected 1 provider, got %d", len(providers))
	}

	// Verify default provider details
	if providers[0].Name != "contexture" {
		t.Errorf("expected provider name 'contexture', got '%s'", providers[0].Name)
	}
	if providers[0].URL != domain.DefaultProviderURL {
		t.Errorf("expected provider URL '%s', got '%s'", domain.DefaultProviderURL, providers[0].URL)
	}

	// Add custom providers with different properties
	_ = registry.Register(&domain.Provider{
		Name:          "custom1",
		URL:           "https://github.com/custom1/rules.git",
		DefaultBranch: "main",
	})
	_ = registry.Register(&domain.Provider{
		Name:          "custom2",
		URL:           "https://github.com/custom2/rules.git",
		DefaultBranch: "production",
	})

	providers = registry.ListProviders()
	if len(providers) != 3 {
		t.Fatalf("expected 3 providers, got %d", len(providers))
	}

	// Verify all providers are returned with their details
	providerMap := make(map[string]*domain.Provider)
	for _, p := range providers {
		providerMap[p.Name] = p
	}

	if providerMap["custom1"].DefaultBranch != "main" {
		t.Errorf("expected custom1 branch 'main', got '%s'", providerMap["custom1"].DefaultBranch)
	}
	if providerMap["custom2"].DefaultBranch != "production" {
		t.Errorf("expected custom2 branch 'production', got '%s'", providerMap["custom2"].DefaultBranch)
	}
}
