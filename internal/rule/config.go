package rule

import (
	"fmt"
	"time"

	"github.com/contextureai/contexture/internal/domain"
	"github.com/go-playground/validator/v10"
)

// Config holds configuration for the rule package
type Config struct {
	// Fetcher configuration
	DefaultRepositoryURL string        `validate:"required,url"`
	FetchTimeout         time.Duration `validate:"required,min=1s"`
	MaxWorkers           int           `validate:"required,min=1,max=100"`

	// Cache configuration
	CacheEnabled bool          `validate:""`
	CacheTTL     time.Duration `validate:"min=0"`

	// Processing configuration
	EnableConcurrency bool          `validate:""`
	ProcessingTimeout time.Duration `validate:"min=0"`
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		DefaultRepositoryURL: "https://github.com/contextureai/rules.git",
		FetchTimeout:         time.Duration(domain.DefaultFetchTimeout) * time.Second,
		MaxWorkers:           domain.DefaultMaxWorkers,
		CacheEnabled:         true,
		CacheTTL:             15 * time.Minute,
		EnableConcurrency:    true,
		ProcessingTimeout:    5 * time.Minute,
	}
}

// Validate validates the configuration
func (c *Config) Validate() error {
	validate := validator.New()
	if err := validate.Struct(c); err != nil {
		return fmt.Errorf("configuration validation failed: %w", err)
	}

	// Additional business logic validation
	if c.CacheEnabled && c.CacheTTL == 0 {
		return fmt.Errorf("cache TTL must be greater than 0 when cache is enabled")
	}

	if c.EnableConcurrency && c.MaxWorkers < 1 {
		return fmt.Errorf("max workers must be at least 1 when concurrency is enabled")
	}

	return nil
}

// FetcherConfig converts Config to FetcherConfig
func (c *Config) FetcherConfig() FetcherConfig {
	return FetcherConfig{
		DefaultURL: c.DefaultRepositoryURL,
	}
}
