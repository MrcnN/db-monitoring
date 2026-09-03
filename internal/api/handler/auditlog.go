package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/dbplatform/api/internal/api/response"
	"github.com/dbplatform/api/internal/audit"
	apperrors "github.com/dbplatform/api/internal/errors"
	"github.com/rs/zerolog"
)

type AuditLogHandler struct {
	repo audit.Repository
	log  zerolog.Logger
}

func NewAuditLogHandler(repo audit.Repository, log zerolog.Logger) *AuditLogHandler {
	return &AuditLogHandler{repo: repo, log: log}
}

type auditLogResponse struct {
	ID           string     `json:"id"`
	UserEmail    string     `json:"user_email"`
	Action       string     `json:"action"`
	ResourceType string     `json:"resource_type"`
	ResourceName string     `json:"resource_name"`
	IPAddress    string     `json:"ip_address"`
	Status       string     `json:"status"`
	CreatedAt    time.Time  `json:"created_at"`
}

func (h *AuditLogHandler) List(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	if offset < 0 {
		offset = 0
	}

	filter := audit.ListFilter{
		Limit:  limit,
		Offset: offset,
		Action: r.URL.Query().Get("action"),
	}

	if from := r.URL.Query().Get("from"); from != "" {
		t, err := time.Parse(time.RFC3339, from)
		if err == nil {
			filter.From = &t
		}
	}
	if to := r.URL.Query().Get("to"); to != "" {
		t, err := time.Parse(time.RFC3339, to)
		if err == nil {
			filter.To = &t
		}
	}

	logs, total, err := h.repo.List(r.Context(), filter)
	if err != nil {
		response.Error(w, r, apperrors.NewInternal(err))
		return
	}

	resp := make([]auditLogResponse, len(logs))
	for i, l := range logs {
		resp[i] = auditLogResponse{
			ID:           l.ID.String(),
			UserEmail:    l.UserEmail,
			Action:       l.Action,
			ResourceType: l.ResourceType,
			ResourceName: l.ResourceName,
			IPAddress:    l.IPAddress,
			Status:       l.Status,
			CreatedAt:    l.CreatedAt,
		}
	}

	response.Paginated(w, r, resp, total, limit, offset)
}

