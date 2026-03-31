package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"phosboard/backend/internal/models"
)

type ConceptRepository interface {
	CreateConcept(ctx context.Context, tenantID, term string) (*models.TenantConcept, error)
	GetActiveConceptsByTenant(ctx context.Context, tenantID string) ([]models.TenantConcept, error)
	DeleteConcept(ctx context.Context, conceptID string) error
}

type PostgresConceptRepository struct {
	pool *pgxpool.Pool
}

func NewConceptRepository(pool *pgxpool.Pool) *PostgresConceptRepository {
	return &PostgresConceptRepository{pool: pool}
}

func (r *PostgresConceptRepository) CreateConcept(ctx context.Context, tenantID, term string) (*models.TenantConcept, error) {
	var concept models.TenantConcept
	err := r.pool.QueryRow(ctx, `
		INSERT INTO tenant_concepts (tenant_id, concept_term, is_active)
		VALUES ($1, $2, TRUE)
		ON CONFLICT (tenant_id, concept_term) DO UPDATE SET is_active = TRUE, updated_at = NOW()
		RETURNING id, tenant_id, concept_term, is_active, created_at, updated_at
	`, tenantID, term).Scan(&concept.ID, &concept.TenantID, &concept.ConceptTerm, &concept.IsActive, &concept.CreatedAt, &concept.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("create concept: %w", err)
	}

	return &concept, nil
}

func (r *PostgresConceptRepository) GetActiveConceptsByTenant(ctx context.Context, tenantID string) ([]models.TenantConcept, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, tenant_id, concept_term, is_active, created_at, updated_at
		FROM tenant_concepts
		WHERE tenant_id = $1 AND is_active = TRUE
		ORDER BY concept_term
	`, tenantID)
	if err != nil {
		return nil, fmt.Errorf("get concepts: %w", err)
	}
	defer rows.Close()

	var concepts []models.TenantConcept
	for rows.Next() {
		var concept models.TenantConcept
		if err := rows.Scan(&concept.ID, &concept.TenantID, &concept.ConceptTerm, &concept.IsActive, &concept.CreatedAt, &concept.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan concept: %w", err)
		}
		concepts = append(concepts, concept)
	}

	return concepts, rows.Err()
}

func (r *PostgresConceptRepository) DeleteConcept(ctx context.Context, conceptID string) error {
	result, err := r.pool.Exec(ctx, `
		UPDATE tenant_concepts SET is_active = FALSE, updated_at = NOW() WHERE id = $1
	`, conceptID)
	if err != nil {
		return fmt.Errorf("delete concept: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("concept not found: %w", err)
	}

	return nil
}
