package subscriber

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"sync"
	"time"

	"cloud.google.com/go/pubsub"
	"google.golang.org/api/option"

	"phosboard/workers/social_probe/internal/publisher"
	"phosboard/workers/social_probe/internal/scraper"
	"phosboard/workers/social_probe/internal/storage"
)

type SocialProbeTask struct {
	DocumentID    string   `json:"document_id"`
	SearchQueries []string `json:"search_queries"`
}

type Worker struct {
	scraper   scraper.SocialScraper
	storage   *storage.GCSStorage
	publisher *publisher.Publisher
}

func NewWorker(
	scraper scraper.SocialScraper,
	storage *storage.GCSStorage,
	pub *publisher.Publisher,
) *Worker {
	return &Worker{
		scraper:   scraper,
		storage:   storage,
		publisher: pub,
	}
}

func (w *Worker) Run(ctx context.Context) error {
	logger := slog.Default()

	projectID := os.Getenv("GOOGLE_PROJECT_ID")
	if projectID == "" {
		projectID = "phosboard"
	}

	pubsubEndpoint := os.Getenv("PUBSUB_EMULATOR_HOST")

	var client *pubsub.Client
	var err error
	if pubsubEndpoint != "" {
		os.Setenv("PUBSUB_EMULATOR_HOST", pubsubEndpoint)
		client, err = pubsub.NewClient(ctx, projectID, option.WithEndpoint(pubsubEndpoint))
	} else {
		client, err = pubsub.NewClient(ctx, projectID)
	}
	if err != nil {
		return fmt.Errorf("create pubsub client: %w", err)
	}
	defer client.Close()

	sub := client.Subscription("social-probe-sub")
	logger.Info("listening on social-probe-sub")

	err = sub.Receive(ctx, func(ctx context.Context, msg *pubsub.Message) {
		if err := w.processMessage(ctx, msg); err != nil {
			logger.Error("failed to process message", "error", err)
			msg.Nack()
		} else {
			msg.Ack()
		}
	})

	return err
}

func (w *Worker) processMessage(ctx context.Context, msg *pubsub.Message) error {
	logger := slog.With("component", "message_processor")

	var task SocialProbeTask
	if err := json.Unmarshal(msg.Data, &task); err != nil {
		return fmt.Errorf("unmarshal message: %w", err)
	}

	logger = logger.With("document_id", task.DocumentID, "queries", len(task.SearchQueries))
	logger.InfoContext(ctx, "received social-probe task")

	var wg sync.WaitGroup
	var mu sync.Mutex
	allMentions := make([]scraper.SocialMention, 0)

	for _, query := range task.SearchQueries {
		wg.Add(1)
		go func(q string) {
			defer wg.Done()

			mentions, err := w.scraper.Scrape(ctx, q)
			if err != nil {
				logger.Error("failed to scrape query", "query", q, "error", err)
				return
			}

			mu.Lock()
			allMentions = append(allMentions, mentions...)
			mu.Unlock()
		}(query)
	}

	wg.Wait()

	logger.InfoContext(ctx, "scraped mentions", "total", len(allMentions))

	result := map[string]interface{}{
		"document_id": task.DocumentID,
		"queries":     task.SearchQueries,
		"mentions":    allMentions,
		"scraped_at":  time.Now().UTC().Format(time.RFC3339),
	}

	jsonData, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("marshal result: %w", err)
	}

	timestamp := time.Now().UTC().Format("20060102T150405Z")
	gcsKey := fmt.Sprintf("mentions/%s/%s.json", task.DocumentID, timestamp)

	if err := w.storage.PutObject(ctx, gcsKey, jsonData); err != nil {
		return fmt.Errorf("save to GCS: %w", err)
	}

	logger.InfoContext(ctx, "saved mentions to GCS", "key", gcsKey)

	if err := w.publisher.PublishClimateAggregate(ctx, task.DocumentID, gcsKey); err != nil {
		return fmt.Errorf("publish climate-aggregate: %w", err)
	}

	logger.InfoContext(ctx, "published to climate-aggregate", "key", gcsKey)

	return nil
}
