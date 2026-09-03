# Collectors Architecture (Phase 2)

## Interface Definition
```go
type Collector interface {
    Collect(ctx context.Context, db *sql.DB) (*Metrics, error)
    Engine() string
}
```

## PostgreSQL Collector Metrics
- Connections (active, idle, waiting).
- Buffer cache hit ratio.
- Dead tuples and bloat estimation.
- Query throughput (TPS).
- Long-running transactions.

## MySQL Collector Metrics
- InnoDB buffer pool efficiency.
- Threads connected/running.
- Slow queries count.
- Table locks and waits.

## Adding New Collectors
1. Implement the `Collector` interface in `internal/collectors/YOUR_ENGINE`.
2. Register the collector in the factory.
3. Update `monitored_databases` enum constraint.

## Concurrency Model
- Worker pool pulls from a Redis task queue.
- Context timeouts strictly enforce collection windows (e.g., max 10s per DB).

## Error Handling
- Connection failures recorded in `monitored_databases` (status = `error`).
- Exponential backoff for failing targets.
