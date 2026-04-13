package processor

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"regexp"
	"strings"
	"time"

	"cloud.google.com/go/storage"
	"github.com/parquet-go/parquet-go"

	"phosboard/workers/parser/internal/models"
	"phosboard/workers/parser/internal/repository"
)

var (
	tagRE   = regexp.MustCompile(`<[^>]+>`)
	spaceRE = regexp.MustCompile(`\s+`)
)

type Config struct {
	GCSBucket string
	TenantID  string
}

type DocumentProcessor struct {
	gcsClient *storage.Client
	bucket    string
	tenantID  string
	repo      repository.DocumentRepository
}

func NewDocumentProcessor(ctx context.Context, cfg Config, repo repository.DocumentRepository) (*DocumentProcessor, error) {
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("create gcs client: %w", err)
	}

	return &DocumentProcessor{
		gcsClient: client,
		bucket:    cfg.GCSBucket,
		tenantID:  cfg.TenantID,
		repo:      repo,
	}, nil
}

func (p *DocumentProcessor) ProcessFile(ctx context.Context, objectName string) (int, error) {
	logger := slog.With("object", objectName)

	logger.InfoContext(ctx, "downloading file from GCS")

	gcsReader, err := p.gcsClient.Bucket(p.bucket).Object(objectName).NewReader(ctx)
	if err != nil {
		return 0, fmt.Errorf("get object from gcs: %w", err)
	}
	defer gcsReader.Close()

	tmpFile, err := os.CreateTemp("", "parquet-*.parquet")
	if err != nil {
		return 0, fmt.Errorf("create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := io.Copy(tmpFile, gcsReader); err != nil {
		tmpFile.Close()
		return 0, fmt.Errorf("copy to temp: %w", err)
	}
	tmpFile.Close()

	f, err := os.Open(tmpFile.Name())
	if err != nil {
		return 0, fmt.Errorf("open temp file: %w", err)
	}
	defer f.Close()

	reader := parquet.NewReader(f)
	defer reader.Close()

	count := 0

	for {
		var payload models.RawPayload
		if err := reader.Read(&payload); err != nil {
			if err == io.EOF {
				break
			}
			return count, fmt.Errorf("read parquet row: %w", err)
		}

		contentText := ExtractText(payload.HTMLContent)

		docID, err := p.insertDocument(ctx, payload.SourceID, payload.URL, payload.ID, contentText)
		if err != nil {
			logger.ErrorContext(ctx, "failed to insert document", "url", payload.URL, "error", err)
			continue
		}

		if docID != "" {
			if err := p.linkToTenant(ctx, docID); err != nil {
				logger.ErrorContext(ctx, "failed to link document to tenant", "doc_id", docID, "error", err)
			}
		}

		count++
	}

	logger.InfoContext(ctx, "file processed", "documents", count)

	return count, nil
}

func (p *DocumentProcessor) insertDocument(ctx context.Context, sourceID, url, title, contentText string) (string, error) {
	rawPayload := map[string]interface{}{
		"source_id":   sourceID,
		"crawled_url": url,
	}
	rawPayloadJSON, _ := json.Marshal(rawPayload)

	doc := repository.GlobalDocument{
		SourceID:    sourceID,
		Title:       title,
		URL:         url,
		ContentText: contentText,
		RawPayload:  rawPayloadJSON,
		CreatedAt:   time.Now(),
	}

	return p.repo.InsertGlobalDocument(ctx, doc)
}

func (p *DocumentProcessor) linkToTenant(ctx context.Context, docID string) error {
	return p.repo.LinkDocumentToTenant(ctx, p.tenantID, docID, nil)
}

func ExtractText(html string) string {
	text := tagRE.ReplaceAllString(html, " ")
	text = spaceRE.ReplaceAllString(text, " ")
	return strings.TrimSpace(text)
}
