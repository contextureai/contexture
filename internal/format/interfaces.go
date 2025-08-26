// Package format provides common interfaces for format implementations
package format

// ConfigInterface defines common methods for all format configurations
type ConfigInterface interface {
	// IsSingleFile returns true if the format outputs to a single file
	IsSingleFile() bool

	// GetFileExtension returns the file extension for the format
	GetFileExtension() string
}
