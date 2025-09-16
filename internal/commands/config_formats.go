// Package commands provides CLI command implementations for format management
package commands

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"github.com/contextureai/contexture/internal/dependencies"
	"github.com/contextureai/contexture/internal/domain"
	"github.com/contextureai/contexture/internal/format"
	"github.com/contextureai/contexture/internal/project"
	"github.com/contextureai/contexture/internal/tui"
	"github.com/contextureai/contexture/internal/ui"
	"github.com/spf13/afero"
	"github.com/urfave/cli/v3"
)

// FormatManager handles format-related operations
type FormatManager struct {
	projectManager *project.Manager
	registry       *format.Registry
	fs             afero.Fs
}

// NewFormatManager creates a new format manager
func NewFormatManager(deps *dependencies.Dependencies) *FormatManager {
	return &FormatManager{
		projectManager: project.NewManager(deps.FS),
		registry:       format.GetDefaultRegistry(deps.FS),
		fs:             deps.FS,
	}
}

// ListFormats displays all configured formats
func (fm *FormatManager) ListFormats(_ context.Context, _ *cli.Command) error {
	// Handle project configuration
	currentDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	configResult, err := fm.projectManager.LoadConfig(currentDir)
	if err != nil {
		fmt.Println("No project configuration found")
		fmt.Println("Initialize a project with: contexture init")
		return err
	}

	return fm.displayProjectFormats(configResult.Config)
}

// AddFormat adds a new format to the project
func (fm *FormatManager) AddFormat(ctx context.Context, cmd *cli.Command, formatType string) error {
	if formatType == "" {
		return fm.interactiveAddFormat(ctx, cmd)
	}

	// Handle project configuration
	currentDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	configResult, err := fm.projectManager.LoadConfig(currentDir)
	if err != nil {
		return fmt.Errorf("no project configuration found. Run 'contexture init' first: %w", err)
	}

	return fm.addFormatToProjectConfig(configResult, formatType, currentDir)
}

// EnableFormat enables an existing format
func (fm *FormatManager) EnableFormat(
	_ context.Context,
	_ *cli.Command,
	formatType string,
) error {
	// Handle project configuration
	currentDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	configResult, err := fm.projectManager.LoadConfig(currentDir)
	if err != nil {
		return fmt.Errorf("no project configuration found. Run 'contexture init' first: %w", err)
	}

	return fm.enableFormatInProjectConfig(configResult, formatType, currentDir)
}

// DisableFormat disables an existing format
func (fm *FormatManager) DisableFormat(
	_ context.Context,
	_ *cli.Command,
	formatType string,
) error {
	// Handle project configuration
	currentDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	configResult, err := fm.projectManager.LoadConfig(currentDir)
	if err != nil {
		return fmt.Errorf("no project configuration found. Run 'contexture init' first: %w", err)
	}

	return fm.disableFormatInProjectConfig(configResult, formatType, currentDir)
}

// RemoveFormat removes a format from the configuration
func (fm *FormatManager) RemoveFormat(
	_ context.Context,
	_ *cli.Command,
	formatType string,
) error {
	// Handle project configuration
	currentDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	configResult, err := fm.projectManager.LoadConfig(currentDir)
	if err != nil {
		return fmt.Errorf("no project configuration found. Run 'contexture init' first: %w", err)
	}

	return fm.removeFormatFromProjectConfig(configResult, formatType, currentDir)
}

