package format

import (
	"fmt"
	"testing"

	"github.com/contextureai/contexture/internal/domain"
	"github.com/contextureai/contexture/internal/format/claude"
	"github.com/contextureai/contexture/internal/format/cursor"
	"github.com/contextureai/contexture/internal/format/windsurf"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewBuilder(t *testing.T) {
	builder := NewBuilder()

	assert.NotNil(t, builder)
	assert.NotNil(t, builder.constructors)

	// Check that built-in formats are registered
	supportedFormats := builder.GetSupportedFormats()
	assert.Contains(t, supportedFormats, domain.FormatClaude)
	assert.Contains(t, supportedFormats, domain.FormatCursor)
	assert.Contains(t, supportedFormats, domain.FormatWindsurf)
}

func TestBuilder_Register(t *testing.T) {
	builder := NewBuilder()

	// Register a custom format
	customFormat := domain.FormatType("custom")
	constructor := func(_ afero.Fs, _ map[string]any) (domain.Format, error) {
		return nil, fmt.Errorf("mock constructor")
	}

	builder.Register(customFormat, constructor)

	// Check that it's registered
	supportedFormats := builder.GetSupportedFormats()
	assert.Contains(t, supportedFormats, customFormat)
}

func TestBuilder_Build(t *testing.T) {
	builder := NewBuilder()
	fs := afero.NewMemMapFs()

	tests := []struct {
		name       string
		formatType domain.FormatType
		options    map[string]any
		wantErr    bool
	}{
		{
			name:       "claude format",
			formatType: domain.FormatClaude,
			options:    nil,
			wantErr:    false,
		},
		{
			name:       "cursor format",
			formatType: domain.FormatCursor,
			options:    nil,
			wantErr:    false,
		},
		{
			name:       "windsurf format",
			formatType: domain.FormatWindsurf,
			options:    nil,
			wantErr:    false,
		},
		{
			name:       "windsurf with single file mode",
			formatType: domain.FormatWindsurf,
			options:    map[string]any{"mode": "single-file"},
			wantErr:    false,
		},
		{
			name:       "windsurf with multi file mode",
			formatType: domain.FormatWindsurf,
			options:    map[string]any{"mode": "multi-file"},
			wantErr:    false,
		},
		{
			name:       "unsupported format",
			formatType: domain.FormatType("unsupported"),
			options:    nil,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			format, err := builder.Build(tt.formatType, fs, tt.options)

			if tt.wantErr {
				require.Error(t, err)
				assert.Nil(t, format)
				assert.Contains(t, err.Error(), "unsupported format type")
			} else {
				require.NoError(t, err)
				assert.NotNil(t, format)
			}
		})
	}
}

func TestBuilder_GetSupportedFormats(t *testing.T) {
	builder := NewBuilder()

	formats := builder.GetSupportedFormats()

	// Should have at least the built-in formats
	assert.GreaterOrEqual(t, len(formats), 3)
	assert.Contains(t, formats, domain.FormatClaude)
	assert.Contains(t, formats, domain.FormatCursor)
	assert.Contains(t, formats, domain.FormatWindsurf)
}

func TestBuiltInConstructors(t *testing.T) {
	fs := afero.NewMemMapFs()

	t.Run("claude constructor", func(t *testing.T) {
		format, err := claude.NewFormatFromOptions(fs, nil)
		require.NoError(t, err)
		assert.NotNil(t, format)
	})

	t.Run("cursor constructor", func(t *testing.T) {
		format, err := cursor.NewFormatFromOptions(fs, nil)
		require.NoError(t, err)
		assert.NotNil(t, format)
	})

	t.Run("windsurf constructor", func(t *testing.T) {
		format, err := windsurf.NewFormatFromOptions(fs, nil)
		require.NoError(t, err)
		assert.NotNil(t, format)
	})

	t.Run("windsurf constructor with options", func(t *testing.T) {
		options := map[string]any{"mode": "single-file"}
		format, err := windsurf.NewFormatFromOptions(fs, options)
		require.NoError(t, err)
		assert.NotNil(t, format)
	})
}
