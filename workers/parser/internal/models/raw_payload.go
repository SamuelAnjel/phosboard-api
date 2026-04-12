package models

import (
	"time"
)

type RawPayload struct {
	ID          string    `parquet:"id"`
	SourceID    string    `parquet:"source_id"`
	URL         string    `parquet:"url"`
	HTMLContent string    `parquet:"html_content"`
	CrawledAt   time.Time `parquet:"crawled_at"`
}
