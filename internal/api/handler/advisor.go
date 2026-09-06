package handler

import (
	"encoding/json"
	"net/http"

	"github.com/dbplatform/api/internal/advisor"
	"github.com/dbplatform/api/internal/api/response"
	apperrors "github.com/dbplatform/api/internal/errors"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

type AdvisorHandler struct {
	advisorSvc *advisor.Service
	log        zerolog.Logger
}

func NewAdvisorHandler(advisorSvc *advisor.Service, log zerolog.Logger) *AdvisorHandler {
	return &AdvisorHandler{
		advisorSvc: advisorSvc,
		log:        log,
	}
}

func (h *AdvisorHandler) ExplainQuery(w http.ResponseWriter, r *http.Request) {
	dbID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, r, apperrors.NewValidation("invalid database ID"))
		return
	}

	var input advisor.ExplainRequest
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.Error(w, r, apperrors.NewValidation("invalid request body"))
		return
	}

	if input.Query == "" {
		response.Error(w, r, apperrors.NewValidation("query is required"))
		return
	}

	res, err := h.advisorSvc.ExplainQuery(r.Context(), dbID, input.Query)
	if err != nil {
		response.Error(w, r, apperrors.NewInternal(err))
		return
	}

	response.Success(w, r, res)
}
