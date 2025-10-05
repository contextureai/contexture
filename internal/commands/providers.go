// Package commands provides CLI command implementations
package commands

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/contextureai/contexture/internal/dependencies"
	"github.com/contextureai/contexture/internal/domain"
	contextureerrors "github.com/contextureai/contexture/internal/errors"
	"github.com/contextureai/contexture/internal/project"
	"github.com/contextureai/contexture/internal/ui"
	"github.com/urfave/cli/v3"
)

// ProvidersCommand implements the providers command
type ProvidersCommand struct {
	projectManager *project.Manager
}

// NewProvidersCommand creates a new providers command
func NewProvidersCommand(deps *dependencies.Dependencies) *ProvidersCommand {
	return &ProvidersCommand{
		projectManager: project.NewManager(deps.FS),
	}
}

// ListAction lists all available providers
func (c *ProvidersCommand) ListAction(_ context.Context, _ *cli.Command, deps *dependencies.Dependencies) error {
	// Show header
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.AdaptiveColor{Light: "#F793FF", Dark: "#AD58B4"})
	fmt.Printf("%s\n\n", headerStyle.Render("Providers"))

	// Get current directory and load configuration
	currentDir, err := os.Getwd()
	if err != nil {
		return contextureerrors.Wrap(err, "get current directory")
	}

	// Load project config to get custom providers
	configResult, err := c.projectManager.LoadConfig(currentDir)
	if err != nil {
		// If no config, just show default providers
		fmt.Println("Default providers:")
	} else {
		// Load providers into registry
		if err := deps.ProviderRegistry.LoadFromProject(configResult.Config); err != nil {
			return contextureerrors.Wrap(err, "load providers")
		}
		fmt.Println("Available providers:")
	}

	// List all providers from registry
	providers := deps.ProviderRegistry.ListProviders()
	if len(providers) == 0 {
		fmt.Println("No providers configured")
		return nil
	}

	theme := ui.DefaultTheme()
	nameStyle := lipgloss.NewStyle().Bold(true).Foreground(theme.Primary)
	urlStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#666666"))
	defaultStyle := lipgloss.NewStyle().Foreground(theme.Success).Italic(true)

	for _, provider := range providers {
		// Build provider name with default indicator if applicable
		nameWithDefault := "@" + provider.Name
		if provider.Name == domain.DefaultProviderName {
			nameWithDefault += " " + defaultStyle.Render("(default)")
		}

		fmt.Printf("  %s\n", nameStyle.Render(nameWithDefault))
		fmt.Printf("    %s\n", urlStyle.Render(provider.URL))
		if provider.DefaultBranch != "" && provider.DefaultBranch != "main" {
			fmt.Printf("    Branch: %s\n", provider.DefaultBranch)
		}
	}

	return nil
}

// AddAction adds a new provider to the project configuration
func (c *ProvidersCommand) AddAction(_ context.Context, _ *cli.Command, _ *dependencies.Dependencies, name, url string) error {
	// Show header
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.AdaptiveColor{Light: "#F793FF", Dark: "#AD58B4"})
	fmt.Printf("%s\n\n", headerStyle.Render("Add Provider"))

	// Validate inputs
	if name == "" {
		return contextureerrors.ValidationErrorf("name", "provider name cannot be empty")
	}
	if url == "" {
		return contextureerrors.ValidationErrorf("url", "provider URL cannot be empty")
	}

	// Check if name is reserved
	if name == "local" {
		return contextureerrors.ValidationErrorf("name", "provider name 'local' is reserved")
	}

	// Get current directory and load configuration
	currentDir, err := os.Getwd()
	if err != nil {
		return contextureerrors.Wrap(err, "get current directory")
	}

	configResult, err := c.projectManager.LoadConfig(currentDir)
	if err != nil {
		return contextureerrors.Wrap(err, "load config")
	}
	config := configResult.Config

	// Check if provider already exists
	if existing := config.GetProviderByName(name); existing != nil {
		return contextureerrors.ValidationErrorf("name", "provider '%s' already exists", name)
	}

	// Create new provider
	newProvider := domain.Provider{
		Name: name,
		URL:  url,
	}

	// Add to config
	config.Providers = append(config.Providers, newProvider)

	// Save config
	location := c.projectManager.GetConfigLocation(currentDir, false)
	if err := c.projectManager.SaveConfig(config, location, currentDir); err != nil {
		return contextureerrors.Wrap(err, "save config")
	}

	theme := ui.DefaultTheme()
	successStyle := lipgloss.NewStyle().Bold(true).Foreground(theme.Success)
	fmt.Println(successStyle.Render("Provider added successfully!"))
	fmt.Printf("  @%s → %s\n", name, url)

	return nil
}

