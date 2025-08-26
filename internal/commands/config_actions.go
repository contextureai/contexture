// Package commands provides CLI action handlers for config commands
package commands

import (
	"context"

	"github.com/contextureai/contexture/internal/dependencies"
	"github.com/urfave/cli/v3"
)

// ConfigFormatsListAction handles the config formats list command
func ConfigFormatsListAction(
	ctx context.Context,
	cmd *cli.Command,
	deps *dependencies.Dependencies,
) error {
	fm := NewFormatManager(deps)
	return fm.ListFormats(ctx, cmd)
}

// ConfigFormatsAddAction handles the config formats add command
func ConfigFormatsAddAction(
	ctx context.Context,
	cmd *cli.Command,
	deps *dependencies.Dependencies,
) error {
	fm := NewFormatManager(deps)
	args := cmd.Args().Slice()

	if len(args) == 0 {
		// Interactive mode
		return fm.interactiveAddFormat(ctx, cmd)
	}

	// Add specific formats
	for _, formatType := range args {
		if err := fm.AddFormat(ctx, cmd, formatType); err != nil {
			return err
		}
	}
	return nil
}

// ConfigFormatsRemoveAction handles the config formats remove command
func ConfigFormatsRemoveAction(
	ctx context.Context,
	cmd *cli.Command,
	deps *dependencies.Dependencies,
) error {
	fm := NewFormatManager(deps)
	args := cmd.Args().Slice()

	if len(args) == 0 {
		// Interactive mode
		return fm.interactiveRemoveFormat(ctx, cmd)
	}

	// Remove specific formats
	for _, formatType := range args {
		if err := fm.RemoveFormat(ctx, cmd, formatType); err != nil {
			return err
		}
	}
	return nil
}

// ConfigFormatsEnableAction handles the config formats enable command
func ConfigFormatsEnableAction(
	ctx context.Context,
	cmd *cli.Command,
	deps *dependencies.Dependencies,
) error {
	fm := NewFormatManager(deps)
	args := cmd.Args().Slice()

	if len(args) == 0 {
		// Interactive mode
		return fm.interactiveEnableFormat(ctx, cmd)
	}

	// Enable specific format
	return fm.EnableFormat(ctx, cmd, args[0])
}

// ConfigFormatsDisableAction handles the config formats disable command
func ConfigFormatsDisableAction(
	ctx context.Context,
	cmd *cli.Command,
	deps *dependencies.Dependencies,
) error {
	fm := NewFormatManager(deps)
	args := cmd.Args().Slice()

	if len(args) == 0 {
		// Interactive mode
		return fm.interactiveDisableFormat(ctx, cmd)
	}

	// Disable specific format
	return fm.DisableFormat(ctx, cmd, args[0])
}

// ConfigFormatsAction handles the base config formats command (defaults to list)
func ConfigFormatsAction(
	ctx context.Context,
	cmd *cli.Command,
	deps *dependencies.Dependencies,
) error {
	// Default to list when no subcommand is provided
	fm := NewFormatManager(deps)
	return fm.ListFormats(ctx, cmd)
}
