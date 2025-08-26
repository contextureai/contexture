package commands

import (
	"context"

	"github.com/contextureai/contexture/internal/dependencies"
	"github.com/urfave/cli/v3"
)

// RulesAction is the CLI action handler for the main rules command
func RulesAction(_ context.Context, _ *cli.Command, _ *dependencies.Dependencies) error {
	// When the rules command is called without subcommands, just show help
	// The CLI framework will automatically show the available subcommands
	return nil
}
