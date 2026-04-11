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
	MaxDepth      int      `json:"maxDepth"`
	MaxPages      int      `json:"maxPages"`
	SameDomain    bool     `json:"sameDomain"`
	IncludePaths  []string `json:"includePaths"`
	ExcludePaths  []string `json:"excludePaths"`
	RespectRobots bool     `json:"respectRobotsTxt"`
	CrawlDelayMS  int      `json:"delayMs"`
}

type CreateSourceRequest struct {
	Name        string       `json:"name"`
	Type        string       `json:"type"`
	URL         string       `json:"url"`
	MaxLinks    int          `json:"max_links"`
	Crawl       *CrawlConfig `json:"crawl,omitempty"`
	CrawlConfig *CrawlConfig `json:"crawlConfig,omitempty"`
}

type UpdateSourceConfigRequest struct {
	Config map[string]interface{} `json:"config"`
}
