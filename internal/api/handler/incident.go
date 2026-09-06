package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/dbplatform/api/internal/api/response"
	apperrors "github.com/dbplatform/api/internal/errors"
	"github.com/dbplatform/api/internal/incident"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

type IncidentHandler struct {
	incidentSvc *incident.Service
	log         zerolog.Logger
}

func NewIncidentHandler(incidentSvc *incident.Service, log zerolog.Logger) *IncidentHandler {
	return &IncidentHandler{
		incidentSvc: incidentSvc,
		log:         log,
	}
}

func (h *IncidentHandler) List(w http.ResponseWriter, r *http.Request) {
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

	incidents, err := h.incidentSvc.List(r.Context(), limit, offset)
	if err != nil {
		response.Error(w, r, apperrors.NewInternal(err))
		return
	}

	if incidents == nil {
		incidents = []incident.Incident{}
	}

	response.Success(w, r, incidents)
}

func (h *IncidentHandler) ListChannels(w http.ResponseWriter, r *http.Request) {
	channels, err := h.incidentSvc.ListChannels(r.Context())
	if err != nil {
		response.Error(w, r, apperrors.NewInternal(err))
		return
	}

	if channels == nil {
		channels = []incident.NotificationChannel{}
	}

	response.Success(w, r, channels)
}

func (h *IncidentHandler) CreateChannel(w http.ResponseWriter, r *http.Request) {
	var input incident.NotificationChannel
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.Error(w, r, apperrors.NewValidation("invalid request body"))
		return
	}

	if input.Name == "" || input.Type == "" {
		response.Error(w, r, apperrors.NewValidation("name and type are required"))
		return
	}

	if err := h.incidentSvc.CreateChannel(r.Context(), &input); err != nil {
		response.Error(w, r, apperrors.NewInternal(err))
		return
	}

	response.Created(w, r, input)
}

func (h *IncidentHandler) DeleteChannel(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, r, apperrors.NewValidation("invalid channel ID"))
		return
	}

	if err := h.incidentSvc.DeleteChannel(r.Context(), id); err != nil {
		response.Error(w, r, apperrors.NewInternal(err))
		return
	}

	response.NoContent(w)
}
