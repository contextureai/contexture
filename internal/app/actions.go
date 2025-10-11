// Package app provides testable command action wrappers
package app

import (
	"context"

	"github.com/contextureai/contexture/internal/commands"
	"github.com/contextureai/contexture/internal/dependencies"
	"github.com/urfave/cli/v3"
)

// CommandActions provides testable command action functions
type CommandActions struct {
	deps *dependencies.Dependencies
}

// NewCommandActions creates a new command actions instance
func NewCommandActions(deps *dependencies.Dependencies) *CommandActions {
	return &CommandActions{
		deps: deps,
	}
}

// InitAction provides a testable wrapper for the init command
func (a *CommandActions) InitAction(ctx context.Context, cmd *cli.Command) error {
	return commands.InitAction(ctx, cmd, a.deps)
}

// AddAction provides a testable wrapper for the add command
func (a *CommandActions) AddAction(ctx context.Context, cmd *cli.Command) error {
	return commands.AddAction(ctx, cmd, a.deps)
}

// RemoveAction provides a testable wrapper for the remove command
func (a *CommandActions) RemoveAction(ctx context.Context, cmd *cli.Command) error {
	return commands.RemoveAction(ctx, cmd, a.deps)
}

// BuildAction provides a testable wrapper for the build command
func (a *CommandActions) BuildAction(ctx context.Context, cmd *cli.Command) error {
	return commands.BuildAction(ctx, cmd, a.deps)
}

// ListAction provides a testable wrapper for the list command
func (a *CommandActions) ListAction(ctx context.Context, cmd *cli.Command) error {
	return commands.ListAction(ctx, cmd, a.deps)
}

// UpdateAction provides a testable wrapper for the update command
func (a *CommandActions) UpdateAction(ctx context.Context, cmd *cli.Command) error {
	return commands.UpdateAction(ctx, cmd, a.deps)
}

// ConfigAction provides a testable wrapper for the config command
func (a *CommandActions) ConfigAction(ctx context.Context, cmd *cli.Command) error {
	return commands.ConfigAction(ctx, cmd, a.deps)
}

// RulesAction provides a testable wrapper for the rules command
func (a *CommandActions) RulesAction(ctx context.Context, cmd *cli.Command) error {
	return commands.RulesAction(ctx, cmd, a.deps)
}

// ConfigFormatsAction provides a testable wrapper for the config formats command
func (a *CommandActions) ConfigFormatsAction(ctx context.Context, cmd *cli.Command) error {
	return commands.ConfigFormatsAction(ctx, cmd, a.deps)
}

// ConfigFormatsListAction provides a testable wrapper for the config formats list command
func (a *CommandActions) ConfigFormatsListAction(
	ctx context.Context,
	cmd *cli.Command,
	deps *dependencies.Dependencies,
) error {
	return commands.ConfigFormatsListAction(ctx, cmd, deps)
}

// ConfigFormatsAddAction provides a testable wrapper for the config formats add command
func (a *CommandActions) ConfigFormatsAddAction(
	ctx context.Context,
	cmd *cli.Command,
	deps *dependencies.Dependencies,
) error {
	return commands.ConfigFormatsAddAction(ctx, cmd, deps)
}

// ConfigFormatsRemoveAction provides a testable wrapper for the config formats remove command
func (a *CommandActions) ConfigFormatsRemoveAction(
	ctx context.Context,
	cmd *cli.Command,
	deps *dependencies.Dependencies,
) error {
	return commands.ConfigFormatsRemoveAction(ctx, cmd, deps)
}

// ConfigFormatsEnableAction provides a testable wrapper for the config formats enable command
func (a *CommandActions) ConfigFormatsEnableAction(
	ctx context.Context,
	cmd *cli.Command,
	deps *dependencies.Dependencies,
) error {
	return commands.ConfigFormatsEnableAction(ctx, cmd, deps)
}

// ConfigFormatsDisableAction provides a testable wrapper for the config formats disable command
func (a *CommandActions) ConfigFormatsDisableAction(
	ctx context.Context,
	cmd *cli.Command,
	deps *dependencies.Dependencies,
) error {
	return commands.ConfigFormatsDisableAction(ctx, cmd, deps)
}

// ProvidersAction provides a testable wrapper for the providers command
func (a *CommandActions) ProvidersAction(ctx context.Context, cmd *cli.Command) error {
	return commands.ProvidersAction(ctx, cmd, a.deps)
}

// ProvidersListAction provides a testable wrapper for the providers list command
func (a *CommandActions) ProvidersListAction(
	ctx context.Context,
	cmd *cli.Command,
	deps *dependencies.Dependencies,
) error {
	return commands.ProvidersListAction(ctx, cmd, deps)
}

// ProvidersAddAction provides a testable wrapper for the providers add command
func (a *CommandActions) ProvidersAddAction(
	ctx context.Context,
	cmd *cli.Command,
	deps *dependencies.Dependencies,
) error {
	return commands.ProvidersAddAction(ctx, cmd, deps)
}

// ProvidersRemoveAction provides a testable wrapper for the providers remove command
func (a *CommandActions) ProvidersRemoveAction(
	ctx context.Context,
	cmd *cli.Command,
	deps *dependencies.Dependencies,
) error {
	return commands.ProvidersRemoveAction(ctx, cmd, deps)
}

// ProvidersShowAction provides a testable wrapper for the providers show command
func (a *CommandActions) ProvidersShowAction(
	ctx context.Context,
	cmd *cli.Command,
	deps *dependencies.Dependencies,
) error {
	return commands.ProvidersShowAction(ctx, cmd, deps)
}

// QueryAction provides a testable wrapper for the query command
func (a *CommandActions) QueryAction(ctx context.Context, cmd *cli.Command) error {
	return commands.QueryAction(ctx, cmd, a.deps)
}
