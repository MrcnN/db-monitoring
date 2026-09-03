package handler

import (
	"net/http"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/dbplatform/api/internal/config"
	"github.com/dbplatform/api/internal/api/response"
)

type HealthHandler struct {
	db    *pgxpool.Pool
	redis *redis.Client
	cfg   *config.Config
}

func NewHealthHandler(db *pgxpool.Pool, redis *redis.Client, cfg *config.Config) *HealthHandler {
	return &HealthHandler{db: db, redis: redis, cfg: cfg}
}

func (h *HealthHandler) Liveness(w http.ResponseWriter, r *http.Request) {
	response.JSON(w, http.StatusOK, map[string]string{"status": "alive"})
}

func (h *HealthHandler) Readiness(w http.ResponseWriter, r *http.Request) {
	dbStatus := "ok"
	if err := h.db.Ping(r.Context()); err != nil {
		dbStatus = "error"
	}
	redisStatus := "ok"
	if err := h.redis.Ping(r.Context()).Err(); err != nil {
		redisStatus = "error"
	}

	resp := map[string]interface{}{
		"status":  "healthy",
		"version": h.cfg.App.Version,
		"checks": map[string]string{
			"database": dbStatus,
			"redis":    redisStatus,
		},
	}
	status := http.StatusOK
	if dbStatus == "error" || redisStatus == "error" {
		status = http.StatusServiceUnavailable
		resp["status"] = "unhealthy"
	}
	response.JSON(w, status, resp)
}
