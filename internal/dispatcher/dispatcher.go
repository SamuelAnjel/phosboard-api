package dispatcher

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"time"

	"cloud.google.com/go/pubsub" //nolint:staticcheck
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/api/option"

	"phosboard/backend/internal/models"
)

const (
	topicID           = "source-discovery"
	tickerInterval    = 1 * time.Minute
	pollQueryInterval = 5 * time.Second
)

type Config struct {
	ProjectID       string
	PubSubEndpoint  string
	IntervalSeconds int
}

type DiscoveryTask struct {
	SourceID  string `json:"source_id"`
	URL       string `json:"url"`
	Timestamp string `json:"timestamp"`
}

type Dispatcher struct {
	pool            *pgxpool.Pool
	client          *pubsub.Client
	topic           *pubsub.Topic
	projectID       string
	intervalSeconds int
}

func NewDispatcher(ctx context.Context, pool *pgxpool.Pool, cfg Config) (*Dispatcher, error) {
	var client *pubsub.Client
	var err error

	if cfg.PubSubEndpoint != "" {
		_ = os.Setenv("PUBSUB_EMULATOR_HOST", cfg.PubSubEndpoint)
		client, err = pubsub.NewClient(ctx, cfg.ProjectID, option.WithEndpoint(cfg.PubSubEndpoint))
	} else {
		client, err = pubsub.NewClient(ctx, cfg.ProjectID)
	}

	if err != nil {
		return nil, fmt.Errorf("create pubsub client: %w", err)
	}

	topic := client.Topic(topicID)
	exists, err := topic.Exists(ctx)
	if err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("check topic exists: %w", err)
	}

	if !exists {
		topic, err = client.CreateTopic(ctx, topicID)
		if err != nil {
			_ = client.Close()
			return nil, fmt.Errorf("create topic: %w", err)
		}
	}

	return &Dispatcher{
		pool:            pool,
		client:          client,
		topic:           topic,
		projectID:       cfg.ProjectID,
		intervalSeconds: cfg.IntervalSeconds,
	}, nil
}

func (d *Dispatcher) Close() {
	d.topic.Stop()
	_ = d.client.Close()
}

func (d *Dispatcher) Start(ctx context.Context) error {
	logger := slog.Default()

	interval := d.intervalSeconds
	if interval <= 0 {
		interval = 900
	}
	tickerInterval := time.Duration(interval) * time.Second

	logger.Info("dispatcher starting", "interval_seconds", interval)

	ticker := time.NewTicker(tickerInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.Info("dispatcher stopping")
			return ctx.Err()
		case <-ticker.C:
			if err := d.dispatchLoop(ctx); err != nil {
				logger.Error("dispatch loop failed", "error", err)
			}
		}
	}
}

func (d *Dispatcher) dispatchLoop(ctx context.Context) error {
	logger := slog.With("loop", "discovery")

	sources, err := d.getActiveSources(ctx)
	if err != nil {
		return fmt.Errorf("get active sources: %w", err)
	}

	if len(sources) == 0 {
		logger.Debug("no active sources for discovery")
		return nil
	}

	logger.Info("publishing discovery intents", "count", len(sources))

	dispatched := 0

	for _, source := range sources {
		task := DiscoveryTask{
			SourceID:  source.ID,
			URL:       source.URL,
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		}

		if err := d.publishDiscovery(ctx, task); err != nil {
			logger.Error("failed to publish discovery", "source_id", source.ID, "error", err)
			continue
		}

		if err := d.updateLastRunAt(ctx, source.ID); err != nil {
			logger.Error("failed to update last_run_at", "source_id", source.ID, "error", err)
			continue
		}

		dispatched++
	}

	logger.Info("discovery loop completed", "dispatched", dispatched, "total", len(sources))

	return nil
}

func (d *Dispatcher) getActiveSources(ctx context.Context) ([]models.SourceForFetch, error) {
	rows, err := d.pool.Query(ctx,
		`SELECT id, url, fetch_strategy 
		 FROM sources 
		 WHERE url IS NOT NULL 
		 AND (
			 -- Fuentes nunca ejecutadas (nuevas)
			 last_run_at IS NULL 
			 -- Fuentes fuera de su intervalo normal  
			 OR last_run_at + (interval_minutes || ' minutes')::interval < NOW()
		 )
		 ORDER BY 
			 -- Prioridad 1: Fuentes nunca ejecutadas (nuevas)
			 CASE WHEN last_run_at IS NULL THEN 1 ELSE 2 END,
			 -- Prioridad 2: Las más antiguas primero entre las que ya se ejecutaron
			 last_run_at ASC NULLS FIRST
		 LIMIT 100`,
	)
	if err != nil {
		return nil, fmt.Errorf("query active sources: %w", err)
	}
	defer rows.Close()

	var sources []models.SourceForFetch
	for rows.Next() {
		var s models.SourceForFetch
		if err := rows.Scan(&s.ID, &s.URL, &s.FetchStrategy); err != nil {
			return nil, fmt.Errorf("scan source: %w", err)
		}
		sources = append(sources, s)
	}

	return sources, rows.Err()
}

func (d *Dispatcher) publishDiscovery(ctx context.Context, task DiscoveryTask) error {
	data, err := json.Marshal(task)
	if err != nil {
		return fmt.Errorf("marshal task: %w", err)
	}

	msg := &pubsub.Message{
		Data: data,
	}

	result := d.topic.Publish(ctx, msg)
	_, err = result.Get(ctx)
	if err != nil {
		return fmt.Errorf("publish message: %w", err)
	}

	return nil
}

func (d *Dispatcher) updateLastRunAt(ctx context.Context, sourceID string) error {
	_, err := d.pool.Exec(ctx,
		`UPDATE sources SET last_run_at = NOW() WHERE id = $1`,
		sourceID,
	)
	if err != nil {
		return fmt.Errorf("update last_run_at: %w", err)
	}
	return nil
}
