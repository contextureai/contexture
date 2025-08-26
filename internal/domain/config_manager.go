package domain

// ConfigManager provides common configuration management functionality
type ConfigManager interface {
	GetEnabledFormats() []FormatConfig
	GetFormatByType(formatType FormatType) *FormatConfig
	HasFormat(formatType FormatType) bool
	GetGeneration() *GenerationConfig
}

// formatConfigContainer contains formats that can be managed
type formatConfigContainer struct {
	formats []FormatConfig
}

// GetEnabledFormats returns only the enabled format configurations
func (f *formatConfigContainer) GetEnabledFormats() []FormatConfig {
	var enabled []FormatConfig
	for _, format := range f.formats {
		if format.Enabled {
			enabled = append(enabled, format)
		}
	}
	return enabled
}

// GetFormatByType returns a format configuration by type
func (f *formatConfigContainer) GetFormatByType(formatType FormatType) *FormatConfig {
	for _, format := range f.formats {
		if format.Type == formatType {
			return &format
		}
	}
	return nil
}

// HasFormat checks if the container has a specific format configured
func (f *formatConfigContainer) HasFormat(formatType FormatType) bool {
	return f.GetFormatByType(formatType) != nil
}

// generationConfigProvider provides generation configuration with defaults
type generationConfigProvider struct {
	generation *GenerationConfig
}

// GetGeneration returns generation config with defaults
func (g *generationConfigProvider) GetGeneration() *GenerationConfig {
	if g.generation == nil {
		return &GenerationConfig{
			ParallelFetches: 5,
			DefaultBranch:   "main",
			CacheEnabled:    true,
			CacheTTL:        "5m",
		}
	}

	// Apply defaults to existing config
	gen := *g.generation
	if gen.ParallelFetches == 0 {
		gen.ParallelFetches = 5
	}
	if gen.DefaultBranch == "" {
		gen.DefaultBranch = DefaultBranch
	}
	if gen.CacheTTL == "" {
		gen.CacheTTL = "5m"
	}

	return &gen
}
