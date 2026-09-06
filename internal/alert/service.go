package alert

import (
	"context"

	"github.com/dbplatform/api/internal/health"
	"github.com/google/uuid"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) ProcessHealthResult(ctx context.Context, dbID uuid.UUID, res health.HealthResult) error {
	var activeTitles []string
	for _, issue := range res.Issues {
		activeTitles = append(activeTitles, issue.Title)

		severity := SeverityWarning
		if res.Status == health.StatusCritical {
			severity = SeverityCritical
		}

		alert := &Alert{
			DatabaseID:   dbID,
			Title:        issue.Title,
			Description:  issue.Description,
			Severity:     severity,
			Status:       StatusActive,
			MetricName:   "",
			CurrentValue: issue.CurrentValue,
		}

		_ = s.repo.Create(ctx, alert)
	}

	return s.repo.ResolveAlertsByDatabase(ctx, dbID, activeTitles)
}

func (s *Service) ListByDatabaseID(ctx context.Context, dbID uuid.UUID, limit, offset int) ([]Alert, error) {
	return s.repo.ListByDatabaseID(ctx, dbID, limit, offset)
}

func (s *Service) ListActive(ctx context.Context, limit, offset int) ([]Alert, error) {
	return s.repo.ListActive(ctx, limit, offset)
}
