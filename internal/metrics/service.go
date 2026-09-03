package metrics

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Record(ctx context.Context, m *Metric) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	return s.repo.Create(ctx, m)
}

func (s *Service) GetLatest(ctx context.Context, dbID uuid.UUID) (*Metric, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	return s.repo.GetLatest(ctx, dbID)
}

func (s *Service) GetTimeSeries(ctx context.Context, dbID uuid.UUID, tr TimeRange) ([]Metric, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	since := time.Now().Add(-tr.Duration())
	return s.repo.GetTimeSeries(ctx, dbID, since, 500)
}
