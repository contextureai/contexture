package rule

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCombineErrors(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		errs     []error
		expected string
		isNil    bool
	}{
		{
			name:     "no errors",
			errs:     []error{},
			expected: "",
			isNil:    true,
		},
		{
			name:     "nil slice",
			errs:     nil,
			expected: "",
			isNil:    true,
		},
		{
			name:     "single error",
			errs:     []error{errors.New("first error")},
			expected: "first error",
			isNil:    false,
		},
		{
			name: "two errors",
			errs: []error{
				errors.New("first error"),
				errors.New("second error"),
			},
			expected: "first error; second error",
			isNil:    false,
		},
		{
			name: "multiple errors",
			errs: []error{
				errors.New("error 1"),
				errors.New("error 2"),
				errors.New("error 3"),
				errors.New("error 4"),
			},
			expected: "error 1; error 2; error 3; error 4",
			isNil:    false,
		},
		{
			name: "errors with special characters",
			errs: []error{
				errors.New("failed to parse: invalid syntax"),
				errors.New("network timeout; connection failed"),
				errors.New("validation error: field 'name' is required"),
			},
			expected: "failed to parse: invalid syntax; network timeout; connection failed; validation error: field 'name' is required",
			isNil:    false,
		},
		{
			name: "errors with empty messages",
			errs: []error{
				errors.New(""),
				errors.New("actual error"),
				errors.New(""),
			},
			expected: "; actual error; ",
			isNil:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := combineErrors(tt.errs)

			if tt.isNil {
				require.NoError(t, result, "should return nil for empty error slice")
			} else {
				require.Error(t, result, "should return non-nil error")
				assert.Equal(t, tt.expected, result.Error(), "error message should match expected")
			}
		})
	}
}

func TestCombineErrors_PreservesOriginalError(t *testing.T) {
	t.Parallel()
	originalErr := errors.New("original error")
	errs := []error{originalErr}

	result := combineErrors(errs)

	assert.Equal(t, originalErr, result, "single error should be returned as-is")
	assert.Same(t, originalErr, result, "should be the exact same error object")
}
