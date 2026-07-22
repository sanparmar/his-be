package postgres

import (
	"context"
	"fmt"

	"github.com/deloitte-us-consulting/his-be/services/auth-service/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type PermissionRepository struct {
	db *DB
}

func NewPermissionRepository(db *DB) *PermissionRepository {
	return &PermissionRepository{db: db}
}

func (r *PermissionRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Permission, error) {
	query := `
		SELECT id, name, resource, action, scope, category, description
		FROM permissions WHERE id = $1
	`
	var perm domain.Permission
	err := r.db.Pool.QueryRow(ctx, query, id).Scan(
		&perm.ID, &perm.Name, &perm.Resource, &perm.Action, &perm.Scope, &perm.Category, &perm.Description,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get permission by ID: %w", err)
	}
	return &perm, nil
}

func (r *PermissionRepository) GetByName(ctx context.Context, name string) (*domain.Permission, error) {
	query := `
		SELECT id, name, resource, action, scope, category, description
		FROM permissions WHERE name = $1
	`
	var perm domain.Permission
	err := r.db.Pool.QueryRow(ctx, query, name).Scan(
		&perm.ID, &perm.Name, &perm.Resource, &perm.Action, &perm.Scope, &perm.Category, &perm.Description,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get permission by name: %w", err)
	}
	return &perm, nil
}

func (r *PermissionRepository) GetByResourceAction(ctx context.Context, resource, action string) ([]domain.Permission, error) {
	query := `
		SELECT id, name, resource, action, scope, category, description
		FROM permissions WHERE resource = $1 AND action = $2
	`
	rows, err := r.db.Pool.Query(ctx, query, resource, action)
	if err != nil {
		return nil, fmt.Errorf("failed to get permissions by resource/action: %w", err)
	}
	defer rows.Close()

	var perms []domain.Permission
	for rows.Next() {
		var perm domain.Permission
		if err := rows.Scan(&perm.ID, &perm.Name, &perm.Resource, &perm.Action, &perm.Scope, &perm.Category, &perm.Description); err != nil {
			return nil, fmt.Errorf("failed to scan permission: %w", err)
		}
		perms = append(perms, perm)
	}
	return perms, nil
}

func (r *PermissionRepository) ListByCategory(ctx context.Context, category string) ([]domain.Permission, error) {
	query := `
		SELECT id, name, resource, action, scope, category, description
		FROM permissions WHERE category = $1 ORDER BY resource, action
	`
	rows, err := r.db.Pool.Query(ctx, query, category)
	if err != nil {
		return nil, fmt.Errorf("failed to list permissions by category: %w", err)
	}
	defer rows.Close()

	var perms []domain.Permission
	for rows.Next() {
		var perm domain.Permission
		if err := rows.Scan(&perm.ID, &perm.Name, &perm.Resource, &perm.Action, &perm.Scope, &perm.Category, &perm.Description); err != nil {
			return nil, fmt.Errorf("failed to scan permission: %w", err)
		}
		perms = append(perms, perm)
	}
	return perms, nil
}

func (r *PermissionRepository) ListAll(ctx context.Context) ([]domain.Permission, error) {
	query := `
		SELECT id, name, resource, action, scope, category, description
		FROM permissions ORDER BY category, resource, action
	`
	rows, err := r.db.Pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list all permissions: %w", err)
	}
	defer rows.Close()

	var perms []domain.Permission
	for rows.Next() {
		var perm domain.Permission
		if err := rows.Scan(&perm.ID, &perm.Name, &perm.Resource, &perm.Action, &perm.Scope, &perm.Category, &perm.Description); err != nil {
			return nil, fmt.Errorf("failed to scan permission: %w", err)
		}
		perms = append(perms, perm)
	}
	return perms, nil
}

func (r *PermissionRepository) Create(ctx context.Context, perm *domain.Permission) error {
	query := `
		INSERT INTO permissions (id, name, resource, action, scope, category, description)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := r.db.Pool.Exec(ctx, query,
		perm.ID, perm.Name, perm.Resource, perm.Action, perm.Scope, perm.Category, perm.Description,
	)
	if err != nil {
		return fmt.Errorf("failed to create permission: %w", err)
	}
	return nil
}

func (r *PermissionRepository) Update(ctx context.Context, perm *domain.Permission) error {
	query := `
		UPDATE permissions SET resource = $2, action = $3, scope = $4, category = $5, description = $6
		WHERE id = $1
	`
	_, err := r.db.Pool.Exec(ctx, query,
		perm.ID, perm.Resource, perm.Action, perm.Scope, perm.Category, perm.Description,
	)
	if err != nil {
		return fmt.Errorf("failed to update permission: %w", err)
	}
	return nil
}

func (r *PermissionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM permissions WHERE id = $1`
	_, err := r.db.Pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete permission: %w", err)
	}
	return nil
}