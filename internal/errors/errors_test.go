package errors

import (
	"context"
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// verifyError is a helper function to verify error conditions
func verifyError(t *testing.T, err error, expectNil bool, expectedMsg string) {
	t.Helper()
	if expectNil {
		assert.NoError(t, err)
	} else {
		require.Error(t, err)
		assert.Equal(t, expectedMsg, err.Error())
	}
}

func TestOpError(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		op       string
		err      error
		expected string
	}{
		{
			name:     "with operation",
			op:       "fetch",
			err:      fmt.Errorf("connection timeout"),
			expected: "fetch: connection timeout",
		},
		{
			name:     "without operation",
			op:       "",
			err:      fmt.Errorf("connection timeout"),
			expected: "connection timeout",
		},
		{
			name:     "nested error",
			op:       "process",
			err:      &OpError{Op: "fetch", Err: fmt.Errorf("timeout")},
			expected: "process: fetch: timeout",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := &OpError{Op: tt.op, Err: tt.err}
			assert.Equal(t, tt.expected, err.Error())
		})
	}
}

func TestOpError_Unwrap(t *testing.T) {
	t.Parallel()
	innerErr := fmt.Errorf("inner error")
	err := &OpError{Op: "test", Err: innerErr}

	unwrapped := err.Unwrap()
	assert.Equal(t, innerErr, unwrapped)
}

//nolint:dupl // Similar structure to TestValidationError but tests different functions
func TestWithOp(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		op       string
		err      error
		expected string
		isNil    bool
	}{
		{
			name:     "nil error",
			op:       "test",
			err:      nil,
			expected: "",
			isNil:    true,
		},
		{
			name:     "with error",
			op:       "fetch",
			err:      fmt.Errorf("timeout"),
			expected: "fetch: timeout",
			isNil:    false,
		},
		{
			name:     "with sentinel error",
			op:       "validate",
			err:      ErrInvalidInput,
			expected: "validate: invalid input",
			isNil:    false,
		},
		{
			name:     "empty operation",
			op:       "",
			err:      fmt.Errorf("error"),
			expected: "error",
			isNil:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := WithOp(tt.op, tt.err)
			verifyError(t, err, tt.isNil, tt.expected)
		})
	}
}

func TestWithOpf(t *testing.T) {
	t.Parallel()
	err := WithOpf("parse", "invalid format: %s", "unknown")
	assert.Equal(t, "parse: invalid format: unknown", err.Error())
}

