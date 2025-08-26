// Package claude provides Claude-specific UI components and format construction
package claude

import (
	"github.com/charmbracelet/huh"
)

// Handler implements the format.Handler interface for Claude format
type Handler struct{}

// GetUIOption returns the UI option for Claude format selection
func (h *Handler) GetUIOption(selected bool) huh.Option[string] {
	return huh.NewOption("Claude (CLAUDE.md)", "claude").Selected(selected)
}

// GetDisplayName returns the display name for Claude format
func (h *Handler) GetDisplayName() string {
	return "Claude (CLAUDE.md)"
}

// GetDescription returns the description for Claude format
func (h *Handler) GetDescription() string {
	return "Single file output for Claude AI assistant"
}
