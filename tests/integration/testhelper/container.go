//go:build integration

package integration

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	tcredis "github.com/testcontainers/testcontainers-go/modules/redis"

	"github.com/dbplatform/api/internal/platform"
)

type TestPostgres struct {
	Container *postgres.PostgresContainer
	Pool      *pgxpool.Pool
	DSN       string
}

func SetupPostgres(ctx context.Context, t *testing.T, migrationsPath string) (*TestPostgres, func()) {
	t.Helper()

	pgContainer, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("dbplatform_test"),
		postgres.WithUsername("test"),
		postgres.WithPassword("test"),
		testcontainers.WithWaitStrategy(
			waitForPostgres(),
		),
	)
	if err != nil {
		t.Fatalf("failed to start postgres container: %v", err)
	}

	dsn, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		pgContainer.Terminate(ctx) //nolint:errcheck
		t.Fatalf("failed to get connection string: %v", err)
	}

	pool, err := platform.ConnectPostgres(ctx, dsn)
	if err != nil {
		pgContainer.Terminate(ctx) //nolint:errcheck
		t.Fatalf("failed to connect to test postgres: %v", err)
	}

	if migrationsPath != "" {
		if err := platform.RunMigrations(dsn, migrationsPath); err != nil {
			pool.Close()
			pgContainer.Terminate(ctx) //nolint:errcheck
			t.Fatalf("failed to run migrations: %v", err)
		}
	}

	cleanup := func() {
		pool.Close()
		if err := pgContainer.Terminate(ctx); err != nil {
			t.Logf("warning: failed to terminate postgres container: %v", err)
		}
	}

	return &TestPostgres{
		Container: pgContainer,
		Pool:      pool,
		DSN:       dsn,
	}, cleanup
}

type TestRedis struct {
	Container *tcredis.RedisContainer
	Client    *redis.Client
}

func SetupRedis(ctx context.Context, t *testing.T) (*TestRedis, func()) {
	t.Helper()

	redisContainer, err := tcredis.Run(ctx, "redis:7-alpine")
	if err != nil {
		t.Fatalf("failed to start redis container: %v", err)
	}

	connStr, err := redisContainer.ConnectionString(ctx)
	if err != nil {
		redisContainer.Terminate(ctx) //nolint:errcheck
		t.Fatalf("failed to get redis connection string: %v", err)
	}

	client, err := platform.ConnectRedis(ctx, "redis://"+connStr, "", 0)
	if err != nil {
		redisContainer.Terminate(ctx) //nolint:errcheck
		t.Fatalf("failed to connect to test redis: %v", err)
	}

	cleanup := func() {
		client.Close()
		if err := redisContainer.Terminate(ctx); err != nil {
			t.Logf("warning: failed to terminate redis container: %v", err)
		}
	}

	return &TestRedis{
		Container: redisContainer,
		Client:    client,
	}, cleanup
}

func waitForPostgres() testcontainers.WaitStrategy {
	return testcontainers.NewWaitStrategyList(
		testcontainers.NewLogStrategy("database system is ready to accept connections").WithStartupTimeout(30 * time.Second),
	)
}

func MigrationsPath() string {
	return fmt.Sprintf("../../%s", "migrations")
}
