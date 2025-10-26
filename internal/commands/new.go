package commands

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/contextureai/contexture/internal/dependencies"
	"github.com/contextureai/contexture/internal/domain"
	contextureerrors "github.com/contextureai/contexture/internal/errors"
	"github.com/contextureai/contexture/internal/project"
	"github.com/spf13/afero"
	"github.com/urfave/cli/v3"
	"gopkg.in/yaml.v3"
)

// NewCommand implements the new command
type NewCommand struct {
	projectManager *project.Manager
	fs             afero.Fs
}

// NewNewCommand creates a new NewCommand instance
func NewNewCommand(deps *dependencies.Dependencies) *NewCommand {
	return &NewCommand{
		projectManager: project.NewManager(deps.FS),
		fs:             deps.FS,
	}
}

// Execute runs the new command
func (c *NewCommand) Execute(_ context.Context, cmd *cli.Command, rulePath, workingDir string) error {
	// Get flags
	name := cmd.String("name")
	description := cmd.String("description")
	tagsStr := cmd.String("tags")
	isGlobal := cmd.Bool("global")

	// Parse tags (only if provided)
	var tags []string
	if tagsStr != "" {
		tags = parseTags(tagsStr)
	}

	// Determine the target path
	targetPath := c.determineTargetPath(workingDir, rulePath, isGlobal)

	// Check if file already exists
	exists, err := afero.Exists(c.fs, targetPath)
	if err != nil {
		return contextureerrors.Wrap(err, "check if file exists")
	}
	if exists {
		return contextureerrors.ValidationErrorf("file", "rule file already exists: %s", targetPath)
	}

	// Create parent directories
	parentDir := filepath.Dir(targetPath)
	if err := c.fs.MkdirAll(parentDir, 0o755); err != nil {
		return contextureerrors.Wrap(err, "create parent directories")
	}

	// Generate rule content
	content, err := c.generateRuleContent(name, description, tags)
	if err != nil {
		return contextureerrors.Wrap(err, "generate rule content")
	}

	// Write file
	if err := afero.WriteFile(c.fs, targetPath, []byte(content), 0o644); err != nil {
		return contextureerrors.Wrap(err, "write rule file")
	}

	// Show success message
	successStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FF69B4"))

	fmt.Printf("\n%s\n", successStyle.Render("Rule created successfully!"))
	fmt.Printf("  Location: %s\n", targetPath)
	if name != "" {
		fmt.Printf("  Title: %s\n", name)
	}
	if description != "" {
		fmt.Printf("  Description: %s\n", description)
	}
	if len(tags) > 0 {
		fmt.Printf("  Tags: %s\n", strings.Join(tags, ", "))
	}
	fmt.Println()

	return nil
}

// determineTargetPath determines where to create the rule file
func (c *NewCommand) determineTargetPath(workingDir, rulePath string, isGlobal bool) string {
	// Normalize the rule path - remove .md extension if provided
	// This ensures consistent handling whether user provides "my-rule" or "my-rule.md"
	rulePath = strings.TrimSuffix(rulePath, ".md")

	// If global flag is set, use global rules directory
	if isGlobal {
		globalDir, err := domain.GetGlobalConfigDir()
		if err == nil {
			rulesDir := filepath.Join(globalDir, domain.LocalRulesDir)
			return filepath.Join(rulesDir, rulePath+".md")
		}
		// If we can't get global dir, fall through to normal logic
	}

	// Try to load project config
	configResult, err := c.projectManager.LoadConfig(workingDir)

	if err == nil {
		// We're in a Contexture project - create in rules directory
		var rulesDir string

		// Determine the rules directory location based on config location
		switch configResult.Location {
		case domain.ConfigLocationContexture:
			// Config is in .contexture/.contexture.yaml
			// Rules should be in .contexture/rules/
			contextureDir := filepath.Dir(configResult.Path)
			rulesDir = filepath.Join(contextureDir, domain.LocalRulesDir)
		case domain.ConfigLocationRoot:
			// Config is in .contexture.yaml at root
			// Rules should be in ./rules/
			rulesDir = filepath.Join(filepath.Dir(configResult.Path), domain.LocalRulesDir)
		case domain.ConfigLocationGlobal:
			// Global config location - use global rules directory
			globalDir := filepath.Dir(configResult.Path)
			rulesDir = filepath.Join(globalDir, domain.LocalRulesDir)
		default:
			// Fallback to current directory
			rulesDir = filepath.Join(filepath.Dir(configResult.Path), domain.LocalRulesDir)
		}

		// Construct the full path (rulePath is already normalized without .md)
		return filepath.Join(rulesDir, rulePath+".md")
	}

	// Not in a Contexture project - use literal path
	// rulePath is already normalized (no .md), so we can safely add it
	if filepath.IsAbs(rulePath) {
		return rulePath + ".md"
	}
	return filepath.Join(workingDir, rulePath+".md")
}

// generateRuleContent generates the rule file content with YAML frontmatter
func (c *NewCommand) generateRuleContent(name, description string, tags []string) (string, error) {
	// Create frontmatter structure - always include title and description
	frontmatter := map[string]any{
		"title":       name,
		"description": description,
		"trigger":     "manual",
	}

	// Only include tags if provided
	if len(tags) > 0 {
		frontmatter["tags"] = tags
	}

	// Marshal to YAML
	yamlBytes, err := yaml.Marshal(frontmatter)
	if err != nil {
		return "", contextureerrors.Wrap(err, "marshal frontmatter")
	}

	// Build the complete content
	var sb strings.Builder
	sb.WriteString("---\n")
	sb.Write(yamlBytes)
	sb.WriteString("---\n\n")

	// Add heading if name is provided
	if name != "" {
		sb.WriteString(fmt.Sprintf("# %s\n\n", name))
	}

	// Add description if provided
	if description != "" {
		sb.WriteString(fmt.Sprintf("%s\n", description))
	}

	return sb.String(), nil
}

// parseTags parses a comma-separated list of tags
func parseTags(tagsStr string) []string {
	parts := strings.Split(tagsStr, ",")
	tags := make([]string, 0, len(parts))
	for _, part := range parts {
		tag := strings.TrimSpace(part)
		if tag != "" {
			tags = append(tags, tag)
		}
	}
	return tags
}

// NewAction is the CLI action handler for the new command
func NewAction(ctx context.Context, cmd *cli.Command, deps *dependencies.Dependencies) error {
	args := cmd.Args().Slice()

	// Validate that we have a path argument
	if len(args) == 0 {
		return contextureerrors.ValidationErrorf("path", "no path provided")
	}

	// Get current working directory
	workingDir, err := os.Getwd()
	if err != nil {
		return contextureerrors.Wrap(err, "get current directory")
	}

	rulePath := args[0]
	newCmd := NewNewCommand(deps)
	return newCmd.Execute(ctx, cmd, rulePath, workingDir)
}
