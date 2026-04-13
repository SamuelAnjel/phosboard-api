package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type GlobalDocument struct {
	ID          string          `json:"id"`
	SourceID    string          `json:"source_id"`
	Title       string          `json:"title"`
	URL         string          `json:"url"`
	ContentText string          `json:"content_text"`
	RawPayload  json.RawMessage `json:"raw_payload,omitempty"`
	CreatedAt   time.Time       `json:"created_at"`
}

type DocumentRepository interface {
	InsertGlobalDocument(ctx context.Context, doc GlobalDocument) (string, error)
	LinkDocumentToTenant(ctx context.Context, tenantID, docID string, matchedKeywords []string) error
}

type PostgresDocumentRepository struct {
	pool *pgxpool.Pool
}

func NewDocumentRepository(ctx context.Context, databaseURL string) (*PostgresDocumentRepository, error) {
	poolConfig, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}

	return &PostgresDocumentRepository{pool: pool}, nil
}

func (r *PostgresDocumentRepository) Close() {
	r.pool.Close()
}

func (r *PostgresDocumentRepository) InsertGlobalDocument(ctx context.Context, doc GlobalDocument) (string, error) {
	var id string
	err := r.pool.QueryRow(ctx,
		`INSERT INTO global_documents (source_id, title, url, content_text, raw_payload, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, COALESCE($6, NOW()), COALESCE($6, NOW()))
		 ON CONFLICT (url) DO UPDATE SET title = EXCLUDED.title, content_text = EXCLUDED.content_text, updated_at = NOW()
		 RETURNING id`,
		doc.SourceID,
		doc.Title,
		doc.URL,
		doc.ContentText,
		doc.RawPayload,
		doc.CreatedAt,
	).Scan(&id)
	if err != nil {
		return "", fmt.Errorf("failed to insert global document: %w", err)
	}
	return id, nil
}

func (r *PostgresDocumentRepository) LinkDocumentToTenant(ctx context.Context, tenantID, docID string, matchedKeywords []string) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO tenant_documents (tenant_id, document_id, matched_keywords, created_at, updated_at)
		 VALUES ($1, $2, $3, NOW(), NOW())
		 ON CONFLICT (tenant_id, document_id) DO UPDATE SET matched_keywords = EXCLUDED.matched_keywords, updated_at = NOW()`,
		tenantID,
		docID,
		matchedKeywords,
	)
	if err != nil {
		return fmt.Errorf("failed to link document to tenant: %w", err)
	}
	return nil
}
