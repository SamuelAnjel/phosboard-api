package scraper

import (
	"context"

	"phosboard/workers/social_probe/internal/apify"
)

type ApifyScraper struct {
	client *apify.ApifyClient
}

func NewApifyScraper() (*ApifyScraper, error) {
	client, err := apify.NewApifyClientFromEnv()
	if err != nil {
		return nil, err
	}
	return &ApifyScraper{client: client}, nil
}

func (s *ApifyScraper) Scrape(ctx context.Context, query string) ([]SocialMention, error) {
	apifyMentions, err := s.client.Scrape(ctx, query)
	if err != nil {
		return nil, err
	}

	// Convert apify.SocialMention to scraper.SocialMention
	mentions := make([]SocialMention, len(apifyMentions))
	for i, am := range apifyMentions {
		mentions[i] = SocialMention{
			Text:            am.Text,
			Author:          am.Author,
			Date:            am.Date,
			Platform:        am.Platform,
			EngagementScore: am.EngagementScore,
		}
	}

	return mentions, nil
}
