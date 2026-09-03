package response

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	apperrors "github.com/dbplatform/api/internal/errors"
)

type Response struct {
	Data  interface{}  `json:"data,omitempty"`
	Meta  *Meta        `json:"meta,omitempty"`
	Error *ErrorDetail `json:"error,omitempty"`
}

type Meta struct {
	RequestID string     `json:"request_id,omitempty"`
	Timestamp time.Time  `json:"timestamp"`
	Page      *PageMeta  `json:"page,omitempty"`
}

type PageMeta struct {
	Total  int `json:"total"`
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func JSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		return
	}
}

func requestID(r *http.Request) string {
	return r.Header.Get("X-Request-ID")
}

func Success(w http.ResponseWriter, r *http.Request, data interface{}) {
	JSON(w, http.StatusOK, Response{
		Data: data,
		Meta: &Meta{RequestID: requestID(r), Timestamp: time.Now()},
	})
}

func Created(w http.ResponseWriter, r *http.Request, data interface{}) {
	JSON(w, http.StatusCreated, Response{
		Data: data,
		Meta: &Meta{RequestID: requestID(r), Timestamp: time.Now()},
	})
}

func NoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

func Paginated(w http.ResponseWriter, r *http.Request, data interface{}, total, limit, offset int) {
	JSON(w, http.StatusOK, Response{
		Data: data,
		Meta: &Meta{
			RequestID: requestID(r),
			Timestamp: time.Now(),
			Page:      &PageMeta{Total: total, Limit: limit, Offset: offset},
		},
	})
}

func Error(w http.ResponseWriter, r *http.Request, err error) {
	var appErr *apperrors.AppError
	if errors.As(err, &appErr) {
		JSON(w, appErr.HTTPStatus, Response{
			Error: &ErrorDetail{
				Code:    string(appErr.Code),
				Message: appErr.Message,
			},
		})
		return
	}

	JSON(w, http.StatusInternalServerError, Response{
		Error: &ErrorDetail{
			Code:    string(apperrors.CodeInternal),
			Message: "An unexpected error occurred",
		},
	})
}
