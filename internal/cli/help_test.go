// Package cli provides CLI utilities and customizations
package cli

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v3"
)

func TestHelpPrinter(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		template string
		data     any
		wantErr  bool
		contains []string
	}{
		{
			name:     "app help template",
			template: AppHelpTemplate,
			data: &cli.Command{
				Name:        "contexture",
				Usage:       "AI assistant rule management",
				Description: "Manage AI assistant rules for your project",
				Commands: []*cli.Command{
					{
						Name:  "init",
						Usage: "Initialize a new project",
					},
					{
						Name:  "rules",
						Usage: "Manage project rules",
					},
				},
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:  "verbose",
						Usage: "Enable verbose output",
					},
				},
			},
			contains: []string{
				"contexture AI assistant rule management",
				"Manage AI assistant rules for your project",
				"Commands:",
				"init",
				"rules",
				"Global Options:",
				"--verbose",
			},
		},
		{
			name:     "command help template",
			template: CommandHelpTemplate,
			data: &cli.Command{
				Name:      "init",
				Usage:     "Initialize a new project",
				ArgsUsage: "[path]",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:  "force",
						Usage: "Force initialization",
					},
				},
			},
			contains: []string{
				"init Initialize a new project",
				"Usage:",
				"contexture init [options] [path]",
				"Options:",
				"--force",
			},
		},
		{
			name:     "default template fallback",
			template: "{{.Name}}",
			data: &cli.Command{
				Name: "test",
			},
			contains: []string{
				"test",
			},
		},
		{
			name:     "non-command data",
			template: "{{.}}",
			data:     "simple string",
			contains: []string{
				"simple string",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			printer := NewHelpPrinter()
			err := printer.Print(&buf, tt.template, tt.data)
			require.NoError(t, err)
			output := buf.String()

			for _, want := range tt.contains {
				assert.Contains(t, output, want, "Expected output to contain %q", want)
			}
		})
	}
}

func TestRenderDefaultTemplate(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		template string
		data     any
		want     string
		wantErr  bool
	}{
		{
			name:     "simple template",
			template: "Hello {{.Name}}",
			data:     struct{ Name string }{"World"},
			want:     "Hello World",
		},
		{
			name:     "template with join function",
			template: "{{join .Items \", \"}}",
			data:     struct{ Items []string }{[]string{"a", "b", "c"}},
			want:     "a, b, c",
		},
		{
			name:     "invalid template syntax",
			template: "{{.Name",
			data:     struct{ Name string }{"Test"},
			wantErr:  true,
		},
		{
			name:     "missing field",
			template: "{{.Missing}}",
			data:     struct{ Name string }{"Test"},
			wantErr:  true, // Go's text/template returns an error for missing fields
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			renderer := NewRenderer(NewStyleProvider())
			result, err := renderer.RenderDefault(tt.template, tt.data)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestRenderAppHelp(t *testing.T) {
	t.Parallel()
	// Note: We can't easily test the private renderAppHelp function directly
	// but we test it through HelpPrinter
	cmd := &cli.Command{
		Name:        "app",
		Usage:       "test application",
		Description: "A test application",
		Commands: []*cli.Command{
			{
				Name:    "cmd1",
				Aliases: []string{"c1"},
				Usage:   "First command",
			},
			{
				Name:   "cmd2",
				Usage:  "Second command",
				Hidden: true, // Should not appear
			},
			{
				Name:  "very-long-command-name",
				Usage: "Command with a long name",
			},
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "config",
				Usage: "Config file",
			},
		},
	}

	var buf bytes.Buffer
	printer := NewHelpPrinter()
	err := printer.Print(&buf, AppHelpTemplate, cmd)
	require.NoError(t, err)
	output := buf.String()

	// Check proper rendering
	assert.Contains(t, output, "app test application")
	assert.Contains(t, output, "A test application")
	assert.Contains(t, output, "Commands:")
	assert.Contains(t, output, "cmd1, c1")
	assert.Contains(t, output, "First command")
	assert.NotContains(t, output, "cmd2") // Hidden command
	assert.Contains(t, output, "very-long-command-name")
	assert.Contains(t, output, "Global Options:")
	assert.Contains(t, output, "--config")
	assert.Contains(t, output, "Use \"contexture [command] --help\"")

	// Check alignment is working (spaces between command and description)
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, "cmd1") && strings.Contains(line, "First command") {
			// Should have proper spacing
			assert.Contains(t, line, "   ", "Command line should have proper spacing")
		}
	}
}

func TestRenderCommandHelp(t *testing.T) {
	t.Parallel()
	cmd := &cli.Command{
		Name:        "test",
		Usage:       "test command",
		Description: "A detailed description",
		UsageText:   "custom usage",
		ArgsUsage:   "<arg1> <arg2>",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "output",
				Aliases: []string{"o"},
				Usage:   "Output file",
			},
		},
	}

	var buf bytes.Buffer
	printer := NewHelpPrinter()
	err := printer.Print(&buf, CommandHelpTemplate, cmd)
	require.NoError(t, err)
	output := buf.String()

	assert.Contains(t, output, "test test command")
	assert.Contains(t, output, "A detailed description")
	assert.Contains(t, output, "Usage:")
	assert.Contains(t, output, "custom usage")
	assert.Contains(t, output, "Options:")
	assert.Contains(t, output, "--output")
}

