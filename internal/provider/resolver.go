package provider

import (
	"fmt"
	"strings"

	"github.com/contextureai/contexture/internal/domain"
)

// Resolver handles resolution of provider references to URLs
type Resolver struct {
	registry *Registry
}

// NewResolver creates a new provider resolver
func NewResolver(registry *Registry) *Resolver {
	return &Resolver{
		registry: registry,
	}
}

// ResolveRuleRef resolves a RuleRef's source to a concrete URL
// Handles: @provider, provider name, direct URL
func (r *Resolver) ResolveRuleRef(ruleRef *domain.RuleRef) (string, error) {
	if ruleRef == nil {
		return "", fmt.Errorf("rule ref cannot be nil")
	}

	source := ruleRef.GetSource()

	// Check if source starts with @ (provider reference)
	if strings.HasPrefix(source, "@") {
		providerName := strings.TrimPrefix(source, "@")
		return r.registry.Resolve(providerName)
	}

	// Check if it's a direct URL (http/https/git@)
	if strings.HasPrefix(source, "http://") ||
		strings.HasPrefix(source, "https://") ||
		strings.HasPrefix(source, "git@") {
		return source, nil
	}

	// Try to resolve as provider name
	url, err := r.registry.Resolve(source)
	if err == nil {
		return url, nil
	}

	// If not found and not a URL, return error
	return "", fmt.Errorf("could not resolve source '%s': not a provider name or URL", source)
}

// ResolveProviderName resolves a provider name (with or without @) to a URL
func (r *Resolver) ResolveProviderName(name string) (string, error) {
	// Strip @ prefix if present
	providerName := strings.TrimPrefix(name, "@")
	return r.registry.Resolve(providerName)
}
