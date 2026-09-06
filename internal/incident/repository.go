package incident

import (
	"context"

	"github.com/dbplatform/api/internal/errors"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, inc *Incident) error {
	query := `
		INSERT INTO incidents (database_id, title, description, status)
		VALUES ($1, $2, $3, $4)
		RETURNING id, started_at
	`
	return r.db.QueryRow(ctx, query, inc.DatabaseID, inc.Title, inc.Description, inc.Status).
		Scan(&inc.ID, &inc.StartedAt)
}

func (r *repository) GetOpenByDatabase(ctx context.Context, dbID uuid.UUID) (*Incident, error) {
	query := `
		SELECT id, database_id, title, description, status, started_at, acknowledged_at, resolved_at
		FROM incidents
		WHERE database_id = $1 AND status != 'resolved'
		ORDER BY started_at DESC LIMIT 1
	`
	var inc Incident
	err := r.db.QueryRow(ctx, query, dbID).Scan(
		&inc.ID, &inc.DatabaseID, &inc.Title, &inc.Description,
		&inc.Status, &inc.StartedAt, &inc.AcknowledgedAt, &inc.ResolvedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, errors.NewNotFound("incident")
		}
		return nil, err
	}
	return &inc, nil
}

func (r *repository) UpdateStatus(ctx context.Context, id uuid.UUID, status Status) error {
	query := `UPDATE incidents SET status = $1`
	
	if status == StatusResolved {
		query += `, resolved_at = NOW()`
	} else if status == StatusAcknowledged {
		query += `, acknowledged_at = NOW()`
	}
	
	query += ` WHERE id = $2`
	
	res, err := r.db.Exec(ctx, query, status, id)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return errors.NewNotFound("incident")
	}
	return nil
}

func (r *repository) List(ctx context.Context, limit, offset int) ([]Incident, error) {
	query := `
		SELECT id, database_id, title, description, status, started_at, acknowledged_at, resolved_at
		FROM incidents
		ORDER BY started_at DESC
		LIMIT $1 OFFSET $2
	`
	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var incidents []Incident
	for rows.Next() {
		var inc Incident
		if err := rows.Scan(
			&inc.ID, &inc.DatabaseID, &inc.Title, &inc.Description,
			&inc.Status, &inc.StartedAt, &inc.AcknowledgedAt, &inc.ResolvedAt,
		); err != nil {
			return nil, err
		}
		incidents = append(incidents, inc)
	}
	return incidents, nil
}

func (r *repository) CreateChannel(ctx context.Context, channel *NotificationChannel) error {
	query := `
		INSERT INTO notification_channels (name, type, config)
		VALUES ($1, $2, $3)
		RETURNING id, created_at, updated_at
	`
	return r.db.QueryRow(ctx, query, channel.Name, channel.Type, channel.Config).
		Scan(&channel.ID, &channel.CreatedAt, &channel.UpdatedAt)
}

func (r *repository) ListChannels(ctx context.Context) ([]NotificationChannel, error) {
	query := `SELECT id, name, type, config, is_active, created_at, updated_at FROM notification_channels ORDER BY created_at DESC`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var channels []NotificationChannel
	for rows.Next() {
		var ch NotificationChannel
		if err := rows.Scan(&ch.ID, &ch.Name, &ch.Type, &ch.Config, &ch.IsActive, &ch.CreatedAt, &ch.UpdatedAt); err != nil {
			return nil, err
		}
		channels = append(channels, ch)
	}
	return channels, nil
}

func (r *repository) DeleteChannel(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM notification_channels WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}
