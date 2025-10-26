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

// ProviderWithSource tracks a provider's source
type ProviderWithSource struct {
	Provider domain.Provider
	Source   string // "built-in", "global", "project"
}

// ListAction lists all available providers
func (c *ProvidersCommand) ListAction(_ context.Context, _ *cli.Command, deps *dependencies.Dependencies) error {
	// Show header
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.AdaptiveColor{Light: "#F793FF", Dark: "#AD58B4"})
	fmt.Printf("%s\n\n", headerStyle.Render("Providers"))

	// Get current directory
	currentDir, err := os.Getwd()
	if err != nil {
		return contextureerrors.Wrap(err, "get current directory")
	}

	// Collect providers with source information
	providersWithSource := make([]ProviderWithSource, 0)
	providerNames := make(map[string]bool)

	// Load global config providers
	globalResult, err := c.projectManager.LoadGlobalConfig()
	if err == nil && globalResult != nil && globalResult.Config != nil {
		for _, provider := range globalResult.Config.Providers {
			providersWithSource = append(providersWithSource, ProviderWithSource{
				Provider: provider,
				Source:   "global",
			})
			providerNames[provider.Name] = true
		}
	}

	// Load project config providers
	projectResult, err := c.projectManager.LoadConfig(currentDir)
	if err == nil && projectResult != nil && projectResult.Config != nil {
		for _, provider := range projectResult.Config.Providers {
			source := "project"
			if providerNames[provider.Name] {
				source = "project (overrides global)"
			}
			providersWithSource = append(providersWithSource, ProviderWithSource{
				Provider: provider,
				Source:   source,
			})
			providerNames[provider.Name] = true
		}
	}

	// Add the built-in default provider if not already overridden
	if !providerNames[domain.DefaultProviderName] {
		// Get default provider from registry
		defaultProvider, err := deps.ProviderRegistry.Get(domain.DefaultProviderName)
		if err == nil {
			providersWithSource = append([]ProviderWithSource{
				{
					Provider: *defaultProvider,
					Source:   "built-in",
				},
			}, providersWithSource...)
		}
	}

	if len(providersWithSource) == 0 {
		fmt.Println("No providers configured")
		return nil
	}

	theme := ui.DefaultTheme()
	nameStyle := lipgloss.NewStyle().Bold(true).Foreground(theme.Primary)
	urlStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#666666"))

	for _, pws := range providersWithSource {
		// Build provider name with source indicator
		nameWithSource := fmt.Sprintf("@%s [%s]", pws.Provider.Name, pws.Source)

		fmt.Printf("  %s\n", nameStyle.Render(nameWithSource))
		fmt.Printf("    %s\n", urlStyle.Render(pws.Provider.URL))
		if pws.Provider.DefaultBranch != "" && pws.Provider.DefaultBranch != "main" {
			fmt.Printf("    Branch: %s\n", pws.Provider.DefaultBranch)
		}
	}

	return nil
}

// AddAction adds a new provider to the project configuration
func (c *ProvidersCommand) AddAction(_ context.Context, cmd *cli.Command, _ *dependencies.Dependencies, name, url string) error {
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

	isGlobal := cmd.Bool("global")

	// Load configuration based on global flag
	var config *domain.Project
	var err error

	if isGlobal {
		// Initialize global config if needed
		if err := c.projectManager.InitializeGlobalConfig(); err != nil {
			return contextureerrors.Wrap(err, "initialize global config")
		}

		// Load global configuration
		globalResult, err := c.projectManager.LoadGlobalConfig()
		if err != nil {
			return contextureerrors.Wrap(err, "load global configuration")
		}
		config = globalResult.Config
	} else {
		// Get current directory and load project configuration
		currentDir, err := os.Getwd()
		if err != nil {
			return contextureerrors.Wrap(err, "get current directory")
		}

		configResult, err := c.projectManager.LoadConfig(currentDir)
		if err != nil {
			return contextureerrors.Wrap(err, "load config")
		}
		config = configResult.Config
	}

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

	// Save config based on global flag
	if isGlobal {
		if err = c.projectManager.SaveGlobalConfig(config); err != nil {
			return contextureerrors.Wrap(err, "save global config")
		}
	} else {
		currentDir, err := os.Getwd()
		if err != nil {
			return contextureerrors.Wrap(err, "get current directory")
		}
		location := c.projectManager.GetConfigLocation(currentDir, false)
		if err = c.projectManager.SaveConfig(config, location, currentDir); err != nil {
			return contextureerrors.Wrap(err, "save config")
		}
	}

	theme := ui.DefaultTheme()
	successStyle := lipgloss.NewStyle().Bold(true).Foreground(theme.Success)
	fmt.Println(successStyle.Render("Provider added successfully!"))
	fmt.Printf("  @%s → %s\n", name, url)

	return nil
}

