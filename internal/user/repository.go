package user

import (
	"context"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5"
	"github.com/dbplatform/api/internal/errors"
)

type Repository interface {
	Create(ctx context.Context, user *User) error
	GetByID(ctx context.Context, id uuid.UUID) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	Update(ctx context.Context, user *User) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type SessionRepository interface {
	Create(ctx context.Context, session *Session) error
	GetByTokenHash(ctx context.Context, hash string) (*Session, error)
	RevokeByTokenHash(ctx context.Context, hash string) error
	RevokeAllForUser(ctx context.Context, userID uuid.UUID) error
	DeleteExpired(ctx context.Context) error
}

type postgresRepository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) Repository {
	return &postgresRepository{db: db}
}

func (r *postgresRepository) Create(ctx context.Context, user *User) error {
	query := `INSERT INTO users (email, password_hash, full_name, role, is_active) 
			  VALUES ($1, $2, $3, $4, $5) RETURNING id, created_at, updated_at`
	err := r.db.QueryRow(ctx, query, user.Email, user.PasswordHash, user.FullName, user.Role, user.IsActive).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return errors.NewDBConnection(err)
	}
	return nil
}

func (r *postgresRepository) GetByID(ctx context.Context, id uuid.UUID) (*User, error) {
	query := `SELECT id, email, password_hash, full_name, role, is_active, last_login_at, created_at, updated_at, deleted_at 
			  FROM users WHERE id = $1 AND deleted_at IS NULL`
	var u User
	err := r.db.QueryRow(ctx, query, id).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.FullName, &u.Role, &u.IsActive, &u.LastLoginAt, &u.CreatedAt, &u.UpdatedAt, &u.DeletedAt)
	if err == pgx.ErrNoRows {
		return nil, errors.NewNotFound("User")
	}
	return &u, err
}

func (r *postgresRepository) GetByEmail(ctx context.Context, email string) (*User, error) {
	query := `SELECT id, email, password_hash, full_name, role, is_active, last_login_at, created_at, updated_at, deleted_at 
			  FROM users WHERE email = $1 AND deleted_at IS NULL`
	var u User
	err := r.db.QueryRow(ctx, query, email).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.FullName, &u.Role, &u.IsActive, &u.LastLoginAt, &u.CreatedAt, &u.UpdatedAt, &u.DeletedAt)
	if err == pgx.ErrNoRows {
		return nil, errors.NewNotFound("User")
	}
	return &u, err
}

func (r *postgresRepository) Update(ctx context.Context, user *User) error {
	query := `UPDATE users SET full_name = $1, role = $2, is_active = $3, last_login_at = $4, updated_at = NOW() 
			  WHERE id = $5 AND deleted_at IS NULL RETURNING updated_at`
	err := r.db.QueryRow(ctx, query, user.FullName, user.Role, user.IsActive, user.LastLoginAt, user.ID).Scan(&user.UpdatedAt)
	return err
}

func (r *postgresRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, `UPDATE users SET deleted_at = NOW() WHERE id = $1`, id)
	return err
}

type postgresSessionRepository struct {
	db *pgxpool.Pool
}

func NewSessionRepository(db *pgxpool.Pool) SessionRepository {
	return &postgresSessionRepository{db: db}
}

func (r *postgresSessionRepository) Create(ctx context.Context, session *Session) error {
	query := `INSERT INTO user_sessions (user_id, refresh_token_hash, user_agent, ip_address, expires_at) 
			  VALUES ($1, $2, $3, $4, $5) RETURNING id, created_at`
	return r.db.QueryRow(ctx, query, session.UserID, session.RefreshTokenHash, session.UserAgent, session.IPAddress, session.ExpiresAt).Scan(&session.ID, &session.CreatedAt)
}

func (r *postgresSessionRepository) GetByTokenHash(ctx context.Context, hash string) (*Session, error) {
	query := `SELECT id, user_id, refresh_token_hash, user_agent, ip_address, expires_at, created_at, revoked_at 
			  FROM user_sessions WHERE refresh_token_hash = $1`
	var s Session
	err := r.db.QueryRow(ctx, query, hash).Scan(&s.ID, &s.UserID, &s.RefreshTokenHash, &s.UserAgent, &s.IPAddress, &s.ExpiresAt, &s.CreatedAt, &s.RevokedAt)
	if err == pgx.ErrNoRows {
		return nil, errors.NewNotFound("Session")
	}
	return &s, err
}

func (r *postgresSessionRepository) RevokeByTokenHash(ctx context.Context, hash string) error {
	_, err := r.db.Exec(ctx, `UPDATE user_sessions SET revoked_at = NOW() WHERE refresh_token_hash = $1`, hash)
	return err
}

func (r *postgresSessionRepository) RevokeAllForUser(ctx context.Context, userID uuid.UUID) error {
	_, err := r.db.Exec(ctx, `UPDATE user_sessions SET revoked_at = NOW() WHERE user_id = $1 AND revoked_at IS NULL`, userID)
	return err
}

func (r *postgresSessionRepository) DeleteExpired(ctx context.Context) error {
	_, err := r.db.Exec(ctx, `DELETE FROM user_sessions WHERE expires_at < NOW()`)
	return err
}
