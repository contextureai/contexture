// Package format provides a registry for managing available formats
package format

import (
	"slices"

	"github.com/charmbracelet/huh"
	"github.com/contextureai/contexture/internal/domain"
	contextureerrors "github.com/contextureai/contexture/internal/errors"
	"github.com/contextureai/contexture/internal/format/claude"
	"github.com/contextureai/contexture/internal/format/cursor"
	"github.com/contextureai/contexture/internal/format/windsurf"
	"github.com/spf13/afero"
)

// Handler defines the interface for format-specific UI operations
type Handler interface {
	GetUIOption(selected bool) huh.Option[string]
	GetDisplayName() string
	GetDescription() string
}

// Registry manages available formats and their implementations
type Registry struct {
	handlers map[domain.FormatType]Handler
	builder  *Builder
	dirMgr   *DirectoryManager
}

// NewRegistry creates a new format registry with afero filesystem
func NewRegistry(fs afero.Fs) *Registry {
	return &Registry{
		handlers: make(map[domain.FormatType]Handler),
		builder:  NewBuilder(),
		dirMgr:   NewDirectoryManager(fs),
	}
}

// NewRegistryWithBuilder creates a new format registry with a custom builder
func NewRegistryWithBuilder(builder *Builder, fs afero.Fs) *Registry {
	return &Registry{
		handlers: make(map[domain.FormatType]Handler),
		builder:  builder,
		dirMgr:   NewDirectoryManager(fs),
	}
}

// GetDefaultRegistry returns a registry with all built-in formats
func GetDefaultRegistry(fs afero.Fs) *Registry {
	registry := NewRegistry(fs)

	// Register built-in formats
	registry.Register(domain.FormatClaude, &claude.Handler{})
	registry.Register(domain.FormatCursor, &cursor.Handler{})
	registry.Register(domain.FormatWindsurf, &windsurf.Handler{})

	return registry
}

// Register adds a format handler to the registry
func (r *Registry) Register(formatType domain.FormatType, handler Handler) {
	r.handlers[formatType] = handler
}

// GetUIOptions returns UI options for all registered formats
func (r *Registry) GetUIOptions(selectedFormats []string) []huh.Option[string] {
	var options []huh.Option[string]

	// Add options in a consistent order
	orderedTypes := []domain.FormatType{
		domain.FormatClaude,
		domain.FormatCursor,
		domain.FormatWindsurf,
	}

	for _, formatType := range orderedTypes {
		if handler, exists := r.handlers[formatType]; exists {
			selected := slices.Contains(selectedFormats, string(formatType))
			options = append(options, handler.GetUIOption(selected))
		}
	}

	return options
}

// GetAvailableFormats returns a list of all available format types
func (r *Registry) GetAvailableFormats() []domain.FormatType {
	var formats []domain.FormatType
	for formatType := range r.handlers {
		formats = append(formats, formatType)
	}
	return formats
}

// CreateFormat creates a format implementation instance using the registry's builder.
// Returns the format instance or an error if the builder is not configured or creation fails.
func (r *Registry) CreateFormat(
	formatType domain.FormatType,
	fs afero.Fs,
	options map[string]any,
) (domain.Format, error) {
	if r.builder == nil {
		return nil, contextureerrors.WithOpf("create_format", "no format builder configured")
	}
	return r.builder.Build(formatType, fs, options)
}

// GetHandler retrieves the UI handler for a specific format type.
// Returns the handler and true if found, or nil and false if not registered.
func (r *Registry) GetHandler(formatType domain.FormatType) (Handler, bool) {
	handler, exists := r.handlers[formatType]
	return handler, exists
}

// IsSupported returns true if the format type is supported
func (r *Registry) IsSupported(formatType domain.FormatType) bool {
	_, exists := r.handlers[formatType]
	return exists
}

// CreateFormatDirectories creates necessary directories for specific formats
func (r *Registry) CreateFormatDirectories(formatType domain.FormatType) error {
	if r.dirMgr == nil {
		return contextureerrors.WithOpf("create_directories", "no directory manager configured")
	}
	return r.dirMgr.CreateFormatDirectories(formatType)
}
