package audit

import (
	"context"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AuditLog struct {
	ID           uuid.UUID
	UserID       *uuid.UUID
	UserEmail    string
	Action       string
	ResourceType string
	ResourceID   string
	ResourceName string
	Details      map[string]interface{}
	IPAddress    string
	UserAgent    string
	Status       string
	CreatedAt    time.Time
}

const (
	ActionLogin      = "user.login"
	ActionLogout     = "user.logout"
	ActionRegister   = "user.register"
	ActionCreateDB   = "database.create"
	ActionUpdateDB   = "database.update"
	ActionDeleteDB   = "database.delete"
	ActionTestConnDB = "database.test_connection"
)

type ListFilter struct {
	UserID *uuid.UUID
	Action string
	From   *time.Time
	To     *time.Time
	Limit  int
	Offset int
}

type Repository interface {
	Create(ctx context.Context, log *AuditLog) error
	List(ctx context.Context, filter ListFilter) ([]AuditLog, int, error)
}

type postgresRepository struct {
	db *pgxpool.Pool
}

var _ Repository = (*postgresRepository)(nil)

func NewRepository(db *pgxpool.Pool) Repository {
	return &postgresRepository{db: db}
}

func (r *postgresRepository) Create(ctx context.Context, log *AuditLog) error {
	query := `
		INSERT INTO audit_logs
			(user_id, user_email, action, resource_type, resource_id, resource_name, details, ip_address, user_agent, status)
		VALUES
			($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, created_at`

	return r.db.QueryRow(ctx, query,
		log.UserID,
		log.UserEmail,
		log.Action,
		log.ResourceType,
		log.ResourceID,
		log.ResourceName,
		log.Details,
		log.IPAddress,
		log.UserAgent,
		log.Status,
	).Scan(&log.ID, &log.CreatedAt)
}

func (r *postgresRepository) List(ctx context.Context, filter ListFilter) ([]AuditLog, int, error) {
	args := []interface{}{}
	where := " WHERE 1=1"
	argIdx := 1

	if filter.UserID != nil {
		where += " AND user_id = $" + itoa(argIdx)
		args = append(args, filter.UserID)
		argIdx++
	}
	if filter.Action != "" {
		where += " AND action = $" + itoa(argIdx)
		args = append(args, filter.Action)
		argIdx++
	}
	if filter.From != nil {
		where += " AND created_at >= $" + itoa(argIdx)
		args = append(args, filter.From)
		argIdx++
	}
	if filter.To != nil {
		where += " AND created_at <= $" + itoa(argIdx)
		args = append(args, filter.To)
		argIdx++
	}

	var total int
	countQuery := "SELECT COUNT(*) FROM audit_logs" + where
	if err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	limit := filter.Limit
	if limit <= 0 {
		limit = 50
	}
	offset := filter.Offset

	args = append(args, limit, offset)
	dataQuery := `SELECT id, user_id, user_email, action, resource_type, resource_id, resource_name, ip_address, user_agent, status, created_at
		FROM audit_logs` + where + ` ORDER BY created_at DESC LIMIT $` + itoa(argIdx) + ` OFFSET $` + itoa(argIdx+1)

	rows, err := r.db.Query(ctx, dataQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var logs []AuditLog
	for rows.Next() {
		var l AuditLog
		if err := rows.Scan(
			&l.ID, &l.UserID, &l.UserEmail,
			&l.Action, &l.ResourceType, &l.ResourceID, &l.ResourceName,
			&l.IPAddress, &l.UserAgent, &l.Status, &l.CreatedAt,
		); err != nil {
			return nil, 0, err
		}
		logs = append(logs, l)
	}

	if rows.Err() != nil {
		return nil, 0, rows.Err()
	}

	return logs, total, nil
}

func itoa(i int) string {
	return strconv.Itoa(i)
}
