package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"cloud.google.com/go/pubsub" //nolint:staticcheck
)

func main() {
	ctx := context.Background()

	projectID := os.Getenv("GOOGLE_PROJECT_ID")
	if projectID == "" {
		projectID = "phosboard"
	}

	// Use emulator only if PUBSUB_EMULATOR_HOST is set
	if emulatorHost := os.Getenv("PUBSUB_EMULATOR_HOST"); emulatorHost != "" {
		_ = os.Setenv("PUBSUB_EMULATOR_HOST", emulatorHost)
	}

	client, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer func() { _ = client.Close() }()

	topics := []string{
		"source-discovery",
		"url-scrape",
		"document-analyze",
		"social-probe",
		"climate-aggregate",
	}

	// Push endpoints for Cloud Run services
	pushEndpoints := map[string]string{
		"source-discovery-sub":  "https://worker-discovery-544990213867.us-east1.run.app/",
		"url-scrape-sub":        "https://worker-scraper-544990213867.us-east1.run.app/",
		"document-analyze-sub":  "https://worker-semantic-544990213867.us-east1.run.app/",
		"social-probe-sub":      "https://worker-social-probe-544990213867.us-east1.run.app/",
		"climate-aggregate-sub": "https://worker-climate-aggregate-544990213867.us-east1.run.app/",
	}

	oldTopic := client.Topic("fetcher-tasks")
	_ = oldTopic.Delete(ctx)
	log.Printf("Deleted old topic: fetcher-tasks")

	oldDLQ := client.Topic("fetcher-tasks-dead-letter")
	_ = oldDLQ.Delete(ctx)
	log.Printf("Deleted old dead letter topic: fetcher-tasks-dead-letter")

	for _, topicName := range topics {
		_, err = client.CreateTopic(ctx, topicName)
		if err != nil {
			log.Printf("Topic %s might already exist: %v", topicName, err)
		} else {
			log.Printf("Created topic: %s", topicName)
		}

		dlqTopicName := topicName + "-dead-letter"
		_, err = client.CreateTopic(ctx, dlqTopicName)
		if err != nil {
			log.Printf("Dead letter topic %s might already exist: %v", dlqTopicName, err)
		} else {
			log.Printf("Created dead letter topic: %s", dlqTopicName)
		}
	}

	for _, topicName := range topics {
		subName := topicName + "-sub"
		dlqTopic := "projects/" + projectID + "/topics/" + topicName + "-dead-letter"

		sub := client.Subscription(subName)
		exists, err := sub.Exists(ctx)
		if err != nil {
			log.Printf("Failed to check subscription %s existence: %v", subName, err)
			continue
		}

		// Check if this subscription should be push
		pushEndpoint, isPush := pushEndpoints[subName]

		if exists {
			config, err := sub.Config(ctx)
			if err != nil {
				log.Printf("Failed to get subscription %s config: %v", subName, err)
				continue
			}

			// Update DLQ and retry policy
			config.DeadLetterPolicy = &pubsub.DeadLetterPolicy{
				DeadLetterTopic:     dlqTopic,
				MaxDeliveryAttempts: 5,
			}
			config.RetryPolicy = &pubsub.RetryPolicy{
				MinimumBackoff: 10 * time.Second,
				MaximumBackoff: 600 * time.Second,
			}

			// Set push config if applicable
			if isPush {
				config.PushConfig = pubsub.PushConfig{
					Endpoint: pushEndpoint,
				}
				log.Printf("Setting push endpoint for %s: %s", subName, pushEndpoint)
			}

			updateConfig := pubsub.SubscriptionConfigToUpdate{
				DeadLetterPolicy: config.DeadLetterPolicy,
				RetryPolicy:      config.RetryPolicy,
			}
			if isPush {
				updateConfig.PushConfig = &config.PushConfig
			}

			_, err = sub.Update(ctx, updateConfig)
			if err != nil {
				log.Printf("Failed to update subscription %s: %v", subName, err)
			} else {
				log.Printf("Updated subscription %s with DLQ and retry policy", subName)
				if isPush {
					log.Printf("  - Push endpoint: %s", pushEndpoint)
				}
			}
		} else {
			// Create new subscription
			subConfig := pubsub.SubscriptionConfig{
				Topic: client.Topic(topicName),
				DeadLetterPolicy: &pubsub.DeadLetterPolicy{
					DeadLetterTopic:     dlqTopic,
					MaxDeliveryAttempts: 5,
				},
				RetryPolicy: &pubsub.RetryPolicy{
					MinimumBackoff: 10 * time.Second,
					MaximumBackoff: 600 * time.Second,
				},
			}

			// Set push config if applicable
			if isPush {
				subConfig.PushConfig = pubsub.PushConfig{
					Endpoint: pushEndpoint,
				}
				log.Printf("Creating push subscription %s with endpoint: %s", subName, pushEndpoint)
			}

			_, err = client.CreateSubscription(ctx, subName, subConfig)
			if err != nil {
				log.Printf("Failed to create subscription %s: %v", subName, err)
			} else {
				log.Printf("Created subscription %s with DLQ", subName)
				if isPush {
					log.Printf("  - Push endpoint: %s", pushEndpoint)
				}
			}
		}
	}

	fmt.Println("Done!")
}
