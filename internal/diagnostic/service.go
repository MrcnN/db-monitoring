package diagnostic

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	"github.com/dbplatform/api/internal/crypto"
	"github.com/dbplatform/api/internal/database"
	apperrors "github.com/dbplatform/api/internal/errors"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type Service struct {
	dbSvc     *database.Service
	encryptor *crypto.Encryptor
}

func NewService(dbSvc *database.Service, encryptor *crypto.Encryptor) *Service {
	return &Service{dbSvc: dbSvc, encryptor: encryptor}
}

type LockInfo struct {
	PID           int    `json:"pid"`
	Query         string `json:"query"`
	State         string `json:"state"`
	BlockedBy     *int   `json:"blocked_by_pid,omitempty"`
	BlockingQuery string `json:"blocking_query,omitempty"`
	WaitEvent     string `json:"wait_event,omitempty"`
	WaitEventType string `json:"wait_event_type,omitempty"`
}

type TableStorageInfo struct {
	TableName    string  `json:"table_name"`
	TotalBytes   int64   `json:"total_bytes"`
	IndexBytes   int64   `json:"index_bytes"`
	LiveTuples   int64   `json:"live_tuples"`
	DeadTuples   int64   `json:"dead_tuples"`
	BloatRatio   float64 `json:"bloat_ratio"` // e.g. 0.15 for 15%
}

func (s *Service) GetActiveLocks(ctx context.Context, dbID uuid.UUID) ([]LockInfo, error) {
	target, err := s.dbSvc.GetTarget(ctx, dbID)
	if err != nil {
		return nil, err
	}
	password, err := s.encryptor.Decrypt(target.EncryptedPassword)
	if err != nil {
		return nil, err
	}

	var locks []LockInfo

	if target.Type == database.TypePostgreSQL {
		dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
			target.Username, password, target.Host, target.Port, target.DatabaseName, target.SSLMode)

		conn, err := pgx.Connect(ctx, dsn)
		if err != nil {
			return nil, apperrors.NewInternal(fmt.Errorf("failed to connect: %w", err))
		}
		defer conn.Close(ctx)

		query := `
		SELECT
			a.pid,
			a.query,
			a.state,
			a.wait_event,
			a.wait_event_type,
			pg_blocking_pids(a.pid)[1] AS blocked_by,
			(SELECT query FROM pg_stat_activity WHERE pid = pg_blocking_pids(a.pid)[1]) AS blocking_query
		FROM pg_stat_activity a
		WHERE a.state IS NOT NULL AND a.pid != pg_backend_pid()
		ORDER BY blocked_by NULLS LAST;`

		rows, err := conn.Query(ctx, query)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		for rows.Next() {
			var l LockInfo
			var blockingQuery *string
			var waitEvent, waitEventType *string

			// We just read the first element of the array safely in SQL using [1] but it might be null
			var blockedByPid *int32

			if err := rows.Scan(&l.PID, &l.Query, &l.State, &waitEvent, &waitEventType, &blockedByPid, &blockingQuery); err != nil {
				continue // skip errors on parsing
			}
			if waitEvent != nil {
				l.WaitEvent = *waitEvent
			}
			if waitEventType != nil {
				l.WaitEventType = *waitEventType
			}
			if blockedByPid != nil && *blockedByPid > 0 {
				pidVal := int(*blockedByPid)
				l.BlockedBy = &pidVal
			}
			if blockingQuery != nil {
				l.BlockingQuery = *blockingQuery
			}
			locks = append(locks, l)
		}
	} else if target.Type == database.TypeMySQL {
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true",
			target.Username, password, target.Host, target.Port, target.DatabaseName)

		dbConn, err := sql.Open("mysql", dsn)
		if err != nil {
			return nil, apperrors.NewInternal(fmt.Errorf("failed to open mysql: %w", err))
		}
		defer dbConn.Close()

		query := `
		SELECT 
			r.trx_mysql_thread_id AS pid,
			r.trx_query AS query,
			r.trx_state AS state,
			b.trx_mysql_thread_id AS blocked_by,
			b.trx_query AS blocking_query
		FROM information_schema.innodb_lock_waits w
		INNER JOIN information_schema.innodb_trx b ON b.trx_id = w.blocking_trx_id
		INNER JOIN information_schema.innodb_trx r ON r.trx_id = w.requesting_trx_id;`

		rows, err := dbConn.QueryContext(ctx, query)
		if err != nil {
			// fallback if innodb_lock_waits fails (MySQL 8+ performance_schema.data_locks is alternative, but let's just return empty)
			return []LockInfo{}, nil
		}
		defer rows.Close()

		for rows.Next() {
			var l LockInfo
			var blockedByPid *int
			var blockingQuery *string
			var queryStr *string
			
			if err := rows.Scan(&l.PID, &queryStr, &l.State, &blockedByPid, &blockingQuery); err != nil {
				continue
			}
			if queryStr != nil {
				l.Query = *queryStr
			}
			if blockedByPid != nil {
				l.BlockedBy = blockedByPid
			}
			if blockingQuery != nil {
				l.BlockingQuery = *blockingQuery
			}
			locks = append(locks, l)
		}
	}

	return locks, nil
}

