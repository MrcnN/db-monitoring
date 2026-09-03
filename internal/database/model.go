package database

import (
	"time"
	"github.com/google/uuid"
)

type DatabaseType string
const (
	TypePostgreSQL DatabaseType = "postgresql"
	TypeMySQL      DatabaseType = "mysql"
)

type SSLMode string
const (
	SSLModeDisable    SSLMode = "disable"
	SSLModeRequire    SSLMode = "require"
	SSLModeVerifyCA   SSLMode = "verify-ca"
	SSLModeVerifyFull SSLMode = "verify-full"
)

type Status string
const (
	StatusActive   Status = "active"
	StatusInactive Status = "inactive"
	StatusError    Status = "error"
)

type MonitoredDatabase struct {
	ID                   uuid.UUID
	Name                 string
	Description          string
	Type                 DatabaseType
	Host                 string
	Port                 int
	DatabaseName         string
	Username             string
	SSLMode              SSLMode
	MonitoringInterval   int
	Status               Status
	IsMonitoringEnabled  bool
	LastCheckedAt        *time.Time
	CreatedBy            uuid.UUID
	CreatedAt            time.Time
	UpdatedAt            time.Time
	DeletedAt            *time.Time
}

type CreateDatabaseRequest struct {
	Name               string       `json:"name" validate:"required,min=1,max=255"`
	Description        string       `json:"description"`
	Type               DatabaseType `json:"type" validate:"required,oneof=postgresql mysql"`
	Host               string       `json:"host" validate:"required"`
	Port               int          `json:"port" validate:"required,min=1,max=65535"`
	DatabaseName       string       `json:"database_name" validate:"required"`
	Username           string       `json:"username" validate:"required"`
	Password           string       `json:"password" validate:"required"`
	SSLMode            SSLMode      `json:"ssl_mode" validate:"oneof=disable require verify-ca verify-full"`
	MonitoringInterval int          `json:"monitoring_interval" validate:"min=5,max=3600"`
}

type UpdateDatabaseRequest struct {
	Name                string       `json:"name" validate:"omitempty,min=1,max=255"`
	Description         string       `json:"description"`
	Host                string       `json:"host"`
	Port                int          `json:"port" validate:"omitempty,min=1,max=65535"`
	DatabaseName        string       `json:"database_name"`
	Username            string       `json:"username"`
	Password            string       `json:"password"`
	SSLMode             SSLMode      `json:"ssl_mode" validate:"omitempty,oneof=disable require verify-ca verify-full"`
	MonitoringInterval  int          `json:"monitoring_interval" validate:"omitempty,min=5,max=3600"`
	IsMonitoringEnabled *bool        `json:"is_monitoring_enabled"`
}

type DatabaseResponse struct {
	ID                   uuid.UUID    `json:"id"`
	Name                 string       `json:"name"`
	Description          string       `json:"description"`
	Type                 DatabaseType `json:"type"`
	Host                 string       `json:"host"`
	Port                 int          `json:"port"`
	DatabaseName         string       `json:"database_name"`
	Username             string       `json:"username"`
	SSLMode              SSLMode      `json:"ssl_mode"`
	MonitoringInterval   int          `json:"monitoring_interval"`
	Status               Status       `json:"status"`
	IsMonitoringEnabled  bool         `json:"is_monitoring_enabled"`
	LastCheckedAt        *time.Time   `json:"last_checked_at"`
	CreatedBy            uuid.UUID    `json:"created_by"`
	CreatedAt            time.Time    `json:"created_at"`
	UpdatedAt            time.Time    `json:"updated_at"`
}
