// Package windsurf provides Windsurf-specific UI components and format construction
package windsurf

import (
	"os"
	"path/filepath"

	"github.com/charmbracelet/huh"
	"github.com/contextureai/contexture/internal/domain"
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

// GetCapabilities returns the capabilities for Windsurf format
func (h *Handler) GetCapabilities() domain.FormatCapabilities {
	homeDir, _ := os.UserHomeDir()
	userRulesPath := filepath.Join(homeDir, ".windsurf", "global_rules.md")

	return domain.FormatCapabilities{
		SupportsUserRules:    true,
		UserRulesPath:        userRulesPath,
		DefaultUserRulesMode: domain.UserRulesNative,
		MaxRuleSize:          12000, // Windsurf supports 12,000 chars per file
	}
}
