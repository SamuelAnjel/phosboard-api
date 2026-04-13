package scraper

import (
	"context"
	"time"
)

type SocialMention struct {
	Text            string `json:"text"`
	Author          string `json:"author"`
	Date            string `json:"date"`
	Platform        string `json:"platform"`
	EngagementScore int    `json:"engagement_score"`
}

type SocialScraper interface {
	Scrape(ctx context.Context, query string) ([]SocialMention, error)
}

type MockScraper struct{}

func NewMockScraper() *MockScraper {
	return &MockScraper{}
}

func (s *MockScraper) Scrape(ctx context.Context, query string) ([]SocialMention, error) {
	count := 3 + len(query)%3
	mentions := make([]SocialMention, count)

	platforms := []string{"twitter", "bluesky", "facebook", "instagram"}

	for i := 0; i < count; i++ {
		mentions[i] = SocialMention{
			Text:            "This is a mock mention about: " + query + " #" + query,
			Author:          "user_" + string(rune('a'+i)) + "_handle",
			Date:            time.Now().Add(-time.Hour * time.Duration(i*24)).Format(time.RFC3339),
			Platform:        platforms[i%len(platforms)],
			EngagementScore: 100 + (i * 50),
		}
	}

	return mentions, nil
}
