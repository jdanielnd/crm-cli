package model

import (
	"errors"
	"fmt"
)

// Sentinel errors for classification.
var (
	ErrNotFound   = errors.New("not found")
	ErrValidation = errors.New("validation error")
	ErrConflict   = errors.New("conflict")
	ErrDatabase   = errors.New("database error")
)

// ExitError wraps a sentinel with a user-facing message.
type ExitError struct {
	Message string
	Err     error
}

func (e *ExitError) Error() string { return e.Message }
func (e *ExitError) Unwrap() error { return e.Err }

// NewExitError creates an ExitError wrapping a sentinel.
func NewExitError(sentinel error, format string, args ...any) *ExitError {
	return &ExitError{
		Message: fmt.Sprintf(format, args...),
		Err:     sentinel,
	}
}

// ExitCode maps an error to a CLI exit code.
func ExitCode(err error) int {
	switch {
	case errors.Is(err, ErrValidation):
		return 2
	case errors.Is(err, ErrNotFound):
		return 3
	case errors.Is(err, ErrConflict):
		return 4
	case errors.Is(err, ErrDatabase):
		return 10
	default:
		return 1
	}
}
