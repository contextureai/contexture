package format

import (
	"slices"
	"testing"

	"github.com/charmbracelet/huh"
	"github.com/contextureai/contexture/internal/domain"
	"github.com/contextureai/contexture/internal/format/claude"
	"github.com/contextureai/contexture/internal/format/cursor"
	"github.com/contextureai/contexture/internal/format/windsurf"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRegistry(t *testing.T) {
	t.Parallel()
	fs := afero.NewMemMapFs()
	registry := NewRegistry(fs)

	assert.NotNil(t, registry)
}

func TestNewRegistryWithBuilder(t *testing.T) {
	t.Parallel()
	fs := afero.NewMemMapFs()
	customBuilder := &Builder{}
	registry := NewRegistryWithBuilder(customBuilder, fs)

	assert.NotNil(t, registry)
}

func TestGetDefaultRegistry(t *testing.T) {
	t.Parallel()
	fs := afero.NewMemMapFs()
	registry := GetDefaultRegistry(fs)

	assert.NotNil(t, registry)

	// Check that all built-in formats are registered
	formats := registry.GetAvailableFormats()
	assert.Len(t, formats, 3)

	expectedFormats := []domain.FormatType{
		domain.FormatClaude,
		domain.FormatCursor,
		domain.FormatWindsurf,
	}

	for _, expected := range expectedFormats {
		assert.Contains(t, formats, expected)
		assert.True(t, registry.IsSupported(expected))
	}
}

func TestRegistry_Register(t *testing.T) {
	t.Parallel()
	fs := afero.NewMemMapFs()
	registry := NewRegistry(fs)
	handler := &TestMockHandler{}

	registry.Register(domain.FormatClaude, handler)

	retrievedHandler, exists := registry.GetHandler(domain.FormatClaude)
	assert.True(t, exists)
	assert.Equal(t, handler, retrievedHandler)
}

func TestRegistry_GetUIOptions(t *testing.T) {
	t.Parallel()
	fs := afero.NewMemMapFs()
	registry := GetDefaultRegistry(fs)

	options := registry.GetUIOptions([]string{"claude"})
	assert.Len(t, options, 3) // claude, cursor, windsurf

	// Check that options are in the expected order
	assert.Equal(t, "claude", options[0].Value)
	assert.Equal(t, "cursor", options[1].Value)
	assert.Equal(t, "windsurf", options[2].Value)
}

func TestRegistry_GetAvailableFormats(t *testing.T) {
	t.Parallel()
	fs := afero.NewMemMapFs()
	registry := NewRegistry(fs)

	// Initially empty
	formats := registry.GetAvailableFormats()
	assert.Empty(t, formats)

	// Add a format
	registry.Register(domain.FormatClaude, &TestMockHandler{})
	formats = registry.GetAvailableFormats()
	assert.Len(t, formats, 1)
	assert.Contains(t, formats, domain.FormatClaude)
}

func TestRegistry_CreateFormat(t *testing.T) {
	t.Parallel()
	fs := afero.NewMemMapFs()
	registry := GetDefaultRegistry(fs)

	t.Run("create Claude format", func(t *testing.T) {
		format, err := registry.CreateFormat(domain.FormatClaude, fs, nil)
		require.NoError(t, err)
		assert.NotNil(t, format)
	})

	t.Run("create Cursor format", func(t *testing.T) {
		format, err := registry.CreateFormat(domain.FormatCursor, fs, nil)
		require.NoError(t, err)
		assert.NotNil(t, format)
	})

	t.Run("create Windsurf format - single file", func(t *testing.T) {
		options := map[string]any{
			"mode": string(windsurf.ModeSingleFile),
		}
		format, err := registry.CreateFormat(domain.FormatWindsurf, fs, options)
		require.NoError(t, err)
		assert.NotNil(t, format)

		windsurfFormat, ok := format.(*windsurf.Format)
		assert.True(t, ok)
		assert.Equal(t, windsurf.ModeSingleFile, windsurfFormat.GetMode())
	})

	t.Run("create Windsurf format - multi file", func(t *testing.T) {
		options := map[string]any{
			"mode": string(windsurf.ModeMultiFile),
		}
		format, err := registry.CreateFormat(domain.FormatWindsurf, fs, options)
		require.NoError(t, err)
		assert.NotNil(t, format)

		windsurfFormat, ok := format.(*windsurf.Format)
		assert.True(t, ok)
		assert.Equal(t, windsurf.ModeMultiFile, windsurfFormat.GetMode())
	})

	t.Run("unsupported format", func(t *testing.T) {
		_, err := registry.CreateFormat("unsupported", fs, nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported format type")
	})
}

func TestRegistry_CreateFormat_NoBuilder(t *testing.T) {
	t.Parallel()
	registry := &Registry{
		handlers: make(map[domain.FormatType]Handler),
		builder:  nil,
	}

	fs := afero.NewMemMapFs()
	_, err := registry.CreateFormat(domain.FormatClaude, fs, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no format builder configured")
}

func TestRegistry_GetHandler(t *testing.T) {
	t.Parallel()
	fs := afero.NewMemMapFs()
	registry := NewRegistry(fs)
	handler := &TestMockHandler{}

	// Handler not registered
	_, exists := registry.GetHandler(domain.FormatClaude)
	assert.False(t, exists)

	// Register and retrieve
	registry.Register(domain.FormatClaude, handler)
	retrievedHandler, exists := registry.GetHandler(domain.FormatClaude)
	assert.True(t, exists)
	assert.Equal(t, handler, retrievedHandler)
}

func TestRegistry_IsSupported(t *testing.T) {
	t.Parallel()
	fs := afero.NewMemMapFs()
	registry := NewRegistry(fs)

	// Not supported initially
	assert.False(t, registry.IsSupported(domain.FormatClaude))

	// Register and check
	registry.Register(domain.FormatClaude, &TestMockHandler{})
	assert.True(t, registry.IsSupported(domain.FormatClaude))

	// Other formats still not supported
	assert.False(t, registry.IsSupported(domain.FormatCursor))
}

func TestDefaultBuilder_Build(t *testing.T) {
	t.Parallel()
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
			name:       "windsurf format default",
			formatType: domain.FormatWindsurf,
			options:    nil,
			wantErr:    false,
		},
		{
			name:       "windsurf format single file",
			formatType: domain.FormatWindsurf,
			options: map[string]any{
				"mode": string(windsurf.ModeSingleFile),
			},
			wantErr: false,
		},
		{
			name:       "windsurf format multi file",
			formatType: domain.FormatWindsurf,
			options: map[string]any{
				"mode": string(windsurf.ModeMultiFile),
			},
			wantErr: false,
		},
		{
			name:       "unsupported format",
			formatType: "unsupported",
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
			} else {
				require.NoError(t, err)
				assert.NotNil(t, format)

				// Verify Windsurf mode is set correctly
				if tt.formatType == domain.FormatWindsurf && tt.options != nil {
					if mode, ok := tt.options["mode"].(string); ok {
						windsurfFormat, ok := format.(*windsurf.Format)
						require.True(t, ok)
						assert.Equal(t, windsurf.OutputMode(mode), windsurfFormat.GetMode())
					}
				}
			}
		})
	}
}