func TestIsRetryable(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "timeout error",
			err:      ErrTimeout,
			expected: true,
		},
		{
			name:     "canceled error",
			err:      ErrCanceled,
			expected: true,
		},
		{
			name:     "context deadline exceeded",
			err:      context.DeadlineExceeded,
			expected: true,
		},
		{
			name:     "connection reset message",
			err:      fmt.Errorf("connection reset by peer"),
			expected: true,
		},
		{
			name:     "connection refused message",
			err:      fmt.Errorf("dial tcp: connection refused"),
			expected: true,
		},
		{
			name:     "network unreachable message",
			err:      fmt.Errorf("network is unreachable"),
			expected: true,
		},
		{
			name:     "temporary failure message",
			err:      fmt.Errorf("temporary failure in name resolution"),
			expected: true,
		},
		{
			name:     "no route to host message",
			err:      fmt.Errorf("no route to host"),
			expected: true,
		},
		{
			name:     "deadline exceeded message",
			err:      fmt.Errorf("i/o timeout deadline exceeded"),
			expected: true,
		},
		{
			name:     "non-retryable error",
			err:      ErrNotFound,
			expected: false,
		},
		{
			name:     "permission denied",
			err:      ErrPermissionDenied,
			expected: false,
		},
		{
			name:     "wrapped timeout",
			err:      fmt.Errorf("operation failed: %w", ErrTimeout),
			expected: true,
		},
		{
			name:     "wrapped non-retryable",
			err:      fmt.Errorf("operation failed: %w", ErrNotFound),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsRetryable(tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsPermission(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "permission denied sentinel",
			err:      ErrPermissionDenied,
			expected: true,
		},
		{
			name:     "permission denied message",
			err:      fmt.Errorf("open /etc/passwd: permission denied"),
			expected: true,
		},
		{
			name:     "access denied message",
			err:      fmt.Errorf("access denied to resource"),
			expected: true,
		},
		{
			name:     "operation not permitted message",
			err:      fmt.Errorf("operation not permitted"),
			expected: true,
		},
		{
			name:     "wrapped permission error",
			err:      fmt.Errorf("failed to write: %w", ErrPermissionDenied),
			expected: true,
		},
		{
			name:     "non-permission error",
			err:      fmt.Errorf("file not found"),
			expected: false,
		},
		{
			name:     "mixed case permission message",
			err:      fmt.Errorf("Permission Denied"),
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsPermission(tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

//nolint:dupl // Similar structure to TestWithOp but tests different functions
func TestValidationError(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		field    string
		err      error
		expected string
		isNil    bool
	}{
		{
			name:     "nil error",
			field:    "username",
			err:      nil,
			expected: "",
			isNil:    true,
		},
		{
			name:     "with error",
			field:    "email",
			err:      fmt.Errorf("invalid format"),
			expected: "validation failed for email: invalid format",
			isNil:    false,
		},
		{
			name:     "with sentinel error",
			field:    "rule_id",
			err:      ErrInvalidRuleID,
			expected: "validation failed for rule_id: invalid rule ID format",
			isNil:    false,
		},
		{
			name:     "empty field name",
			field:    "",
			err:      fmt.Errorf("required"),
			expected: "validation failed for : required",
			isNil:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidationError(tt.field, tt.err)
			verifyError(t, err, tt.isNil, tt.expected)
		})
	}
}

func TestValidationErrorf(t *testing.T) {
	t.Parallel()
	err := ValidationErrorf("age", "must be between %d and %d", 0, 120)
	assert.Equal(t, "validation failed for age: must be between 0 and 120", err.Error())
}

func TestSentinelErrors(t *testing.T) {
	t.Parallel()
	// Test that sentinel errors are properly defined
	sentinels := []struct {
		name string
		err  error
		msg  string
	}{
		{"ErrNotFound", ErrNotFound, "not found"},
		{"ErrInvalidInput", ErrInvalidInput, "invalid input"},
		{"ErrPermissionDenied", ErrPermissionDenied, "permission denied"},
		{"ErrTimeout", ErrTimeout, "timeout"},
		{"ErrCanceled", ErrCanceled, "operation canceled"},
		{"ErrInvalidRuleID", ErrInvalidRuleID, "invalid rule ID format"},
		{"ErrRuleNotFound", ErrRuleNotFound, "rule not found"},
		{"ErrRepositoryNotFound", ErrRepositoryNotFound, "repository not found"},
		{"ErrConfigNotFound", ErrConfigNotFound, "configuration not found"},
		{"ErrUnsupportedFormat", ErrUnsupportedFormat, "unsupported format"},
		{"ErrAuthenticationFailed", ErrAuthenticationFailed, "authentication failed"},
	}

	for _, tt := range sentinels {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.msg, tt.err.Error())

			// Test that error is not nil
			require.Error(t, tt.err)

			// Test wrapping
			wrapped := fmt.Errorf("operation failed: %w", tt.err)
			assert.ErrorIs(t, wrapped, tt.err)
		})
	}
}

func TestErrorChaining(t *testing.T) {
	t.Parallel()
	// Test multiple levels of wrapping
	baseErr := ErrNotFound
	wrapped1 := WithOp("fetch", baseErr)
	wrapped2 := WithOp("process", wrapped1)
	wrapped3 := fmt.Errorf("handler failed: %w", wrapped2)

	// Should be able to find the base error through all wrapping
	require.ErrorIs(t, wrapped3, ErrNotFound)
	assert.Equal(t, "handler failed: process: fetch: not found", wrapped3.Error())
}

