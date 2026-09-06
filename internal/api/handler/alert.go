package handler

import (
	"net/http"
	"strconv"

	"github.com/dbplatform/api/internal/alert"
	"github.com/dbplatform/api/internal/api/response"
	apperrors "github.com/dbplatform/api/internal/errors"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

type AlertHandler struct {
	alertSvc *alert.Service
	log      zerolog.Logger
}

func NewAlertHandler(alertSvc *alert.Service, log zerolog.Logger) *AlertHandler {
	return &AlertHandler{
		alertSvc: alertSvc,
		log:      log,
	}
}

func (h *AlertHandler) ListByDatabase(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, r, apperrors.NewValidation("invalid database ID"))
		return
	}

	limit := 50
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	offset := 0
	if o := r.URL.Query().Get("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	alerts, err := h.alertSvc.ListByDatabaseID(r.Context(), id, limit, offset)
	if err != nil {
		response.Error(w, r, apperrors.NewInternal(err))
		return
	}

	if alerts == nil {
		alerts = []alert.Alert{}
	}

	response.Success(w, r, alerts)
}

func (h *AlertHandler) ListActive(w http.ResponseWriter, r *http.Request) {
	limit := 100
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	offset := 0
	if o := r.URL.Query().Get("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	alerts, err := h.alertSvc.ListActive(r.Context(), limit, offset)
	if err != nil {
		response.Error(w, r, apperrors.NewInternal(err))
		return
	}

	if alerts == nil {
		alerts = []alert.Alert{}
	}

	response.Success(w, r, alerts)
}