func TestDefaultBuilder_GetSupportedFormats(t *testing.T) {
	t.Parallel()
	builder := NewBuilder()

	formats := builder.GetSupportedFormats()
	assert.Len(t, formats, 3)

	expectedFormats := []domain.FormatType{
		domain.FormatClaude,
		domain.FormatCursor,
		domain.FormatWindsurf,
	}

	for _, expected := range expectedFormats {
		assert.Contains(t, formats, expected)
	}
}

func TestHandlers(t *testing.T) {
	t.Parallel()
	t.Run("claude handler", func(t *testing.T) {
		handler := &claude.Handler{}

		option := handler.GetUIOption(true)
		assert.Equal(t, "claude", option.Value)

		assert.NotEmpty(t, handler.GetDisplayName())
		assert.NotEmpty(t, handler.GetDescription())
	})

	t.Run("cursor handler", func(t *testing.T) {
		handler := &cursor.Handler{}

		option := handler.GetUIOption(false)
		assert.Equal(t, "cursor", option.Value)

		assert.NotEmpty(t, handler.GetDisplayName())
		assert.NotEmpty(t, handler.GetDescription())
	})

	t.Run("windsurf handler", func(t *testing.T) {
		handler := &windsurf.Handler{}

		option := handler.GetUIOption(true)
		assert.Equal(t, "windsurf", option.Value)

		assert.NotEmpty(t, handler.GetDisplayName())
		assert.NotEmpty(t, handler.GetDescription())
	})
}

func TestContains(t *testing.T) {
	t.Parallel()
	slice := []string{"a", "b", "c"}

	assert.True(t, slices.Contains(slice, "a"))
	assert.True(t, slices.Contains(slice, "b"))
	assert.True(t, slices.Contains(slice, "c"))
	assert.False(t, slices.Contains(slice, "d"))
	assert.False(t, slices.Contains([]string{}, "a"))
}

// Mock implementations for testing

type TestMockHandler struct{}

func (m *TestMockHandler) GetUIOption(selected bool) huh.Option[string] {
	return huh.NewOption("Mock Format", "mock").Selected(selected)
}

func (m *TestMockHandler) GetDisplayName() string {
	return "Mock Format"
}

func (m *TestMockHandler) GetDescription() string {
	return "Mock format for testing"
}

func (m *TestMockHandler) GetCapabilities() domain.FormatCapabilities {
	return domain.FormatCapabilities{
		SupportsUserRules:    false,
		UserRulesPath:        "",
		DefaultUserRulesMode: domain.UserRulesProject,
		MaxRuleSize:          0,
	}
}

type MockBuilder struct{}

func (m *MockBuilder) Build(
	_ domain.FormatType,
	_ afero.Fs,
	_ map[string]any,
) (domain.Format, error) {
	return &MockFormat{}, nil
}

func (m *MockBuilder) GetSupportedFormats() []domain.FormatType {
	return []domain.FormatType{"mock"}
}

type MockFormat struct{}

func (m *MockFormat) Transform(processedRule *domain.ProcessedRule) (*domain.TransformedRule, error) {
	return &domain.TransformedRule{Rule: processedRule.Rule, Content: "mock content"}, nil
}

func (m *MockFormat) Validate(_ *domain.Rule) (*domain.ValidationResult, error) {
	return &domain.ValidationResult{Valid: true}, nil
}

func (m *MockFormat) Write(_ []*domain.TransformedRule, _ *domain.FormatConfig) error {
	return nil
}

func (m *MockFormat) Remove(_ string, _ *domain.FormatConfig) error {
	return nil
}

func (m *MockFormat) List(_ *domain.FormatConfig) ([]*domain.InstalledRule, error) {
	return []*domain.InstalledRule{}, nil
}

func (m *MockFormat) GetOutputPath(_ *domain.FormatConfig) string {
	return "mock/path"
}

func (m *MockFormat) CleanupEmptyDirectories(_ *domain.FormatConfig) error {
	return nil
}

func (m *MockFormat) CreateDirectories(_ *domain.FormatConfig) error {
	return nil
}

func (m *MockFormat) GetMetadata() *domain.FormatMetadata {
	return &domain.FormatMetadata{
		Type:        "mock",
		DisplayName: "Mock Format",
		Description: "Mock format for testing",
		IsDirectory: false,
	}
}
