package handler

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"phosboard/workers/social_probe/internal/publisher"
	"phosboard/workers/social_probe/internal/scraper"
	"phosboard/workers/social_probe/internal/storage"
)

type PubSubMessage struct {
	Message struct {
		Data        string            `json:"data"`
		MessageID   string            `json:"messageId"`
		PublishTime string            `json:"publishTime"`
		Attributes  map[string]string `json:"attributes"`
	} `json:"message"`
	Subscription string `json:"subscription"`
}

type SocialProbeTask struct {
	DocumentID    string   `json:"document_id"`
	SearchQueries []string `json:"search_queries"`
}

type Handler struct {
	scraper   scraper.SocialScraper
	storage   *storage.GCSStorage
	publisher *publisher.Publisher
	logger    *slog.Logger
}

func NewHandler(scraper scraper.SocialScraper, storage *storage.GCSStorage, pub *publisher.Publisher) *Handler {
	return &Handler{
		scraper:   scraper,
		storage:   storage,
		publisher: pub,
		logger:    slog.Default(),
	}
}

func (h *Handler) HandlePubSubPush(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := h.logger.With("component", "pubsub_handler")

	if r.Method != http.MethodPost {
		logger.Error("invalid HTTP method", "method", r.Method)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var pubsubMsg PubSubMessage
	if err := json.NewDecoder(r.Body).Decode(&pubsubMsg); err != nil {
		logger.Error("failed to decode Pub/Sub message", "error", err)
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	data, err := base64.StdEncoding.DecodeString(pubsubMsg.Message.Data)
	if err != nil {
		logger.Error("failed to decode base64 data", "error", err)
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	var task SocialProbeTask
	if err := json.Unmarshal(data, &task); err != nil {
		logger.Error("failed to unmarshal task", "error", err)
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	logger = logger.With(
		"message_id", pubsubMsg.Message.MessageID,
		"document_id", task.DocumentID,
		"queries", len(task.SearchQueries),
	)

	logger.Info("received social-probe task")

	go func() {
		if err := h.processTask(ctx, task); err != nil {
			logger.Error("failed to process task", "error", err)
		} else {
			logger.Info("task processed successfully")
		}
	}()

	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte("OK")); err != nil {
		logger.Error("failed to write response", "error", err)
	}
}

func (h *Handler) processTask(ctx context.Context, task SocialProbeTask) error {
	logger := h.logger.With("component", "task_processor", "document_id", task.DocumentID)

	var wg sync.WaitGroup
	var mu sync.Mutex
	allMentions := make([]scraper.SocialMention, 0)

	for _, query := range task.SearchQueries {
		wg.Add(1)
		go func(q string) {
			defer wg.Done()

			mentions, err := h.scraper.Scrape(ctx, q)
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

	logger.Info("scraped mentions", "total", len(allMentions))

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

	if err := h.storage.PutObject(ctx, gcsKey, jsonData); err != nil {
		return fmt.Errorf("save to GCS: %w", err)
	}

	logger.Info("saved mentions to GCS", "key", gcsKey)

	if err := h.publisher.PublishClimateAggregate(ctx, task.DocumentID, gcsKey); err != nil {
		return fmt.Errorf("publish climate-aggregate: %w", err)
	}

	logger.Info("published to climate-aggregate", "key", gcsKey)

	return nil
}
