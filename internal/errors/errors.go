// Package errors provides unified error handling for the Contexture CLI
package errors

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
)

// ErrorCode represents standardized error codes
type ErrorCode int

const (
	// ExitSuccess indicates successful completion
	ExitSuccess ErrorCode = iota
	// ExitError indicates a general error
	ExitError
	// ExitUsageError indicates incorrect usage
	ExitUsageError
	// ExitConfigError indicates configuration error
	ExitConfigError
	// ExitPermError indicates permission error
	ExitPermError
	// ExitNetworkError indicates network error
	ExitNetworkError
	// ExitNotFound indicates resource not found
	ExitNotFound
	// ExitValidation indicates validation error
	ExitValidation
	// ExitFormat indicates format error
	ExitFormat
)

// Error represents a unified error with user-friendly messaging
type Error struct {
	// Core error information
	Op   string    // Operation that failed
	Kind ErrorKind // Error kind/category
	Code ErrorCode // Exit code
	Err  error     // Underlying error

	// User-friendly information
	Message     string   // User-friendly message
	Suggestions []string // Helpful suggestions
	Field       string   // Field name for validation errors
}

// ErrorKind represents the category of error
type ErrorKind int

const (
	// KindOther represents other/unclassified errors
	KindOther ErrorKind = iota
	// KindNotFound represents not found errors
	KindNotFound
	// KindValidation represents validation errors
	KindValidation
	// KindPermission represents permission errors
	KindPermission
	// KindNetwork represents network errors
	KindNetwork
	// KindConfig represents configuration errors
	KindConfig
	// KindFormat represents format errors
	KindFormat
	// KindRepository represents repository errors
	KindRepository
	// KindTimeout represents timeout errors
	KindTimeout
	// KindCanceled represents canceled operation errors
	KindCanceled
)

// Error implements the error interface
func (e *Error) Error() string {
	if e.Message != "" {
		return e.Message
	}

	var b strings.Builder
	if e.Op != "" {
		b.WriteString(e.Op)
		b.WriteString(": ")
	}

	if e.Field != "" {
		fmt.Fprintf(&b, "%s: ", e.Field)
	}

	if e.Err != nil {
		b.WriteString(e.Err.Error())
	} else {
		b.WriteString(e.kindString())
	}

	return b.String()
}

// Unwrap returns the underlying error
func (e *Error) Unwrap() error {
	return e.Err
}

// ExitCode returns the appropriate exit code
func (e *Error) ExitCode() int {
	if e.Code != 0 {
		return int(e.Code)
	}

	// Map kind to exit code
	switch e.Kind {
	case KindNotFound:
		return int(ExitNotFound)
	case KindValidation:
		return int(ExitValidation)
	case KindPermission:
		return int(ExitPermError)
	case KindNetwork, KindTimeout:
		return int(ExitNetworkError)
	case KindConfig:
		return int(ExitConfigError)
	case KindFormat:
		return int(ExitFormat)
	case KindOther, KindRepository, KindCanceled:
		return int(ExitError)
	default:
		return int(ExitError)
	}
}

// IsRetryable returns true if the error is retryable
func (e *Error) IsRetryable() bool {
	return e.Kind == KindTimeout || e.Kind == KindNetwork
}

// WithSuggestions adds suggestions to the error
func (e *Error) WithSuggestions(suggestions ...string) *Error {
	e.Suggestions = append(e.Suggestions, suggestions...)
	return e
}

// kindString returns a string representation of the error kind
func (e *Error) kindString() string {
	switch e.Kind {
	case KindNotFound:
		return "not found"
	case KindValidation:
		return "validation error"
	case KindPermission:
		return "permission denied"
	case KindNetwork:
		return "network error"
	case KindConfig:
		return "configuration error"
	case KindFormat:
		return "format error"
	case KindRepository:
		return "repository error"
	case KindTimeout:
		return "timeout"
	case KindCanceled:
		return "canceled"
	case KindOther:
		return "other"
	default:
		return "error"
	}
}

// Wrap wraps an error with additional context
func Wrap(err error, op string) *Error {
	if err == nil {
		return nil
	}

	// If it's already our error type, create a new wrapper to preserve the chain
	var e *Error
	if errors.As(err, &e) {
		return &Error{
			Op:   op,
			Err:  err, // Wrap the existing error to preserve the chain
			Kind: e.Kind,
		}
	}

	// Create new error with detected kind
	return &Error{
		Op:   op,
		Err:  err,
		Kind: detectKind(err),
	}
}

// Validation creates a validation error
func Validation(field, message string) *Error {
	return &Error{
		Kind:    KindValidation,
		Field:   field,
		Message: fmt.Sprintf("validation failed for %s: %s", field, message),
		Code:    ExitValidation,
	}
}