// RemoveAction removes a provider from the project configuration
func (c *ProvidersCommand) RemoveAction(_ context.Context, cmd *cli.Command, _ *dependencies.Dependencies, name string) error {
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

	isGlobal := cmd.Bool("global")

	// Load configuration based on global flag
	var config *domain.Project

	if isGlobal {
		// Load global configuration
		globalResult, err := c.projectManager.LoadGlobalConfig()
		if err != nil {
			return contextureerrors.Wrap(err, "load global configuration")
		}
		if globalResult == nil || globalResult.Config == nil {
			return contextureerrors.ValidationErrorf("global config", "global configuration not found")
		}
		config = globalResult.Config
	} else {
		// Get current directory and load project configuration
		currentDir, err := os.Getwd()
		if err != nil {
			return contextureerrors.Wrap(err, "get current directory")
		}

		configResult, err := c.projectManager.LoadConfig(currentDir)
		if err != nil {
			return contextureerrors.Wrap(err, "load config")
		}
		config = configResult.Config
	}

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

	// Save config based on global flag
	if isGlobal {
		if err := c.projectManager.SaveGlobalConfig(config); err != nil {
			return contextureerrors.Wrap(err, "save global config")
		}
	} else {
		currentDir, saveErr := os.Getwd()
		if saveErr != nil {
			return contextureerrors.Wrap(saveErr, "get current directory")
		}
		location := c.projectManager.GetConfigLocation(currentDir, false)
		if err := c.projectManager.SaveConfig(config, location, currentDir); err != nil {
			return contextureerrors.Wrap(err, "save config")
		}
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

	// Get current directory
	currentDir, err := os.Getwd()
	if err != nil {
		return contextureerrors.Wrap(err, "get current directory")
	}

	// Load global config providers first
	var providerSource string
	globalResult, err := c.projectManager.LoadGlobalConfig()
	if err == nil && globalResult != nil && globalResult.Config != nil {
		if err := deps.ProviderRegistry.LoadFromProject(globalResult.Config); err != nil {
			return contextureerrors.Wrap(err, "load global providers")
		}
	}

	// Load project config providers (these will override global if same name)
	configResult, err := c.projectManager.LoadConfig(currentDir)
	if err == nil && configResult != nil && configResult.Config != nil {
		// Check if provider exists in project config
		projectProvider := configResult.Config.GetProviderByName(name)
		if projectProvider != nil {
			providerSource = "project"
		}

		if err := deps.ProviderRegistry.LoadFromProject(configResult.Config); err != nil {
			return contextureerrors.Wrap(err, "load project providers")
		}
	}

	// Get provider from registry
	provider, err := deps.ProviderRegistry.Get(name)
	if err != nil {
		return contextureerrors.ValidationErrorf("name", "provider '@%s' not found", name)
	}

	// Determine source if not already set
	if providerSource == "" {
		if provider.Name == domain.DefaultProviderName {
			providerSource = "built-in"
		} else if globalResult != nil && globalResult.Config != nil {
			globalProvider := globalResult.Config.GetProviderByName(name)
			if globalProvider != nil {
				providerSource = "global"
			}
		}
	}

	theme := ui.DefaultTheme()
	nameStyle := lipgloss.NewStyle().Bold(true).Foreground(theme.Primary)
	labelStyle := lipgloss.NewStyle().Bold(true)
	sourceStyle := lipgloss.NewStyle().Foreground(theme.Muted).Italic(true)

	fmt.Printf("%s\n", nameStyle.Render("@"+provider.Name))
	fmt.Printf("  %s %s\n", labelStyle.Render("URL:"), provider.URL)
	if provider.DefaultBranch != "" {
		fmt.Printf("  %s %s\n", labelStyle.Render("Branch:"), provider.DefaultBranch)
	}
	if provider.Auth != nil {
		fmt.Printf("  %s %s\n", labelStyle.Render("Auth:"), provider.Auth.Type)
	}
	if providerSource != "" {
		fmt.Printf("  %s %s\n", labelStyle.Render("Source:"), sourceStyle.Render(providerSource))
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
