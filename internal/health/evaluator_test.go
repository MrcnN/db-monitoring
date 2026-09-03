package health_test

import (
	"errors"
	"testing"

	"github.com/dbplatform/api/internal/collector"
	"github.com/dbplatform/api/internal/health"
)

func TestEvaluator_UnreachableDatabase(t *testing.T) {
	eval := health.NewEvaluator()
	res := eval.Evaluate(nil, false, errors.New("connection refused"))

	if res.Status != health.StatusCritical {
		t.Fatalf("expected critical status for unreachable db, got: %s", res.Status)
	}
	if res.Score != 0 {
		t.Fatalf("expected score 0, got: %d", res.Score)
	}
	if len(res.Issues) == 0 {
		t.Fatal("expected issues to be populated")
	}
}

func TestEvaluator_HealthyDatabase(t *testing.T) {
	eval := health.NewEvaluator()
	snapshot := &collector.MetricsSnapshot{
		ConnectionsTotal:   100,
		ConnectionsActive:  10,
		ConnectionUsagePct: 10.0,
		CacheHitRatio:      99.5,
		P95LatencyMs:       1.5,
		ActiveLocksCount:   0,
		DeadTuplesCount:    500,
	}

	res := eval.Evaluate(snapshot, true, nil)
	if res.Status != health.StatusHealthy {
		t.Fatalf("expected healthy status, got: %s", res.Status)
	}
	if res.Score < 85 {
		t.Fatalf("expected score >= 85, got: %d", res.Score)
	}
	if len(res.Issues) != 0 {
		t.Fatalf("expected 0 issues, got: %d", len(res.Issues))
	}
}

func TestEvaluator_Warning_HighConnectionUsage(t *testing.T) {
	eval := health.NewEvaluator()
	snapshot := &collector.MetricsSnapshot{
		ConnectionsTotal:   100,
		ConnectionsActive:  85,
		ConnectionUsagePct: 85.0, // > 80%
		CacheHitRatio:      99.0,
		P95LatencyMs:       2.0,
	}

	res := eval.Evaluate(snapshot, true, nil)
	if res.Status != health.StatusWarning {
		t.Fatalf("expected warning status, got: %s", res.Status)
	}
	if len(res.Issues) == 0 {
		t.Fatal("expected at least 1 issue")
	}
	if res.Issues[0].Title != "Elevated Connection Usage" {
		t.Fatalf("expected 'Elevated Connection Usage', got %q", res.Issues[0].Title)
	}
}

func TestEvaluator_Critical_ConnectionExhaustion(t *testing.T) {
	eval := health.NewEvaluator()
	snapshot := &collector.MetricsSnapshot{
		ConnectionsTotal:   100,
		ConnectionsActive:  96,
		ConnectionUsagePct: 96.0, // > 92%
		CacheHitRatio:      99.0,
		P95LatencyMs:       2.0,
	}

	res := eval.Evaluate(snapshot, true, nil)
	if res.Status != health.StatusCritical {
		t.Fatalf("expected critical status, got: %s", res.Status)
	}
}

func TestEvaluator_Critical_PoorCacheHitRatio(t *testing.T) {
	eval := health.NewEvaluator()
	snapshot := &collector.MetricsSnapshot{
		ConnectionsTotal:   100,
		ConnectionsActive:  20,
		ConnectionUsagePct: 20.0,
		CacheHitRatio:      78.0, // < 85%
		P95LatencyMs:       5.0,
	}

	res := eval.Evaluate(snapshot, true, nil)
	if res.Status != health.StatusCritical {
		t.Fatalf("expected critical status for bad cache hit ratio, got: %s", res.Status)
	}
}