// interactiveAddFormat provides an interactive interface to add formats
func (fm *FormatManager) interactiveAddFormat(_ context.Context, _ *cli.Command) error {
	// Get current directory and load configuration
	currentDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	configResult, err := fm.projectManager.LoadConfig(currentDir)
	if err != nil {
		return fmt.Errorf("no project configuration found. Run 'contexture init' first: %w", err)
	}

	config := configResult.Config
	theme := ui.DefaultTheme()

	// Show current formats
	fmt.Println(ui.CommandHeader("add formats"))

	if len(config.Formats) > 0 {
		headerStyle := lipgloss.NewStyle().
			Bold(true).
			Foreground(theme.Info)

		fmt.Println(headerStyle.Render("Current Formats"))

		// Use exactly the same styling as main config command
		enabledStyle := lipgloss.NewStyle().
			Foreground(theme.Success).
			Bold(true)

		disabledStyle := lipgloss.NewStyle().
			Foreground(theme.Muted)

		darkMutedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#666666"))

		// Calculate proper column widths for better alignment (same as main config)
		maxTypeWidth := 0
		maxNameWidth := 0
		for _, formatConfig := range config.Formats {
			handler, exists := fm.registry.GetHandler(formatConfig.Type)
			if !exists {
				continue
			}
			if len(string(formatConfig.Type)) > maxTypeWidth {
				maxTypeWidth = len(string(formatConfig.Type))
			}
			if len(handler.GetDisplayName()) > maxNameWidth {
				maxNameWidth = len(handler.GetDisplayName())
			}
		}

		for _, formatConfig := range config.Formats {
			handler, exists := fm.registry.GetHandler(formatConfig.Type)
			if !exists {
				continue
			}

			var status string
			var statusStyle lipgloss.Style
			if formatConfig.Enabled {
				status = "[enabled]"
				statusStyle = enabledStyle
			} else {
				status = "[disabled]"
				statusStyle = disabledStyle
			}

			// Use raw text for width calculation, then apply styling (same as main config)
			typeText := string(formatConfig.Type)
			nameText := handler.GetDisplayName()

			fmt.Printf("  %s%s %s%s %s\n",
				darkMutedStyle.Render(typeText),
				strings.Repeat(" ", maxTypeWidth-len(typeText)),
				nameText,
				strings.Repeat(" ", maxNameWidth-len(nameText)),
				statusStyle.Render(status))
		}
		fmt.Println()
	}

	// Get available formats to add
	existingTypes := make(map[domain.FormatType]bool)
	for _, format := range config.Formats {
		existingTypes[format.Type] = true
	}

	var availableFormats []tui.SelectOption
	for _, formatType := range fm.registry.GetAvailableFormats() {
		if !existingTypes[formatType] {
			if handler, exists := fm.registry.GetHandler(formatType); exists {
				availableFormats = append(availableFormats, tui.SelectOption{
					Label:       handler.GetDisplayName(),
					Value:       string(formatType),
					Description: handler.GetDescription(),
				})
			}
		}
	}

	if len(availableFormats) == 0 {
		fmt.Println("All supported formats are already configured in your project.")
		fmt.Println()
		fmt.Println("You can:")
		fmt.Println("• Enable/disable existing formats")
		fmt.Println("• Remove formats to add them again")
		fmt.Println("• Run 'contexture build' to update format files")
		return nil
	}

	// Interactive selection
	selectedFormats, err := tui.MultiSelect(tui.MultiSelectOptions{
		Title:       "Select formats to add",
		Description: "Choose one or more formats to add to your project\nPress 'q' or 'esc' to exit",
		Options:     availableFormats,
	})
	if err != nil {
		return fmt.Errorf("format selection cancelled: %w", err)
	}

	if len(selectedFormats) == 0 {
		log.Info("No formats selected")
		return nil
	}

	// Add selected formats
	var addedFormats []string
	for _, selectedFormat := range selectedFormats {
		newFormat := domain.FormatConfig{
			Type:    domain.FormatType(selectedFormat),
			Enabled: true,
		}
		config.Formats = append(config.Formats, newFormat)

		// Create format directories if needed
		if err := fm.registry.CreateFormatDirectories(domain.FormatType(selectedFormat)); err != nil {
			log.Warn("Failed to create format directories", "format", selectedFormat, "error", err)
		}

		addedFormats = append(addedFormats, selectedFormat)
	}

	// Save configuration
	if err := fm.projectManager.SaveConfig(config, configResult.Location, currentDir); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	// Show success message
	theme = ui.DefaultTheme()
	successStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(theme.Success)

	if len(addedFormats) == 1 {
		displayName := fm.getFormatDisplayName(domain.FormatType(addedFormats[0]))
		fmt.Println(successStyle.Render("Format added: " + displayName))
	} else {
		fmt.Println(successStyle.Render(fmt.Sprintf("Added %d formats", len(addedFormats))))
		for _, formatType := range addedFormats {
			displayName := fm.getFormatDisplayName(domain.FormatType(formatType))
			fmt.Printf("  • %s\n", displayName)
		}
	}

	return nil
}

