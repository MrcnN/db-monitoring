package handler

import (
	"net/http"

	"github.com/dbplatform/api/internal/api/response"
	"github.com/dbplatform/api/internal/collector"
	"github.com/dbplatform/api/internal/database"
	apperrors "github.com/dbplatform/api/internal/errors"
	"github.com/dbplatform/api/internal/health"
	"github.com/dbplatform/api/internal/metrics"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

type MetricsHandler struct {
	metricSvc *metrics.Service
	dbSvc     *database.Service
	evaluator *health.Evaluator
	log       zerolog.Logger
}

func NewMetricsHandler(
	metricSvc *metrics.Service,
	dbSvc *database.Service,
	evaluator *health.Evaluator,
	log zerolog.Logger,
) *MetricsHandler {
	return &MetricsHandler{
		metricSvc: metricSvc,
		dbSvc:     dbSvc,
		evaluator: evaluator,
		log:       log,
	}
}

func (h *MetricsHandler) GetTimeSeries(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, r, apperrors.NewValidation("invalid database ID"))
		return
	}

	rangeParam := r.URL.Query().Get("range")
	tr := metrics.ParseTimeRange(rangeParam)

	points, err := h.metricSvc.GetTimeSeries(r.Context(), id, tr)
	if err != nil {
		response.Error(w, r, apperrors.NewInternal(err))
		return
	}

	response.Success(w, r, points)
}

func (h *MetricsHandler) GetLatest(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, r, apperrors.NewValidation("invalid database ID"))
		return
	}

	metric, err := h.metricSvc.GetLatest(r.Context(), id)
	if err != nil {
		if apperrors.IsNotFound(err) {
			response.Success(w, r, nil)
			return
		}
		response.Error(w, r, apperrors.NewInternal(err))
		return
	}

	response.Success(w, r, metric)
}

func (h *MetricsHandler) GetHealth(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, r, apperrors.NewValidation("invalid database ID"))
		return
	}

	db, err := h.dbSvc.GetByID(r.Context(), id)
	if err != nil {
		response.Error(w, r, err)
		return
	}

	latest, err := h.metricSvc.GetLatest(r.Context(), id)
	if err != nil && !apperrors.IsNotFound(err) {
		response.Error(w, r, apperrors.NewInternal(err))
		return
	}

	if latest == nil {
		res := h.evaluator.Evaluate(nil, db.Status != database.StatusError, nil)
		response.Success(w, r, res)
		return
	}

	snapshot := &collector.MetricsSnapshot{
		CPUUsage:              latest.CPUUsage,
		MemoryUsage:           latest.MemoryUsage,
		ConnectionsTotal:      latest.ConnectionsTotal,
		ConnectionsActive:     latest.ConnectionsActive,
		ConnectionsIdle:       latest.ConnectionsIdle,
		ConnectionUsagePct:    latest.ConnectionUsagePct,
		QueryRate:             latest.QueryRate,
		P95LatencyMs:          latest.P95LatencyMs,
		CacheHitRatio:         latest.CacheHitRatio,
		DatabaseSizeBytes:     latest.DatabaseSizeBytes,
		DeadTuplesCount:       latest.DeadTuplesCount,
		ActiveLocksCount:      latest.ActiveLocksCount,
		ReplicationLagSeconds: latest.ReplicationLagSeconds,
	}

	res := h.evaluator.Evaluate(snapshot, db.Status != database.StatusError, nil)
	response.Success(w, r, res)
}
