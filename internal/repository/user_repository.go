package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

type UserRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

type User struct {
	ID           string `json:"id"`
	Email        string `json:"email"`
	PasswordHash string `json:"-"`
	FullName     string `json:"full_name,omitempty"`
	IsActive     bool   `json:"is_active"`
	CreatedAt    string `json:"created_at,omitempty"`
	UpdatedAt    string `json:"updated_at,omitempty"`
}

type UserWithRoles struct {
	User
	Roles []UserRole `json:"roles"`
}

type UserRole struct {
	RoleID   string  `json:"role_id"`
	RoleName string  `json:"role_name"`
	TenantID *string `json:"tenant_id,omitempty"` // NULL for global roles
}

func (r *UserRepository) GetUserByEmail(ctx context.Context, email string) (*UserWithRoles, error) {
	var user UserWithRoles

	// Get user basic info
	err := r.pool.QueryRow(ctx, `
		SELECT id, email, password_hash, full_name, is_active, created_at, updated_at
		FROM users
		WHERE email = $1 AND is_active = TRUE
	`, email).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.FullName, &user.IsActive, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("get user by email: %w", err)
	}

	// Get user roles
	rows, err := r.pool.Query(ctx, `
		SELECT r.id, r.name, ur.tenant_id
		FROM user_roles ur
		JOIN roles r ON ur.role_id = r.id
		WHERE ur.user_id = $1
	`, user.ID)

	if err != nil {
		return nil, fmt.Errorf("get user roles: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var role UserRole
		var tenantID *string
		if err := rows.Scan(&role.RoleID, &role.RoleName, &tenantID); err != nil {
			return nil, fmt.Errorf("scan user role: %w", err)
		}
		role.TenantID = tenantID
		user.Roles = append(user.Roles, role)
	}

	return &user, nil
}

func (r *UserRepository) ValidateCredentials(ctx context.Context, email, password string) (*UserWithRoles, error) {
	user, err := r.GetUserByEmail(ctx, email)
	if err != nil {
		// Don't leak whether user exists or not
		return nil, fmt.Errorf("invalid credentials")
	}

	// Compare password hash
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	return user, nil
}

func (r *UserRepository) HasPermission(ctx context.Context, userID, endpoint, method string) (bool, error) {
	var hasPermission bool

	// Check if user has permission via any of their roles
	// This query checks for:
	// 1. Exact endpoint match OR wildcard endpoint match
	// 2. Exact method match OR wildcard method (*)
	// 3. Either global role (tenant_id IS NULL) or tenant-specific role
	err := r.pool.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1
			FROM user_roles ur
			JOIN role_permissions rp ON ur.role_id = rp.role_id
			JOIN permissions p ON rp.permission_id = p.id
			WHERE ur.user_id = $1
			AND (
				-- Exact endpoint match
				p.endpoint = $2
				OR
				-- Wildcard endpoint match (e.g., /api/v1/*)
				($2 LIKE REPLACE(p.endpoint, '*', '%') AND p.endpoint LIKE '%*%')
			)
			AND (
				-- Exact method match
				p.method = $3
				OR
				-- Wildcard method
				p.method = '*'
			)
		)
	`, userID, endpoint, method).Scan(&hasPermission)

	if err != nil {
		return false, fmt.Errorf("check permission: %w", err)
	}

	return hasPermission, nil
}

func (r *UserRepository) CreateUser(ctx context.Context, email, password, fullName string) (*User, error) {
	// Hash password
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	var user User
	err = r.pool.QueryRow(ctx, `
		INSERT INTO users (email, password_hash, full_name, is_active)
		VALUES ($1, $2, $3, TRUE)
		RETURNING id, email, full_name, is_active, created_at, updated_at
	`, email, string(passwordHash), fullName).Scan(
		&user.ID, &user.Email, &user.FullName, &user.IsActive, &user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}

	return &user, nil
}
