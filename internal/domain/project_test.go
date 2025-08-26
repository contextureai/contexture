package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProject_GetEnabledFormats(t *testing.T) {
	t.Parallel()
	project := &Project{
		Formats: []FormatConfig{
			{Type: FormatClaude, Enabled: true},
			{Type: FormatCursor, Enabled: false},
			{Type: FormatWindsurf, Enabled: true},
		},
	}

	enabled := project.GetEnabledFormats()
	assert.Len(t, enabled, 2)
	assert.Equal(t, FormatClaude, enabled[0].Type)
	assert.Equal(t, FormatWindsurf, enabled[1].Type)
}

func TestProject_GetFormatByType(t *testing.T) {
	t.Parallel()
	project := &Project{
		Formats: []FormatConfig{
			{Type: FormatClaude, Enabled: true},
			{Type: FormatCursor, Enabled: false},
		},
	}

	t.Run("existing format", func(t *testing.T) {
		format := project.GetFormatByType(FormatClaude)
		assert.NotNil(t, format)
		assert.Equal(t, FormatClaude, format.Type)
		assert.True(t, format.Enabled)
	})

	t.Run("non-existing format", func(t *testing.T) {
		format := project.GetFormatByType(FormatWindsurf)
		assert.Nil(t, format)
	})
}

func TestProject_HasFormat(t *testing.T) {
	t.Parallel()
	project := &Project{
		Formats: []FormatConfig{
			{Type: FormatClaude, Enabled: true},
			{Type: FormatCursor, Enabled: false},
		},
	}

	assert.True(t, project.HasFormat(FormatClaude))
	assert.True(t, project.HasFormat(FormatCursor))
	assert.False(t, project.HasFormat(FormatWindsurf))
}

func TestProject_GetEnabledSources(t *testing.T) {
	t.Parallel()
	project := &Project{
		Sources: []Source{
			{Name: "enabled-source", Enabled: true},
			{Name: "disabled-source", Enabled: false},
			{Name: "another-enabled", Enabled: true},
		},
	}

	enabled := project.GetEnabledSources()
	assert.Len(t, enabled, 2)
	assert.Equal(t, "enabled-source", enabled[0].Name)
	assert.Equal(t, "another-enabled", enabled[1].Name)
}

func TestProject_GetSourceByName(t *testing.T) {
	t.Parallel()
	project := &Project{
		Sources: []Source{
			{Name: "source1", URL: "url1"},
			{Name: "source2", URL: "url2"},
		},
	}

	t.Run("existing source", func(t *testing.T) {
		source := project.GetSourceByName("source1")
		assert.NotNil(t, source)
		assert.Equal(t, "source1", source.Name)
		assert.Equal(t, "url1", source.URL)
	})

	t.Run("non-existing source", func(t *testing.T) {
		source := project.GetSourceByName("nonexistent")
		assert.Nil(t, source)
	})
}

func TestProject_GetGeneration(t *testing.T) {
	t.Parallel()
	t.Run("nil generation config", func(t *testing.T) {
		project := &Project{}
		gen := project.GetGeneration()

		assert.NotNil(t, gen)
		assert.Equal(t, 5, gen.ParallelFetches)
		assert.Equal(t, "main", gen.DefaultBranch)
		assert.True(t, gen.CacheEnabled)
		assert.Equal(t, "5m", gen.CacheTTL)
	})

	t.Run("existing generation config with defaults", func(t *testing.T) {
		project := &Project{
			Generation: &GenerationConfig{
				ParallelFetches: 10,
				CacheEnabled:    false,
			},
		}
		gen := project.GetGeneration()

		assert.NotNil(t, gen)
		assert.Equal(t, 10, gen.ParallelFetches)
		assert.Equal(t, DefaultBranch, gen.DefaultBranch)
		assert.False(t, gen.CacheEnabled)
		assert.Equal(t, "5m", gen.CacheTTL)
	})

	t.Run("complete generation config", func(t *testing.T) {
		project := &Project{
			Generation: &GenerationConfig{
				ParallelFetches: 3,
				DefaultBranch:   "develop",
				CacheEnabled:    true,
				CacheTTL:        "10m",
			},
		}
		gen := project.GetGeneration()

		assert.NotNil(t, gen)
		assert.Equal(t, 3, gen.ParallelFetches)
		assert.Equal(t, "develop", gen.DefaultBranch)
		assert.True(t, gen.CacheEnabled)
		assert.Equal(t, "10m", gen.CacheTTL)
	})
}

func TestGetConfigFileName(t *testing.T) {
	t.Parallel()
	assert.Equal(t, ".contexture.yaml", GetConfigFileName())
}

func TestGetContextureDir(t *testing.T) {
	t.Parallel()
	assert.Equal(t, ".contexture", GetContextureDir())
}

func TestGetConfigPath(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		baseDir  string
		location ConfigLocation
		expected string
	}{
		{
			name:     "root location",
			baseDir:  "/home/user/project",
			location: ConfigLocationRoot,
			expected: "/home/user/project/.contexture.yaml",
		},
		{
			name:     "contexture location",
			baseDir:  "/home/user/project",
			location: ConfigLocationContexture,
			expected: "/home/user/project/.contexture/.contexture.yaml",
		},
		{
			name:     "invalid location",
			baseDir:  "/home/user/project",
			location: ConfigLocation("invalid"),
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetConfigPath(tt.baseDir, tt.location)
			assert.Equal(t, tt.expected, result)
		})
	}
}
