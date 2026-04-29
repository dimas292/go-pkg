package apperror

import (
	"fmt"
	"net/http"
)

// AppError is a structured error that separates internal details from client-facing messages.
// Internal errors are logged but never exposed to clients.
type AppError struct {
	Code    int    // HTTP status code
	Message string // Safe message for the client
	Err     error  // Internal error (logged, never sent to client)
}

// Error implements the error interface.
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

// Unwrap allows errors.Is / errors.As to work with the wrapped error.
func (e *AppError) Unwrap() error {
	return e.Err
}

// --- Constructors ---

// BadRequest creates a 400 error.
func BadRequest(message string) *AppError {
	return &AppError{Code: http.StatusBadRequest, Message: message}
}

// Unauthorized creates a 401 error.
func Unauthorized(message string) *AppError {
	return &AppError{Code: http.StatusUnauthorized, Message: message}
}

// Forbidden creates a 403 error.
func Forbidden(message string) *AppError {
	return &AppError{Code: http.StatusForbidden, Message: message}
}

// NotFound creates a 404 error.
func NotFound(message string) *AppError {
	return &AppError{Code: http.StatusNotFound, Message: message}
}

// Conflict creates a 409 error.
func Conflict(message string) *AppError {
	return &AppError{Code: http.StatusConflict, Message: message}
}

// Internal creates a 500 error. The internal error is logged but the message
// shown to the client is safe and generic.
func Internal(message string, err error) *AppError {
	return &AppError{Code: http.StatusInternalServerError, Message: message, Err: err}
}

// Wrap wraps an existing error with a specific HTTP status code and safe message.
func Wrap(code int, message string, err error) *AppError {
	return &AppError{Code: code, Message: message, Err: err}
}
