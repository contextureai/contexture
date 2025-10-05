// Package provider implements the provider registry system for managing
// named rule repository references.
package provider

import (
	"fmt"

	"github.com/contextureai/contexture/internal/domain"
)

// Registry manages provider name to URL mappings
type Registry struct {
	providers map[string]*domain.Provider
}

// NewRegistry creates a provider registry with default providers
func NewRegistry() *Registry {
	r := &Registry{
		providers: make(map[string]*domain.Provider),
	}
	r.registerDefaults()
	return r
}

// Register adds or updates a provider
func (r *Registry) Register(provider *domain.Provider) error {
	if provider == nil {
		return fmt.Errorf("provider cannot be nil")
	}
	if provider.Name == "" {
		return fmt.Errorf("provider name cannot be empty")
	}
	if provider.URL == "" {
		return fmt.Errorf("provider URL cannot be empty")
	}

	r.providers[provider.Name] = provider
	return nil
}

// Resolve converts a provider name to a URL
func (r *Registry) Resolve(providerName string) (string, error) {
	provider, exists := r.providers[providerName]
	if !exists {
		return "", fmt.Errorf("provider not found: %s", providerName)
	}
	return provider.URL, nil
}

// Get returns a provider by name
func (r *Registry) Get(providerName string) (*domain.Provider, error) {
	provider, exists := r.providers[providerName]
	if !exists {
		return nil, fmt.Errorf("provider not found: %s", providerName)
	}
	return provider, nil
}

// LoadFromProject registers providers from project configuration
func (r *Registry) LoadFromProject(project *domain.Project) error {
	if project == nil {
		return nil
	}

	for i := range project.Providers {
		provider := &project.Providers[i]
		if err := r.Register(provider); err != nil {
			return fmt.Errorf("failed to register provider %s: %w", provider.Name, err)
		}
	}

	return nil
}

// List returns all registered provider names
func (r *Registry) List() []string {
	names := make([]string, 0, len(r.providers))
	for name := range r.providers {
		names = append(names, name)
	}
	return names
}

// ListProviders returns all registered providers
func (r *Registry) ListProviders() []*domain.Provider {
	providers := make([]*domain.Provider, 0, len(r.providers))
	for _, provider := range r.providers {
		providers = append(providers, provider)
	}
	return providers
}

// registerDefaults adds built-in providers
func (r *Registry) registerDefaults() {
	r.providers[domain.DefaultProviderName] = &domain.Provider{
		Name:          domain.DefaultProviderName,
		URL:           domain.DefaultProviderURL,
		DefaultBranch: domain.DefaultBranch,
	}
}
