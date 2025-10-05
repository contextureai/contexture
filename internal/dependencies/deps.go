// Package dependencies provides minimal dependency injection for the application
package dependencies

import (
	"context"

	"github.com/contextureai/contexture/internal/provider"
	"github.com/spf13/afero"
)

// Dependencies holds the minimal set of dependencies needed by commands.
type Dependencies struct {
	// FS provides filesystem operations
	FS afero.Fs

	// Context for the application lifecycle
	Context context.Context

	// ProviderRegistry manages rule providers
	ProviderRegistry *provider.Registry
}

// New creates a new Dependencies instance with production defaults.
func New(ctx context.Context) *Dependencies {
	if ctx == nil {
		ctx = context.Background()
	}

	return &Dependencies{
		FS:               afero.NewOsFs(),
		Context:          ctx,
		ProviderRegistry: provider.NewRegistry(),
	}
}

// NewForTesting creates a new Dependencies instance optimized for testing.
// It uses an in-memory filesystem to avoid side effects.
func NewForTesting(ctx context.Context) *Dependencies {
	if ctx == nil {
		ctx = context.Background()
	}

	return &Dependencies{
		FS:               afero.NewMemMapFs(),
		Context:          ctx,
		ProviderRegistry: provider.NewRegistry(),
	}
}

// WithContext returns a new Dependencies instance with the given context.
// This enables proper context propagation without modifying the original.
func (d *Dependencies) WithContext(ctx context.Context) *Dependencies {
	return &Dependencies{
		FS:               d.FS,
		Context:          ctx,
		ProviderRegistry: d.ProviderRegistry,
	}
}

// WithFS returns a new Dependencies instance with the given filesystem.
// This is useful for testing or using different filesystem implementations.
func (d *Dependencies) WithFS(fs afero.Fs) *Dependencies {
	return &Dependencies{
		FS:               fs,
		Context:          d.Context,
		ProviderRegistry: d.ProviderRegistry,
	}
}

// WithProviderRegistry returns a new Dependencies instance with the given provider registry.
// This is useful for testing or using different provider configurations.
func (d *Dependencies) WithProviderRegistry(registry *provider.Registry) *Dependencies {
	return &Dependencies{
		FS:               d.FS,
		Context:          d.Context,
		ProviderRegistry: registry,
	}
}
