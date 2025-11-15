// Package errors provides custom error types with HTTP status codes
// for the REST API v2 application.
package errors

import "fmt"

// AppError represents an application error with HTTP status code and error code.
// It wraps underlying errors while providing structured error information
// for API responses.
type AppError struct {
	Code    string // Error code for client reference (e.g., "NOT_FOUND")
	Message string // Human-readable message
	Err     error  // Underlying error (optional)
	Status  int    // HTTP status code
}

// Error implements the error interface.
// Returns the message with underlying error if present.
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

// Unwrap implements the errors.Unwrap interface.
// Allows errors.Is and errors.As to work with wrapped errors.
func (e *AppError) Unwrap() error {
	return e.Err
}

// Predefined common errors
var (
	// ErrNotFound indicates a requested resource was not found
	ErrNotFound = &AppError{
		Code:    "NOT_FOUND",
		Message: "Resource not found",
		Status:  404,
	}

	// ErrBadRequest indicates invalid request parameters
	ErrBadRequest = &AppError{
		Code:    "BAD_REQUEST",
		Message: "Invalid request",
		Status:  400,
	}

	// ErrInternal indicates an internal server error
	ErrInternal = &AppError{
		Code:    "INTERNAL_ERROR",
		Message: "Internal server error",
		Status:  500,
	}
)

// Wrap creates a new AppError by wrapping an underlying error
// with the context from a base error.
// This preserves the status code and error code while adding context.
func Wrap(base *AppError, err error) *AppError {
	return &AppError{
		Code:    base.Code,
		Message: base.Message,
		Err:     err,
		Status:  base.Status,
	}
}

// New creates a new AppError with the given code, message, and status.
func New(code, message string, status int) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Status:  status,
	}
}
