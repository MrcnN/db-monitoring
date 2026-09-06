package incident

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/dbplatform/api/internal/errors"
	"github.com/dbplatform/api/internal/health"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

type Service struct {
	repo Repository
	log  zerolog.Logger
}

func NewService(repo Repository, log zerolog.Logger) *Service {
	return &Service{
		repo: repo,
		log:  log,
	}
}

func (s *Service) EvaluateHealth(ctx context.Context, dbID uuid.UUID, res health.HealthResult) error {
	existing, err := s.repo.GetOpenByDatabase(ctx, dbID)
	if err != nil && !errors.IsNotFound(err) {
		return err
	}

	if res.Status == health.StatusCritical {
		if existing == nil {
			title := "Database Critical Failure"
			desc := "Health score dropped below critical threshold."
			if len(res.Issues) > 0 {
				title = res.Issues[0].Title
				desc = res.Issues[0].Description
			}

			newInc := &Incident{
				DatabaseID:  dbID,
				Title:       title,
				Description: desc,
				Status:      StatusOpen,
			}
			if err := s.repo.Create(ctx, newInc); err != nil {
				return err
			}
			go s.notifyChannels(newInc)
		}
	} else if res.Status == health.StatusHealthy || res.Status == health.StatusWarning {
		if existing != nil {
			_ = s.repo.UpdateStatus(ctx, existing.ID, StatusResolved)
		}
	}

	return nil
}

func (s *Service) List(ctx context.Context, limit, offset int) ([]Incident, error) {
	return s.repo.List(ctx, limit, offset)
}

func (s *Service) CreateChannel(ctx context.Context, channel *NotificationChannel) error {
	return s.repo.CreateChannel(ctx, channel)
}

func (s *Service) ListChannels(ctx context.Context) ([]NotificationChannel, error) {
	return s.repo.ListChannels(ctx)
}

func (s *Service) DeleteChannel(ctx context.Context, id uuid.UUID) error {
	return s.repo.DeleteChannel(ctx, id)
}

func (s *Service) notifyChannels(inc *Incident) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	channels, err := s.repo.ListChannels(ctx)
	if err != nil {
		s.log.Error().Err(err).Msg("Failed to list notification channels")
		return
	}

	for _, ch := range channels {
		if !ch.IsActive {
			continue
		}
		
		if ch.Type == ChannelSlack || ch.Type == ChannelWebhook {
			webhookURL, ok := ch.Config["url"].(string)
			if !ok || webhookURL == "" {
				continue
			}

			payload := map[string]interface{}{
				"text": fmt.Sprintf("🚨 *Incident Opened*: %s\n%s", inc.Title, inc.Description),
			}
			body, _ := json.Marshal(payload)

			req, _ := http.NewRequestWithContext(ctx, "POST", webhookURL, bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				s.log.Error().Err(err).Str("channel", ch.Name).Msg("Failed to send webhook")
			} else {
				resp.Body.Close()
			}
		}
	}
}
