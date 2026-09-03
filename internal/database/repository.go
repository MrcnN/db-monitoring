package database

import (
	"context"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5"
	"github.com/dbplatform/api/internal/errors"
)

type Repository interface {
	Create(ctx context.Context, db *MonitoredDatabase, encryptedPassword string) error
	GetByID(ctx context.Context, id uuid.UUID) (*MonitoredDatabase, string, error)
	List(ctx context.Context, userID *uuid.UUID) ([]MonitoredDatabase, error)
	Update(ctx context.Context, db *MonitoredDatabase, encryptedPassword *string) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetCredentials(ctx context.Context, dbID uuid.UUID) (string, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status Status) error
}

type postgresRepository struct {
	pool *pgxpool.Pool
}

var _ Repository = (*postgresRepository)(nil)

func NewRepository(pool *pgxpool.Pool) Repository {
	return &postgresRepository{pool: pool}
}

func (r *postgresRepository) Create(ctx context.Context, db *MonitoredDatabase, encryptedPassword string) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	query := `INSERT INTO monitored_databases (name, description, type, host, port, database_name, username, ssl_mode, monitoring_interval, created_by) 
			  VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) RETURNING id, created_at, updated_at`
	err = tx.QueryRow(ctx, query, db.Name, db.Description, db.Type, db.Host, db.Port, db.DatabaseName, db.Username, db.SSLMode, db.MonitoringInterval, db.CreatedBy).Scan(&db.ID, &db.CreatedAt, &db.UpdatedAt)
	if err != nil {
		return err
	}

	credQuery := `INSERT INTO database_credentials (database_id, encrypted_password) VALUES ($1, $2)`
	if _, err := tx.Exec(ctx, credQuery, db.ID, encryptedPassword); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *postgresRepository) GetByID(ctx context.Context, id uuid.UUID) (*MonitoredDatabase, string, error) {
	query := `SELECT d.id, d.name, d.description, d.type, d.host, d.port, d.database_name, d.username, d.ssl_mode, d.monitoring_interval, d.status, d.is_monitoring_enabled, d.last_checked_at, d.created_by, d.created_at, d.updated_at, c.encrypted_password
			  FROM monitored_databases d 
			  JOIN database_credentials c ON d.id = c.database_id 
			  WHERE d.id = $1 AND d.deleted_at IS NULL`
	var db MonitoredDatabase
	var encPwd string
	err := r.pool.QueryRow(ctx, query, id).Scan(&db.ID, &db.Name, &db.Description, &db.Type, &db.Host, &db.Port, &db.DatabaseName, &db.Username, &db.SSLMode, &db.MonitoringInterval, &db.Status, &db.IsMonitoringEnabled, &db.LastCheckedAt, &db.CreatedBy, &db.CreatedAt, &db.UpdatedAt, &encPwd)
	if err == pgx.ErrNoRows {
		return nil, "", errors.NewNotFound("Database")
	}
	return &db, encPwd, err
}

func (r *postgresRepository) List(ctx context.Context, userID *uuid.UUID) ([]MonitoredDatabase, error) {
	query := `SELECT id, name, description, type, host, port, database_name, username, ssl_mode, monitoring_interval, status, is_monitoring_enabled, last_checked_at, created_by, created_at, updated_at
			  FROM monitored_databases WHERE deleted_at IS NULL`
	var args []interface{}
	if userID != nil {
		query += ` AND created_by = $1`
		args = append(args, *userID)
	}
	
	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var dbs []MonitoredDatabase
	for rows.Next() {
		var db MonitoredDatabase
		if err := rows.Scan(&db.ID, &db.Name, &db.Description, &db.Type, &db.Host, &db.Port, &db.DatabaseName, &db.Username, &db.SSLMode, &db.MonitoringInterval, &db.Status, &db.IsMonitoringEnabled, &db.LastCheckedAt, &db.CreatedBy, &db.CreatedAt, &db.UpdatedAt); err != nil {
			return nil, err
		}
		dbs = append(dbs, db)
	}
	return dbs, nil
}

func (r *postgresRepository) Update(ctx context.Context, db *MonitoredDatabase, encryptedPassword *string) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	query := `UPDATE monitored_databases SET name=$1, description=$2, host=$3, port=$4, database_name=$5, username=$6, ssl_mode=$7, monitoring_interval=$8, is_monitoring_enabled=$9, updated_at=NOW() WHERE id=$10 AND deleted_at IS NULL`
	_, err = tx.Exec(ctx, query, db.Name, db.Description, db.Host, db.Port, db.DatabaseName, db.Username, db.SSLMode, db.MonitoringInterval, db.IsMonitoringEnabled, db.ID)
	if err != nil {
		return err
	}

	if encryptedPassword != nil {
		credQuery := `UPDATE database_credentials SET encrypted_password=$1, updated_at=NOW() WHERE database_id=$2`
		if _, err := tx.Exec(ctx, credQuery, *encryptedPassword, db.ID); err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func (r *postgresRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `UPDATE monitored_databases SET deleted_at = NOW() WHERE id = $1`, id)
	return err
}

func (r *postgresRepository) GetCredentials(ctx context.Context, dbID uuid.UUID) (string, error) {
	var pwd string
	err := r.pool.QueryRow(ctx, `SELECT encrypted_password FROM database_credentials WHERE database_id = $1`, dbID).Scan(&pwd)
	return pwd, err
}

func (r *postgresRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status Status) error {
	query := `UPDATE monitored_databases SET status = $1, last_checked_at = NOW(), updated_at = NOW() WHERE id = $2`
	_, err := r.pool.Exec(ctx, query, status, id)
	return err
}

