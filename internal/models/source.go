package models

import (
	"encoding/json"
	"time"
)

type SourceForFetch struct {
	ID            string
	URL           string
	FetchStrategy string
}

type Source struct {
	ID        string          `json:"id"`
	Name      string          `json:"name"`
	Type      string          `json:"type"`
	Config    json.RawMessage `json:"config"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

type CreateSourceRequest struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	MaxLinks int    `json:"max_links"`
}

type UpdateSourceConfigRequest struct {
	Config map[string]interface{} `json:"config"`
}
