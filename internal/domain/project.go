package domain

import (
	"path/filepath"
)

// Project represents the main project configuration
type Project struct {
	// Version for configuration compatibility
	Version int `yaml:"version,omitempty" json:"version,omitempty"`

	// Sources for external rule repositories (optional)
	Sources []Source `yaml:"sources,omitempty" json:"sources,omitempty"`

	// Format configurations
	Formats []FormatConfig `yaml:"formats" json:"formats"`

	// Rule references
	Rules []RuleRef `yaml:"rules" json:"rules"`

	// Generation settings (optional)
	Generation *GenerationConfig `yaml:"generation,omitempty" json:"generation,omitempty"`

	// Embedded format config functionality
	formatContainer formatConfigContainer `yaml:"-" json:"-"`
	// Embedded generation config functionality
	genProvider generationConfigProvider `yaml:"-" json:"-"`
}

// Source represents an external rule source
type Source struct {
	Name    string      `yaml:"name"             json:"name"`
	Type    string      `yaml:"type"             json:"type"` // Currently only "git"
	URL     string      `yaml:"url"              json:"url"`
	Branch  string      `yaml:"branch,omitempty" json:"branch,omitempty"`
	Tag     string      `yaml:"tag,omitempty"    json:"tag,omitempty"`
	Enabled bool        `yaml:"enabled"          json:"enabled"`
	Auth    *SourceAuth `yaml:"auth,omitempty"   json:"auth,omitempty"`
}

// SourceAuth represents authentication configuration for a source
type SourceAuth struct {
	Type  string `yaml:"type"            json:"type"` // "token" or "ssh"
	Token string `yaml:"token,omitempty" json:"token,omitempty"`
}

// GenerationConfig represents settings for rule generation
type GenerationConfig struct {
	ParallelFetches int    `yaml:"parallelFetches,omitempty" json:"parallelFetches,omitempty"`
	DefaultBranch   string `yaml:"defaultBranch,omitempty"   json:"defaultBranch,omitempty"`
	CacheEnabled    bool   `yaml:"cacheEnabled,omitempty"    json:"cacheEnabled,omitempty"`
	CacheTTL        string `yaml:"cacheTTL,omitempty"        json:"cacheTTL,omitempty"` // Duration string like "5m"
}

// GetEnabledFormats returns only the enabled format configurations for Project
func (p *Project) GetEnabledFormats() []FormatConfig {
	p.formatContainer.formats = p.Formats
	return p.formatContainer.GetEnabledFormats()
}

// GetFormatByType returns a format configuration by type for Project
func (p *Project) GetFormatByType(formatType FormatType) *FormatConfig {
	p.formatContainer.formats = p.Formats
	return p.formatContainer.GetFormatByType(formatType)
}

// HasFormat checks if the project has a specific format configured
func (p *Project) HasFormat(formatType FormatType) bool {
	p.formatContainer.formats = p.Formats
	return p.formatContainer.HasFormat(formatType)
}

// GetEnabledSources returns only the enabled sources
func (p *Project) GetEnabledSources() []Source {
	var enabled []Source
	for _, source := range p.Sources {
		if source.Enabled {
			enabled = append(enabled, source)
		}
	}
	return enabled
}

// GetSourceByName returns a source by name
func (p *Project) GetSourceByName(name string) *Source {
	for _, source := range p.Sources {
		if source.Name == name {
			return &source
		}
	}
	return nil
}

// GetGeneration returns generation config with defaults for Project
func (p *Project) GetGeneration() *GenerationConfig {
	p.genProvider.generation = p.Generation
	return p.genProvider.GetGeneration()
}

// ConfigLocation represents where configuration is stored
type ConfigLocation string

const (
	// ConfigLocationRoot indicates config is stored in project root
	ConfigLocationRoot ConfigLocation = "root"
	// ConfigLocationContexture indicates config is stored in .contexture/ directory
	ConfigLocationContexture ConfigLocation = "contexture"
)

// ConfigResult represents the result of loading configuration
type ConfigResult struct {
	Config   *Project       `json:"config"`
	Location ConfigLocation `json:"location"`
	Path     string         `json:"path"`
}

// GetConfigFileName returns the config file name
func GetConfigFileName() string {
	return ".contexture.yaml"
}

// GetContextureDir returns the contexture directory name
func GetContextureDir() string {
	return ".contexture"
}

// GetConfigPath returns the full config path for a given location
func GetConfigPath(baseDir string, location ConfigLocation) string {
	switch location {
	case ConfigLocationRoot:
		return filepath.Join(baseDir, GetConfigFileName())
	case ConfigLocationContexture:
		return filepath.Join(baseDir, GetContextureDir(), GetConfigFileName())
	default:
		return ""
	}
}
