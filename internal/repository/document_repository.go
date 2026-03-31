package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"phosboard/backend/internal/models"
)

type DocumentRepository interface {
	InsertGlobalDocument(ctx context.Context, doc models.GlobalDocument) (string, error)
	LinkDocumentToTenant(ctx context.Context, tenantID, docID string, matchedKeywords []string) error
	GetLatestByTenant(ctx context.Context, tenantID string) ([]models.DocumentWithSource, error)
	GetDocumentsByTenant(ctx context.Context, tenantID string, limit, offset int) ([]models.DocumentWithAnalysis, int, error)
	GetOrCreateSource(ctx context.Context, name, sourceType string) (string, error)
	TrackDocument(ctx context.Context, tenantID, url, sourceType string, priority int) (string, error)
}

type PostgresDocumentRepository struct {
	pool *pgxpool.Pool
}

func NewDocumentRepository(pool *pgxpool.Pool) *PostgresDocumentRepository {
	return &PostgresDocumentRepository{pool: pool}
}

func (r *PostgresDocumentRepository) InsertGlobalDocument(ctx context.Context, doc models.GlobalDocument) (string, error) {
	var id string
	err := r.pool.QueryRow(ctx,
		`INSERT INTO global_documents (source_id, title, url, content_text, raw_payload, content_embedding, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, COALESCE($7, NOW()), COALESCE($7, NOW()))
		 ON CONFLICT (url) DO UPDATE SET title = EXCLUDED.title, content_text = EXCLUDED.content_text, updated_at = NOW()
		 RETURNING id`,
		doc.SourceID,
		doc.Title,
		doc.URL,
		doc.ContentText,
		doc.RawPayload,
		doc.ContentEmbedding,
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

func (r *PostgresDocumentRepository) GetLatestByTenant(ctx context.Context, tenantID string) ([]models.DocumentWithSource, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT gd.id, gd.title, gd.url, s.name as source_name
		 FROM global_documents gd
		 JOIN tenant_documents td ON gd.id = td.document_id
		 JOIN sources s ON gd.source_id = s.id
		 WHERE td.tenant_id = $1
		 ORDER BY td.created_at DESC
		 LIMIT 10`,
		tenantID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get latest documents by tenant: %w", err)
	}
	defer rows.Close()

	docs, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (models.DocumentWithSource, error) {
		var doc models.DocumentWithSource
		err := row.Scan(&doc.ID, &doc.Title, &doc.URL, &doc.SourceName)
		return doc, err
	})
	if err != nil {
		return nil, fmt.Errorf("failed to collect document rows: %w", err)
	}

	return docs, nil
}

func (r *PostgresDocumentRepository) GetOrCreateSource(ctx context.Context, name, sourceType string) (string, error) {
	var id string
	err := r.pool.QueryRow(ctx,
		`INSERT INTO sources (name, type, created_at, updated_at)
		 VALUES ($1, $2, NOW(), NOW())
		 ON CONFLICT (name) DO UPDATE SET updated_at = NOW()
		 RETURNING id`,
		name,
		sourceType,
	).Scan(&id)
	if err != nil {
		return "", fmt.Errorf("failed to get or create source: %w", err)
	}
	return id, nil
}

func (r *PostgresDocumentRepository) GetDocumentsByTenant(ctx context.Context, tenantID string, limit, offset int) ([]models.DocumentWithAnalysis, int, error) {
	var total int
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM tenant_documents WHERE tenant_id = $1`,
		tenantID,
	).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count documents: %w", err)
	}

	rows, err := r.pool.Query(ctx,
		`SELECT gd.id, gd.title, gd.url, s.name as source_name, gd.semantic_analysis, gd.social_temperature, gd.created_at
		 FROM global_documents gd
		 JOIN tenant_documents td ON gd.id = td.document_id
		 JOIN sources s ON gd.source_id = s.id
		 WHERE td.tenant_id = $1
		 ORDER BY gd.created_at DESC
		 LIMIT $2 OFFSET $3`,
		tenantID,
		limit,
		offset,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get documents by tenant: %w", err)
	}
	defer rows.Close()

	docs, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (models.DocumentWithAnalysis, error) {
		var doc models.DocumentWithAnalysis
		err := row.Scan(&doc.ID, &doc.Title, &doc.URL, &doc.SourceName, &doc.SemanticAnalysis, &doc.SocialTemperature, &doc.CreatedAt)
		return doc, err
	})
	if err != nil {
		return nil, 0, fmt.Errorf("failed to collect document rows: %w", err)
	}

	return docs, total, nil
}

func (r *PostgresDocumentRepository) TrackDocument(ctx context.Context, tenantID, url, sourceType string, priority int) (string, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return "", fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	sourceID, err := r.getOrCreateSourceTx(ctx, tx, sourceType)
	if err != nil {
		return "", fmt.Errorf("get or create source: %w", err)
	}

	var docID string
	err = tx.QueryRow(ctx,
		`INSERT INTO global_documents (source_id, title, url, content_text, created_at, updated_at)
		 VALUES ($1, '', $2, '', NOW(), NOW())
		 ON CONFLICT (url) DO UPDATE SET updated_at = NOW()
		 RETURNING id`,
		sourceID, url,
	).Scan(&docID)
	if err != nil {
		return "", fmt.Errorf("insert global document: %w", err)
	}

	_, err = tx.Exec(ctx,
		`INSERT INTO tenant_documents (tenant_id, document_id, created_at, updated_at)
		 VALUES ($1, $2, NOW(), NOW())
		 ON CONFLICT (tenant_id, document_id) DO NOTHING`,
		tenantID, docID,
	)
	if err != nil {
		return "", fmt.Errorf("insert tenant document: %w", err)
	}

	var taskID string
	err = tx.QueryRow(ctx,
		`INSERT INTO discovery_tasks (url, source_type, status, priority)
		 VALUES ($1, $2, 'pending', $3)
		 ON CONFLICT (url) DO UPDATE SET updated_at = NOW()
		 RETURNING id`,
		url, sourceType, priority,
	).Scan(&taskID)
	if err != nil {
		return "", fmt.Errorf("insert discovery task: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return "", fmt.Errorf("commit transaction: %w", err)
	}

	return taskID, nil
}

func (r *PostgresDocumentRepository) getOrCreateSourceTx(ctx context.Context, tx pgx.Tx, name string) (string, error) {
	var id string
	err := tx.QueryRow(ctx,
		`INSERT INTO sources (name, type, created_at, updated_at)
		 VALUES ($1, 'manual', NOW(), NOW())
		 ON CONFLICT (name) DO UPDATE SET updated_at = NOW()
		 RETURNING id`,
		name,
	).Scan(&id)
	if err != nil {
		return "", fmt.Errorf("get or create source: %w", err)
	}
	return id, nil
}
