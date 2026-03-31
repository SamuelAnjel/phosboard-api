package dispatcher

import (
	"encoding/json"
	"testing"

	"phosboard/backend/internal/models"
)

func TestDiscoveryTask_JSONMarshal(t *testing.T) {
	task := DiscoveryTask{
		SourceID:  "source-123",
		URL:       "https://example.com/feed.xml",
		Timestamp: "2024-01-01T12:00:00Z",
	}

	data, err := json.Marshal(task)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded DiscoveryTask
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if decoded.SourceID != task.SourceID {
		t.Errorf("expected source_id %s, got %s", task.SourceID, decoded.SourceID)
	}
	if decoded.URL != task.URL {
		t.Errorf("expected url %s, got %s", task.URL, decoded.URL)
	}
	if decoded.Timestamp != task.Timestamp {
		t.Errorf("expected timestamp %s, got %s", task.Timestamp, decoded.Timestamp)
	}
}

func TestConstants(t *testing.T) {
	if topicID == "" {
		t.Error("topicID should not be empty")
	}
	if tickerInterval == 0 {
		t.Error("tickerInterval should not be zero")
	}
}

func TestSourceForFetch_Fields(t *testing.T) {
	source := models.SourceForFetch{
		ID:            "test-id",
		URL:           "https://example.com",
		FetchStrategy: "rss",
	}

	if source.ID == "" {
		t.Error("ID should not be empty")
	}
	if source.URL == "" {
		t.Error("URL should not be empty")
	}
	if source.FetchStrategy != "rss" {
		t.Errorf("expected strategy 'rss', got %s", source.FetchStrategy)
	}
}
