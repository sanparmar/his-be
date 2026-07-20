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

func scanPermission(row interface {
	Scan(dest ...interface{}) error
}) (*domain.Permission, error) {
	var perm domain.Permission
	if err := row.Scan(&perm.ID, &perm.Resource, &perm.Action, &perm.Description, &perm.CreatedAt); err != nil {
		return nil, err
	}
	perm.Name = perm.Resource + ":" + perm.Action
	return &perm, nil
}

func (r *PermissionRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Permission, error) {
	query := `SELECT id, resource, action, description, created_at FROM permissions WHERE id = $1`
	perm, err := scanPermission(r.db.Pool.QueryRow(ctx, query, id))
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get permission by ID: %w", err)
	}
	return perm, nil
}

func (r *PermissionRepository) GetByResourceAction(ctx context.Context, resource, action string) ([]domain.Permission, error) {
	query := `SELECT id, resource, action, description, created_at FROM permissions WHERE resource = $1 AND action = $2`
	rows, err := r.db.Pool.Query(ctx, query, resource, action)
	if err != nil {
		return nil, fmt.Errorf("failed to get permissions by resource/action: %w", err)
	}
	defer rows.Close()
	return scanPermissionRows(rows)
}

func (r *PermissionRepository) ListAll(ctx context.Context) ([]domain.Permission, error) {
	query := `SELECT id, resource, action, description, created_at FROM permissions ORDER BY resource, action`
	rows, err := r.db.Pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list all permissions: %w", err)
	}
	defer rows.Close()
	return scanPermissionRows(rows)
}

// GetByRoleIDs resolves permissions for a batch of roles via the real
// role_permissions join (migrations/000004_add_rbac_tables.up.sql).
func (r *PermissionRepository) GetByRoleIDs(ctx context.Context, roleIDs []uuid.UUID) ([]domain.Permission, error) {
	if len(roleIDs) == 0 {
		return []domain.Permission{}, nil
	}

	query := `
		SELECT DISTINCT p.id, p.resource, p.action, p.description, p.created_at
		FROM permissions p
		JOIN role_permissions rp ON rp.permission_id = p.id
		WHERE rp.role_id = ANY($1)
	`
	rows, err := r.db.Pool.Query(ctx, query, roleIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to get permissions by role IDs: %w", err)
	}
	defer rows.Close()
	return scanPermissionRows(rows)
}

func (r *PermissionRepository) Create(ctx context.Context, perm *domain.Permission) error {
	query := `INSERT INTO permissions (id, resource, action, description) VALUES ($1, $2, $3, $4)`
	_, err := r.db.Pool.Exec(ctx, query, perm.ID, perm.Resource, perm.Action, perm.Description)
	if err != nil {
		return fmt.Errorf("failed to create permission: %w", err)
	}
	return nil
}

func scanPermissionRows(rows pgx.Rows) ([]domain.Permission, error) {
	var perms []domain.Permission
	for rows.Next() {
		var perm domain.Permission
		if err := rows.Scan(&perm.ID, &perm.Resource, &perm.Action, &perm.Description, &perm.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan permission: %w", err)
		}
		perm.Name = perm.Resource + ":" + perm.Action
		perms = append(perms, perm)
	}
	return perms, nil
}