// interactiveRemoveFormat provides an interactive interface to remove formats
func (fm *FormatManager) interactiveRemoveFormat(_ context.Context, _ *cli.Command) error {
	// Get current directory and load configuration
	currentDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	configResult, err := fm.projectManager.LoadConfig(currentDir)
	if err != nil {
		return fmt.Errorf("no project configuration found. Run 'contexture init' first: %w", err)
	}

	config := configResult.Config

	if len(config.Formats) == 0 {
		fmt.Println("No formats configured to remove")
		return nil
	}

	// Show header
	fmt.Println(ui.CommandHeader("remove formats"))

	// Show current formats
	theme := ui.DefaultTheme()
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(theme.Info)

	fmt.Println(headerStyle.Render("Current Formats"))

	// Use exactly the same styling as main config command
	enabledStyle := lipgloss.NewStyle().
		Foreground(theme.Success).
		Bold(true)

	disabledStyle := lipgloss.NewStyle().
		Foreground(theme.Muted)

	darkMutedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#666666"))

	// Calculate proper column widths for better alignment (same as main config)
	maxTypeWidth := 0
	maxNameWidth := 0
	for _, formatConfig := range config.Formats {
		handler, exists := fm.registry.GetHandler(formatConfig.Type)
		if !exists {
			continue
		}
		if len(string(formatConfig.Type)) > maxTypeWidth {
			maxTypeWidth = len(string(formatConfig.Type))
		}
		if len(handler.GetDisplayName()) > maxNameWidth {
			maxNameWidth = len(handler.GetDisplayName())
		}
	}

	var availableFormats []tui.SelectOption
	for _, formatConfig := range config.Formats {
		handler, exists := fm.registry.GetHandler(formatConfig.Type)
		if !exists {
			continue
		}

		var status string
		var statusStyle lipgloss.Style
		if formatConfig.Enabled {
			status = statusEnabled
			statusStyle = enabledStyle
		} else {
			status = statusDisabled
			statusStyle = disabledStyle
		}

		// Use raw text for width calculation, then apply styling (same as main config)
		typeText := string(formatConfig.Type)
		nameText := handler.GetDisplayName()

		fmt.Printf("  %s%s %s%s %s\n",
			darkMutedStyle.Render(typeText),
			strings.Repeat(" ", maxTypeWidth-len(typeText)),
			nameText,
			strings.Repeat(" ", maxNameWidth-len(nameText)),
			statusStyle.Render(status))

		// Add to available for removal
		availableFormats = append(availableFormats, tui.SelectOption{
			Label:       handler.GetDisplayName(),
			Value:       string(formatConfig.Type),
			Description: fmt.Sprintf("Currently %s", strings.Trim(status, "[]")),
		})
	}
	fmt.Println()

	// Interactive selection
	selectedFormats, err := tui.MultiSelect(tui.MultiSelectOptions{
		Title:       "Select formats to remove",
		Description: "Choose one or more formats to remove from your project\nPress 'q' or 'esc' to exit",
		Options:     availableFormats,
	})
	if err != nil {
		return fmt.Errorf("format selection cancelled: %w", err)
	}

	if len(selectedFormats) == 0 {
		log.Info("No formats selected for removal")
		return nil
	}

	// Check if we're removing all enabled formats
	enabledCount := 0
	removingEnabledCount := 0
	for _, format := range config.Formats {
		if format.Enabled {
			enabledCount++
			for _, selected := range selectedFormats {
				if format.Type == domain.FormatType(selected) {
					removingEnabledCount++
					break
				}
			}
		}
	}

	if enabledCount-removingEnabledCount <= 0 {
		return fmt.Errorf(
			"cannot remove all enabled formats. At least one format must remain enabled",
		)
	}

	// Remove selected formats
	var removedFormats []string
	for _, selectedFormat := range selectedFormats {
		for i, format := range config.Formats {
			if format.Type == domain.FormatType(selectedFormat) {
				// Remove format from slice
				config.Formats = append(config.Formats[:i], config.Formats[i+1:]...)
				removedFormats = append(removedFormats, selectedFormat)
				break
			}
		}
	}

	// Save configuration
	if err := fm.projectManager.SaveConfig(config, configResult.Location, currentDir); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	// Show success message
	successStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(theme.Success)

	if len(removedFormats) == 1 {
		displayName := fm.getFormatDisplayName(domain.FormatType(removedFormats[0]))
		fmt.Println(successStyle.Render("Format removed: " + displayName))
	} else {
		fmt.Println(successStyle.Render(fmt.Sprintf("Removed %d formats", len(removedFormats))))
		for _, formatType := range removedFormats {
			displayName := fm.getFormatDisplayName(domain.FormatType(formatType))
			fmt.Printf("  • %s\n", displayName)
		}
	}

	return nil
}

