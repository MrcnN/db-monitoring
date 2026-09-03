package collector

import (
	"context"
	"fmt"

	"github.com/dbplatform/api/internal/database"
)

type MetricsSnapshot struct {
	CPUUsage              float64                `json:"cpu_usage"`
	MemoryUsage           float64                `json:"memory_usage"`
	ConnectionsTotal      int                    `json:"connections_total"`
	ConnectionsActive     int                    `json:"connections_active"`
	ConnectionsIdle       int                    `json:"connections_idle"`
	ConnectionUsagePct    float64                `json:"connection_usage_pct"`
	QueryRate             float64                `json:"query_rate"`
	P95LatencyMs          float64                `json:"p95_latency_ms"`
	CacheHitRatio         float64                `json:"cache_hit_ratio"`
	DatabaseSizeBytes     int64                  `json:"database_size_bytes"`
	DeadTuplesCount       int64                  `json:"dead_tuples_count"`
	ActiveLocksCount      int                    `json:"active_locks_count"`
	ReplicationLagSeconds float64                `json:"replication_lag_seconds"`
	RawDetails            map[string]interface{} `json:"raw_details,omitempty"`
}

type Collector interface {
	Collect(ctx context.Context) (*MetricsSnapshot, error)
	Type() database.DatabaseType
	Close() error
}

func NewCollector(db *database.MonitoredDatabase, password string) (Collector, error) {
	switch db.Type {
	case database.TypePostgreSQL:
		return NewPostgresCollector(db, password)
	case database.TypeMySQL:
		return NewMySQLCollector(db, password)
	default:
		return nil, fmt.Errorf("unsupported database type: %s", db.Type)
	}
}
