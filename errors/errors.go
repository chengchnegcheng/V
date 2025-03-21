package errors

import (
	"fmt"
	"net/http"
)

// Error represents a custom error type
type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Err     error  `json:"-"`
}

// Error implements the error interface
func (e *Error) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

// Unwrap implements the unwrap interface
func (e *Error) Unwrap() error {
	return e.Err
}

// New creates a new Error
func New(code int, message string, err error) *Error {
	return &Error{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

// Common error codes
const (
	CodeBadRequest         = http.StatusBadRequest
	CodeUnauthorized       = http.StatusUnauthorized
	CodeForbidden          = http.StatusForbidden
	CodeNotFound           = http.StatusNotFound
	CodeInternalServer     = http.StatusInternalServerError
	CodeServiceUnavailable = http.StatusServiceUnavailable
)

// Common error messages
var (
	ErrInvalidInput       = New(CodeBadRequest, "Invalid input", nil)
	ErrUnauthorizedAccess = New(CodeUnauthorized, "Unauthorized", nil)
	ErrForbiddenAccess    = New(CodeForbidden, "Forbidden", nil)
	ErrResourceNotFound   = New(CodeNotFound, "Not found", nil)
	ErrInternal           = New(CodeInternalServer, "Internal server error", nil)
	ErrNotFound           = New(CodeNotFound, "Resource not found", nil)
	ErrInternalServer     = New(CodeInternalServer, "Internal server error", nil)
)

// IsNotFound checks if the error is a not found error
func IsNotFound(err error) bool {
	if e, ok := err.(*Error); ok {
		return e.Code == CodeNotFound
	}
	return false
}

// IsUnauthorized checks if the error is an unauthorized error
func IsUnauthorized(err error) bool {
	if e, ok := err.(*Error); ok {
		return e.Code == CodeUnauthorized
	}
	return false
}

// IsForbidden checks if the error is a forbidden error
func IsForbidden(err error) bool {
	if e, ok := err.(*Error); ok {
		return e.Code == CodeForbidden
	}
	return false
}

// IsBadRequest checks if the error is a bad request error
func IsBadRequest(err error) bool {
	if e, ok := err.(*Error); ok {
		return e.Code == CodeBadRequest
	}
	return false
}

// IsInternal checks if the error is an internal server error
func IsInternal(err error) bool {
	if e, ok := err.(*Error); ok {
		return e.Code == CodeInternalServer
	}
	return false
}
