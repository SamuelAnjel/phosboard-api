package models

import (
	"encoding/json"
	"time"
)

type GlobalDocument struct {
	ID               string          `json:"id"`
	SourceID         string          `json:"source_id"`
	Title            string          `json:"title"`
	URL              string          `json:"url"`
	ContentText      string          `json:"content_text"`
	RawPayload       json.RawMessage `json:"raw_payload,omitempty"`
	ContentEmbedding []float64       `json:"content_embedding,omitempty"`
	CreatedAt        time.Time       `json:"created_at"`
	UpdatedAt        time.Time       `json:"updated_at,omitempty"`
}

type DocumentWithSource struct {
	ID         string `json:"id"`
	Title      string `json:"title"`
	URL        string `json:"url"`
	SourceName string `json:"source_name"`
}

type DocumentWithAnalysis struct {
	ID                string          `json:"id"`
	Title             string          `json:"title"`
	URL               string          `json:"url"`
	SourceName        string          `json:"source_name"`
	SemanticAnalysis  json.RawMessage `json:"semantic_analysis,omitempty"`
	SocialTemperature *float64        `json:"social_temperature,omitempty"`
	CreatedAt         time.Time       `json:"created_at"`
}
