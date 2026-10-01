// Package apperror defines sentinel errors shared by the service and handler
// layers. Go has no exceptions, so services return these (wrapped with
// fmt.Errorf("...: %w", apperror.ErrNotFound) when they need more context) and
// handlers map them to HTTP status codes with errors.Is.
package apperror

import "errors"

var (
	// ErrNotFound means the requested resource does not exist.
	ErrNotFound = errors.New("resource not found")
	// ErrConflict means the request conflicts with existing state (e.g. duplicate email).
	ErrConflict = errors.New("resource conflict")
	// ErrInvalidInput means the caller sent a request that failed validation.
	ErrInvalidInput = errors.New("invalid input")
	// ErrUnauthorized means authentication failed or is missing.
	ErrUnauthorized = errors.New("unauthorized")
)
