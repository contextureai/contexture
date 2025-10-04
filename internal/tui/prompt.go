// Package tui provides terminal user interface components for Contexture.
package tui

import (
	"errors"

	"github.com/charmbracelet/huh"
	contextureerrors "github.com/contextureai/contexture/internal/errors"
	"github.com/contextureai/contexture/internal/ui"
)

const (
	// Default titles for prompts
	defaultSelectTitle       = "Select options"
	defaultSelectOptionTitle = "Select an option"
)

// ErrUserCancelled indicates the user cancelled the operation (e.g., pressed ESC)
var ErrUserCancelled = errors.New("operation cancelled by user")

// HandleFormError checks if the error is from user cancellation and returns a more user-friendly error
func HandleFormError(err error) error {
	if err == nil {
		return nil
	}

	// Check if this is a cancellation error from huh
	if errors.Is(err, huh.ErrUserAborted) {
		return ErrUserCancelled
	}

	// Return the original error if it's not a cancellation
	return err
}

// SelectOptions represents options for selection prompts
type SelectOptions struct {
	Title       string
	Description string
	Options     []SelectOption
	Default     string
}

// SelectOption represents a single option in a select prompt
type SelectOption struct {
	Label       string
	Value       string
	Description string
}

// Select prompts the user to select from a list of options
func Select(opts SelectOptions) (string, error) {
	if opts.Title == "" {
		opts.Title = "Select an option"
	}

	if len(opts.Options) == 0 {
		return "", contextureerrors.ValidationErrorf("options", "no options provided")
	}

	var selected string

	// Create huh options
	huhOptions := make([]huh.Option[string], len(opts.Options))
	for i, opt := range opts.Options {
		huhOptions[i] = huh.NewOption(opt.Label, opt.Value)
		if opt.Value == opts.Default {
			huhOptions[i] = huhOptions[i].Selected(true)
		}
	}

	selectPrompt := huh.NewSelect[string]().
		Title(opts.Title).
		Options(huhOptions...).
		Value(&selected)

	if opts.Description != "" {
		selectPrompt = selectPrompt.Description(opts.Description)
	}

	form := ui.ConfigureHuhForm(huh.NewForm(
		huh.NewGroup(selectPrompt),
	))

	if err := HandleFormError(form.Run()); err != nil {
		return "", err
	}

	return selected, nil
}

// MultiSelectOptions represents options for multi-selection prompts
type MultiSelectOptions struct {
	Title       string
	Description string
	Options     []SelectOption
	Default     []string
}

// MultiSelect prompts the user to select multiple options from a list
func MultiSelect(opts MultiSelectOptions) ([]string, error) {
	if opts.Title == "" {
		opts.Title = defaultSelectTitle
	}

	if len(opts.Options) == 0 {
		return nil, contextureerrors.ValidationErrorf("options", "no options provided")
	}

	var selected []string

	// Create huh options
	huhOptions := make([]huh.Option[string], len(opts.Options))
	for i, opt := range opts.Options {
		huhOptions[i] = huh.NewOption(opt.Label, opt.Value)
		// Check if this option should be selected by default
		for _, defaultVal := range opts.Default {
			if opt.Value == defaultVal {
				huhOptions[i] = huhOptions[i].Selected(true)
				break
			}
		}
	}

	multiSelectPrompt := huh.NewMultiSelect[string]().
		Title(opts.Title).
		Options(huhOptions...).
		Value(&selected)

	if opts.Description != "" {
		multiSelectPrompt = multiSelectPrompt.Description(opts.Description)
	}

	form := ui.ConfigureHuhForm(huh.NewForm(
		huh.NewGroup(multiSelectPrompt),
	))

	if err := HandleFormError(form.Run()); err != nil {
		return nil, err
	}

	return selected, nil
}
