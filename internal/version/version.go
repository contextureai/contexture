// Package version provides comprehensive build-time version information
package version

import (
	"runtime"
)

const (
	defaultVersion = "dev"
	defaultValue   = "unknown"
)

var (
	// Version is the build version set via ldflags
	Version = defaultVersion
	// Commit is the git commit hash set via ldflags
	Commit = defaultValue
	// BuildDate is the build date set via ldflags
	BuildDate = defaultValue
	// BuildBy is who built the binary set via ldflags
	BuildBy = defaultValue
)

// Info represents comprehensive version information
type Info struct {
	// Version is the build version set via ldflags
	Version string `json:"version"`
	// Commit is the git commit hash set via ldflags
	Commit string `json:"commit"`
	// BuildDate is the build date set via ldflags
	BuildDate string `json:"build_date"`
	// BuildBy is who built the binary set via ldflags
	BuildBy string `json:"build_by"`
	// GoVersion is the version of Go used to build the binary
	GoVersion string `json:"go_version"`
	// Platform is the platform the binary was built on
	Platform string `json:"platform"`
}

// Get returns comprehensive version information
func Get() Info {
	return Info{
		Version:   ifTrueReturnDefault(Version, defaultVersion),
		Commit:    ifTrueReturnDefault(Commit, defaultValue),
		BuildDate: ifTrueReturnDefault(BuildDate, defaultValue),
		BuildBy:   ifTrueReturnDefault(BuildBy, defaultValue),
		GoVersion: runtime.Version(),
		Platform:  runtime.GOOS + "/" + runtime.GOARCH,
	}
}

// GetShort returns just the version string for basic usage
func GetShort() string {
	return ifTrueReturnDefault(Version, defaultVersion)
}

// ifTrueReturnDefault returns the value if it is not empty, otherwise returns the default value
func ifTrueReturnDefault(value string, defaultValue string) string {
	if value == "" {
		return defaultValue
	}
	return value
}

// String returns a formatted version string
func (i Info) String() string {
	if i.Version == "dev" {
		return "contexture development build"
	}
	return "contexture version " + i.Version
}

// Detailed returns a detailed version string with all information
func (i Info) Detailed() string {
	base := i.String()
	if i.Commit != defaultValue {
		base += " (" + i.Commit
		if i.BuildDate != "unknown" {
			base += ", built " + i.BuildDate
		}
		if i.BuildBy != "unknown" {
			base += " by " + i.BuildBy
		}
		base += ")"
	}
	base += "\nGo version: " + i.GoVersion
	base += "\nPlatform: " + i.Platform
	return base
}
