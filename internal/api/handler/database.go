package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/dbplatform/api/internal/api/middleware"
	"github.com/dbplatform/api/internal/api/response"
	"github.com/dbplatform/api/internal/audit"
	"github.com/dbplatform/api/internal/crypto"
	"github.com/dbplatform/api/internal/database"
	apperrors "github.com/dbplatform/api/internal/errors"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

type DatabaseHandler struct {
	dbSvc     *database.Service
	auditSvc  *audit.Service
	encryptor *crypto.Encryptor
	validate  *validator.Validate
	log       zerolog.Logger
}

func NewDatabaseHandler(
	dbSvc *database.Service,
	auditSvc *audit.Service,
	encryptor *crypto.Encryptor,
	log zerolog.Logger,
) *DatabaseHandler {
	return &DatabaseHandler{
		dbSvc:     dbSvc,
		auditSvc:  auditSvc,
		encryptor: encryptor,
		validate:  validator.New(),
		log:       log,
	}
}


type createDatabaseRequest struct {
	Name               string               `json:"name"                validate:"required,min=1,max=255"`
	Description        string               `json:"description"`
	Type               database.DatabaseType `json:"type"               validate:"required,oneof=postgresql mysql"`
	Host               string               `json:"host"                validate:"required,max=255"`
	Port               int                  `json:"port"               validate:"required,min=1,max=65535"`
	DatabaseName       string               `json:"database_name"       validate:"required,max=255"`
	Username           string               `json:"username"            validate:"required,max=255"`
	Password           string               `json:"password"            validate:"required"`
	SSLMode            database.SSLMode     `json:"ssl_mode"            validate:"omitempty,oneof=disable require verify-ca verify-full"`
	MonitoringInterval int                  `json:"monitoring_interval" validate:"omitempty,min=5,max=3600"`
}

type updateDatabaseRequest struct {
	Name               string               `json:"name"                validate:"omitempty,min=1,max=255"`
	Description        string               `json:"description"`
	Host               string               `json:"host"                validate:"omitempty,max=255"`
	Port               int                  `json:"port"                validate:"omitempty,min=1,max=65535"`
	DatabaseName       string               `json:"database_name"       validate:"omitempty,max=255"`
	Username           string               `json:"username"            validate:"omitempty,max=255"`
	Password           string               `json:"password"`
	SSLMode            database.SSLMode     `json:"ssl_mode"            validate:"omitempty,oneof=disable require verify-ca verify-full"`
	MonitoringInterval int                  `json:"monitoring_interval" validate:"omitempty,min=5,max=3600"`
	IsMonitoringEnabled *bool               `json:"is_monitoring_enabled"`
}

type databaseResponse struct {
	ID                  string               `json:"id"`
	Name                string               `json:"name"`
	Description         string               `json:"description"`
	Type                database.DatabaseType `json:"type"`
	Host                string               `json:"host"`
	Port                int                  `json:"port"`
	DatabaseName        string               `json:"database_name"`
	Username            string               `json:"username"`
	SSLMode             database.SSLMode     `json:"ssl_mode"`
	MonitoringInterval  int                  `json:"monitoring_interval"`
	Status              database.Status      `json:"status"`
	IsMonitoringEnabled bool                 `json:"is_monitoring_enabled"`
	LastCheckedAt       *time.Time           `json:"last_checked_at,omitempty"`
	CreatedAt           time.Time            `json:"created_at"`
	UpdatedAt           time.Time            `json:"updated_at"`
}

func toDatabaseResponse(db *database.MonitoredDatabase) databaseResponse {
	return databaseResponse{
		ID:                  db.ID.String(),
		Name:                db.Name,
		Description:         db.Description,
		Type:                db.Type,
		Host:                db.Host,
		Port:                db.Port,
		DatabaseName:        db.DatabaseName,
		Username:            db.Username,
		SSLMode:             db.SSLMode,
		MonitoringInterval:  db.MonitoringInterval,
		Status:              db.Status,
		IsMonitoringEnabled: db.IsMonitoringEnabled,
		LastCheckedAt:       db.LastCheckedAt,
		CreatedAt:           db.CreatedAt,
		UpdatedAt:           db.UpdatedAt,
	}
}


