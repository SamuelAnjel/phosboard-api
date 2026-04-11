package repository

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"phosboard/backend/internal/models"
)

type SourceRepository interface {
	CreateSource(ctx context.Context, name, sourceType, url string) (*models.Source, error)
	GetSources(ctx context.Context) ([]models.Source, error)
	GetSourceByID(ctx context.Context, id string) (*models.Source, error)
	UpdateSourceConfig(ctx context.Context, id string, config map[string]interface{}) (*models.Source, error)
	DeleteSource(ctx context.Context, id string) error
}

type PostgresSourceRepository struct {
	pool *pgxpool.Pool
}

func NewSourceRepository(pool *pgxpool.Pool) *PostgresSourceRepository {
	return &PostgresSourceRepository{pool: pool}
}

func (r *PostgresSourceRepository) CreateSource(ctx context.Context, name, sourceType, url string) (*models.Source, error) {
	var source models.Source

	// Determinar fetch_strategy basado en el tipo de source
	fetchStrategy := "rss"
	if sourceType == "web-crawl" {
		fetchStrategy = "web-crawl"
	}

	err := r.pool.QueryRow(ctx, `
		INSERT INTO sources (name, type, url, fetch_strategy, config, created_at, updated_at)
		VALUES ($1, $2, $3, $4, '{}'::jsonb, NOW(), NOW())
		ON CONFLICT (name) DO UPDATE SET updated_at = NOW()
		RETURNING id, name, type, url, config, created_at, updated_at
	`, name, sourceType, url, fetchStrategy).Scan(&source.ID, &source.Name, &source.Type, &source.URL, &source.Config, &source.CreatedAt, &source.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("create source: %w", err)
	}

	return &source, nil
}

func (r *PostgresSourceRepository) GetSources(ctx context.Context) ([]models.Source, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, name, type, url, config, created_at, updated_at
		FROM sources
		ORDER BY name
	`)
	if err != nil {
		return nil, fmt.Errorf("get sources: %w", err)
	}
	defer rows.Close()

	var sources []models.Source
	for rows.Next() {
		var source models.Source
		if err := rows.Scan(&source.ID, &source.Name, &source.Type, &source.URL, &source.Config, &source.CreatedAt, &source.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan source: %w", err)
		}
		sources = append(sources, source)
	}

	return sources, rows.Err()
}

func (r *PostgresSourceRepository) GetSourceByID(ctx context.Context, id string) (*models.Source, error) {
	var source models.Source
	err := r.pool.QueryRow(ctx, `
		SELECT id, name, type, url, config, created_at, updated_at
		FROM sources
		WHERE id = $1
	`, id).Scan(&source.ID, &source.Name, &source.Type, &source.URL, &source.Config, &source.CreatedAt, &source.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("get source by id: %w", err)
	}

	return &source, nil
}

func (r *PostgresSourceRepository) UpdateSourceConfig(ctx context.Context, id string, config map[string]interface{}) (*models.Source, error) {
	configJSON, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("marshal config: %w", err)
	}

	var source models.Source
	err = r.pool.QueryRow(ctx, `
		UPDATE sources SET config = $1, updated_at = NOW()
		WHERE id = $2
		RETURNING id, name, type, url, config, created_at, updated_at
	`, configJSON, id).Scan(&source.ID, &source.Name, &source.Type, &source.URL, &source.Config, &source.CreatedAt, &source.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("update source config: %w", err)
	}

	return &source, nil
}

func (r *PostgresSourceRepository) DeleteSource(ctx context.Context, id string) error {
	result, err := r.pool.Exec(ctx, `DELETE FROM sources WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete source: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("source not found")
	}

	return nil
}
