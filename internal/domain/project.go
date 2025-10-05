package domain

import (
	"path/filepath"
)

// Project represents the main project configuration
type Project struct {
	// Version for configuration compatibility
	Version int `yaml:"version,omitempty" json:"version,omitempty"`

	// Providers for external rule repositories (optional)
	Providers []Provider `yaml:"providers,omitempty" json:"providers,omitempty"`

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

// Provider represents a named rule repository
type Provider struct {
	Name          string        `yaml:"name"                     json:"name"                     validate:"required"`
	URL           string        `yaml:"url"                      json:"url"                      validate:"required,url"`
	DefaultBranch string        `yaml:"defaultBranch,omitempty"  json:"defaultBranch,omitempty"`
	Auth          *ProviderAuth `yaml:"auth,omitempty"           json:"auth,omitempty"`
}

// ProviderAuth represents authentication configuration for a provider
type ProviderAuth struct {
	Type  string `yaml:"type"            json:"type"            validate:"required,oneof=token ssh"`
	Token string `yaml:"token,omitempty" json:"token,omitempty" validate:"required_if=Type token"`
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

// GetProviderByName returns a provider by name
func (p *Project) GetProviderByName(name string) *Provider {
	for i := range p.Providers {
		if p.Providers[i].Name == name {
			return &p.Providers[i]
		}
	}
	return nil
}

// GetProviders returns all configured providers
func (p *Project) GetProviders() []Provider {
	return p.Providers
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
