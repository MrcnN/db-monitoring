package config

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
