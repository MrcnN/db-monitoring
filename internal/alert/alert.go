package alert

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Severity string
type Status string

const (
	SeverityWarning  Severity = "warning"
	SeverityCritical Severity = "critical"

	StatusActive   Status = "active"
	StatusResolved Status = "resolved"
)

type Alert struct {
	ID           uuid.UUID `json:"id"`
	DatabaseID   uuid.UUID `json:"database_id"`
	Title        string    `json:"title"`
	Description  string    `json:"description"`
	Severity     Severity  `json:"severity"`
	Status       Status    `json:"status"`
	MetricName   string    `json:"metric_name,omitempty"`
	CurrentValue string    `json:"current_value,omitempty"`
	ResolvedAt   *time.Time `json:"resolved_at,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}

type Repository interface {
	Create(ctx context.Context, alert *Alert) error
	ListByDatabaseID(ctx context.Context, dbID uuid.UUID, limit, offset int) ([]Alert, error)
	ListActive(ctx context.Context, limit, offset int) ([]Alert, error)
	ResolveAlertsByDatabase(ctx context.Context, dbID uuid.UUID, exceptTitles []string) error
}
