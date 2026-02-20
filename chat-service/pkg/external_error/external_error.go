package externalerror

import "net/http"

// ExternalError is an error safe to expose to the client.
// HTTPCode is the HTTP status code, Message is user-facing, Err is the internal cause (never exposed).
type ExternalError struct {
	HTTPCode int
	Message  string
	Err      error
}

func (e *ExternalError) Error() string { return e.Message }
func (e *ExternalError) Unwrap() error { return e.Err }

func New(code int, message string) *ExternalError {
	return &ExternalError{HTTPCode: code, Message: message}
}

func Wrap(code int, message string, err error) *ExternalError {
	return &ExternalError{HTTPCode: code, Message: message, Err: err}
}

// Common constructors

func BadRequest(msg string) *ExternalError {
	return New(http.StatusBadRequest, msg)
}

func Unauthorized(msg string) *ExternalError {
	return New(http.StatusUnauthorized, msg)
}

func Forbidden(msg string) *ExternalError {
	return New(http.StatusForbidden, msg)
}

func NotFound(msg string) *ExternalError {
	return New(http.StatusNotFound, msg)
}

func Conflict(msg string) *ExternalError {
	return New(http.StatusConflict, msg)
}

func UnprocessableEntity(msg string) *ExternalError {
	return New(http.StatusUnprocessableEntity, msg)
}

func Internal() *ExternalError {
	return New(http.StatusInternalServerError, "internal server error")
}
