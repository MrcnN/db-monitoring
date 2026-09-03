package metrics

import (
	"time"

	"github.com/google/uuid"
)

type Metric struct {
	ID                    uuid.UUID `json:"id"`
	DatabaseID            uuid.UUID `json:"database_id"`
	CPUUsage              float64   `json:"cpu_usage"`
	MemoryUsage           float64   `json:"memory_usage"`
	ConnectionsTotal      int       `json:"connections_total"`
	ConnectionsActive     int       `json:"connections_active"`
	ConnectionsIdle       int       `json:"connections_idle"`
	ConnectionUsagePct    float64   `json:"connection_usage_pct"`
	QueryRate             float64   `json:"query_rate"` // Transactions/Queries per second
	P95LatencyMs          float64   `json:"p95_latency_ms"`
	CacheHitRatio         float64   `json:"cache_hit_ratio"` // Percentage (0 - 100)
	DatabaseSizeBytes     int64     `json:"database_size_bytes"`
	DeadTuplesCount       int64     `json:"dead_tuples_count"`
	ActiveLocksCount      int       `json:"active_locks_count"`
	ReplicationLagSeconds float64   `json:"replication_lag_seconds"`
	CreatedAt             time.Time `json:"created_at"`
}

type TimeRange string

const (
	Range15M TimeRange = "15m"
	Range1H  TimeRange = "1h"
	Range6H  TimeRange = "6h"
	Range24H TimeRange = "24h"
	Range7D  TimeRange = "7d"
)

func (tr TimeRange) Duration() time.Duration {
	switch tr {
	case Range15M:
		return 15 * time.Minute
	case Range1H:
		return 1 * time.Hour
	case Range6H:
		return 6 * time.Hour
	case Range24H:
		return 24 * time.Hour
	case Range7D:
		return 7 * 24 * time.Hour
	default:
		return 1 * time.Hour
	}
}

func ParseTimeRange(s string) TimeRange {
	switch s {
	case "15m":
		return Range15M
	case "1h":
		return Range1H
	case "6h":
		return Range6H
	case "24h":
		return Range24H
	case "7d":
		return Range7D
	default:
		return Range1H
	}
}
