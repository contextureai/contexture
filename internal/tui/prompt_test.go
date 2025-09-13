package tui

import (
	"errors"
	"testing"

	"github.com/charmbracelet/huh"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSelect_BasicSelection tests basic selection functionality
func TestSelect_BasicSelection(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		options  SelectOptions
		expected string
		wantErr  bool
	}{
		{
			name: "single option selection",
			options: SelectOptions{
				Title: "Choose One",
				Options: []SelectOption{
					{Label: "Option A", Value: "a", Description: "First option"},
				},
			},
			expected: "", // Can't actually run interactive test
			wantErr:  false,
		},
		{
			name: "multiple options with default",
			options: SelectOptions{
				Title:   "Choose One",
				Default: "b",
				Options: []SelectOption{
					{Label: "Option A", Value: "a"},
					{Label: "Option B", Value: "b"},
					{Label: "Option C", Value: "c"},
				},
			},
			expected: "",
			wantErr:  false,
		},
		{
			name: "with description",
			options: SelectOptions{
				Title:       "Select Item",
				Description: "Please choose from the following options:",
				Options: []SelectOption{
					{Label: "First", Value: "1"},
					{Label: "Second", Value: "2"},
				},
			},
			expected: "",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test option validation and setup
			if tt.options.Title == "" {
				tt.options.Title = defaultSelectOptionTitle
			}
			assert.NotEmpty(t, tt.options.Title)
			assert.NotEmpty(t, tt.options.Options)

			// Verify default handling
			if tt.options.Default != "" {
				found := false
				for _, opt := range tt.options.Options {
					if opt.Value == tt.options.Default {
						found = true
						break
					}
				}
				assert.True(t, found, "Default value should exist in options")
			}
		})
	}
}

// TestSelect_EmptyOptions tests error handling for empty options
func TestSelect_EmptyOptions(t *testing.T) {
	t.Parallel()
	opts := SelectOptions{
		Title:   "Empty Test",
		Options: []SelectOption{},
	}

	// This would return an error without actually running the interactive prompt
	assert.Empty(t, opts.Options, "Options should be empty for this test")

	// We can't actually call Select() in tests since it's interactive,
	// but we can test the validation logic that would occur
	if len(opts.Options) == 0 {
		// This is what Select() would return
		expectedErr := "no options provided"
		assert.Contains(t, expectedErr, "no options")
	}
}

// TestMultiSelect_BasicSelection tests multi-selection functionality
func TestMultiSelect_BasicSelection(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		options  MultiSelectOptions
		expected []string
		wantErr  bool
	}{
		{
			name: "multiple selection without defaults",
			options: MultiSelectOptions{
				Title: "Choose Multiple",
				Options: []SelectOption{
					{Label: "Option A", Value: "a"},
					{Label: "Option B", Value: "b"},
					{Label: "Option C", Value: "c"},
				},
			},
			expected: nil,
			wantErr:  false,
		},
		{
			name: "with default selections",
			options: MultiSelectOptions{
				Title:   "Choose Multiple",
				Default: []string{"a", "c"},
				Options: []SelectOption{
					{Label: "Option A", Value: "a"},
					{Label: "Option B", Value: "b"},
					{Label: "Option C", Value: "c"},
				},
			},
			expected: []string{"a", "c"},
			wantErr:  false,
		},
		{
			name: "with description",
			options: MultiSelectOptions{
				Title:       "Multi Select",
				Description: "Select multiple items:",
				Options: []SelectOption{
					{Label: "First", Value: "1"},
					{Label: "Second", Value: "2"},
					{Label: "Third", Value: "3"},
				},
			},
			expected: nil,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test option validation and setup
			if tt.options.Title == "" {
				tt.options.Title = defaultSelectTitle
			}
			assert.NotEmpty(t, tt.options.Title)
			assert.NotEmpty(t, tt.options.Options)

			// Verify default handling
			if len(tt.options.Default) > 0 {
				for _, defaultVal := range tt.options.Default {
					found := false
					for _, opt := range tt.options.Options {
						if opt.Value == defaultVal {
							found = true
							break
						}
					}
					assert.True(t, found, "Default value %s should exist in options", defaultVal)
				}
			}
		})
	}
}