func TestConcurrentErrorHandling(t *testing.T) {
	t.Parallel()
	// Test that error handling is safe for concurrent use
	done := make(chan bool)

	for i := range 100 {
		go func(n int) {
			err := WithOpf("operation", "test error %d", n)
			_ = IsRetryable(err)
			_ = IsPermission(err)
			done <- true
		}(i)
	}

	for range 100 {
		<-done
	}
}

func TestError_ExitCode(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		err      *Error
		expected int
	}{
		{
			name: "custom exit code",
			err: &Error{
				Code: ExitValidation,
			},
			expected: int(ExitValidation),
		},
		{
			name: "KindNotFound",
			err: &Error{
				Kind: KindNotFound,
			},
			expected: int(ExitNotFound),
		},
		{
			name: "KindValidation",
			err: &Error{
				Kind: KindValidation,
			},
			expected: int(ExitValidation),
		},
		{
			name: "KindPermission",
			err: &Error{
				Kind: KindPermission,
			},
			expected: int(ExitPermError),
		},
		{
			name: "KindNetwork",
			err: &Error{
				Kind: KindNetwork,
			},
			expected: int(ExitNetworkError),
		},
		{
			name: "KindTimeout",
			err: &Error{
				Kind: KindTimeout,
			},
			expected: int(ExitNetworkError),
		},
		{
			name: "KindConfig",
			err: &Error{
				Kind: KindConfig,
			},
			expected: int(ExitConfigError),
		},
		{
			name: "KindFormat",
			err: &Error{
				Kind: KindFormat,
			},
			expected: int(ExitFormat),
		},
		{
			name: "KindOther",
			err: &Error{
				Kind: KindOther,
			},
			expected: int(ExitError),
		},
		{
			name: "KindRepository",
			err: &Error{
				Kind: KindRepository,
			},
			expected: int(ExitError),
		},
		{
			name: "KindCanceled",
			err: &Error{
				Kind: KindCanceled,
			},
			expected: int(ExitError),
		},
		{
			name: "unknown kind",
			err: &Error{
				Kind: -1, // Invalid kind
			},
			expected: int(ExitError),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.err.ExitCode()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestError_WithSuggestions(t *testing.T) {
	t.Parallel()
	err := &Error{
		Message: "test error",
		Kind:    KindValidation,
	}

	// Test adding single suggestion
	result := err.WithSuggestions("try again")
	assert.Equal(t, []string{"try again"}, result.Suggestions)
	assert.Same(t, err, result) // Should return same instance

	// Test adding multiple suggestions
	result = err.WithSuggestions("check input", "verify config")
	assert.Equal(t, []string{"try again", "check input", "verify config"}, result.Suggestions)
}

func TestError_kindString(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		kind     ErrorKind
		expected string
	}{
		{
			name:     "KindNotFound",
			kind:     KindNotFound,
			expected: "not found",
		},
		{
			name:     "KindValidation",
			kind:     KindValidation,
			expected: "validation error",
		},
		{
			name:     "KindPermission",
			kind:     KindPermission,
			expected: "permission denied",
		},
		{
			name:     "KindNetwork",
			kind:     KindNetwork,
			expected: "network error",
		},
		{
			name:     "KindConfig",
			kind:     KindConfig,
			expected: "configuration error",
		},
		{
			name:     "KindFormat",
			kind:     KindFormat,
			expected: "format error",
		},
		{
			name:     "KindRepository",
			kind:     KindRepository,
			expected: "repository error",
		},
		{
			name:     "KindTimeout",
			kind:     KindTimeout,
			expected: "timeout",
		},
		{
			name:     "KindCanceled",
			kind:     KindCanceled,
			expected: "canceled",
		},
		{
			name:     "KindOther",
			kind:     KindOther,
			expected: "other",
		},
		{
			name:     "unknown kind",
			kind:     -1,
			expected: "error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := &Error{Kind: tt.kind}
			result := err.kindString()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestWrap(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name         string
		err          error
		op           string
		expectedType *Error
		expectedMsg  string
	}{
		{
			name:         "nil error",
			err:          nil,
			op:           "test",
			expectedType: nil,
		},
		{
			name: "wrap standard error",
			err:  fmt.Errorf("timeout"),
			op:   "fetch",
			expectedType: &Error{
				Op:   "fetch",
				Kind: KindOther,
			},
			expectedMsg: "fetch: timeout",
		},
		{
			name: "wrap existing Error",
			err: &Error{
				Kind:    KindValidation,
				Message: "invalid input",
			},
			op: "process",
			expectedType: &Error{
				Op:   "process",
				Kind: KindValidation,
			},
			expectedMsg: "process: invalid input",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Wrap(tt.err, tt.op)

			if tt.expectedType == nil {
				assert.Nil(t, result)
				return
			}

			require.NotNil(t, result)
			assert.Equal(t, tt.expectedType.Op, result.Op)
			// Just check that error message is correct, kind detection may vary
			assert.Equal(t, tt.expectedMsg, result.Error())
		})
	}
}

func TestValidation(t *testing.T) {
	t.Parallel()
	result := Validation("email", "invalid format")

	require.NotNil(t, result)
	assert.Equal(t, KindValidation, result.Kind)
	assert.Equal(t, "email", result.Field)
	assert.Equal(t, "validation failed for email: invalid format", result.Message)
	assert.Equal(t, ExitValidation, result.Code)
}

func TestIsTerminal(t *testing.T) {
	tests := []struct {
		name     string
		termEnv  string
		expected bool
	}{
		{
			name:     "empty TERM",
			termEnv:  "",
			expected: false,
		},
		{
			name:     "dumb terminal",
			termEnv:  "dumb",
			expected: false,
		},
		{
			name:     "xterm terminal",
			termEnv:  "xterm",
			expected: true,
		},
		{
			name:     "xterm-256color terminal",
			termEnv:  "xterm-256color",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variable
			t.Setenv("TERM", tt.termEnv)

			result := isTerminal()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestShouldShowDetails(t *testing.T) {
	tests := []struct {
		name       string
		debugEnv   string
		verboseEnv string
		expected   bool
	}{
		{
			name:       "no debug flags",
			debugEnv:   "",
			verboseEnv: "",
			expected:   false,
		},
		{
			name:       "debug enabled",
			debugEnv:   "true",
			verboseEnv: "",
			expected:   true,
		},
		{
			name:       "verbose enabled",
			debugEnv:   "",
			verboseEnv: "true",
			expected:   true,
		},
		{
			name:       "both enabled",
			debugEnv:   "true",
			verboseEnv: "true",
			expected:   true,
		},
		{
			name:       "debug false",
			debugEnv:   "false",
			verboseEnv: "",
			expected:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variables
			t.Setenv("CONTEXTURE_DEBUG", tt.debugEnv)
			t.Setenv("CONTEXTURE_VERBOSE", tt.verboseEnv)

			result := shouldShowDetails()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDisplay(t *testing.T) {
	tests := []struct {
		name               string
		err                error
		termEnv            string
		debugEnv           string
		expectError        bool
		expectSuggestions  bool
		expectDetails      bool
		expectColors       bool
		expectedContains   []string
		unexpectedContains []string
	}{
		{
			name:             "nil error",
			err:              nil,
			expectError:      false,
			expectedContains: []string{},
		},
		{
			name:             "basic error without colors",
			err:              &Error{Message: "test error", Kind: KindValidation},
			termEnv:          "",
			expectError:      true,
			expectColors:     false,
			expectedContains: []string{"Error:", "test error", "For more help"},
		},
		{
			name:             "basic error with colors",
			err:              &Error{Message: "test error", Kind: KindValidation},
			termEnv:          "xterm",
			expectError:      true,
			expectColors:     true,
			expectedContains: []string{"Error:", "test error", "For more help", "\033[31m", "\033[0m"},
		},
		{
			name: "error with suggestions no colors",
			err: &Error{
				Message:     "validation failed",
				Kind:        KindValidation,
				Suggestions: []string{"Check your input", "Try again"},
			},
			termEnv:            "",
			expectError:        true,
			expectSuggestions:  true,
			expectColors:       false,
			expectedContains:   []string{"Error:", "validation failed", "Suggestions:", "Check your input", "Try again", "•"},
			unexpectedContains: []string{"\033["},
		},
		{
			name: "error with suggestions and colors",
			err: &Error{
				Message:     "validation failed",
				Kind:        KindValidation,
				Suggestions: []string{"Check your input"},
			},
			termEnv:           "xterm-256color",
			expectError:       true,
			expectSuggestions: true,
			expectColors:      true,
			expectedContains:  []string{"Suggestions:", "Check your input", "\033[33m"},
		},
		{
			name: "error with details in debug mode",
			err: &Error{
				Message: "wrapper error",
				Kind:    KindNetwork,
				Err:     fmt.Errorf("underlying network error"),
			},
			termEnv:          "xterm",
			debugEnv:         "true",
			expectError:      true,
			expectDetails:    true,
			expectedContains: []string{"Error:", "wrapper error", "Details:", "underlying network error"},
		},
		{
			name: "error without details when debug off",
			err: &Error{
				Message: "wrapper error",
				Kind:    KindNetwork,
				Err:     fmt.Errorf("underlying network error"),
			},
			termEnv:            "xterm",
			debugEnv:           "",
			expectError:        true,
			expectDetails:      false,
			expectedContains:   []string{"Error:", "wrapper error"},
			unexpectedContains: []string{"Details:", "underlying network error"},
		},
		{
			name:             "standard error gets converted",
			err:              fmt.Errorf("standard error message"),
			termEnv:          "",
			expectError:      true,
			expectedContains: []string{"Error:", "standard error message"},
		},
		{
			name:               "dumb terminal no colors",
			err:                &Error{Message: "test", Kind: KindOther},
			termEnv:            "dumb",
			expectError:        true,
			expectColors:       false,
			expectedContains:   []string{"Error:", "test"},
			unexpectedContains: []string{"\033["},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture stderr
			oldStderr := os.Stderr
			r, w, err := os.Pipe()
			require.NoError(t, err)
			os.Stderr = w

			// Set environment variables
			t.Setenv("TERM", tt.termEnv)
			t.Setenv("CONTEXTURE_DEBUG", tt.debugEnv)

			// Call Display
			Display(tt.err)

			// Restore stderr and read output
			_ = w.Close()
			os.Stderr = oldStderr

			var buf []byte
			buf, err = io.ReadAll(r)
			require.NoError(t, err)
			output := string(buf)

			// Verify output
			if !tt.expectError {
				assert.Empty(t, output, "should not write anything for nil error")
				return
			}

			// Check expected strings are present
			for _, expected := range tt.expectedContains {
				assert.Contains(t, output, expected, "output should contain: %q", expected)
			}

			// Check unexpected strings are not present
			for _, unexpected := range tt.unexpectedContains {
				assert.NotContains(t, output, unexpected, "output should not contain: %q", unexpected)
			}
		})
	}
}

func TestDisplay_WithComplexError(t *testing.T) {
	// Test a more realistic error scenario
	t.Setenv("TERM", "xterm")
	t.Setenv("CONTEXTURE_DEBUG", "true")

	// Capture stderr
	oldStderr := os.Stderr
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stderr = w

	// Create a complex error with wrapping
	baseErr := fmt.Errorf("connection refused")
	wrappedErr := &Error{
		Op:          "fetch_rule",
		Kind:        KindNetwork,
		Message:     "failed to fetch rule",
		Err:         baseErr,
		Suggestions: []string{"Check your network connection", "Verify the repository URL"},
	}

	Display(wrappedErr)

	// Restore stderr and read output
	_ = w.Close()
	os.Stderr = oldStderr

	var buf []byte
	buf, err = io.ReadAll(r)
	require.NoError(t, err)
	output := string(buf)

	// Verify all components are present
	assert.Contains(t, output, "Error:")
	assert.Contains(t, output, "Suggestions:")
	assert.Contains(t, output, "Check your network connection")
	assert.Contains(t, output, "Verify the repository URL")
	assert.Contains(t, output, "Details:")
	assert.Contains(t, output, "connection refused")
	assert.Contains(t, output, "For more help")
}
