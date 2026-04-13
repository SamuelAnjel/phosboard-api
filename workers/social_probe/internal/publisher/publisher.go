package publisher

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
)

const topicID = "climate-aggregate"

type ClimateAggregateTask struct {
	DocumentID     string    `json:"document_id"`
	GCSMentionsKey string    `json:"gcs_mentions_key"`
	Timestamp      time.Time `json:"timestamp"`
}

type Publisher struct {
	projectID      string
	pubsubEndpoint string
	client         *pubsub.Client
	topic          *pubsub.Topic
	mu             sync.Mutex
	initialized    bool
}

func NewPublisher(ctx context.Context, projectID, pubsubEndpoint string) (*Publisher, error) {
	logger := slog.With("component", "publisher")

	logger.Info("publisher initialized (lazy mode)", "topic", topicID)

	return &Publisher{
		projectID:      projectID,
		pubsubEndpoint: pubsubEndpoint,
		initialized:    false,
	}, nil
}

func (p *Publisher) init(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.initialized {
		return nil
	}

	logger := slog.With("component", "publisher")

	var client *pubsub.Client
	var err error

	if p.pubsubEndpoint != "" {
		os.Setenv("PUBSUB_EMULATOR_HOST", p.pubsubEndpoint)
		client, err = pubsub.NewClient(ctx, p.projectID, option.WithEndpoint(p.pubsubEndpoint))
	} else {
		client, err = pubsub.NewClient(ctx, p.projectID)
	}

	if err != nil {
		return fmt.Errorf("create pubsub client: %w", err)
	}

	p.client = client
	p.topic = client.Topic(topicID)
	p.initialized = true

	logger.Info("pubsub client initialized", "topic", topicID)

	return nil
}

func (p *Publisher) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.topic != nil {
		p.topic.Stop()
	}
	if p.client != nil {
		p.client.Close()
	}
}

func (p *Publisher) PublishClimateAggregate(ctx context.Context, documentID, gcsMentionsKey string) error {
	if err := p.init(ctx); err != nil {
		return fmt.Errorf("initialize publisher: %w", err)
	}

	task := ClimateAggregateTask{
		DocumentID:     documentID,
		GCSMentionsKey: gcsMentionsKey,
		Timestamp:      time.Now().UTC(),
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