func (h *DatabaseHandler) List(w http.ResponseWriter, r *http.Request) {
	dbs, err := h.dbSvc.List(r.Context())
	if err != nil {
		response.Error(w, r, err)
		return
	}

	resp := make([]databaseResponse, len(dbs))
	for i := range dbs {
		resp[i] = toDatabaseResponse(&dbs[i])
	}
	response.Success(w, r, resp)
}

func (h *DatabaseHandler) Create(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		response.Error(w, r, apperrors.NewUnauthorized())
		return
	}

	var req createDatabaseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, r, apperrors.NewValidation("invalid request body"))
		return
	}
	if err := h.validate.Struct(&req); err != nil {
		response.Error(w, r, apperrors.NewValidation(err.Error()))
		return
	}

	if req.SSLMode == "" {
		req.SSLMode = database.SSLModeDisable
	}
	if req.MonitoringInterval == 0 {
		req.MonitoringInterval = 15
	}

	encPwd, err := h.encryptor.Encrypt(req.Password)
	if err != nil {
		response.Error(w, r, apperrors.NewInternal(err))
		return
	}

	createdBy := mustParseUUID(claims.UserID)
	db := &database.MonitoredDatabase{
		Name:                req.Name,
		Description:         req.Description,
		Type:                req.Type,
		Host:                req.Host,
		Port:                req.Port,
		DatabaseName:        req.DatabaseName,
		Username:            req.Username,
		SSLMode:             req.SSLMode,
		MonitoringInterval:  req.MonitoringInterval,
		IsMonitoringEnabled: true,
		Status:              database.StatusActive,
		CreatedBy:           createdBy,
	}

	if err := h.dbSvc.Create(r.Context(), db, encPwd); err != nil {
		response.Error(w, r, err)
		return
	}

	h.auditSvc.LogAsync(&audit.AuditLog{
		UserID:       &createdBy,
		UserEmail:    claims.Email,
		Action:       audit.ActionCreateDB,
		ResourceType: "database",
		ResourceID:   db.ID.String(),
		ResourceName: db.Name,
		IPAddress:    clientIP(r),
		UserAgent:    r.UserAgent(),
		Status:       "success",
	})

	response.Created(w, r, toDatabaseResponse(db))
}

func (h *DatabaseHandler) Get(w http.ResponseWriter, r *http.Request) {
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

	response.Success(w, r, toDatabaseResponse(db))
}

func (h *DatabaseHandler) Update(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		response.Error(w, r, apperrors.NewUnauthorized())
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, r, apperrors.NewValidation("invalid database ID"))
		return
	}

	var req updateDatabaseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, r, apperrors.NewValidation("invalid request body"))
		return
	}
	if err := h.validate.Struct(&req); err != nil {
		response.Error(w, r, apperrors.NewValidation(err.Error()))
		return
	}

	existing, err := h.dbSvc.GetByID(r.Context(), id)
	if err != nil {
		response.Error(w, r, err)
		return
	}

	if req.Name != "" {
		existing.Name = req.Name
	}
	if req.Description != "" {
		existing.Description = req.Description
	}
	if req.Host != "" {
		existing.Host = req.Host
	}
	if req.Port != 0 {
		existing.Port = req.Port
	}
	if req.DatabaseName != "" {
		existing.DatabaseName = req.DatabaseName
	}
	if req.Username != "" {
		existing.Username = req.Username
	}
	if req.SSLMode != "" {
		existing.SSLMode = req.SSLMode
	}
	if req.MonitoringInterval != 0 {
		existing.MonitoringInterval = req.MonitoringInterval
	}
	if req.IsMonitoringEnabled != nil {
		existing.IsMonitoringEnabled = *req.IsMonitoringEnabled
	}

	var encPwd *string
	if req.Password != "" {
		ep, err := h.encryptor.Encrypt(req.Password)
		if err != nil {
			response.Error(w, r, apperrors.NewInternal(err))
			return
		}
		encPwd = &ep
	}

	if err := h.dbSvc.Update(r.Context(), existing, encPwd); err != nil {
		response.Error(w, r, err)
		return
	}

	userID := mustParseUUID(claims.UserID)
	h.auditSvc.LogAsync(&audit.AuditLog{
		UserID:       &userID,
		UserEmail:    claims.Email,
		Action:       audit.ActionUpdateDB,
		ResourceType: "database",
		ResourceID:   id.String(),
		ResourceName: existing.Name,
		IPAddress:    clientIP(r),
		UserAgent:    r.UserAgent(),
		Status:       "success",
	})

	response.Success(w, r, toDatabaseResponse(existing))
}

