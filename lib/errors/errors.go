package errors

import (
	"errors"
	"fmt"
	"net/http"
)

// Standard error types
var (
	ErrBadRequest          = errors.New("bad request")
	ErrUnauthorized        = errors.New("unauthorized")
	ErrForbidden           = errors.New("forbidden")
	ErrNotFound            = errors.New("not found")
	ErrInternalServerError = errors.New("internal server error")
	ErrValidation          = errors.New("validation error")
	ErrDatabaseError       = errors.New("database error")
	ErrAlreadyExists       = errors.New("resource already exists")
)

// AppError represents an application error with context
type AppError struct {
	Err        error  // The underlying error
	StatusCode int    // HTTP status code
	Message    string // User-friendly error message
	Field      string // Optional field name for validation errors
	Internal   bool   // Whether this error should be logged but not exposed to client
}

// New creates a new application error
func New(err error, statusCode int, message string) *AppError {
	return &AppError{
		Err:        err,
		StatusCode: statusCode,
		Message:    message,
	}
}

// WithField adds a field name to an AppError (useful for validation errors)
func (e *AppError) WithField(field string) *AppError {
	e.Field = field
	return e
}

// AsInternal marks an error as internal (not to be exposed to clients)
func (e *AppError) AsInternal() *AppError {
	e.Internal = true
	return e
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("%s: %s (%s)", e.Err.Error(), e.Message, e.Field)
	}
	return fmt.Sprintf("%s: %s", e.Err.Error(), e.Message)
}

// Unwrap implements the errors.Unwrap interface
func (e *AppError) Unwrap() error {
	return e.Err
}

// Is implements the errors.Is interface for better error comparison
func (e *AppError) Is(target error) bool {
	return errors.Is(e.Err, target)
}

// Factory functions for common errors

// BadRequest creates a new 400 Bad Request error
func BadRequest(message string) *AppError {
	return New(ErrBadRequest, http.StatusBadRequest, message)
}

// ValidationError creates a new 400 Bad Request error for validation failures
func ValidationError(message string, field string) *AppError {
	return New(ErrValidation, http.StatusBadRequest, message).WithField(field)
}

// Unauthorized creates a new 401 Unauthorized error
func Unauthorized(message string) *AppError {
	return New(ErrUnauthorized, http.StatusUnauthorized, message)
}

// Forbidden creates a new 403 Forbidden error
func Forbidden(message string) *AppError {
	return New(ErrForbidden, http.StatusForbidden, message)
}

// NotFound creates a new 404 Not Found error
func NotFound(message string) *AppError {
	return New(ErrNotFound, http.StatusNotFound, message)
}

// InternalServerError creates a new 500 Internal Server Error
func InternalServerError(message string) *AppError {
	return New(ErrInternalServerError, http.StatusInternalServerError, message).AsInternal()
}

// DatabaseError creates a new 500 Internal Server Error for database failures
func DatabaseError(message string) *AppError {
	return New(ErrDatabaseError, http.StatusInternalServerError, message).AsInternal()
}

// AlreadyExists creates a new 409 Conflict error for duplicate resources
func AlreadyExists(message string) *AppError {
	return New(ErrAlreadyExists, http.StatusConflict, message)
}

// FromError converts a standard error into an AppError using reasonable defaults
func FromError(err error) *AppError {
	if err == nil {
		return nil
	}

	// Check if it's already an AppError
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr
	}

	// Default to internal server error for unknown errors
	return InternalServerError(err.Error())
}
