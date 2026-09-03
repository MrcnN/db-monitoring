package audit

import (
	"context"
	"time"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Log(ctx context.Context, entry *AuditLog) error {
	return s.repo.Create(ctx, entry)
}

func (s *Service) LogAsync(entry *AuditLog) {
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = s.repo.Create(ctx, entry)
	}()
}
