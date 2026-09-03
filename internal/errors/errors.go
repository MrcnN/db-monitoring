package errors

import "net/http"

type Code string

const (
	CodeInternal           Code = "INTERNAL_ERROR"
	CodeValidation         Code = "VALIDATION_ERROR"
	CodeNotFound           Code = "NOT_FOUND"
	CodeUnauthorized       Code = "UNAUTHORIZED"
	CodeForbidden          Code = "FORBIDDEN"
	CodeConflict           Code = "CONFLICT"
	CodeDBConnection       Code = "DATABASE_CONNECTION_FAILED"
	CodeDBQuery            Code = "DATABASE_QUERY_ERROR"
	CodeTimeout            Code = "TIMEOUT"
	CodeRateLimit          Code = "RATE_LIMIT_EXCEEDED"
	CodeInvalidCredentials Code = "INVALID_CREDENTIALS"
)

type AppError struct {
	Code       Code
	Message    string
	HTTPStatus int
	Err        error
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}
	return e.Message
}

func (e *AppError) Unwrap() error { return e.Err }

func NewInternal(err error) *AppError {
	return &AppError{Code: CodeInternal, Message: "Internal server error", HTTPStatus: http.StatusInternalServerError, Err: err}
}

func NewValidation(msg string) *AppError {
	return &AppError{Code: CodeValidation, Message: msg, HTTPStatus: http.StatusBadRequest}
}

func NewNotFound(resource string) *AppError {
	return &AppError{Code: CodeNotFound, Message: resource + " not found", HTTPStatus: http.StatusNotFound}
}

func NewUnauthorized() *AppError {
	return &AppError{Code: CodeUnauthorized, Message: "Unauthorized", HTTPStatus: http.StatusUnauthorized}
}

func NewForbidden() *AppError {
	return &AppError{Code: CodeForbidden, Message: "Forbidden", HTTPStatus: http.StatusForbidden}
}

func NewConflict(msg string) *AppError {
	return &AppError{Code: CodeConflict, Message: msg, HTTPStatus: http.StatusConflict}
}

func NewDBConnection(err error) *AppError {
	return &AppError{Code: CodeDBConnection, Message: "Database connection failed", HTTPStatus: http.StatusInternalServerError, Err: err}
}

func NewInvalidCredentials() *AppError {
	return &AppError{Code: CodeInvalidCredentials, Message: "Invalid credentials", HTTPStatus: http.StatusUnauthorized}
}

func IsNotFound(err error) bool {
	if e, ok := IsAppError(err); ok {
		return e.Code == CodeNotFound
	}
	return false
}

func IsAppError(err error) (*AppError, bool) {
	e, ok := err.(*AppError)
	return e, ok
}
