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
	URL       string          `json:"url"`
	Config    json.RawMessage `json:"config"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

type CrawlConfig struct {
	MaxDepth      int      `json:"max_depth"`
	MaxPages      int      `json:"max_pages"`
	SameDomain    bool     `json:"same_domain"`
	IncludePaths  []string `json:"include_paths"`
	ExcludePaths  []string `json:"exclude_paths"`
	RespectRobots bool     `json:"respect_robots"`
	CrawlDelayMS  int      `json:"crawl_delay_ms"`
}

type CreateSourceRequest struct {
	Name     string       `json:"name"`
	Type     string       `json:"type"`
	URL      string       `json:"url"`
	MaxLinks int          `json:"max_links"`
	Crawl    *CrawlConfig `json:"crawl,omitempty"`
}

type UpdateSourceConfigRequest struct {
	Config map[string]interface{} `json:"config"`
}
