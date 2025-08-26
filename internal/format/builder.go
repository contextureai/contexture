package format

import (
	"fmt"

	"github.com/contextureai/contexture/internal/domain"
	"github.com/contextureai/contexture/internal/format/claude"
	"github.com/contextureai/contexture/internal/format/cursor"
	"github.com/contextureai/contexture/internal/format/windsurf"
	"github.com/spf13/afero"
)

// Constructor is a function that creates a format implementation
type Constructor func(fs afero.Fs, options map[string]any) (domain.Format, error)

// Builder provides a more Go-idiomatic way to create format implementations
type Builder struct {
	constructors map[domain.FormatType]Constructor
}

// NewBuilder creates a new format builder with built-in formats
func NewBuilder() *Builder {
	builder := &Builder{
		constructors: make(map[domain.FormatType]Constructor),
	}

	// Register built-in format constructors
	builder.Register(domain.FormatClaude, claude.NewFormatFromOptions)
	builder.Register(domain.FormatCursor, cursor.NewFormatFromOptions)
	builder.Register(domain.FormatWindsurf, windsurf.NewFormatFromOptions)

	return builder
}

// Register adds a format constructor to the builder
func (fb *Builder) Register(formatType domain.FormatType, constructor Constructor) {
	fb.constructors[formatType] = constructor
}

// Build creates a format implementation based on the format type
func (fb *Builder) Build(
	formatType domain.FormatType,
	fs afero.Fs,
	options map[string]any,
) (domain.Format, error) {
	constructor, exists := fb.constructors[formatType]
	if !exists {
		return nil, fmt.Errorf("unsupported format type: %s", formatType)
	}

	return constructor(fs, options)
}

// GetSupportedFormats returns the list of supported format types
func (fb *Builder) GetSupportedFormats() []domain.FormatType {
	formats := make([]domain.FormatType, 0, len(fb.constructors))
	for formatType := range fb.constructors {
		formats = append(formats, formatType)
	}
	return formats
}