// detectKind attempts to detect the error kind from the error message
func detectKind(err error) ErrorKind {
	if err == nil {
		return KindOther
	}

	msg := strings.ToLower(err.Error())

	switch {
	case strings.Contains(msg, "not found"):
		return KindNotFound
	case strings.Contains(msg, "validation") || strings.Contains(msg, "invalid"):
		return KindValidation
	case strings.Contains(msg, "permission denied") || strings.Contains(msg, "access denied"):
		return KindPermission
	case strings.Contains(msg, "timeout") || strings.Contains(msg, "deadline exceeded"):
		return KindTimeout
	case strings.Contains(msg, "connection") || strings.Contains(msg, "network"):
		return KindNetwork
	case strings.Contains(msg, "config"):
		return KindConfig
	case strings.Contains(msg, "yaml") || strings.Contains(msg, "json") || strings.Contains(msg, "parse"):
		return KindFormat
	case strings.Contains(msg, "repository") || strings.Contains(msg, "git"):
		return KindRepository
	default:
		return KindOther
	}
}

// Display shows the error in a user-friendly format
func Display(err error) {
	if err == nil {
		return
	}

	var e *Error
	if !errors.As(err, &e) {
		// Convert to our error type
		e = Wrap(err, "")
	}

	// Use colors if terminal supports it
	errorColor := "\033[31m"      // Red
	suggestionColor := "\033[33m" // Yellow
	resetColor := "\033[0m"

	// Check if we should use colors
	if !isTerminal() {
		errorColor = ""
		suggestionColor = ""
		resetColor = ""
	}

	// Display main error message
	fmt.Fprintf(os.Stderr, "%sError:%s %s\n", errorColor, resetColor, e.Error())

	// Display suggestions if available
	if len(e.Suggestions) > 0 {
		fmt.Fprintf(os.Stderr, "\n%sSuggestions:%s\n", suggestionColor, resetColor)
		for _, suggestion := range e.Suggestions {
			fmt.Fprintf(os.Stderr, "  • %s\n", suggestion)
		}
	}

	// Show technical details in debug mode
	if shouldShowDetails() && e.Err != nil {
		fmt.Fprintf(os.Stderr, "\n%sDetails:%s %s\n", suggestionColor, resetColor, e.Err.Error())
	}

	fmt.Fprintf(os.Stderr, "\nFor more help, run: contexture --help\n")
}

// isTerminal checks if the output is a terminal
func isTerminal() bool {
	term := os.Getenv("TERM")
	return term != "" && term != "dumb"
}

// shouldShowDetails checks if technical details should be shown
func shouldShowDetails() bool {
	return os.Getenv("CONTEXTURE_DEBUG") == "true" ||
		os.Getenv("CONTEXTURE_VERBOSE") == "true"
}

// Sentinel errors for common cases (kept for backward compatibility)
var (
	// General errors
	ErrNotFound         = errors.New("not found")
	ErrInvalidInput     = errors.New("invalid input")
	ErrPermissionDenied = errors.New("permission denied")
	ErrTimeout          = errors.New("timeout")
	ErrCanceled         = errors.New("operation canceled")

	// Domain-specific errors
	ErrInvalidRuleID        = errors.New("invalid rule ID format")
	ErrRuleNotFound         = errors.New("rule not found")
	ErrRepositoryNotFound   = errors.New("repository not found")
	ErrConfigNotFound       = errors.New("configuration not found")
	ErrUnsupportedFormat    = errors.New("unsupported format")
	ErrAuthenticationFailed = errors.New("authentication failed")
)

// OpError is now an alias to Error for backward compatibility
type OpError = Error

// WithOp wraps an error with operation context
func WithOp(op string, err error) error {
	if err == nil {
		return nil
	}
	return Wrap(err, op)
}

// WithOpf wraps an error with operation context and formatting
func WithOpf(op string, format string, args ...any) error {
	return WithOp(op, fmt.Errorf(format, args...))
}

// IsRetryable checks if an error is retryable (global function for convenience)
func IsRetryable(err error) bool {
	if err == nil {
		return false
	}

	// Check for common retryable errors
	if errors.Is(err, ErrTimeout) ||
		errors.Is(err, ErrCanceled) ||
		errors.Is(err, context.DeadlineExceeded) {
		return true
	}

	// Check if it's our error type and use its method
	var e *Error
	if errors.As(err, &e) {
		return e.IsRetryable()
	}

	// Check error message for retryable conditions
	msg := strings.ToLower(err.Error())
	retryableConditions := []string{
		"timeout",
		"connection reset",
		"connection refused",
		"temporary failure",
		"network is unreachable",
		"no route to host",
		"deadline exceeded",
	}

	for _, condition := range retryableConditions {
		if strings.Contains(msg, condition) {
			return true
		}
	}

	return false
}

// IsPermission checks if an error is permission-related
func IsPermission(err error) bool {
	if err == nil {
		return false
	}

	if errors.Is(err, ErrPermissionDenied) {
		return true
	}

	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "permission denied") ||
		strings.Contains(msg, "access denied") ||
		strings.Contains(msg, "operation not permitted")
}

// ValidationError wraps a validation error with field context
func ValidationError(field string, err error) error {
	if err == nil {
		return nil
	}
	return Validation(field, err.Error())
}

// ValidationErrorf creates a formatted validation error
func ValidationErrorf(field, format string, args ...any) error {
	return Validation(field, fmt.Sprintf(format, args...))
}
