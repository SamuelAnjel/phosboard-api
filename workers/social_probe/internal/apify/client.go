package apify

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

type SocialMention struct {
	Text            string `json:"text"`
	Author          string `json:"author"`
	Date            string `json:"date"`
	Platform        string `json:"platform"`
	EngagementScore int    `json:"engagement_score"`
}

type ApifyClient struct {
	apiToken string
	actorID  string
	client   *http.Client
}

type ApifyRunRequest struct {
	StartURLs []ApifyStartURL `json:"startUrls"`
}

type ApifyStartURL struct {
	URL string `json:"url"`
}

type ApifyRunResponse struct {
	ID string `json:"id"`
}

type ApifyDatasetItem struct {
	URL         string `json:"url"`
	Text        string `json:"text"`
	Author      string `json:"author"`
	Date        string `json:"date"`
	Platform    string `json:"platform"`
	Likes       int    `json:"likes"`
	Retweets    int    `json:"retweets"`
	Replies     int    `json:"replies"`
	Quotes      int    `json:"quotes"`
	Bookmarks   int    `json:"bookmarks"`
	Impressions int    `json:"impressions"`
}

func NewApifyClient(apiToken, actorID string) *ApifyClient {
	return &ApifyClient{
		apiToken: apiToken,
		actorID:  actorID,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func NewApifyClientFromEnv() (*ApifyClient, error) {
	apiToken := os.Getenv("APIFY_API_TOKEN")
	actorID := os.Getenv("APIFY_ACTOR_ID")

	if apiToken == "" || actorID == "" {
		return nil, fmt.Errorf("APIFY_API_TOKEN and APIFY_ACTOR_ID environment variables are required")
	}

	return NewApifyClient(apiToken, actorID), nil
}

func (c *ApifyClient) Scrape(ctx context.Context, query string) ([]SocialMention, error) {
	// Create search URL for Twitter/X
	searchURL := fmt.Sprintf("https://twitter.com/search?q=%s&src=typed_query", query)

	// Start Apify actor run
	runID, err := c.startActorRun(ctx, searchURL)
	if err != nil {
		return nil, fmt.Errorf("failed to start Apify actor: %w", err)
	}

	// Wait for completion and get results
	items, err := c.waitForResults(ctx, runID)
	if err != nil {
		return nil, fmt.Errorf("failed to get Apify results: %w", err)
	}

	// Convert to SocialMention format
	mentions := make([]SocialMention, 0, len(items))
	for _, item := range items {
		engagementScore := calculateEngagementScore(item)

		mentions = append(mentions, SocialMention{
			Text:            item.Text,
			Author:          item.Author,
			Date:            item.Date,
			Platform:        item.Platform,
			EngagementScore: engagementScore,
		})
	}

	return mentions, nil
}

func (c *ApifyClient) startActorRun(ctx context.Context, searchURL string) (string, error) {
	reqBody := ApifyRunRequest{
		StartURLs: []ApifyStartURL{
			{URL: searchURL},
		},
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("https://api.apify.com/v2/acts/%s/runs?token=%s", c.actorID, c.apiToken)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("Apify API error: %s - %s", resp.Status, string(body))
	}

	var runResp ApifyRunResponse
	if err := json.NewDecoder(resp.Body).Decode(&runResp); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	return runResp.ID, nil
}

func (c *ApifyClient) waitForResults(ctx context.Context, runID string) ([]ApifyDatasetItem, error) {
	// Poll for completion (simplified - in production would use Webhooks or longer polling)
	time.Sleep(30 * time.Second)

	// Get dataset items
	url := fmt.Sprintf("https://api.apify.com/v2/actor-runs/%s/dataset/items?token=%s", runID, c.apiToken)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Apify API error: %s - %s", resp.Status, string(body))
	}

	var items []ApifyDatasetItem
	if err := json.NewDecoder(resp.Body).Decode(&items); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return items, nil
}

func calculateEngagementScore(item ApifyDatasetItem) int {
	// Simple engagement score calculation
	score := item.Likes*1 + item.Retweets*2 + item.Replies*3 + item.Quotes*2 + item.Bookmarks*1
	if score == 0 {
		score = 10 // Default minimum score
	}
	return score
}
