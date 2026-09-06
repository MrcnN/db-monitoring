package alert

import (
	"context"

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

func (r *repository) Create(ctx context.Context, alert *Alert) error {
	query := `
		INSERT INTO alerts (database_id, title, description, severity, status, metric_name, current_value)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at
	`
	return r.db.QueryRow(ctx, query,
		alert.DatabaseID,
		alert.Title,
		alert.Description,
		alert.Severity,
		alert.Status,
		alert.MetricName,
		alert.CurrentValue,
	).Scan(&alert.ID, &alert.CreatedAt)
}

func (r *repository) ListByDatabaseID(ctx context.Context, dbID uuid.UUID, limit, offset int) ([]Alert, error) {
	query := `
		SELECT id, database_id, title, description, severity, status, metric_name, current_value, resolved_at, created_at
		FROM alerts
		WHERE database_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.Query(ctx, query, dbID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanAlerts(rows)
}

func (r *repository) ListActive(ctx context.Context, limit, offset int) ([]Alert, error) {
	query := `
		SELECT id, database_id, title, description, severity, status, metric_name, current_value, resolved_at, created_at
		FROM alerts
		WHERE status = 'active'
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`
	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanAlerts(rows)
}

func (r *repository) ResolveAlertsByDatabase(ctx context.Context, dbID uuid.UUID, exceptTitles []string) error {
	if len(exceptTitles) == 0 {
		query := `UPDATE alerts SET status = 'resolved', resolved_at = NOW() WHERE database_id = $1 AND status = 'active'`
		_, err := r.db.Exec(ctx, query, dbID)
		return err
	}

	query := `
		UPDATE alerts 
		SET status = 'resolved', resolved_at = NOW() 
		WHERE database_id = $1 AND status = 'active' AND title != ALL($2)
	`
	_, err := r.db.Exec(ctx, query, dbID, exceptTitles)
	return err
}

func (r *repository) scanAlerts(rows pgx.Rows) ([]Alert, error) {
	var alerts []Alert
	for rows.Next() {
		var a Alert
		err := rows.Scan(
			&a.ID,
			&a.DatabaseID,
			&a.Title,
			&a.Description,
			&a.Severity,
			&a.Status,
			&a.MetricName,
			&a.CurrentValue,
			&a.ResolvedAt,
			&a.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		alerts = append(alerts, a)
	}
	return alerts, nil
}
