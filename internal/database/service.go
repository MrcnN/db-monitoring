package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/dbplatform/api/internal/crypto"
	apperrors "github.com/dbplatform/api/internal/errors"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type Service struct {
	repo      Repository
	encryptor *crypto.Encryptor
}

func NewService(repo Repository, encryptor *crypto.Encryptor) *Service {
	return &Service{repo: repo, encryptor: encryptor}
}

func (s *Service) Create(ctx context.Context, db *MonitoredDatabase, encryptedPassword string) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if err := s.repo.Create(ctx, db, encryptedPassword); err != nil {
		return apperrors.NewInternal(err)
	}
	return nil
}

func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (*MonitoredDatabase, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	db, _, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return db, nil
}

func (s *Service) GetWithCredentials(ctx context.Context, id uuid.UUID) (*MonitoredDatabase, string, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	return s.repo.GetByID(ctx, id)
}

func (s *Service) List(ctx context.Context) ([]MonitoredDatabase, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	dbs, err := s.repo.List(ctx, nil)
	if err != nil {
		return nil, apperrors.NewInternal(err)
	}
	return dbs, nil
}

func (s *Service) Update(ctx context.Context, db *MonitoredDatabase, encryptedPassword *string) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if err := s.repo.Update(ctx, db, encryptedPassword); err != nil {
		return apperrors.NewInternal(err)
	}
	return nil
}

func (s *Service) Delete(ctx context.Context, id uuid.UUID) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if err := s.repo.Delete(ctx, id); err != nil {
		return apperrors.NewInternal(err)
	}
	return nil
}

func (s *Service) UpdateStatus(ctx context.Context, id uuid.UUID, status Status) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	return s.repo.UpdateStatus(ctx, id, status)
}

func (s *Service) TestConnection(ctx context.Context, db *MonitoredDatabase, password string) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	switch db.Type {
	case TypePostgreSQL:
		dsn := fmt.Sprintf(
			"postgres://%s:%s@%s:%d/%s?sslmode=%s&connect_timeout=4",
			db.Username, password, db.Host, db.Port, db.DatabaseName, db.SSLMode,
		)
		conn, err := pgx.Connect(ctx, dsn)
		if err != nil {
			return fmt.Errorf("could not connect to PostgreSQL instance")
		}
		defer conn.Close(ctx)
		if err := conn.Ping(ctx); err != nil {
			return fmt.Errorf("ping failed")
		}
		return nil

	case TypeMySQL:
		tlsMode := "false"
		if db.SSLMode != SSLModeDisable {
			tlsMode = "true"
		}
		dsn := fmt.Sprintf(
			"%s:%s@tcp(%s:%d)/%s?tls=%s&timeout=4s",
			db.Username, password, db.Host, db.Port, db.DatabaseName, tlsMode,
		)
		conn, err := sql.Open("mysql", dsn)
		if err != nil {
			return fmt.Errorf("could not connect to MySQL instance")
		}
		defer conn.Close()
		if err := conn.PingContext(ctx); err != nil {
			return fmt.Errorf("ping failed")
		}
		return nil

	default:
		return fmt.Errorf("unsupported database type: %s", db.Type)
	}
}