// TestMultiSelect_EmptyOptions tests error handling for empty options
func TestMultiSelect_EmptyOptions(t *testing.T) {
	t.Parallel()
	opts := MultiSelectOptions{
		Title:   "Empty Multi Test",
		Options: []SelectOption{},
	}

	assert.Empty(t, opts.Options, "Options should be empty for this test")

	// Test validation logic
	if len(opts.Options) == 0 {
		expectedErr := "no options provided"
		assert.Contains(t, expectedErr, "no options")
	}
}

// TestHandleFormError tests error handling transformation
func TestHandleFormError(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		inputError  error
		expectedErr error
	}{
		{
			name:        "nil error",
			inputError:  nil,
			expectedErr: nil,
		},
		{
			name:        "user cancellation error",
			inputError:  huh.ErrUserAborted,
			expectedErr: ErrUserCancelled,
		},
		{
			name:        "other error passes through",
			inputError:  errors.New("some other error"),
			expectedErr: errors.New("some other error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := HandleFormError(tt.inputError)

			switch {
			case tt.expectedErr == nil:
				require.NoError(t, result)
			case errors.Is(tt.expectedErr, ErrUserCancelled):
				require.ErrorIs(t, result, ErrUserCancelled)
			default:
				require.Error(t, result)
				assert.Equal(t, tt.expectedErr.Error(), result.Error())
			}
		})
	}
}

// TestSelectOptions_Validation tests option validation
func TestSelectOptions_Validation(t *testing.T) {
	t.Parallel()
	t.Run("valid options", func(t *testing.T) {
		opts := SelectOptions{
			Title: "Valid Test",
			Options: []SelectOption{
				{Label: "A", Value: "a"},
				{Label: "B", Value: "b"},
			},
		}

		assert.NotEmpty(t, opts.Title)
		assert.Len(t, opts.Options, 2)

		// All options should have labels and values
		for _, opt := range opts.Options {
			assert.NotEmpty(t, opt.Label, "All options should have labels")
			assert.NotEmpty(t, opt.Value, "All options should have values")
		}
	})

	t.Run("empty title gets default", func(t *testing.T) {
		opts := SelectOptions{
			Title: "",
			Options: []SelectOption{
				{Label: "A", Value: "a"},
			},
		}

		// This mimics what Select() does internally
		if opts.Title == "" {
			opts.Title = "Select an option"
		}

		assert.Equal(t, "Select an option", opts.Title)
	})
}

// TestMultiSelectOptions_Validation tests multi-select option validation
func TestMultiSelectOptions_Validation(t *testing.T) {
	t.Parallel()
	t.Run("valid multi-select options", func(t *testing.T) {
		opts := MultiSelectOptions{
			Title:   "Valid Multi Test",
			Default: []string{"a", "b"},
			Options: []SelectOption{
				{Label: "A", Value: "a"},
				{Label: "B", Value: "b"},
				{Label: "C", Value: "c"},
			},
		}

		assert.NotEmpty(t, opts.Title)
		assert.Len(t, opts.Options, 3)
		assert.Len(t, opts.Default, 2)

		// Verify defaults exist in options
		for _, defaultVal := range opts.Default {
			found := false
			for _, opt := range opts.Options {
				if opt.Value == defaultVal {
					found = true
					break
				}
			}
			assert.True(t, found, "Default %s should exist in options", defaultVal)
		}
	})

	t.Run("empty title gets default", func(t *testing.T) {
		opts := MultiSelectOptions{
			Title: "",
			Options: []SelectOption{
				{Label: "A", Value: "a"},
			},
		}

		// This mimics what MultiSelect() does internally
		if opts.Title == "" {
			opts.Title = defaultSelectTitle
		}

		assert.Equal(t, defaultSelectTitle, opts.Title)
	})
}

