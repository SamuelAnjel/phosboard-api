package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type TenantRepository struct {
	pool *pgxpool.Pool
}

func NewTenantRepository(pool *pgxpool.Pool) *TenantRepository {
	return &TenantRepository{pool: pool}
}

type Tenant struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	CreatedAt string `json:"created_at,omitempty"`
	UpdatedAt string `json:"updated_at,omitempty"`
}

func (r *TenantRepository) CreateTenant(ctx context.Context, name string) (*Tenant, error) {
	var tenant Tenant
	err := r.pool.QueryRow(ctx, `
		INSERT INTO tenants (name)
		VALUES ($1)
		RETURNING id, name, created_at, updated_at
	`, name).Scan(&tenant.ID, &tenant.Name, &tenant.CreatedAt, &tenant.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("create tenant: %w", err)
	}

	return &tenant, nil
}

func (r *TenantRepository) GetTenantByID(ctx context.Context, id string) (*Tenant, error) {
	var tenant Tenant
	err := r.pool.QueryRow(ctx, `
		SELECT id, name, created_at, updated_at
		FROM tenants
		WHERE id = $1
	`, id).Scan(&tenant.ID, &tenant.Name, &tenant.CreatedAt, &tenant.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("get tenant by id: %w", err)
	}

	return &tenant, nil
}

func (r *TenantRepository) TenantExists(ctx context.Context, id string) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx, `
		SELECT EXISTS(SELECT 1 FROM tenants WHERE id = $1)
	`, id).Scan(&exists)

	if err != nil {
		return false, fmt.Errorf("check tenant exists: %w", err)
	}

	return exists, nil
}
