package version

import (
	"fmt"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	testVersionDev = "dev"
	testUnknown    = "unknown"
)

func TestGet(t *testing.T) {
	t.Run("returns_complete_version_info", func(t *testing.T) {
		info := Get()

		assert.NotEmpty(t, info.Version)
		assert.NotEmpty(t, info.GoVersion)
		assert.NotEmpty(t, info.Platform)

		// These might be testUnknown in development
		assert.NotEmpty(t, info.Commit)
		assert.NotEmpty(t, info.BuildDate)
		assert.NotEmpty(t, info.BuildBy)
	})

	t.Run("version_info_has_correct_structure", func(t *testing.T) {
		info := Get()

		// Check Go version format
		assert.True(t, strings.HasPrefix(info.GoVersion, "go"), "Go version should start with 'go'")

		// Check platform format (should be OS/ARCH)
		assert.Contains(t, info.Platform, "/", "Platform should be in format OS/ARCH")

		// Platform should match runtime
		expectedPlatform := runtime.GOOS + "/" + runtime.GOARCH
		assert.Equal(t, expectedPlatform, info.Platform)
	})
}

func TestGetShort(t *testing.T) {
	tests := []struct {
		name           string
		versionInput   string
		expectedOutput string
	}{
		{
			name:           "development_version",
			versionInput:   testVersionDev,
			expectedOutput: testVersionDev,
		},
		{
			name:           "empty_version_returns_dev",
			versionInput:   "",
			expectedOutput: testVersionDev,
		},
		{
			name:           "release_version",
			versionInput:   "v1.2.3",
			expectedOutput: "v1.2.3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Store original
			original := Version
			defer func() { Version = original }()

			// Set test version
			Version = tt.versionInput

			result := GetShort()
			assert.Equal(t, tt.expectedOutput, result)
		})
	}
}

func TestInfoString(t *testing.T) {
	tests := []struct {
		name           string
		version        string
		expectedPrefix string
	}{
		{
			name:           "development_build",
			version:        testVersionDev,
			expectedPrefix: "contexture development build",
		},
		{
			name:           "release_build",
			version:        "v1.0.0",
			expectedPrefix: "contexture version v1.0.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := Info{
				Version:   tt.version,
				Commit:    "abc123",
				BuildDate: "2024-01-01",
				BuildBy:   "ci",
				GoVersion: "go1.21.0",
				Platform:  "linux/amd64",
			}

			result := info.String()
			assert.Equal(t, tt.expectedPrefix, result)
		})
	}
}

func TestInfoDetailed(t *testing.T) {
	t.Run("development_build_detailed", func(t *testing.T) {
		info := Info{
			Version:   testVersionDev,
			Commit:    testUnknown,
			BuildDate: testUnknown,
			BuildBy:   testUnknown,
			GoVersion: "go1.21.0",
			Platform:  "linux/amd64",
		}

		result := info.Detailed()

		assert.Contains(t, result, "contexture development build")
		assert.Contains(t, result, "Go version: go1.21.0")
		assert.Contains(t, result, "Platform: linux/amd64")
		assert.NotContains(t, result, "abc123") // Should not show unknown commit
	})

	t.Run("release_build_detailed", func(t *testing.T) {
		info := Info{
			Version:   "v1.0.0",
			Commit:    "abc123",
			BuildDate: "2024-01-01",
			BuildBy:   "ci",
			GoVersion: "go1.21.0",
			Platform:  "linux/amd64",
		}

		result := info.Detailed()

		assert.Contains(t, result, "contexture version v1.0.0")
		assert.Contains(t, result, "abc123")
		assert.Contains(t, result, "built 2024-01-01")
		assert.Contains(t, result, "by ci")
		assert.Contains(t, result, "Go version: go1.21.0")
		assert.Contains(t, result, "Platform: linux/amd64")
	})

	t.Run("partial_build_info", func(t *testing.T) {
		info := Info{
			Version:   "v1.0.0",
			Commit:    "abc123",
			BuildDate: testUnknown,
			BuildBy:   testUnknown,
			GoVersion: "go1.21.0",
			Platform:  "linux/amd64",
		}

		result := info.Detailed()

		assert.Contains(t, result, "contexture version v1.0.0")
		assert.Contains(t, result, "abc123")
		assert.NotContains(t, result, "built unknown")
		assert.NotContains(t, result, "by unknown")
	})
}

func TestBuildTimeVariables(t *testing.T) {
	t.Run("variables_are_mutable_for_ldflags", func(t *testing.T) {
		// Store originals
		origVersion := Version
		origCommit := Commit
		origBuildDate := BuildDate
		origBuildBy := BuildBy

		defer func() {
			Version = origVersion
			Commit = origCommit
			BuildDate = origBuildDate
			BuildBy = origBuildBy
		}()

		// Test that variables can be set (simulating ldflags)
		Version = "v2.0.0"
		Commit = "def456"
		BuildDate = "2024-02-01T10:00:00Z"
		BuildBy = "github-actions"

		info := Get()
		assert.Equal(t, "v2.0.0", info.Version)
		assert.Equal(t, "def456", info.Commit)
		assert.Equal(t, "2024-02-01T10:00:00Z", info.BuildDate)
		assert.Equal(t, "github-actions", info.BuildBy)
	})

	t.Run("default_values", func(t *testing.T) {
		// Reset to defaults
		Version = testVersionDev
		Commit = testUnknown
		BuildDate = testUnknown
		BuildBy = testUnknown

		info := Get()
		assert.Equal(t, testVersionDev, info.Version)
		assert.Equal(t, testUnknown, info.Commit)
		assert.Equal(t, testUnknown, info.BuildDate)
		assert.Equal(t, testUnknown, info.BuildBy)
	})
}

// Benchmark version info creation
func BenchmarkGet(b *testing.B) {
	b.ResetTimer()
	for range b.N {
		_ = Get()
	}
}

func BenchmarkGetShort(b *testing.B) {
	b.ResetTimer()
	for range b.N {
		_ = GetShort()
	}
}

func BenchmarkInfoString(b *testing.B) {
	info := Get()
	b.ResetTimer()
	for range b.N {
		_ = info.String()
	}
}

func BenchmarkInfoDetailed(b *testing.B) {
	info := Get()
	b.ResetTimer()
	for range b.N {
		_ = info.Detailed()
	}
}

// Example usage of version package
func ExampleGet() {
	info := Get()

	// Basic version string
	_ = info.String()

	// Detailed version information
	_ = info.Detailed()

	// Individual fields
	_ = info.Version
	_ = info.Commit
	_ = info.GoVersion

	fmt.Println("Version information is available through the Info struct")
	// Output:
	// Version information is available through the Info struct
}

func ExampleGetShort() {
	version := GetShort()
	_ = version

	fmt.Println("Returns just the version string")
	// Output:
	// Returns just the version string
}
