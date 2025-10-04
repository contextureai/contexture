package cli

import (
	"io"
	"strings"
	"text/template"

	"github.com/charmbracelet/lipgloss"
	contextureerrors "github.com/contextureai/contexture/internal/errors"
	"github.com/urfave/cli/v3"
)

// Constants for template patterns and layout
const (
	// Template patterns for detection
	appTemplatePattern     = "{{.Name}} - {{.Usage}}"
	commandTemplatePattern = "{{.HelpName}} - {{.Usage}}"

	// Layout constants
	commandPadding  = 3
	defaultMaxWidth = 80

	// Footer messages
	commandHelpFooter = "Use \"contexture [command] --help\" for more information about a command."
)

// HelpPrinter defines the interface for help printing
type HelpPrinter interface {
	Print(w io.Writer, templ string, data any) error
}

// Renderer defines the interface for rendering help content
type Renderer interface {
	RenderApp(cmd *cli.Command) (string, error)
	RenderCommand(cmd *cli.Command) (string, error)
	RenderDefault(templ string, data any) (string, error)
}

// StyleProvider defines the interface for providing styles
type StyleProvider interface {
	TitleStyle() lipgloss.Style
	HeaderStyle() lipgloss.Style
	CommandStyle() lipgloss.Style
	DescriptionStyle() lipgloss.Style
}

// helpPrinter implements the HelpPrinter interface
type helpPrinter struct {
	renderer Renderer
}

// NewHelpPrinter creates a new help printer with default configuration
func NewHelpPrinter() HelpPrinter {
	return &helpPrinter{
		renderer: NewRenderer(NewStyleProvider()),
	}
}

// NewHelpPrinterWithRenderer creates a help printer with a custom renderer
func NewHelpPrinterWithRenderer(renderer Renderer) HelpPrinter {
	return &helpPrinter{
		renderer: renderer,
	}
}

// Print renders and writes help content to the writer
func (p *helpPrinter) Print(w io.Writer, templ string, data any) error {
	content, err := p.render(templ, data)
	if err != nil {
		return contextureerrors.Wrap(err, "render help")
	}

	n, err := w.Write([]byte(content))
	if err != nil {
		return contextureerrors.Wrap(err, "write help")
	}

	if n != len(content) {
		return io.ErrShortWrite
	}

	return nil
}

// render determines the template type and renders accordingly
func (p *helpPrinter) render(templ string, data any) (string, error) {
	cmd, ok := data.(*cli.Command)
	if !ok {
		return p.renderer.RenderDefault(templ, data)
	}

	// Always use our custom renderer for known commands to avoid template issues
	if cmd.Name == "config" || cmd.Name == "contexture" || cmd.Name == "formats" ||
		cmd.Name == "enable" || cmd.Name == "disable" || cmd.Name == "rules" {
		if len(cmd.Commands) > 0 {
			return p.renderer.RenderApp(cmd)
		}
		return p.renderer.RenderCommand(cmd)
	}

	// Determine template type for other commands
	switch {
	case strings.Contains(templ, appTemplatePattern):
		return p.renderer.RenderApp(cmd)
	case strings.Contains(templ, commandTemplatePattern):
		return p.renderer.RenderCommand(cmd)
	case cmd.Name == "contexture" && len(cmd.Commands) > 0:
		// Main command with subcommands
		return p.renderer.RenderApp(cmd)
	default:
		return p.renderer.RenderDefault(templ, data)
	}
}

// renderer implements the Renderer interface
type renderer struct {
	styles StyleProvider
}

// NewRenderer creates a new renderer with the given style provider
func NewRenderer(styles StyleProvider) Renderer {
	return &renderer{
		styles: styles,
	}
}

