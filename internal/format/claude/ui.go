// Package claude provides Claude-specific UI components and format construction
package claude

import (
	"os"
	"path/filepath"

	"github.com/charmbracelet/huh"
	"github.com/contextureai/contexture/internal/domain"
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

// GetCapabilities returns the capabilities for Claude format
func (h *Handler) GetCapabilities() domain.FormatCapabilities {
	homeDir, _ := os.UserHomeDir()
	userRulesPath := filepath.Join(homeDir, ".claude", "CLAUDE.md")

	return domain.FormatCapabilities{
		SupportsUserRules:    true,
		UserRulesPath:        userRulesPath,
		DefaultUserRulesMode: domain.UserRulesNative,
		MaxRuleSize:          0, // No specific limit for Claude
	}
}
