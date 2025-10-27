// Package cursor provides Cursor-specific UI components and format construction
package cursor

import (
	"github.com/charmbracelet/huh"
	"github.com/contextureai/contexture/internal/domain"
)

// Handler implements the format.Handler interface for Cursor format
type Handler struct{}

// GetUIOption returns the UI option for Cursor format selection
func (h *Handler) GetUIOption(selected bool) huh.Option[string] {
	return huh.NewOption("Cursor (.cursor/rules/)", "cursor").Selected(selected)
}

// GetDisplayName returns the display name for Cursor format
func (h *Handler) GetDisplayName() string {
	return "Cursor (.cursor/rules/)"
}

// GetDescription returns the description for Cursor format
func (h *Handler) GetDescription() string {
	return "Multi-file output for Cursor AI code editor"
}

// GetCapabilities returns the capabilities for Cursor format
func (h *Handler) GetCapabilities() domain.FormatCapabilities {
	return domain.FormatCapabilities{
		SupportsUserRules:    false,                   // Cursor doesn't support native user rules
		UserRulesPath:        "",                      // No user rules path
		DefaultUserRulesMode: domain.UserRulesProject, // Default to including user rules in project
		MaxRuleSize:          0,                       // No specific limit
	}
}