// RenderApp renders application-level help
func (r *renderer) RenderApp(cmd *cli.Command) (string, error) {
	var help strings.Builder

	// Pre-allocate capacity for better performance
	help.Grow(1024)

	// Title
	r.writeTitle(&help, cmd)

	// Description
	if cmd.Description != "" {
		r.writeDescription(&help, cmd.Description)
	}

	// Commands
	if len(cmd.Commands) > 0 {
		r.writeCommands(&help, cmd.Commands)
	}

	// Global options
	if len(cmd.Flags) > 0 {
		if err := r.writeGlobalOptions(&help, cmd.Flags); err != nil {
			return "", err
		}
	}

	// Footer
	help.WriteString(r.styles.DescriptionStyle().Render(commandHelpFooter))
	help.WriteString("\n")

	return help.String(), nil
}

// RenderCommand renders command-specific help
func (r *renderer) RenderCommand(cmd *cli.Command) (string, error) {
	var help strings.Builder

	// Pre-allocate capacity
	help.Grow(512)

	// Title
	r.writeTitle(&help, cmd)

	// Description
	if cmd.Description != "" {
		r.writeDescription(&help, cmd.Description)
	}

	// Usage
	r.writeUsage(&help, cmd)

	// Options
	if len(cmd.Flags) > 0 {
		if err := r.writeOptions(&help, cmd.Flags); err != nil {
			return "", err
		}
	}

	return help.String(), nil
}

// RenderDefault renders using the default template engine
func (r *renderer) RenderDefault(templ string, data any) (string, error) {
	t := template.New("help")
	t.Funcs(template.FuncMap{
		"join": strings.Join,
		"wrap": func(text string, width int) string {
			// Simple word wrapping implementation
			if len(text) <= width {
				return text
			}

			words := strings.Fields(text)
			if len(words) == 0 {
				return text
			}

			var lines []string
			currentLine := words[0]

			for _, word := range words[1:] {
				if len(currentLine)+1+len(word) <= width {
					currentLine += " " + word
				} else {
					lines = append(lines, currentLine)
					currentLine = word
				}
			}

			if currentLine != "" {
				lines = append(lines, currentLine)
			}

			return strings.Join(lines, "\n")
		},
	})

	t, err := t.Parse(templ)
	if err != nil {
		return "", contextureerrors.Wrap(err, "parse help template")
	}

	var buf strings.Builder
	if err := t.Execute(&buf, data); err != nil {
		return "", contextureerrors.Wrap(err, "execute help template")
	}

	return buf.String(), nil
}

// Helper methods for rendering components

func (r *renderer) writeTitle(w *strings.Builder, cmd *cli.Command) {
	// Bold purple command name, darker gray description
	w.WriteString(r.styles.TitleStyle().Render(cmd.Name))
	w.WriteString(" ")
	darkerGrayStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#808080"))
	w.WriteString(darkerGrayStyle.Render(cmd.Usage))
	w.WriteString("\n\n")
}

func (r *renderer) writeDescription(w *strings.Builder, description string) {
	// White for long description
	whiteStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))
	w.WriteString(whiteStyle.Render(description))
	w.WriteString("\n\n")
}

func (r *renderer) writeCommands(w *strings.Builder, commands []*cli.Command) {
	w.WriteString(r.styles.HeaderStyle().Render("Commands:"))
	w.WriteString("\n")

	// Calculate alignment
	maxWidth := r.calculateMaxCommandWidth(commands)
	darkerGrayStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#808080"))

	// Render visible commands
	for _, cmd := range commands {
		if cmd.Hidden {
			continue
		}

		names := r.getCommandNames(cmd)
		nameStr := strings.Join(names, ", ")

		w.WriteString("    ")
		w.WriteString(r.styles.CommandStyle().Render(nameStr))
		w.WriteString(strings.Repeat(" ", maxWidth-len(nameStr)+commandPadding))
		w.WriteString(darkerGrayStyle.Render(cmd.Usage))
		w.WriteString("\n")
	}

	w.WriteString("\n")
}

func (r *renderer) writeGlobalOptions(w *strings.Builder, flags []cli.Flag) error {
	w.WriteString(r.styles.HeaderStyle().Render("Global Options:"))
	w.WriteString("\n")
	return r.writeFlags(w, flags)
}

