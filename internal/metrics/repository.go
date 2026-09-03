package metrics

import (
	"context"
	"time"

	apperrors "github.com/dbplatform/api/internal/errors"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	Create(ctx context.Context, m *Metric) error
	GetLatest(ctx context.Context, databaseID uuid.UUID) (*Metric, error)
	GetTimeSeries(ctx context.Context, databaseID uuid.UUID, since time.Time, limit int) ([]Metric, error)
}

type postgresRepository struct {
	pool *pgxpool.Pool
}

var _ Repository = (*postgresRepository)(nil)

func NewRepository(pool *pgxpool.Pool) Repository {
	return &postgresRepository{pool: pool}
}

func (r *postgresRepository) Create(ctx context.Context, m *Metric) error {
	query := `
		INSERT INTO metrics (
			database_id, cpu_usage, memory_usage,
			connections_total, connections_active, connections_idle, connection_usage_pct,
			query_rate, p95_latency_ms, cache_hit_ratio,
			database_size_bytes, dead_tuples_count, active_locks_count,
			replication_lag_seconds, created_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15
		) RETURNING id`

	if m.CreatedAt.IsZero() {
		m.CreatedAt = time.Now()
	}

	return r.pool.QueryRow(ctx, query,
		m.DatabaseID, m.CPUUsage, m.MemoryUsage,
		m.ConnectionsTotal, m.ConnectionsActive, m.ConnectionsIdle, m.ConnectionUsagePct,
		m.QueryRate, m.P95LatencyMs, m.CacheHitRatio,
		m.DatabaseSizeBytes, m.DeadTuplesCount, m.ActiveLocksCount,
		m.ReplicationLagSeconds, m.CreatedAt,
	).Scan(&m.ID)
}

func (r *postgresRepository) GetLatest(ctx context.Context, databaseID uuid.UUID) (*Metric, error) {
	query := `
		SELECT id, database_id, cpu_usage, memory_usage,
		       connections_total, connections_active, connections_idle, connection_usage_pct,
		       query_rate, p95_latency_ms, cache_hit_ratio,
		       database_size_bytes, dead_tuples_count, active_locks_count,
		       replication_lag_seconds, created_at
		FROM metrics
		WHERE database_id = $1
		ORDER BY created_at DESC
		LIMIT 1`

	var m Metric
	err := r.pool.QueryRow(ctx, query, databaseID).Scan(
		&m.ID, &m.DatabaseID, &m.CPUUsage, &m.MemoryUsage,
		&m.ConnectionsTotal, &m.ConnectionsActive, &m.ConnectionsIdle, &m.ConnectionUsagePct,
		&m.QueryRate, &m.P95LatencyMs, &m.CacheHitRatio,
		&m.DatabaseSizeBytes, &m.DeadTuplesCount, &m.ActiveLocksCount,
		&m.ReplicationLagSeconds, &m.CreatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, apperrors.NewNotFound("Metric")
		}
		return nil, err
	}
	return &m, nil
}

func (r *postgresRepository) GetTimeSeries(ctx context.Context, databaseID uuid.UUID, since time.Time, limit int) ([]Metric, error) {
	if limit <= 0 || limit > 1000 {
		limit = 300
	}

	query := `
		SELECT id, database_id, cpu_usage, memory_usage,
		       connections_total, connections_active, connections_idle, connection_usage_pct,
		       query_rate, p95_latency_ms, cache_hit_ratio,
		       database_size_bytes, dead_tuples_count, active_locks_count,
		       replication_lag_seconds, created_at
		FROM metrics
		WHERE database_id = $1 AND created_at >= $2
		ORDER BY created_at ASC
		LIMIT $3`

	rows, err := r.pool.Query(ctx, query, databaseID, since, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []Metric
	for rows.Next() {
		var m Metric
		if err := rows.Scan(
			&m.ID, &m.DatabaseID, &m.CPUUsage, &m.MemoryUsage,
			&m.ConnectionsTotal, &m.ConnectionsActive, &m.ConnectionsIdle, &m.ConnectionUsagePct,
			&m.QueryRate, &m.P95LatencyMs, &m.CacheHitRatio,
			&m.DatabaseSizeBytes, &m.DeadTuplesCount, &m.ActiveLocksCount,
			&m.ReplicationLagSeconds, &m.CreatedAt,
		); err != nil {
			return nil, err
		}
		result = append(result, m)
	}

	return result, rows.Err()
}