// interactiveEnableFormat provides an interactive interface to enable formats
func (fm *FormatManager) interactiveEnableFormat(_ context.Context, _ *cli.Command) error {
	// Get current directory and load configuration
	currentDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	configResult, err := fm.projectManager.LoadConfig(currentDir)
	if err != nil {
		return fmt.Errorf("no project configuration found. Run 'contexture init' first: %w", err)
	}

	config := configResult.Config

	if len(config.Formats) == 0 {
		fmt.Println("No formats configured to enable")
		return nil
	}

	// Show header
	fmt.Println(ui.CommandHeader("enable format"))

	// Show current formats
	theme := ui.DefaultTheme()
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(theme.Info)

	fmt.Println(headerStyle.Render("Current Formats"))

	// Use exactly the same styling as main config command
	enabledStyle := lipgloss.NewStyle().
		Foreground(theme.Success).
		Bold(true)

	disabledStyle := lipgloss.NewStyle().
		Foreground(theme.Muted)

	darkMutedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#666666"))

	// Calculate proper column widths for better alignment (same as main config)
	maxTypeWidth := 0
	maxNameWidth := 0
	for _, formatConfig := range config.Formats {
		handler, exists := fm.registry.GetHandler(formatConfig.Type)
		if !exists {
			continue
		}
		if len(string(formatConfig.Type)) > maxTypeWidth {
			maxTypeWidth = len(string(formatConfig.Type))
		}
		if len(handler.GetDisplayName()) > maxNameWidth {
			maxNameWidth = len(handler.GetDisplayName())
		}
	}

	var availableFormats []tui.SelectOption
	for _, formatConfig := range config.Formats {
		handler, exists := fm.registry.GetHandler(formatConfig.Type)
		if !exists {
			continue
		}

		var status string
		var statusStyle lipgloss.Style
		if formatConfig.Enabled {
			status = statusEnabled
			statusStyle = enabledStyle
		} else {
			status = statusDisabled
			statusStyle = disabledStyle
		}

		// Use raw text for width calculation, then apply styling (same as main config)
		typeText := string(formatConfig.Type)
		nameText := handler.GetDisplayName()

		fmt.Printf("  %s%s %s%s %s\n",
			darkMutedStyle.Render(typeText),
			strings.Repeat(" ", maxTypeWidth-len(typeText)),
			nameText,
			strings.Repeat(" ", maxNameWidth-len(nameText)),
			statusStyle.Render(status))

		// Only add disabled formats to available for enabling
		if !formatConfig.Enabled {
			availableFormats = append(availableFormats, tui.SelectOption{
				Label:       handler.GetDisplayName(),
				Value:       string(formatConfig.Type),
				Description: fmt.Sprintf("Currently %s", strings.Trim(status, "[]")),
			})
		}
	}
	fmt.Println()

	if len(availableFormats) == 0 {
		fmt.Println("All formats are already enabled.")
		return nil
	}

	// Interactive selection
	selectedFormat, err := tui.Select(tui.SelectOptions{
		Title:       "Select format to enable",
		Description: "Choose a format to enable in your project\nPress 'q' or 'esc' to exit",
		Options:     availableFormats,
	})
	if err != nil {
		return fmt.Errorf("format selection cancelled: %w", err)
	}

	if selectedFormat == "" {
		log.Info("No format selected")
		return nil
	}

	// Enable the selected format
	return fm.EnableFormat(context.Background(), &cli.Command{}, selectedFormat)
}