func TestHelpPrinterEdgeCases(t *testing.T) {
	t.Parallel()
	t.Run("empty command", func(t *testing.T) {
		var buf bytes.Buffer
		cmd := &cli.Command{}
		printer := NewHelpPrinter()
		err := printer.Print(&buf, AppHelpTemplate, cmd)
		require.NoError(t, err)
		output := buf.String()
		assert.NotEmpty(t, output)
	})

	t.Run("command with no visible flags", func(t *testing.T) {
		var buf bytes.Buffer
		cmd := &cli.Command{
			Name:  "test",
			Usage: "test command",
			Flags: []cli.Flag{}, // Empty flags
		}
		printer := NewHelpPrinter()
		err := printer.Print(&buf, CommandHelpTemplate, cmd)
		require.NoError(t, err)
		output := buf.String()
		assert.NotContains(t, output, "OPTIONS:")
	})

	t.Run("command with no commands", func(t *testing.T) {
		var buf bytes.Buffer
		cmd := &cli.Command{
			Name:     "test",
			Usage:    "test command",
			Commands: []*cli.Command{}, // Empty commands
		}
		printer := NewHelpPrinter()
		err := printer.Print(&buf, AppHelpTemplate, cmd)
		require.NoError(t, err)
		output := buf.String()
		assert.NotContains(t, output, "COMMANDS:")
	})

	t.Run("command with custom UsageText", func(t *testing.T) {
		var buf bytes.Buffer
		cmd := &cli.Command{
			Name:      "test",
			Usage:     "test command",
			UsageText: "test [custom] [usage]",
		}
		printer := NewHelpPrinter()
		err := printer.Print(&buf, CommandHelpTemplate, cmd)
		require.NoError(t, err)
		output := buf.String()
		assert.Contains(t, output, "test [custom] [usage]")
		assert.NotContains(t, output, "[options]") // Should not add default options text
	})
}

func TestHelpPrinterWithWriter(t *testing.T) {
	t.Parallel()
	// Test with different writers
	t.Run("bytes.Buffer", func(t *testing.T) {
		var buf bytes.Buffer
		cmd := &cli.Command{Name: "test"}
		printer := NewHelpPrinter()
		err := printer.Print(&buf, "{{.Name}}", cmd)
		require.NoError(t, err)
		assert.Equal(t, "test", buf.String())
	})

	t.Run("error writer", func(t *testing.T) {
		// Test with a writer that returns errors
		w := &errorWriter{err: errors.New("write error")}
		cmd := &cli.Command{Name: "test"}
		printer := NewHelpPrinter()
		err := printer.Print(w, "{{.Name}}", cmd)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "write error")
		// The printer should attempt to write at least once
		assert.GreaterOrEqual(t, w.writeCount, 1)
	})
}

// errorWriter is a test writer that returns errors
type errorWriter struct {
	writeCount int
	err        error
}

func (w *errorWriter) Write(p []byte) (int, error) {
	w.writeCount++
	if w.err != nil {
		return 0, w.err
	}
	return len(p), nil
}

func TestTemplateDetection(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name         string
		template     string
		expectApp    bool
		expectCmd    bool
		expectCustom bool
	}{
		{
			name:      "app template",
			template:  "{{.Name}} - {{.Usage}}\n{{if .VisibleCommands}}",
			expectApp: true,
		},
		{
			name:      "command template",
			template:  "{{.HelpName}} - {{.Usage}}\nUSAGE:",
			expectCmd: true,
		},
		{
			name:         "custom template",
			template:     "Custom: {{.Name}}",
			expectCustom: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			cmd := &cli.Command{
				Name:  "test",
				Usage: "test usage",
			}

			printer := NewHelpPrinter()
			err := printer.Print(&buf, tt.template, cmd)
			require.NoError(t, err)
			output := buf.String()

			// Basic check that something was rendered
			assert.NotEmpty(t, output)
			assert.Contains(t, output, "test")
		})
	}
}

// Benchmarks
func BenchmarkHelpPrinter(b *testing.B) {
	cmd := &cli.Command{
		Name:        "bench",
		Usage:       "benchmark app",
		Description: "A benchmark application",
		Commands: []*cli.Command{
			{Name: "cmd1", Usage: "Command 1"},
			{Name: "cmd2", Usage: "Command 2"},
			{Name: "cmd3", Usage: "Command 3"},
		},
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "flag1", Usage: "Flag 1"},
			&cli.StringFlag{Name: "flag2", Usage: "Flag 2"},
		},
	}

	b.ResetTimer()
	for range b.N {
		var buf bytes.Buffer
		printer := NewHelpPrinter()
		err := printer.Print(&buf, AppHelpTemplate, cmd)
		require.NoError(b, err)
	}
}

func BenchmarkRenderAppHelp(b *testing.B) {
	cmd := &cli.Command{
		Name:        "bench",
		Usage:       "benchmark app",
		Description: "A benchmark application with many commands",
		Commands:    make([]*cli.Command, 20),
		Flags:       make([]cli.Flag, 10),
	}

	// Initialize commands and flags
	for i := range 20 {
		cmd.Commands[i] = &cli.Command{
			Name:  strings.Repeat("cmd", i+1),
			Usage: strings.Repeat("Description ", i+1),
		}
	}

	for i := range 10 {
		cmd.Flags[i] = &cli.StringFlag{
			Name:  strings.Repeat("flag", i+1),
			Usage: strings.Repeat("Flag description ", i+1),
		}
	}

	var buf bytes.Buffer
	b.ResetTimer()
	for range b.N {
		buf.Reset()
		printer := NewHelpPrinter()
		err := printer.Print(&buf, AppHelpTemplate, cmd)
		require.NoError(b, err)
	}
}