func (r *renderer) writeOptions(w *strings.Builder, flags []cli.Flag) error {
	w.WriteString(r.styles.HeaderStyle().Render("Options:"))
	w.WriteString("\n")
	return r.writeFlags(w, flags)
}

func (r *renderer) writeFlags(w *strings.Builder, flags []cli.Flag) error {
	// Calculate max width for alignment
	maxWidth := r.calculateMaxFlagWidth(flags)
	darkerGrayStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#808080"))

	for _, flag := range flags {
		if flagStr := flag.String(); flagStr != "" {
			// Parse flag string to separate flag name from description
			flagName, flagDesc := r.parseFlagString(flagStr)

			w.WriteString("    ")
			w.WriteString(r.styles.CommandStyle().Render(flagName))
			w.WriteString(strings.Repeat(" ", maxWidth-len(flagName)+commandPadding))
			w.WriteString(darkerGrayStyle.Render(flagDesc))
			w.WriteString("\n")
		}
	}
	w.WriteString("\n")
	return nil
}

func (r *renderer) writeUsage(w *strings.Builder, cmd *cli.Command) {
	w.WriteString(r.styles.HeaderStyle().Render("Usage:"))
	w.WriteString("\n    ")

	if cmd.UsageText != "" {
		w.WriteString(cmd.UsageText)
	} else {
		w.WriteString(r.styles.CommandStyle().Render("contexture " + cmd.Name))
		whiteStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))
		if len(cmd.Flags) > 0 {
			w.WriteString(whiteStyle.Render(" [options]"))
		}
		if cmd.ArgsUsage != "" {
			w.WriteString(" ")
			w.WriteString(whiteStyle.Render(cmd.ArgsUsage))
		}
	}

	w.WriteString("\n\n")
}

func (r *renderer) calculateMaxCommandWidth(commands []*cli.Command) int {
	maxWidth := 0
	for _, cmd := range commands {
		if cmd.Hidden {
			continue
		}
		names := r.getCommandNames(cmd)
		nameStr := strings.Join(names, ", ")
		if len(nameStr) > maxWidth {
			maxWidth = len(nameStr)
		}
	}
	return maxWidth
}

func (r *renderer) getCommandNames(cmd *cli.Command) []string {
	names := []string{cmd.Name}
	names = append(names, cmd.Aliases...)
	return names
}

func (r *renderer) calculateMaxFlagWidth(flags []cli.Flag) int {
	maxWidth := 0
	for _, flag := range flags {
		if flagStr := flag.String(); flagStr != "" {
			flagName, _ := r.parseFlagString(flagStr)
			if len(flagName) > maxWidth {
				maxWidth = len(flagName)
			}
		}
	}
	return maxWidth
}

func (r *renderer) parseFlagString(flagStr string) (string, string) {
	// Flag strings typically have the format: "--flag, -f\tdescription (default: value)"
	// or just "--flag\tdescription"
	parts := strings.Split(flagStr, "\t")
	if len(parts) >= 2 {
		return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
	}
	// If no tab separator, treat the whole string as the flag name
	return strings.TrimSpace(flagStr), ""
}

// styleProvider implements the StyleProvider interface using the UI package
type styleProvider struct{}

// NewStyleProvider creates a new style provider using the UI theme
func NewStyleProvider() StyleProvider {
	return &styleProvider{}
}

func (s *styleProvider) TitleStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#9370DB")) // Bold purple
}

func (s *styleProvider) HeaderStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#9370DB")) // Bold purple
}

func (s *styleProvider) CommandStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FF69B4")) // Bold pink
}

func (s *styleProvider) DescriptionStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("#808080")) // Darker gray
}

// AppHelpTemplate is the default app help template for backward compatibility
var AppHelpTemplate = GetAppHelpTemplate()

// CommandHelpTemplate is the default command help template for backward compatibility
var CommandHelpTemplate = GetCommandHelpTemplate()
