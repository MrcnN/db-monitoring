package worker

import (
	"context"
	"sync"
	"time"

	"github.com/dbplatform/api/internal/alert"
	"github.com/dbplatform/api/internal/api/ws"
	"github.com/dbplatform/api/internal/collector"
	"github.com/dbplatform/api/internal/crypto"
	"github.com/dbplatform/api/internal/database"
	"github.com/dbplatform/api/internal/health"
	"github.com/dbplatform/api/internal/incident"
	"github.com/dbplatform/api/internal/metrics"
	"github.com/dbplatform/api/internal/observability"
	"github.com/rs/zerolog"
)

type CollectorWorker struct {
	dbSvc       *database.Service
	metricSvc   *metrics.Service
	evaluator   *health.Evaluator
	encryptor   *crypto.Encryptor
	alertSvc    *alert.Service
	incidentSvc *incident.Service
	wsHub       *ws.Hub
	interval    time.Duration
	log         zerolog.Logger
}

func NewCollectorWorker(
	dbSvc *database.Service,
	metricSvc *metrics.Service,
	evaluator *health.Evaluator,
	encryptor *crypto.Encryptor,
	alertSvc *alert.Service,
	incidentSvc *incident.Service,
	wsHub *ws.Hub,
	interval time.Duration,
	log zerolog.Logger,
) *CollectorWorker {
	if interval <= 0 {
		interval = 15 * time.Second
	}
	return &CollectorWorker{
		dbSvc:       dbSvc,
		metricSvc:   metricSvc,
		evaluator:   evaluator,
		encryptor:   encryptor,
		alertSvc:    alertSvc,
		incidentSvc: incidentSvc,
		wsHub:       wsHub,
		interval:    interval,
		log:         log.With().Str("component", "collector_worker").Logger(),
	}
}

func (w *CollectorWorker) Start(ctx context.Context) {
	w.log.Info().Dur("interval", w.interval).Msg("Starting metrics collector worker")

	w.collectAll(ctx)

	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			w.log.Info().Msg("Metrics collector worker shutting down")
			return
		case <-ticker.C:
			w.collectAll(ctx)
		}
	}
}

func (w *CollectorWorker) collectAll(ctx context.Context) {
	dbs, err := w.dbSvc.List(ctx)
	if err != nil {
		w.log.Error().Err(err).Msg("Failed to list databases for collection")
		return
	}

	var wg sync.WaitGroup
	for i := range dbs {
		db := dbs[i]
		if !db.IsMonitoringEnabled || db.Status == database.StatusInactive {
			continue
		}

		wg.Add(1)
		go func(target database.MonitoredDatabase) {
			defer wg.Done()
			w.collectOne(ctx, target)
		}(db)
	}

	wg.Wait()
}

func (w *CollectorWorker) collectOne(parentCtx context.Context, target database.MonitoredDatabase) {
	collectCtx, cancel := context.WithTimeout(parentCtx, 6*time.Second)
	defer cancel()

	_, encPwd, err := w.dbSvc.GetWithCredentials(collectCtx, target.ID)
	if err != nil {
		w.log.Error().Str("database", target.Name).Err(err).Msg("Failed to retrieve credentials")
		w.recordFailure(collectCtx, target, err)
		return
	}

	password, err := w.encryptor.Decrypt(encPwd)
	if err != nil {
		w.log.Error().Str("database", target.Name).Err(err).Msg("Failed to decrypt database password")
		w.recordFailure(collectCtx, target, err)
		return
	}

	col, err := collector.NewCollector(&target, password)
	if err != nil {
		w.log.Error().Str("database", target.Name).Err(err).Msg("Failed to instantiate collector")
		w.recordFailure(collectCtx, target, err)
		return
	}
	defer col.Close()

	snapshot, err := col.Collect(collectCtx)
	if err != nil {
		w.log.Warn().Str("database", target.Name).Err(err).Msg("Collection failed")
		w.recordFailure(collectCtx, target, err)
		return
	}

	m := &metrics.Metric{
		DatabaseID:            target.ID,
		CPUUsage:              snapshot.CPUUsage,
		MemoryUsage:           snapshot.MemoryUsage,
		ConnectionsTotal:      snapshot.ConnectionsTotal,
		ConnectionsActive:     snapshot.ConnectionsActive,
		ConnectionsIdle:       snapshot.ConnectionsIdle,
		ConnectionUsagePct:    snapshot.ConnectionUsagePct,
		QueryRate:             snapshot.QueryRate,
		P95LatencyMs:          snapshot.P95LatencyMs,
		CacheHitRatio:         snapshot.CacheHitRatio,
		DatabaseSizeBytes:     snapshot.DatabaseSizeBytes,
		DeadTuplesCount:       snapshot.DeadTuplesCount,
		ActiveLocksCount:      snapshot.ActiveLocksCount,
		ReplicationLagSeconds: snapshot.ReplicationLagSeconds,
		CreatedAt:             time.Now(),
	}

	if err := w.metricSvc.Record(collectCtx, m); err != nil {
		w.log.Error().Str("database", target.Name).Err(err).Msg("Failed to persist metrics record")
	} else {
		select {
		case w.wsHub.Broadcast <- ws.Message{
			DatabaseID: target.ID,
			Payload: map[string]interface{}{
				"type": "metric_update",
				"data": m,
			},
		}:
		default:
		}
	}

	healthResult := w.evaluator.Evaluate(snapshot, true, nil)
	newStatus := database.StatusActive
	if healthResult.Status == health.StatusCritical {
		newStatus = database.StatusError
	}

	_ = w.dbSvc.UpdateStatus(collectCtx, target.ID, newStatus)
	_ = w.alertSvc.ProcessHealthResult(collectCtx, target.ID, healthResult)
	_ = w.incidentSvc.EvaluateHealth(collectCtx, target.ID, healthResult)

	select {
	case w.wsHub.Broadcast <- ws.Message{
		DatabaseID: target.ID,
		Payload: map[string]interface{}{
			"type": "health_update",
			"data": healthResult,
		},
	}:
	default:
	}

	observability.HTTPRequestsTotal.WithLabelValues("collector", string(target.Type), "success").Inc()

	w.log.Debug().
		Str("database", target.Name).
		Str("status", string(healthResult.Status)).
		Float64("cache_hit", snapshot.CacheHitRatio).
		Int("active_conn", snapshot.ConnectionsActive).
		Msg("Collected metrics successfully")
}

func (w *CollectorWorker) recordFailure(ctx context.Context, target database.MonitoredDatabase, lastErr error) {
	_ = w.dbSvc.UpdateStatus(ctx, target.ID, database.StatusError)
	observability.HTTPRequestsTotal.WithLabelValues("collector", string(target.Type), "failure").Inc()

	res := w.evaluator.Evaluate(nil, false, lastErr)
	_ = w.alertSvc.ProcessHealthResult(ctx, target.ID, res)
	_ = w.incidentSvc.EvaluateHealth(ctx, target.ID, res)

	select {
	case w.wsHub.Broadcast <- ws.Message{
		DatabaseID: target.ID,
		Payload: map[string]interface{}{
			"type": "health_update",
			"data": res,
		},
	}:
	default:
	}
}
