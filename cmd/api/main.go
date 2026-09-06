package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog"

	mainapi "github.com/dbplatform/api/internal/api"
	"github.com/dbplatform/api/internal/api/handler"
	"github.com/dbplatform/api/internal/api/middleware"
	"github.com/dbplatform/api/internal/api/ws"
	"github.com/dbplatform/api/internal/audit"
	"github.com/dbplatform/api/internal/auth"
	"github.com/dbplatform/api/internal/config"
	"github.com/dbplatform/api/internal/crypto"
	"github.com/dbplatform/api/internal/database"
	"github.com/dbplatform/api/internal/health"
	"github.com/dbplatform/api/internal/metrics"
	"github.com/dbplatform/api/internal/observability"
	"github.com/dbplatform/api/internal/platform"
	"github.com/dbplatform/api/internal/user"
	"github.com/dbplatform/api/internal/worker"
)

var version = "0.1.0" // overridden by ldflags at build time

func main() {
	_ = godotenv.Load()

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "FATAL: config error: %v\n", err)
		os.Exit(1)
	}

	logger := setupLogger(cfg)
	logger.Info().
		Str("version", version).
		Str("env", cfg.App.Env).
		Str("port", cfg.Server.Port).
		Msg("Starting DB Health Platform API")

	observability.Register()

	ctx := context.Background()

	pool, err := platform.ConnectPostgres(ctx, cfg.Database.URL)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to connect to platform PostgreSQL")
	}
	defer pool.Close()
	logger.Info().Msg("Connected to platform PostgreSQL")

	if err := platform.RunMigrations(cfg.Database.URL, cfg.Database.MigrationsPath); err != nil {
		logger.Fatal().Err(err).Msg("Database migration failed")
	}
	logger.Info().Msg("Database migrations applied")

	redisClient, err := platform.ConnectRedis(ctx, cfg.Redis.URL, cfg.Redis.Password, cfg.Redis.DB)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to connect to Redis")
	}
	defer redisClient.Close()
	logger.Info().Msg("Connected to Redis")

	encryptor, err := crypto.NewEncryptor(cfg.Crypto.EncryptionKey)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to initialize encryptor")
	}

	userRepo := user.NewRepository(pool)
	sessionRepo := user.NewSessionRepository(pool)
	dbRepo := database.NewRepository(pool)
	auditRepo := audit.NewRepository(pool)
	metricsRepo := metrics.NewRepository(pool)

	jwtSvc := auth.NewJWTService(cfg.JWT)
	authSvc := auth.NewService(userRepo, sessionRepo, jwtSvc, cfg.JWT.RefreshTokenTTL)
	dbSvc := database.NewService(dbRepo, encryptor)
	auditSvc := audit.NewService(auditRepo)
	metricsSvc := metrics.NewService(metricsRepo)
	healthEvaluator := health.NewEvaluator()

	workerCtx, cancelWorker := context.WithCancel(context.Background())
	wsHub := ws.NewHub(logger)
	go wsHub.Run(workerCtx)

	collectorWorker := worker.NewCollectorWorker(
		dbSvc,
		metricsSvc,
		healthEvaluator,
		encryptor,
		wsHub,
		15*time.Second,
		logger,
	)
	go collectorWorker.Start(workerCtx)

	healthHandler := handler.NewHealthHandler(pool, redisClient, cfg)
	authHandler := handler.NewAuthHandler(authSvc, auditSvc, logger)
	dbHandler := handler.NewDatabaseHandler(dbSvc, auditSvc, encryptor, logger)
	auditHandler := handler.NewAuditLogHandler(auditRepo, logger)
	metricsHandler := handler.NewMetricsHandler(metricsSvc, dbSvc, healthEvaluator, encryptor, wsHub, logger)

	rateLimiter := middleware.NewRateLimiter(cfg.Server.RateLimit, cfg.Server.RateBurst)

	router := mainapi.NewRouter(mainapi.RouterConfig{
		HealthHandler:   healthHandler,
		AuthHandler:     authHandler,
		DatabaseHandler: dbHandler,
		AuditHandler:    auditHandler,
		MetricsHandler:  metricsHandler,
		JWTService:      jwtSvc,
		RateLimiter:     rateLimiter,
		Log:             logger,
	})

	srv := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  60 * time.Second,
	}

	serverErrors := make(chan error, 1)
	go func() {
		logger.Info().Str("addr", srv.Addr).Msg("HTTP server listening")
		serverErrors <- srv.ListenAndServe()
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		logger.Fatal().Err(err).Msg("Server error")
	case sig := <-quit:
		logger.Info().Str("signal", sig.String()).Msg("Shutdown signal received")
	}

	logger.Info().Msg("Shutting down server and background workers...")
	cancelWorker()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error().Err(err).Msg("Server shutdown error")
	}

	logger.Info().Msg("Server stopped gracefully")
}

func setupLogger(cfg *config.Config) zerolog.Logger {
	level, err := zerolog.ParseLevel(cfg.Log.Level)
	if err != nil {
		level = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(level)

	if cfg.App.Env == "development" {
		return zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339}).
			With().Timestamp().
			Str("service", cfg.App.Name).
			Str("version", version).
			Logger()
	}

	return zerolog.New(os.Stderr).
		With().Timestamp().
		Str("service", cfg.App.Name).
		Str("version", version).
		Logger()
}



