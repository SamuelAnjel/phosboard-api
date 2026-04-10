//nolint:staticcheck
package publisher

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"cloud.google.com/go/pubsub"
	"google.golang.org/api/option"
)

const topicID = "url-scrape"

type URLScrapeTask struct {
	DocumentID string `json:"document_id"`
	URL        string `json:"url"`
}

type Publisher struct {
	client *pubsub.Client
	topic  *pubsub.Topic
}

func NewPublisher(ctx context.Context, projectID, pubsubEndpoint string) (*Publisher, error) {
	logger := slog.With("component", "publisher")

	var client *pubsub.Client
	var err error

	if pubsubEndpoint != "" {
		_ = os.Setenv("PUBSUB_EMULATOR_HOST", pubsubEndpoint)
		client, err = pubsub.NewClient(ctx, projectID, option.WithEndpoint(pubsubEndpoint))
	} else {
		client, err = pubsub.NewClient(ctx, projectID)
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

	logger.Info("publisher initialized", "topic", topicID)

	return &Publisher{
		client: client,
		topic:  topic,
	}, nil
}

func (p *Publisher) Close() {
	p.topic.Stop()
	_ = p.client.Close()
}

func (p *Publisher) PublishURLScrape(ctx context.Context, documentID, url string) error {
	task := URLScrapeTask{
		DocumentID: documentID,
		URL:        url,
	}

	data, err := json.Marshal(task)
	if err != nil {
		return fmt.Errorf("marshal task: %w", err)
	}

	msg := &pubsub.Message{
		Data: data,
	}

	result := p.topic.Publish(ctx, msg)
	_, err = result.Get(ctx)
	if err != nil {
		return fmt.Errorf("publish message: %w", err)
	}

	return nil
}