func (h *DatabaseHandler) Delete(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		response.Error(w, r, apperrors.NewUnauthorized())
		return
	}

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

	if err := h.dbSvc.Delete(r.Context(), id); err != nil {
		response.Error(w, r, err)
		return
	}

	userID := mustParseUUID(claims.UserID)
	h.auditSvc.LogAsync(&audit.AuditLog{
		UserID:       &userID,
		UserEmail:    claims.Email,
		Action:       audit.ActionDeleteDB,
		ResourceType: "database",
		ResourceID:   id.String(),
		ResourceName: db.Name,
		IPAddress:    clientIP(r),
		UserAgent:    r.UserAgent(),
		Status:       "success",
	})

	response.NoContent(w, r)
}

func (h *DatabaseHandler) ExecuteQuery(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, r, apperrors.NewValidation("invalid id"))
		return
	}

	var input struct {
		Query string `json:"query"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.Error(w, r, apperrors.NewValidation("invalid request body"))
		return
	}
	if input.Query == "" {
		response.Error(w, r, apperrors.NewValidation("query is required"))
		return
	}

	res, err := h.dbSvc.ExecuteQuery(r.Context(), id, input.Query)
	if err != nil {
		response.Error(w, r, err)
		return
	}

	response.Success(w, r, res)
}

func (h *DatabaseHandler) TestConnection(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, r, apperrors.NewValidation("invalid database ID"))
		return
	}

	db, encPwd, err := h.dbSvc.GetWithCredentials(r.Context(), id)
	if err != nil {
		response.Error(w, r, err)
		return
	}

	pwd, err := h.encryptor.Decrypt(encPwd)
	if err != nil {
		response.Error(w, r, apperrors.NewInternal(err))
		return
	}

	testErr := h.dbSvc.TestConnection(r.Context(), db, pwd)

	type testResult struct {
		Success bool   `json:"success"`
		Message string `json:"message"`
	}

	claims, _ := middleware.GetClaims(r.Context())
	if claims != nil {
		userID := mustParseUUID(claims.UserID)
		status := "success"
		if testErr != nil {
			status = "failure"
		}
		h.auditSvc.LogAsync(&audit.AuditLog{
			UserID:       &userID,
			UserEmail:    claims.Email,
			Action:       audit.ActionTestConnDB,
			ResourceType: "database",
			ResourceID:   id.String(),
			ResourceName: db.Name,
			IPAddress:    clientIP(r),
			UserAgent:    r.UserAgent(),
			Status:       status,
		})
	}

	if testErr != nil {
		response.Success(w, r, testResult{
			Success: false,
			Message: "Connection failed: " + safeErrorMessage(testErr),
		})
		return
	}

	response.Success(w, r, testResult{
		Success: true,
		Message: "Connection successful",
	})
}

func safeErrorMessage(err error) string {
	if err == nil {
		return ""
	}
	msg := err.Error()
	if len(msg) > 200 {
		msg = msg[:200] + "..."
	}
	return msg
}
