package commands

import (
	"context"

	"github.com/contextureai/contexture/internal/dependencies"
	"github.com/urfave/cli/v3"
)

// RulesAction is the CLI action handler for the main rules command
func RulesAction(ctx context.Context, cmd *cli.Command, deps *dependencies.Dependencies) error {
	// When the rules command is called without subcommands, default to list action
	return ListAction(ctx, cmd, deps)
}
