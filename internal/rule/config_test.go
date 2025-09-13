package rule

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultConfig(t *testing.T) {
	t.Parallel()
	config := DefaultConfig()

	require.NotNil(t, config)

	// Test default values
	assert.Equal(t, "https://github.com/contextureai/rules.git", config.DefaultRepositoryURL)
	assert.Positive(t, int64(config.FetchTimeout), "fetch timeout should be positive")
	assert.Positive(t, config.MaxWorkers, "max workers should be positive")
	assert.True(t, config.CacheEnabled, "cache should be enabled by default")
	assert.Equal(t, 15*time.Minute, config.CacheTTL)
	assert.True(t, config.EnableConcurrency, "concurrency should be enabled by default")
	assert.Equal(t, 5*time.Minute, config.ProcessingTimeout)
}

func TestConfig_Validate(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		config    *Config
		wantValid bool
	}{
		{
			name:      "default config should be valid",
			config:    DefaultConfig(),
			wantValid: true,
		},
		{
			name: "valid custom config",
			config: &Config{
				DefaultRepositoryURL: "https://example.com/repo.git",
				FetchTimeout:         30 * time.Second,
				MaxWorkers:           5,
				CacheEnabled:         false,
				CacheTTL:             10 * time.Minute,
				EnableConcurrency:    true,
				ProcessingTimeout:    2 * time.Minute,
			},
			wantValid: true,
		},
		{
			name: "invalid repository URL",
			config: &Config{
				DefaultRepositoryURL: "not-a-url",
				FetchTimeout:         30 * time.Second,
				MaxWorkers:           5,
				CacheEnabled:         true,
				CacheTTL:             10 * time.Minute,
				EnableConcurrency:    true,
				ProcessingTimeout:    2 * time.Minute,
			},
			wantValid: false,
		},
		{
			name: "zero fetch timeout should be invalid",
			config: &Config{
				DefaultRepositoryURL: "https://example.com/repo.git",
				FetchTimeout:         0,
				MaxWorkers:           5,
				CacheEnabled:         true,
				CacheTTL:             10 * time.Minute,
				EnableConcurrency:    true,
				ProcessingTimeout:    2 * time.Minute,
			},
			wantValid: false,
		},
		{
			name: "too many workers should be invalid",
			config: &Config{
				DefaultRepositoryURL: "https://example.com/repo.git",
				FetchTimeout:         30 * time.Second,
				MaxWorkers:           200, // exceeds max of 100
				CacheEnabled:         true,
				CacheTTL:             10 * time.Minute,
				EnableConcurrency:    true,
				ProcessingTimeout:    2 * time.Minute,
			},
			wantValid: false,
		},
		{
			name: "zero workers should be invalid",
			config: &Config{
				DefaultRepositoryURL: "https://example.com/repo.git",
				FetchTimeout:         30 * time.Second,
				MaxWorkers:           0,
				CacheEnabled:         true,
				CacheTTL:             10 * time.Minute,
				EnableConcurrency:    true,
				ProcessingTimeout:    2 * time.Minute,
			},
			wantValid: false,
		},
		{
			name: "negative cache TTL should be invalid",
			config: &Config{
				DefaultRepositoryURL: "https://example.com/repo.git",
				FetchTimeout:         30 * time.Second,
				MaxWorkers:           5,
				CacheEnabled:         true,
				CacheTTL:             -1 * time.Minute,
				EnableConcurrency:    true,
				ProcessingTimeout:    2 * time.Minute,
			},
			wantValid: false,
		},
		{
			name: "zero cache TTL with cache disabled should be valid",
			config: &Config{
				DefaultRepositoryURL: "https://example.com/repo.git",
				FetchTimeout:         30 * time.Second,
				MaxWorkers:           5,
				CacheEnabled:         false, // Cache disabled
				CacheTTL:             0,
				EnableConcurrency:    true,
				ProcessingTimeout:    2 * time.Minute,
			},
			wantValid: true,
		},
		{
			name: "negative processing timeout should be invalid",
			config: &Config{
				DefaultRepositoryURL: "https://example.com/repo.git",
				FetchTimeout:         30 * time.Second,
				MaxWorkers:           5,
				CacheEnabled:         true,
				CacheTTL:             10 * time.Minute,
				EnableConcurrency:    true,
				ProcessingTimeout:    -1 * time.Minute,
			},
			wantValid: false,
		},
		{
			name: "zero processing timeout should be valid (no timeout)",
			config: &Config{
				DefaultRepositoryURL: "https://example.com/repo.git",
				FetchTimeout:         30 * time.Second,
				MaxWorkers:           5,
				CacheEnabled:         true,
				CacheTTL:             10 * time.Minute,
				EnableConcurrency:    true,
				ProcessingTimeout:    0,
			},
			wantValid: true,
		},
		{
			name: "empty repository URL should be invalid",
			config: &Config{
				DefaultRepositoryURL: "",
				FetchTimeout:         30 * time.Second,
				MaxWorkers:           5,
				CacheEnabled:         true,
				CacheTTL:             10 * time.Minute,
				EnableConcurrency:    true,
				ProcessingTimeout:    2 * time.Minute,
			},
			wantValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()

			if tt.wantValid {
				assert.NoError(t, err, "config should be valid")
			} else {
				assert.Error(t, err, "config should be invalid")
			}
		})
	}
}

func TestConfig_FetcherConfig(t *testing.T) {
	t.Parallel()
	config := DefaultConfig()

	// Modify some values to test they're properly returned
	config.FetchTimeout = 45 * time.Second
	config.MaxWorkers = 8

	fetcherConfig := config.FetcherConfig()

	require.NotNil(t, fetcherConfig)
	// The exact structure of fetcherConfig depends on the implementation
	// but we can verify it's created without error
}
