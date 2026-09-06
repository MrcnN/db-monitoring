package incident

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Status string
type ChannelType string

const (
	StatusOpen         Status = "open"
	StatusAcknowledged Status = "acknowledged"
	StatusResolved     Status = "resolved"

	ChannelSlack   ChannelType = "slack"
	ChannelWebhook ChannelType = "webhook"
	ChannelEmail   ChannelType = "email"
)

type Incident struct {
	ID             uuid.UUID  `json:"id"`
	DatabaseID     uuid.UUID  `json:"database_id"`
	Title          string     `json:"title"`
	Description    string     `json:"description"`
	Status         Status     `json:"status"`
	StartedAt      time.Time  `json:"started_at"`
	AcknowledgedAt *time.Time `json:"acknowledged_at,omitempty"`
	ResolvedAt     *time.Time `json:"resolved_at,omitempty"`
}

type NotificationChannel struct {
	ID        uuid.UUID              `json:"id"`
	Name      string                 `json:"name"`
	Type      ChannelType            `json:"type"`
	Config    map[string]interface{} `json:"config"`
	IsActive  bool                   `json:"is_active"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
}

type Repository interface {
	Create(ctx context.Context, incident *Incident) error
	GetOpenByDatabase(ctx context.Context, dbID uuid.UUID) (*Incident, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status Status) error
	List(ctx context.Context, limit, offset int) ([]Incident, error)

	CreateChannel(ctx context.Context, channel *NotificationChannel) error
	ListChannels(ctx context.Context) ([]NotificationChannel, error)
	DeleteChannel(ctx context.Context, id uuid.UUID) error
}