// interactiveDisableFormat provides an interactive interface to disable formats
func (fm *FormatManager) interactiveDisableFormat(_ context.Context, _ *cli.Command) error {
	// Get current directory and load configuration
	currentDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	configResult, err := fm.projectManager.LoadConfig(currentDir)
	if err != nil {
		return fmt.Errorf("no project configuration found. Run 'contexture init' first: %w", err)
	}

	config := configResult.Config

	if len(config.Formats) == 0 {
		fmt.Println("No formats configured to disable")
		return nil
	}

	// Show header
	fmt.Println(ui.CommandHeader("disable format"))

	// Show current formats
	theme := ui.DefaultTheme()
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(theme.Info)

	fmt.Println(headerStyle.Render("Current Formats"))

	// Use exactly the same styling as main config command
	enabledStyle := lipgloss.NewStyle().
		Foreground(theme.Success).
		Bold(true)

	disabledStyle := lipgloss.NewStyle().
		Foreground(theme.Muted)

	darkMutedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#666666"))

	// Calculate proper column widths for better alignment (same as main config)
	maxTypeWidth := 0
	maxNameWidth := 0
	for _, formatConfig := range config.Formats {
		handler, exists := fm.registry.GetHandler(formatConfig.Type)
		if !exists {
			continue
		}
		if len(string(formatConfig.Type)) > maxTypeWidth {
			maxTypeWidth = len(string(formatConfig.Type))
		}
		if len(handler.GetDisplayName()) > maxNameWidth {
			maxNameWidth = len(handler.GetDisplayName())
		}
	}

	// Count enabled formats first
	enabledCount := 0
	for _, formatConfig := range config.Formats {
		if formatConfig.Enabled {
			enabledCount++
		}
	}

	if enabledCount <= 1 {
		fmt.Println("Cannot disable formats - at least one format must remain enabled.")
		return nil
	}

	var availableFormats []tui.SelectOption
	for _, formatConfig := range config.Formats {
		handler, exists := fm.registry.GetHandler(formatConfig.Type)
		if !exists {
			continue
		}

		var status string
		var statusStyle lipgloss.Style
		if formatConfig.Enabled {
			status = statusEnabled
			statusStyle = enabledStyle
		} else {
			status = statusDisabled
			statusStyle = disabledStyle
		}

		// Use raw text for width calculation, then apply styling (same as main config)
		typeText := string(formatConfig.Type)
		nameText := handler.GetDisplayName()

		fmt.Printf("  %s%s %s%s %s\n",
			darkMutedStyle.Render(typeText),
			strings.Repeat(" ", maxTypeWidth-len(typeText)),
			nameText,
			strings.Repeat(" ", maxNameWidth-len(nameText)),
			statusStyle.Render(status))

		// Only add enabled formats to available for disabling
		if formatConfig.Enabled {
			availableFormats = append(availableFormats, tui.SelectOption{
				Label:       handler.GetDisplayName(),
				Value:       string(formatConfig.Type),
				Description: fmt.Sprintf("Currently %s", strings.Trim(status, "[]")),
			})
		}
	}
	fmt.Println()

	if len(availableFormats) == 0 {
		fmt.Println("No formats available to disable.")
		return nil
	}

	// Interactive selection
	selectedFormat, err := tui.Select(tui.SelectOptions{
		Title:       "Select format to disable",
		Description: "Choose a format to disable in your project\nPress 'q' or 'esc' to exit",
		Options:     availableFormats,
	})
	if err != nil {
		return fmt.Errorf("format selection cancelled: %w", err)
	}

	if selectedFormat == "" {
		log.Info("No format selected")
		return nil
	}

	// Disable the selected format
	return fm.DisableFormat(context.Background(), &cli.Command{}, selectedFormat)
}

// Helper methods

// displayProjectFormats displays project format configuration
func (fm *FormatManager) displayProjectFormats(config *domain.Project) error {
	theme := ui.DefaultTheme()

	// Style definitions
	sectionStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(theme.Primary).
		MarginTop(1)

	enabledStyle := lipgloss.NewStyle().
		Foreground(theme.Success).
		Bold(true)

	disabledStyle := lipgloss.NewStyle().
		Foreground(theme.Muted)

	// Display formats configuration
	fmt.Println(ui.CommandHeader("output formats"))
	fmt.Println(sectionStyle.Render("Output Formats"))

	if len(config.Formats) == 0 {
		fmt.Println("No formats currently configured")
		fmt.Println()
		return nil
	}

	return fm.displayFormatsTable(config.Formats, enabledStyle, disabledStyle)
}