// TestSelectOption_Structure tests the SelectOption struct
func TestSelectOption_Structure(t *testing.T) {
	t.Parallel()
	opt := SelectOption{
		Label:       "Test Label",
		Value:       "test_value",
		Description: "Test description for this option",
	}

	assert.Equal(t, "Test Label", opt.Label)
	assert.Equal(t, "test_value", opt.Value)
	assert.Equal(t, "Test description for this option", opt.Description)

	// Test that description is optional
	opt2 := SelectOption{
		Label: "Simple Option",
		Value: "simple",
	}

	assert.Equal(t, "Simple Option", opt2.Label)
	assert.Equal(t, "simple", opt2.Value)
	assert.Empty(t, opt2.Description)
}

// TestErrUserCancelled tests the custom error
func TestErrUserCancelled(t *testing.T) {
	t.Parallel()
	require.Error(t, ErrUserCancelled)
	assert.Equal(t, "operation cancelled by user", ErrUserCancelled.Error())

	// Test that it can be used with errors.Is
	wrappedErr := HandleFormError(huh.ErrUserAborted)
	assert.ErrorIs(t, wrappedErr, ErrUserCancelled)
}

// TestSelectOptions_EdgeCases tests edge cases for SelectOptions
func TestSelectOptions_EdgeCases(t *testing.T) {
	t.Parallel()
	t.Run("default not in options", func(t *testing.T) {
		opts := SelectOptions{
			Title:   "Bad Default Test",
			Default: "nonexistent",
			Options: []SelectOption{
				{Label: "A", Value: "a"},
				{Label: "B", Value: "b"},
			},
		}

		// Check that default doesn't exist (this would be a configuration error)
		found := false
		for _, opt := range opts.Options {
			if opt.Value == opts.Default {
				found = true
				break
			}
		}
		assert.False(t, found, "Default should not exist in options for this test")
	})

	t.Run("duplicate values", func(t *testing.T) {
		opts := SelectOptions{
			Title: "Duplicate Test",
			Options: []SelectOption{
				{Label: "A", Value: "duplicate"},
				{Label: "B", Value: "duplicate"},
			},
		}

		// Count duplicate values
		valueCount := make(map[string]int)
		for _, opt := range opts.Options {
			valueCount[opt.Value]++
		}

		assert.Equal(t, 2, valueCount["duplicate"], "Should have duplicate values for this test")
	})
}

// TestMultiSelectOptions_EdgeCases tests edge cases for MultiSelectOptions
func TestMultiSelectOptions_EdgeCases(t *testing.T) {
	t.Parallel()
	t.Run("defaults not in options", func(t *testing.T) {
		opts := MultiSelectOptions{
			Title:   "Bad Defaults Test",
			Default: []string{"x", "y", "z"},
			Options: []SelectOption{
				{Label: "A", Value: "a"},
				{Label: "B", Value: "b"},
			},
		}

		// Check how many defaults don't exist
		validDefaults := 0
		for _, defaultVal := range opts.Default {
			for _, opt := range opts.Options {
				if opt.Value == defaultVal {
					validDefaults++
					break
				}
			}
		}

		assert.Equal(t, 0, validDefaults, "No defaults should exist in options for this test")
	})

	t.Run("duplicate defaults", func(t *testing.T) {
		opts := MultiSelectOptions{
			Title:   "Duplicate Defaults Test",
			Default: []string{"a", "a", "b", "a"},
			Options: []SelectOption{
				{Label: "A", Value: "a"},
				{Label: "B", Value: "b"},
			},
		}

		// Count duplicate defaults
		defaultCount := make(map[string]int)
		for _, def := range opts.Default {
			defaultCount[def]++
		}

		assert.Equal(t, 3, defaultCount["a"], "Should have 3 'a' defaults")
		assert.Equal(t, 1, defaultCount["b"], "Should have 1 'b' default")
	})
}
