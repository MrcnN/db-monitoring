CREATE TABLE IF NOT EXISTS metrics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    database_id UUID NOT NULL REFERENCES monitored_databases(id) ON DELETE CASCADE,
    cpu_usage DOUBLE PRECISION NOT NULL DEFAULT 0,
    memory_usage DOUBLE PRECISION NOT NULL DEFAULT 0,
    connections_total INT NOT NULL DEFAULT 0,
    connections_active INT NOT NULL DEFAULT 0,
    connections_idle INT NOT NULL DEFAULT 0,
    connection_usage_pct DOUBLE PRECISION NOT NULL DEFAULT 0,
    query_rate DOUBLE PRECISION NOT NULL DEFAULT 0,
    p95_latency_ms DOUBLE PRECISION NOT NULL DEFAULT 0,
    cache_hit_ratio DOUBLE PRECISION NOT NULL DEFAULT 0,
    database_size_bytes BIGINT NOT NULL DEFAULT 0,
    dead_tuples_count BIGINT NOT NULL DEFAULT 0,
    active_locks_count INT NOT NULL DEFAULT 0,
    replication_lag_seconds DOUBLE PRECISION NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_metrics_db_created ON metrics (database_id, created_at DESC);
