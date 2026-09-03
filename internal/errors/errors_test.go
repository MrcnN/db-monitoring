package errors_test

import (
	"errors"
	"net/http"
	"testing"

	apperrors "github.com/dbplatform/api/internal/errors"
)

func TestNewInternal(t *testing.T) {
	underlying := errors.New("db is down")
	err := apperrors.NewInternal(underlying)

	if err.Code != apperrors.CodeInternal {
		t.Errorf("expected code %q, got %q", apperrors.CodeInternal, err.Code)
	}
	if err.HTTPStatus != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", err.HTTPStatus)
	}
	if !errors.Is(err, underlying) {
		t.Error("expected errors.Is to find underlying error")
	}
}

func TestNewValidation(t *testing.T) {
	err := apperrors.NewValidation("field is required")
	if err.Code != apperrors.CodeValidation {
		t.Errorf("expected code %q, got %q", apperrors.CodeValidation, err.Code)
	}
	if err.HTTPStatus != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", err.HTTPStatus)
	}
}

func TestNewNotFound(t *testing.T) {
	err := apperrors.NewNotFound("User")
	if err.Code != apperrors.CodeNotFound {
		t.Errorf("expected code %q, got %q", apperrors.CodeNotFound, err.Code)
	}
	if err.HTTPStatus != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", err.HTTPStatus)
	}
}

func TestNewUnauthorized(t *testing.T) {
	err := apperrors.NewUnauthorized()
	if err.HTTPStatus != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", err.HTTPStatus)
	}
}

func TestNewForbidden(t *testing.T) {
	err := apperrors.NewForbidden()
	if err.HTTPStatus != http.StatusForbidden {
		t.Errorf("expected status 403, got %d", err.HTTPStatus)
	}
}

func TestNewConflict(t *testing.T) {
	err := apperrors.NewConflict("email already exists")
	if err.Code != apperrors.CodeConflict {
		t.Errorf("expected code %q, got %q", apperrors.CodeConflict, err.Code)
	}
	if err.HTTPStatus != http.StatusConflict {
		t.Errorf("expected status 409, got %d", err.HTTPStatus)
	}
}

func TestIsNotFound_True(t *testing.T) {
	err := apperrors.NewNotFound("Database")
	if !apperrors.IsNotFound(err) {
		t.Error("expected IsNotFound to return true")
	}
}

func TestIsNotFound_False(t *testing.T) {
	err := apperrors.NewInternal(errors.New("oops"))
	if apperrors.IsNotFound(err) {
		t.Error("expected IsNotFound to return false for internal error")
	}
}

func TestIsAppError_True(t *testing.T) {
	err := apperrors.NewValidation("bad input")
	appErr, ok := apperrors.IsAppError(err)
	if !ok {
		t.Fatal("expected IsAppError to return true")
	}
	if appErr == nil {
		t.Fatal("expected non-nil AppError")
	}
}

func TestIsAppError_False(t *testing.T) {
	err := errors.New("plain error")
	_, ok := apperrors.IsAppError(err)
	if ok {
		t.Error("expected IsAppError to return false for plain error")
	}
}

func TestAppError_Error_WithUnderlying(t *testing.T) {
	underlying := errors.New("connection refused")
	err := apperrors.NewInternal(underlying)
	if err.Error() != underlying.Error() {
		t.Errorf("expected error message %q, got %q", underlying.Error(), err.Error())
	}
}

func TestAppError_Error_WithoutUnderlying(t *testing.T) {
	err := apperrors.NewValidation("required field missing")
	if err.Error() != "required field missing" {
		t.Errorf("expected message %q, got %q", "required field missing", err.Error())
	}
}