// displayFormatsTable displays a formatted table of formats
func (fm *FormatManager) displayFormatsTable(formats []domain.FormatConfig, enabledStyle, disabledStyle lipgloss.Style) error {
	darkMutedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#666666"))

	// Calculate proper column widths for better alignment (same as main config)
	maxTypeWidth := 0
	maxNameWidth := 0
	for _, formatConfig := range formats {
		handler, exists := fm.registry.GetHandler(formatConfig.Type)
		if !exists {
			continue
		}
		if len(string(formatConfig.Type)) > maxTypeWidth {
			maxTypeWidth = len(string(formatConfig.Type))
		}
		if len(handler.GetDisplayName()) > maxNameWidth {
			maxNameWidth = len(handler.GetDisplayName())
		}
	}

	for _, formatConfig := range formats {
		handler, exists := fm.registry.GetHandler(formatConfig.Type)
		if !exists {
			continue
		}

		var status string
		var statusStyle lipgloss.Style
		if formatConfig.Enabled {
			status = statusEnabled
			statusStyle = enabledStyle
		} else {
			status = statusDisabled
			statusStyle = disabledStyle
		}

		// Use raw text for width calculation, then apply styling (same as main config)
		typeText := string(formatConfig.Type)
		nameText := handler.GetDisplayName()

		fmt.Printf("  %s%s %s%s %s\n",
			darkMutedStyle.Render(typeText),
			strings.Repeat(" ", maxTypeWidth-len(typeText)),
			nameText,
			strings.Repeat(" ", maxNameWidth-len(nameText)),
			statusStyle.Render(status))
	}

	fmt.Println()
	return nil
}

// addFormatToProjectConfig adds a format to project configuration
func (fm *FormatManager) addFormatToProjectConfig(configResult *domain.ConfigResult, formatType, currentDir string) error {
	config := configResult.Config

	// Validate format type
	if !fm.registry.IsSupported(domain.FormatType(formatType)) {
		return fmt.Errorf("unsupported format: %s", formatType)
	}

	// Check if format already exists
	for i, format := range config.Formats {
		if format.Type == domain.FormatType(formatType) {
			// Enable if disabled
			if !format.Enabled {
				config.Formats[i].Enabled = true
			} else {
				fmt.Printf("Format %s is already enabled\n", formatType)
				return nil
			}
			break
		}
	}

	// Add new format if not found
	var found bool
	for _, format := range config.Formats {
		if format.Type == domain.FormatType(formatType) {
			found = true
			break
		}
	}

	if !found {
		newFormat := domain.FormatConfig{
			Type:    domain.FormatType(formatType),
			Enabled: true,
		}
		config.Formats = append(config.Formats, newFormat)
	}

	// Save configuration
	if err := fm.projectManager.SaveConfig(config, configResult.Location, currentDir); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	// Create format directories if needed
	if err := fm.registry.CreateFormatDirectories(domain.FormatType(formatType)); err != nil {
		log.Warn("Failed to create format directories", "error", err)
	}

	theme := ui.DefaultTheme()
	successStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(theme.Success)

	displayName := fm.getFormatDisplayName(domain.FormatType(formatType))
	fmt.Println(successStyle.Render("Format added: " + displayName))
	return nil
}

// getFormatDisplayName returns the display name for a format type
func (fm *FormatManager) getFormatDisplayName(formatType domain.FormatType) string {
	if handler, exists := fm.registry.GetHandler(formatType); exists {
		return handler.GetDisplayName()
	}
	return string(formatType)
}