func (s *Service) GetStorageStats(ctx context.Context, dbID uuid.UUID) ([]TableStorageInfo, error) {
	target, err := s.dbSvc.GetTarget(ctx, dbID)
	if err != nil {
		return nil, err
	}
	password, err := s.encryptor.Decrypt(target.EncryptedPassword)
	if err != nil {
		return nil, err
	}

	var stats []TableStorageInfo

	if target.Type == database.TypePostgreSQL {
		dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
			target.Username, password, target.Host, target.Port, target.DatabaseName, target.SSLMode)

		conn, err := pgx.Connect(ctx, dsn)
		if err != nil {
			return nil, apperrors.NewInternal(fmt.Errorf("failed to connect: %w", err))
		}
		defer conn.Close(ctx)

		query := `
		SELECT
			relname AS table_name,
			pg_total_relation_size(c.oid) AS total_bytes,
			pg_indexes_size(c.oid) AS index_bytes,
			COALESCE(stat.n_live_tup, 0) AS live_tuples,
			COALESCE(stat.n_dead_tup, 0) AS dead_tuples
		FROM pg_class c
		LEFT JOIN pg_namespace n ON n.oid = c.relnamespace
		LEFT JOIN pg_stat_user_tables stat ON stat.relid = c.oid
		WHERE c.relkind = 'r' AND n.nspname NOT IN ('pg_catalog', 'information_schema')
		ORDER BY pg_total_relation_size(c.oid) DESC
		LIMIT 100;`

		rows, err := conn.Query(ctx, query)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		for rows.Next() {
			var t TableStorageInfo
			if err := rows.Scan(&t.TableName, &t.TotalBytes, &t.IndexBytes, &t.LiveTuples, &t.DeadTuples); err != nil {
				continue
			}
			totalTups := t.LiveTuples + t.DeadTuples
			if totalTups > 0 {
				t.BloatRatio = float64(t.DeadTuples) / float64(totalTups)
			}
			stats = append(stats, t)
		}
	} else if target.Type == database.TypeMySQL {
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true",
			target.Username, password, target.Host, target.Port, target.DatabaseName)

		dbConn, err := sql.Open("mysql", dsn)
		if err != nil {
			return nil, apperrors.NewInternal(fmt.Errorf("failed to open mysql: %w", err))
		}
		defer dbConn.Close()

		query := `
		SELECT 
			TABLE_NAME,
			(DATA_LENGTH + INDEX_LENGTH) AS total_bytes,
			INDEX_LENGTH AS index_bytes,
			TABLE_ROWS AS live_tuples,
			DATA_FREE AS dead_bytes
		FROM information_schema.TABLES 
		WHERE TABLE_SCHEMA = ?
		ORDER BY (DATA_LENGTH + INDEX_LENGTH) DESC
		LIMIT 100;`

		rows, err := dbConn.QueryContext(ctx, query, target.DatabaseName)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		for rows.Next() {
			var t TableStorageInfo
			var deadBytes int64
			var liveTup *int64
			
			if err := rows.Scan(&t.TableName, &t.TotalBytes, &t.IndexBytes, &liveTup, &deadBytes); err != nil {
				continue
			}
			if liveTup != nil {
				t.LiveTuples = *liveTup
			}
			if t.TotalBytes > 0 {
				t.BloatRatio = float64(deadBytes) / float64(t.TotalBytes)
			}
			stats = append(stats, t)
		}
	}

	return stats, nil
}
