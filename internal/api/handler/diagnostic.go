package handler

import (
	"net/http"

	"github.com/dbplatform/api/internal/api/response"
	"github.com/dbplatform/api/internal/diagnostic"
	apperrors "github.com/dbplatform/api/internal/errors"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type DiagnosticHandler struct {
	svc *diagnostic.Service
}

func NewDiagnosticHandler(svc *diagnostic.Service) *DiagnosticHandler {
	return &DiagnosticHandler{svc: svc}
}

func (h *DiagnosticHandler) GetLocks(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(w, r, apperrors.NewValidation("Invalid database ID"))
		return
	}

	locks, err := h.svc.GetActiveLocks(r.Context(), id)
	if err != nil {
		response.Error(w, r, err)
		return
	}

	response.Success(w, r, locks)
}

func (h *DiagnosticHandler) GetStorage(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(w, r, apperrors.NewValidation("Invalid database ID"))
		return
	}

	stats, err := h.svc.GetStorageStats(r.Context(), id)
	if err != nil {
		response.Error(w, r, err)
		return
	}

	response.Success(w, r, stats)
}
