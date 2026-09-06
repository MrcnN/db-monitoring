package health

import (
	"fmt"

	"github.com/dbplatform/api/internal/collector"
)

type Status string

const (
	StatusHealthy  Status = "healthy"
	StatusWarning  Status = "warning"
	StatusCritical Status = "critical"
	StatusUnknown  Status = "unknown"
)

type IssueSeverity string

const (
	SeverityInfo     IssueSeverity = "info"
	SeverityWarning  IssueSeverity = "warning"
	SeverityCritical IssueSeverity = "critical"
)

type HealthIssue struct {
	Severity     IssueSeverity `json:"severity"`
	Title        string        `json:"title"`
	Description  string        `json:"description"`
	CurrentValue string        `json:"current_value"`
	Threshold    string        `json:"threshold"`
}

type HealthResult struct {
	Status  Status        `json:"status"`
	Score   int           `json:"score"` // 0 - 100 health score
	Issues  []HealthIssue `json:"issues"`
	Summary string        `json:"summary"`
}

type Evaluator struct{}

func NewEvaluator() *Evaluator {
	return &Evaluator{}
}

func (e *Evaluator) Evaluate(snapshot *collector.MetricsSnapshot, isReachable bool, lastErr error) HealthResult {
	if !isReachable {
		errMsg := "Database is unreachable"
		if lastErr != nil {
			errMsg = fmt.Sprintf("Connection failed: %s", lastErr.Error())
		}
		return HealthResult{
			Status: StatusCritical,
			Score:  0,
			Issues: []HealthIssue{
				{
					Severity:     SeverityCritical,
					Title:        "Database Connection Down",
					Description:  errMsg,
					CurrentValue: "Offline",
					Threshold:    "Online",
				},
			},
			Summary: "Database is unreachable and failing health checks.",
		}
	}

	if snapshot == nil {
		return HealthResult{
			Status:  StatusUnknown,
			Score:   50,
			Issues:  nil,
			Summary: "No metrics collected yet.",
		}
	}

	var issues []HealthIssue
	score := 100

	if snapshot.ConnectionUsagePct >= 92.0 {
		score -= 45
		issues = append(issues, HealthIssue{
			Severity:     SeverityCritical,
			Title:        "Connection Pool Exhaustion",
			Description:  "Database connection pool is almost completely full.",
			CurrentValue: fmt.Sprintf("%.1f%% (%d / %d)", snapshot.ConnectionUsagePct, snapshot.ConnectionsActive, snapshot.ConnectionsTotal),
			Threshold:    "< 90%",
		})
	} else if snapshot.ConnectionUsagePct >= 80.0 {
		score -= 20
		issues = append(issues, HealthIssue{
			Severity:     SeverityWarning,
			Title:        "Elevated Connection Usage",
			Description:  "Connection pool usage is running high.",
			CurrentValue: fmt.Sprintf("%.1f%% (%d / %d)", snapshot.ConnectionUsagePct, snapshot.ConnectionsActive, snapshot.ConnectionsTotal),
			Threshold:    "< 80%",
		})
	}

	if snapshot.CacheHitRatio > 0 && snapshot.CacheHitRatio < 85.0 {
		score -= 45
		issues = append(issues, HealthIssue{
			Severity:     SeverityCritical,
			Title:        "Poor Cache Hit Ratio",
			Description:  "Significant disk I/O bottleneck detected. Memory buffer pool is undersized or cache is missing.",
			CurrentValue: fmt.Sprintf("%.1f%%", snapshot.CacheHitRatio),
			Threshold:    "> 95%",
		})
	} else if snapshot.CacheHitRatio > 0 && snapshot.CacheHitRatio < 95.0 {
		score -= 20
		issues = append(issues, HealthIssue{
			Severity:     SeverityWarning,
			Title:        "Suboptimal Cache Hit Ratio",
			Description:  "Frequent disk reads detected. Queries could benefit from increased buffer cache.",
			CurrentValue: fmt.Sprintf("%.1f%%", snapshot.CacheHitRatio),
			Threshold:    "> 95%",
		})
	}

	if snapshot.P95LatencyMs >= 1000.0 {
		score -= 45
		issues = append(issues, HealthIssue{
			Severity:     SeverityCritical,
			Title:        "Critical Query Latency",
			Description:  "Database response latency is severely degraded.",
			CurrentValue: fmt.Sprintf("%.1f ms", snapshot.P95LatencyMs),
			Threshold:    "< 200 ms",
		})
	} else if snapshot.P95LatencyMs >= 200.0 {
		score -= 20
		issues = append(issues, HealthIssue{
			Severity:     SeverityWarning,
			Title:        "Elevated Query Latency",
			Description:  "Response latency exceeds typical baseline.",
			CurrentValue: fmt.Sprintf("%.1f ms", snapshot.P95LatencyMs),
			Threshold:    "< 200 ms",
		})
	}

	if snapshot.ActiveLocksCount >= 50 {
		score -= 45
		issues = append(issues, HealthIssue{
			Severity:     SeverityCritical,
			Title:        "High Lock Contention",
			Description:  "Multiple concurrent transactions are blocked waiting on table or row locks.",
			CurrentValue: fmt.Sprintf("%d waiting locks", snapshot.ActiveLocksCount),
			Threshold:    "< 10 locks",
		})
	} else if snapshot.ActiveLocksCount >= 10 {
		score -= 20
		issues = append(issues, HealthIssue{
			Severity:     SeverityWarning,
			Title:        "Moderate Lock Contention",
			Description:  "Several queries are currently queued waiting for locks.",
			CurrentValue: fmt.Sprintf("%d waiting locks", snapshot.ActiveLocksCount),
			Threshold:    "< 10 locks",
		})
	}

	if snapshot.DeadTuplesCount >= 500000 {
		score -= 20
		issues = append(issues, HealthIssue{
			Severity:     SeverityWarning,
			Title:        "High Dead Tuples Bloat",
			Description:  "Vacuuming may be falling behind. Table bloat can degrade query scan performance.",
			CurrentValue: fmt.Sprintf("%d dead tuples", snapshot.DeadTuplesCount),
			Threshold:    "< 100,000 dead tuples",
		})
	}

	if snapshot.ReplicationLagSeconds >= 60.0 {
		score -= 45
		issues = append(issues, HealthIssue{
			Severity:     SeverityCritical,
			Title:        "High Replication Lag",
			Description:  "Replica is falling significantly behind the primary node.",
			CurrentValue: fmt.Sprintf("%.1f s", snapshot.ReplicationLagSeconds),
			Threshold:    "< 10 s",
		})
	} else if snapshot.ReplicationLagSeconds >= 10.0 {
		score -= 20
		issues = append(issues, HealthIssue{
			Severity:     SeverityWarning,
			Title:        "Noticeable Replication Lag",
			Description:  "Replica delay detected.",
			CurrentValue: fmt.Sprintf("%.1f s", snapshot.ReplicationLagSeconds),
			Threshold:    "< 10 s",
		})
	}

	if score < 0 {
		score = 0
	}

	var status Status
	var summary string
	if score >= 85 {
		status = StatusHealthy
		summary = "All monitored health parameters are within normal operating thresholds."
	} else if score >= 60 {
		status = StatusWarning
		summary = fmt.Sprintf("System is operational with %d warning(s) requiring attention.", len(issues))
	} else {
		status = StatusCritical
		summary = fmt.Sprintf("Critical performance or availability issues detected (%d issue(s)).", len(issues))
	}

	return HealthResult{
		Status:  status,
		Score:   score,
		Issues:  issues,
		Summary: summary,
	}
}