// RemoveAction removes a provider from the project configuration
func (c *ProvidersCommand) RemoveAction(_ context.Context, _ *cli.Command, _ *dependencies.Dependencies, name string) error {
	// Show header
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.AdaptiveColor{Light: "#F793FF", Dark: "#AD58B4"})
	fmt.Printf("%s\n\n", headerStyle.Render("Remove Provider"))

	// Validate input
	if name == "" {
		return contextureerrors.ValidationErrorf("name", "provider name cannot be empty")
	}

	// Cannot remove default provider
	if name == domain.DefaultProviderName {
		return contextureerrors.ValidationErrorf("name", "cannot remove default provider '%s'", domain.DefaultProviderName)
	}

	// Get current directory and load configuration
	currentDir, err := os.Getwd()
	if err != nil {
		return contextureerrors.Wrap(err, "get current directory")
	}

	configResult, err := c.projectManager.LoadConfig(currentDir)
	if err != nil {
		return contextureerrors.Wrap(err, "load config")
	}
	config := configResult.Config

	// Find and remove provider
	found := false
	newProviders := make([]domain.Provider, 0, len(config.Providers))
	for _, provider := range config.Providers {
		if provider.Name == name {
			found = true
			continue
		}
		newProviders = append(newProviders, provider)
	}

	if !found {
		return contextureerrors.ValidationErrorf("name", "provider '%s' not found", name)
	}

	config.Providers = newProviders

	// Save config
	location := c.projectManager.GetConfigLocation(currentDir, false)
	if err := c.projectManager.SaveConfig(config, location, currentDir); err != nil {
		return contextureerrors.Wrap(err, "save config")
	}

	theme := ui.DefaultTheme()
	successStyle := lipgloss.NewStyle().Bold(true).Foreground(theme.Success)
	fmt.Println(successStyle.Render("Provider removed successfully!"))
	fmt.Printf("  @%s\n", name)

	return nil
}

// ShowAction shows details for a specific provider
func (c *ProvidersCommand) ShowAction(_ context.Context, _ *cli.Command, deps *dependencies.Dependencies, name string) error {
	// Show header
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.AdaptiveColor{Light: "#F793FF", Dark: "#AD58B4"})
	fmt.Printf("%s\n\n", headerStyle.Render("Provider Details"))

	// Validate input
	if name == "" {
		return contextureerrors.ValidationErrorf("name", "provider name cannot be empty")
	}

	// Strip @ prefix if provided
	name = strings.TrimPrefix(name, "@")

	// Get current directory and load configuration
	currentDir, err := os.Getwd()
	if err != nil {
		return contextureerrors.Wrap(err, "get current directory")
	}

	// Load project config to get custom providers
	configResult, err := c.projectManager.LoadConfig(currentDir)
	if err == nil {
		// Load providers into registry
		if err := deps.ProviderRegistry.LoadFromProject(configResult.Config); err != nil {
			return contextureerrors.Wrap(err, "load providers")
		}
	}

	// Get provider from registry
	provider, err := deps.ProviderRegistry.Get(name)
	if err != nil {
		return contextureerrors.ValidationErrorf("name", "provider '@%s' not found", name)
	}

	theme := ui.DefaultTheme()
	nameStyle := lipgloss.NewStyle().Bold(true).Foreground(theme.Primary)
	labelStyle := lipgloss.NewStyle().Bold(true)
	defaultStyle := lipgloss.NewStyle().Foreground(theme.Success).Italic(true)

	fmt.Printf("%s\n", nameStyle.Render("@"+provider.Name))
	fmt.Printf("  %s %s\n", labelStyle.Render("URL:"), provider.URL)
	if provider.DefaultBranch != "" {
		fmt.Printf("  %s %s\n", labelStyle.Render("Branch:"), provider.DefaultBranch)
	}
	if provider.Auth != nil {
		fmt.Printf("  %s %s\n", labelStyle.Render("Auth:"), provider.Auth.Type)
	}
	if provider.Name == domain.DefaultProviderName {
		fmt.Printf("  %s\n", defaultStyle.Render("(default provider)"))
	}

	return nil
}

// ProvidersAction is the default action when running 'contexture providers'
func ProvidersAction(ctx context.Context, cmd *cli.Command, deps *dependencies.Dependencies) error {
	providersCmd := NewProvidersCommand(deps)
	return providersCmd.ListAction(ctx, cmd, deps)
}

// ProvidersListAction handles 'contexture providers list'
func ProvidersListAction(ctx context.Context, cmd *cli.Command, deps *dependencies.Dependencies) error {
	providersCmd := NewProvidersCommand(deps)
	return providersCmd.ListAction(ctx, cmd, deps)
}

// ProvidersAddAction handles 'contexture providers add <name> <url>'
func ProvidersAddAction(ctx context.Context, cmd *cli.Command, deps *dependencies.Dependencies) error {
	args := cmd.Args().Slice()
	if len(args) < 2 {
		return contextureerrors.ValidationErrorf("args", "usage: contexture providers add <name> <url>")
	}

	name := args[0]
	url := args[1]

	providersCmd := NewProvidersCommand(deps)
	return providersCmd.AddAction(ctx, cmd, deps, name, url)
}

// ProvidersRemoveAction handles 'contexture providers remove <name>'
func ProvidersRemoveAction(ctx context.Context, cmd *cli.Command, deps *dependencies.Dependencies) error {
	args := cmd.Args().Slice()
	if len(args) < 1 {
		return contextureerrors.ValidationErrorf("args", "usage: contexture providers remove <name>")
	}

	name := args[0]

	providersCmd := NewProvidersCommand(deps)
	return providersCmd.RemoveAction(ctx, cmd, deps, name)
}

// ProvidersShowAction handles 'contexture providers show <name>'
func ProvidersShowAction(ctx context.Context, cmd *cli.Command, deps *dependencies.Dependencies) error {
	args := cmd.Args().Slice()
	if len(args) < 1 {
		return contextureerrors.ValidationErrorf("args", "usage: contexture providers show <name>")
	}

	name := args[0]

	providersCmd := NewProvidersCommand(deps)
	return providersCmd.ShowAction(ctx, cmd, deps, name)
}
