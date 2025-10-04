// Package commands provides CLI command implementations for the main config command
package commands

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"github.com/contextureai/contexture/internal/dependencies"
	contextureerrors "github.com/contextureai/contexture/internal/errors"
	"github.com/contextureai/contexture/internal/format"
	"github.com/contextureai/contexture/internal/project"
	"github.com/contextureai/contexture/internal/ui"
	"github.com/spf13/afero"
	"github.com/urfave/cli/v3"
)

const (
	// Status display constants
	statusEnabled  = "[enabled]"
	statusDisabled = "[disabled]"
)

// MainConfigCommand implements the main config command for showing project configuration
type MainConfigCommand struct {
	projectManager *project.Manager
	registry       *format.Registry
	fs             afero.Fs
}

// NewMainConfigCommand creates a new main config command
func NewMainConfigCommand(deps *dependencies.Dependencies) *MainConfigCommand {
	return &MainConfigCommand{
		projectManager: project.NewManager(deps.FS),
		registry:       format.GetDefaultRegistry(deps.FS),
		fs:             deps.FS,
	}
}

// Execute runs the main config command (shows project configuration)
func (c *MainConfigCommand) Execute(_ context.Context, _ *cli.Command) error {
	// Get current directory and load configuration
	currentDir, err := os.Getwd()
	if err != nil {
		return contextureerrors.Wrap(err, "get current directory")
	}

	configResult, err := c.projectManager.LoadConfigWithLocalRules(currentDir)
	if err != nil {
		fmt.Println("No project configuration found")
		fmt.Println("Initialize a project with: contexture init")
		return err
	}

	config := configResult.Config
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

	darkMutedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#666666"))

	// Display project configuration
	fmt.Println(ui.CommandHeader("project configuration"))

	// Display formats configuration
	fmt.Println(sectionStyle.Render("Output Formats"))
	if len(config.Formats) == 0 {
		fmt.Println("No formats currently configured")
		fmt.Println()
	} else {
		// Calculate proper column widths for better alignment
		maxTypeWidth := 0
		maxNameWidth := 0
		for _, formatConfig := range config.Formats {
			handler, exists := c.registry.GetHandler(formatConfig.Type)
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
			handler, exists := c.registry.GetHandler(formatConfig.Type)
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

			// Use raw text for width calculation, then apply styling
			typeText := string(formatConfig.Type)
			nameText := handler.GetDisplayName()

			fmt.Printf("  %s%s %s%s %s\n",
				darkMutedStyle.Render(typeText),
				strings.Repeat(" ", maxTypeWidth-len(typeText)),
				nameText,
				strings.Repeat(" ", maxNameWidth-len(nameText)),
				statusStyle.Render(status))
		}
	}

	// Display local rules
	localRules := []string{}
	for _, rule := range config.Rules {
		if rule.Source == "local" {
			localRules = append(localRules, rule.ID)
		}
	}

	if len(localRules) > 0 {
		fmt.Println(sectionStyle.Render("Local Rules"))

		for _, ruleID := range localRules {
			fmt.Printf("  %s %s\n", darkMutedStyle.Render("•"), ruleID)
		}
	}

	// Display cache configuration
	fmt.Println(sectionStyle.Render("Cache Configuration"))
	cacheDir := filepath.Join(os.TempDir(), "contexture")

	fmt.Printf("  %s %s\n",
		darkMutedStyle.Render("directory:"),
		cacheDir)

	// Check if cache directory exists and count cached repositories
	if info, err := c.fs.Stat(cacheDir); err == nil && info.IsDir() {
		// Count cached repositories
		files, err := afero.ReadDir(c.fs, cacheDir)
		if err == nil {
			repoCount := 0
			for _, file := range files {
				if file.IsDir() {
					repoCount++
				}
			}
			fmt.Printf("  %s %d cached repositories\n",
				darkMutedStyle.Render("status:"),
				repoCount)
		} else {
			fmt.Printf("  %s directory exists\n",
				darkMutedStyle.Render("status:"))
		}
	} else {
		fmt.Printf("  %s no cache directory (will be created on first use)\n",
			darkMutedStyle.Render("status:"))
	}

	log.Debug("Configuration displayed",
		"config_path", configResult.Path,
		"location", configResult.Location,
		"rules_count", len(config.Rules),
		"formats_count", len(config.Formats),
		"cache_dir", cacheDir)

	return nil
}

// ConfigAction is the CLI action handler for the main config command
func ConfigAction(ctx context.Context, cmd *cli.Command, deps *dependencies.Dependencies) error {
	configCmd := NewMainConfigCommand(deps)
	return configCmd.Execute(ctx, cmd)
}