// getFormatOutputPath returns the output file/directory path for a given format type
func (fm *FormatManager) getFormatOutputPath(formatType domain.FormatType) string {
	// Create a format instance to get the output path
	format, err := fm.registry.CreateFormat(formatType, fm.fs, nil)
	if err != nil {
		log.Debug("Failed to create format for output path", "formatType", formatType, "error", err)
		return "unknown"
	}

	// Use empty config to get the default path
	config := &domain.FormatConfig{Type: formatType}
	outputPath := format.GetOutputPath(config)

	// Add trailing slash for directory-based formats
	metadata := format.GetMetadata()
	if metadata.IsDirectory && !strings.HasSuffix(outputPath, "/") {
		return outputPath + "/"
	}

	return outputPath
}

// enableFormatInProjectConfig enables a format in project configuration
func (fm *FormatManager) enableFormatInProjectConfig(configResult *domain.ConfigResult, formatType, currentDir string) error {
	config := configResult.Config

	// Find and enable the format
	var found bool
	for i, format := range config.Formats {
		if format.Type == domain.FormatType(formatType) {
			found = true
			if format.Enabled {
				fmt.Printf("Format %s is already enabled\n", formatType)
				return nil
			}
			config.Formats[i].Enabled = true
			break
		}
	}

	if !found {
		return fmt.Errorf(
			"format %s is not configured. Add it first with 'contexture config formats add %s'",
			formatType,
			formatType,
		)
	}

	// Save configuration
	if err := fm.projectManager.SaveConfig(config, configResult.Location, currentDir); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	// Create format directories if needed
	if err := fm.registry.CreateFormatDirectories(domain.FormatType(formatType)); err != nil {
		log.Warn("Failed to create format directories", "error", err)
	}

	theme := ui.DefaultTheme()
	successStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(theme.Success)

	displayName := fm.getFormatDisplayName(domain.FormatType(formatType))
	fmt.Println(successStyle.Render("Format enabled: " + displayName))
	return nil
}

// disableFormatInProjectConfig disables a format in project configuration
func (fm *FormatManager) disableFormatInProjectConfig(configResult *domain.ConfigResult, formatType, currentDir string) error {
	config := configResult.Config

	// Count enabled formats
	enabledCount := 0
	for _, format := range config.Formats {
		if format.Enabled {
			enabledCount++
		}
	}

	if enabledCount <= 1 {
		return fmt.Errorf(
			"cannot disable the last enabled format. At least one format must remain enabled",
		)
	}

	// Find and disable the format
	var found bool
	for i, format := range config.Formats {
		if format.Type == domain.FormatType(formatType) {
			found = true
			if !format.Enabled {
				fmt.Printf("Format %s is already disabled\n", formatType)
				return nil
			}
			config.Formats[i].Enabled = false
			break
		}
	}

	if !found {
		return fmt.Errorf("format %s is not configured", formatType)
	}

	// Save configuration
	if err := fm.projectManager.SaveConfig(config, configResult.Location, currentDir); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	theme := ui.DefaultTheme()
	successStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(theme.Success)

	displayName := fm.getFormatDisplayName(domain.FormatType(formatType))
	fmt.Println(successStyle.Render("Format disabled: " + displayName))
	return nil
}

// removeFormatFromProjectConfig removes a format from project configuration
func (fm *FormatManager) removeFormatFromProjectConfig(configResult *domain.ConfigResult, formatType, currentDir string) error {
	config := configResult.Config

	// Find and remove the format
	var found bool
	var removedEnabled bool
	for i, format := range config.Formats {
		if format.Type == domain.FormatType(formatType) {
			found = true
			removedEnabled = format.Enabled

			// Remove format from slice
			config.Formats = append(config.Formats[:i], config.Formats[i+1:]...)
			break
		}
	}

	if !found {
		return fmt.Errorf("format %s is not configured", formatType)
	}

	// Check if we're removing the last enabled format
	if removedEnabled {
		enabledCount := 0
		for _, format := range config.Formats {
			if format.Enabled {
				enabledCount++
			}
		}
		if enabledCount == 0 {
			return fmt.Errorf(
				"cannot remove the last enabled format. Enable another format first or add a new one",
			)
		}
	}

	// Save configuration
	if err := fm.projectManager.SaveConfig(config, configResult.Location, currentDir); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	theme := ui.DefaultTheme()
	successStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(theme.Success)

	displayName := fm.getFormatDisplayName(domain.FormatType(formatType))
	fmt.Println(successStyle.Render("Format removed: " + displayName))
	return nil
}
