package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"

	"github.com/dbplatform/api/internal/api/handler"
	"github.com/dbplatform/api/internal/api/middleware"
	"github.com/dbplatform/api/internal/auth"
	"github.com/dbplatform/api/internal/user"
)

type RouterConfig struct {
	HealthHandler   *handler.HealthHandler
	AuthHandler     *handler.AuthHandler
	DatabaseHandler *handler.DatabaseHandler
	AuditHandler    *handler.AuditLogHandler
	MetricsHandler  *handler.MetricsHandler
	JWTService      *auth.JWTService
	RateLimiter     *middleware.RateLimiter
	Log             zerolog.Logger
	AllowedOrigins  []string
}

func NewRouter(cfg RouterConfig) http.Handler {
	r := chi.NewRouter()


	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(chimiddleware.Recoverer)

	origins := cfg.AllowedOrigins
	if len(origins) == 0 {
		origins = []string{"http://localhost:3000", "http://localhost:5173"}
	}
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   origins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Request-ID"},
		ExposedHeaders:   []string{"X-Request-ID"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	if cfg.RateLimiter != nil {
		r.Use(cfg.RateLimiter.Limit())
	}


	r.Get("/health", cfg.HealthHandler.Liveness)
	r.Get("/ready", cfg.HealthHandler.Readiness)
	r.Handle("/metrics", promhttp.Handler())


	r.Route("/api/v1", func(r chi.Router) {

		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", cfg.AuthHandler.Register)
			r.Post("/login", cfg.AuthHandler.Login)
			r.Post("/refresh", cfg.AuthHandler.Refresh)
		})

		r.Group(func(r chi.Router) {
			r.Use(middleware.Authenticate(cfg.JWTService))

			r.Post("/auth/logout", cfg.AuthHandler.Logout)

			r.Get("/users/me", cfg.AuthHandler.Me)

			r.Get("/databases", cfg.DatabaseHandler.List)
			r.Get("/databases/{id}", cfg.DatabaseHandler.Get)

			r.With(middleware.RequireRole(user.RoleAdmin, user.RoleOperator)).Post("/databases", cfg.DatabaseHandler.Create)
			r.With(middleware.RequireRole(user.RoleAdmin, user.RoleOperator)).Put("/databases/{id}", cfg.DatabaseHandler.Update)
			r.With(middleware.RequireRole(user.RoleAdmin, user.RoleOperator)).Delete("/databases/{id}", cfg.DatabaseHandler.Delete)
			r.Post("/databases/{id}/test-connection", cfg.DatabaseHandler.TestConnection)

			r.Get("/databases/{id}/metrics", cfg.MetricsHandler.GetTimeSeries)
			r.Get("/databases/{id}/metrics/latest", cfg.MetricsHandler.GetLatest)
			r.Get("/databases/{id}/health", cfg.MetricsHandler.GetHealth)

			r.With(middleware.RequireRole(user.RoleAdmin)).Get("/audit-logs", cfg.AuditHandler.List)
		})
	})

	return r
}
