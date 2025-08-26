// Package windsurf provides Windsurf-specific UI components and format construction
package windsurf

import (
	"github.com/charmbracelet/huh"
)

// Handler implements the format.Handler interface for Windsurf format
type Handler struct{}

// GetUIOption returns the UI option for Windsurf format selection
func (h *Handler) GetUIOption(selected bool) huh.Option[string] {
	return huh.NewOption("Windsurf (.windsurf/rules/)", "windsurf").Selected(selected)
}

// GetDisplayName returns the display name for Windsurf format
func (h *Handler) GetDisplayName() string {
	return "Windsurf (.windsurf/rules/)"
}

// GetDescription returns the description for Windsurf format
func (h *Handler) GetDescription() string {
	return "Flexible output for Windsurf AI code editor (directory or single-file)"
}
