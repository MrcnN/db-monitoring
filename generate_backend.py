import os
import sys

base_dir = r"c:\Users\Erdi\Desktop\e\projeler\new"

files = {
    "go.mod": """module github.com/dbplatform/api

go 1.23

require (
\tgithub.com/go-chi/chi/v5 v5.2.1
\tgithub.com/go-chi/cors v1.2.1
\tgithub.com/go-playground/validator/v10 v10.24.0
\tgithub.com/go-sql-driver/mysql v1.8.1
\tgithub.com/golang-jwt/jwt/v5 v5.2.2
\tgithub.com/golang-migrate/migrate/v4 v4.18.2
\tgithub.com/google/uuid v1.6.0
\tgithub.com/jackc/pgx/v5 v5.7.2
\tgithub.com/joho/godotenv v1.5.1
\tgithub.com/prometheus/client_golang v1.21.1
\tgithub.com/redis/go-redis/v9 v9.7.3
\tgithub.com/rs/zerolog v1.33.0
\tgithub.com/testcontainers/testcontainers-go v0.35.0
\tgithub.com/testcontainers/testcontainers-go/modules/postgres v0.35.0
\tgithub.com/testcontainers/testcontainers-go/modules/redis v0.35.0
\tgolang.org/x/crypto v0.36.0
\tgolang.org/x/time v0.11.0
)
""",
    "migrations/000001_create_users.up.sql": """CREATE TYPE user_role AS ENUM ('admin', 'operator', 'viewer');

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    full_name VARCHAR(255) NOT NULL,
    role user_role NOT NULL DEFAULT 'viewer',
    is_active BOOLEAN NOT NULL DEFAULT true,
    last_login_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_users_email ON users(email) WHERE deleted_at IS NULL;
CREATE INDEX idx_users_role ON users(role) WHERE deleted_at IS NULL;
""",
    "migrations/000001_create_users.down.sql": """DROP TABLE users; DROP TYPE user_role;""",
    "migrations/000002_create_sessions.up.sql": """CREATE TABLE user_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    refresh_token_hash VARCHAR(255) NOT NULL UNIQUE,
    user_agent TEXT,
    ip_address VARCHAR(45),
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    revoked_at TIMESTAMPTZ
);

CREATE INDEX idx_sessions_user_id ON user_sessions(user_id);
CREATE INDEX idx_sessions_refresh_token ON user_sessions(refresh_token_hash);
""",
    "migrations/000002_create_sessions.down.sql": """DROP TABLE user_sessions;""",
    "migrations/000003_create_monitored_databases.up.sql": """CREATE TYPE database_type AS ENUM ('postgresql', 'mysql');
CREATE TYPE database_status AS ENUM ('active', 'inactive', 'error');
CREATE TYPE ssl_mode AS ENUM ('disable', 'require', 'verify-ca', 'verify-full');

CREATE TABLE monitored_databases (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    type database_type NOT NULL,
    host VARCHAR(255) NOT NULL,
    port INTEGER NOT NULL,
    database_name VARCHAR(255) NOT NULL,
    username VARCHAR(255) NOT NULL,
    ssl_mode ssl_mode NOT NULL DEFAULT 'disable',
    monitoring_interval INTEGER NOT NULL DEFAULT 15,
    status database_status NOT NULL DEFAULT 'active',
    is_monitoring_enabled BOOLEAN NOT NULL DEFAULT true,
    last_checked_at TIMESTAMPTZ,
    created_by UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE TABLE database_credentials (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    database_id UUID NOT NULL UNIQUE REFERENCES monitored_databases(id) ON DELETE CASCADE,
    encrypted_password TEXT NOT NULL,
    encryption_key_id VARCHAR(50) NOT NULL DEFAULT 'v1',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_monitored_databases_status ON monitored_databases(status) WHERE deleted_at IS NULL;
CREATE INDEX idx_monitored_databases_type ON monitored_databases(type) WHERE deleted_at IS NULL;
""",
    "migrations/000003_create_monitored_databases.down.sql": """DROP TABLE database_credentials; DROP TABLE monitored_databases; DROP TYPE ssl_mode; DROP TYPE database_status; DROP TYPE database_type;""",
    "migrations/000004_create_audit_logs.up.sql": """CREATE TABLE audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id),
    user_email VARCHAR(255),
    action VARCHAR(100) NOT NULL,
    resource_type VARCHAR(100),
    resource_id VARCHAR(255),
    resource_name VARCHAR(255),
    details JSONB,
    ip_address VARCHAR(45),
    user_agent TEXT,
    status VARCHAR(20) NOT NULL DEFAULT 'success',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_audit_logs_user_id ON audit_logs(user_id);
CREATE INDEX idx_audit_logs_action ON audit_logs(action);
CREATE INDEX idx_audit_logs_created_at ON audit_logs(created_at DESC);
CREATE INDEX idx_audit_logs_resource ON audit_logs(resource_type, resource_id);
""",
    "migrations/000004_create_audit_logs.down.sql": """DROP TABLE audit_logs;""",
    "internal/config/config.go": """package config

import (
    "fmt"
    "os"
    "strconv"
    "time"
)

type Config struct {
    App      AppConfig
    Database DatabaseConfig
    Redis    RedisConfig
    JWT      JWTConfig
    Crypto   CryptoConfig
    Log      LogConfig
    Server   ServerConfig
}

type AppConfig struct {
    Env     string // development, production, test
    Name    string
    Version string
}

type DatabaseConfig struct {
    URL             string
    MaxOpenConns    int
    MaxIdleConns    int
    ConnMaxLifetime time.Duration
    MigrationsPath  string
}

type RedisConfig struct {
    URL      string
    Password string
    DB       int
}

type JWTConfig struct {
    Secret            string
    AccessTokenTTL    time.Duration
    RefreshTokenTTL   time.Duration
}

type CryptoConfig struct {
    EncryptionKey string // 32-byte hex string
}

type LogConfig struct {
    Level string
}

type ServerConfig struct {
    Port            string
    ReadTimeout     time.Duration
    WriteTimeout    time.Duration
    ShutdownTimeout time.Duration
    RateLimit       float64
    RateBurst       int
}

func Load() (*Config, error) {
    cfg := &Config{}
    
    cfg.App.Env = getEnv("APP_ENV", "development")
    cfg.App.Name = getEnv("APP_NAME", "dbplatform")
    cfg.App.Version = getEnv("APP_VERSION", "0.1.0")
    
    cfg.Database.URL = getEnv("DATABASE_URL", "")
    if cfg.Database.URL == "" {
        return nil, fmt.Errorf("DATABASE_URL is required")
    }
    cfg.Database.MaxOpenConns = getEnvInt("DB_MAX_OPEN_CONNS", 25)
    cfg.Database.MaxIdleConns = getEnvInt("DB_MAX_IDLE_CONNS", 5)
    cfg.Database.ConnMaxLifetime = getEnvDuration("DB_CONN_MAX_LIFETIME", 5*time.Minute)
    cfg.Database.MigrationsPath = getEnv("MIGRATIONS_PATH", "migrations")
    
    cfg.Redis.URL = getEnv("REDIS_URL", "redis://localhost:6379")
    cfg.Redis.Password = getEnv("REDIS_PASSWORD", "")
    cfg.Redis.DB = getEnvInt("REDIS_DB", 0)
    
    cfg.JWT.Secret = getEnv("JWT_SECRET", "")
    if cfg.JWT.Secret == "" {
        return nil, fmt.Errorf("JWT_SECRET is required")
    }
    if len(cfg.JWT.Secret) < 32 {
        return nil, fmt.Errorf("JWT_SECRET must be at least 32 characters")
    }
    cfg.JWT.AccessTokenTTL = getEnvDuration("JWT_ACCESS_TTL", 15*time.Minute)
    cfg.JWT.RefreshTokenTTL = getEnvDuration("JWT_REFRESH_TTL", 7*24*time.Hour)
    
    cfg.Crypto.EncryptionKey = getEnv("ENCRYPTION_KEY", "")
    if cfg.Crypto.EncryptionKey == "" {
        return nil, fmt.Errorf("ENCRYPTION_KEY is required")
    }
    if len(cfg.Crypto.EncryptionKey) != 64 {
        return nil, fmt.Errorf("ENCRYPTION_KEY must be 64 hex characters (32 bytes)")
    }
    
    cfg.Log.Level = getEnv("LOG_LEVEL", "info")
    
    cfg.Server.Port = getEnv("APP_PORT", "8080")
    cfg.Server.ReadTimeout = getEnvDuration("SERVER_READ_TIMEOUT", 30*time.Second)
    cfg.Server.WriteTimeout = getEnvDuration("SERVER_WRITE_TIMEOUT", 30*time.Second)
    cfg.Server.ShutdownTimeout = getEnvDuration("SERVER_SHUTDOWN_TIMEOUT", 15*time.Second)
    cfg.Server.RateLimit = getEnvFloat("RATE_LIMIT", 100)
    cfg.Server.RateBurst = getEnvInt("RATE_BURST", 200)
    
    return cfg, nil
}

func getEnv(key, defaultVal string) string {
    if v := os.Getenv(key); v != "" {
        return v
    }
    return defaultVal
}

func getEnvInt(key string, defaultVal int) int {
    if v := os.Getenv(key); v != "" {
        if i, err := strconv.Atoi(v); err == nil {
            return i
        }
    }
    return defaultVal
}

func getEnvFloat(key string, defaultVal float64) float64 {
    if v := os.Getenv(key); v != "" {
        if f, err := strconv.ParseFloat(v, 64); err == nil {
            return f
        }
    }
    return defaultVal
}

func getEnvDuration(key string, defaultVal time.Duration) time.Duration {
    if v := os.Getenv(key); v != "" {
        if d, err := time.ParseDuration(v); err == nil {
            return d
        }
    }
    return defaultVal
}
""",
    "cmd/api/main.go": """package main

import (
\t"context"
\t"fmt"
\t"net/http"
\t"os"
\t"os/signal"
\t"syscall"
\t"time"

\t"github.com/dbplatform/api/internal/config"
\t"github.com/joho/godotenv"
\t"github.com/rs/zerolog"
\t"github.com/rs/zerolog/log"
)

func main() {
\t_ = godotenv.Load()

\tcfg, err := config.Load()
\tif err != nil {
\t\tfmt.Printf("Failed to load config: %v\\n", err)
\t\tos.Exit(1)
\t}

\tsetupLogger(cfg)
\tlog.Info().Str("version", cfg.App.Version).Str("env", cfg.App.Env).Msg("Starting application")

\tctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
\tdefer cancel()

\t// TODO: Initialize DB, Redis, Repositories, Services, Handlers, Router here
\tsrv := &http.Server{
\t\tAddr:         ":" + cfg.Server.Port,
\t\tReadTimeout:  cfg.Server.ReadTimeout,
\t\tWriteTimeout: cfg.Server.WriteTimeout,
\t}

\tgo func() {
\t\tlog.Info().Str("port", cfg.Server.Port).Msg("Server listening")
\t\tif err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
\t\t\tlog.Fatal().Err(err).Msg("Server failed")
\t\t}
\t}()

\t<-ctx.Done()
\tlog.Info().Msg("Shutting down gracefully...")

\tshutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
\tdefer shutdownCancel()

\tif err := srv.Shutdown(shutdownCtx); err != nil {
\t\tlog.Error().Err(err).Msg("Server shutdown error")
\t}
}

func setupLogger(cfg *config.Config) {
\tlevel, err := zerolog.ParseLevel(cfg.Log.Level)
\tif err != nil {
\t\tlevel = zerolog.InfoLevel
\t}
\tzerolog.SetGlobalLevel(level)

\tif cfg.App.Env == "development" {
\t\tlog.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339})
\t}
}
"""
}

for path, content in files.items():
    full_path = os.path.join(base_dir, path)
    os.makedirs(os.path.dirname(full_path), exist_ok=True)
    with open(full_path, "w", encoding="utf-8") as f:
        f.write(content)

print("Created essential base files")
