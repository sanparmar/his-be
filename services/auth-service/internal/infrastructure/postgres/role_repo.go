package postgres

import (
	"context"
	"fmt"

	"github.com/deloitte-us-consulting/his-be/services/auth-service/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type RoleRepository struct {
	db *DB
}

func NewRoleRepository(db *DB) *RoleRepository {
	return &RoleRepository{db: db}
}

func (r *RoleRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Role, error) {
	query := `
		SELECT id, name, display_name, category, parent_role_id, is_system, description
		FROM roles WHERE id = $1
	`
	var role domain.Role
	err := r.db.Pool.QueryRow(ctx, query, id).Scan(
		&role.ID, &role.Name, &role.DisplayName, &role.Category, &role.ParentID, &role.IsSystem, &role.Description,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get role by ID: %w", err)
	}
	return &role, nil
}

func (r *RoleRepository) GetByName(ctx context.Context, name string) (*domain.Role, error) {
	query := `
		SELECT id, name, display_name, category, parent_role_id, is_system, description
		FROM roles WHERE name = $1
	`
	var role domain.Role
	err := r.db.Pool.QueryRow(ctx, query, name).Scan(
		&role.ID, &role.Name, &role.DisplayName, &role.Category, &role.ParentID, &role.IsSystem, &role.Description,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get role by name: %w", err)
	}
	return &role, nil
}

func (r *RoleRepository) GetWithPermissions(ctx context.Context, id uuid.UUID) (*domain.Role, error) {
	role, err := r.GetByID(ctx, id)
	if err != nil || role == nil {
		return role, err
	}

	query := `
		SELECT p.id, p.name, p.resource, p.action, p.scope, p.category, p.description
		FROM permissions p
		JOIN role_permissions rp ON p.id = rp.permission_id
		WHERE rp.role_id = $1
		ORDER BY p.category, p.resource, p.action
	`
	rows, err := r.db.Pool.Query(ctx, query, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get role permissions: %w", err)
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
	role.Permissions = perms
	return role, nil
}

func (r *RoleRepository) GetInheritedRoles(ctx context.Context, roleID uuid.UUID) ([]domain.Role, error) {
	query := `
		WITH RECURSIVE role_hierarchy AS (
			SELECT id, name, display_name, category, parent_role_id, is_system, description, 0 as level
			FROM roles WHERE id = $1
			UNION ALL
			SELECT r.id, r.name, r.display_name, r.category, r.parent_role_id, r.is_system, r.description, rh.level + 1
			FROM roles r
			JOIN role_hierarchy rh ON r.id = rh.parent_role_id
			WHERE rh.level < 10
		)
		SELECT id, name, display_name, category, parent_role_id, is_system, description
		FROM role_hierarchy
		ORDER BY level
	`
	rows, err := r.db.Pool.Query(ctx, query, roleID)
	if err != nil {
		return nil, fmt.Errorf("failed to get inherited roles: %w", err)
	}
	defer rows.Close()

	var roles []domain.Role
	for rows.Next() {
		var role domain.Role
		if err := rows.Scan(&role.ID, &role.Name, &role.DisplayName, &role.Category, &role.ParentID, &role.IsSystem, &role.Description); err != nil {
			return nil, fmt.Errorf("failed to scan role: %w", err)
		}
		roles = append(roles, role)
	}
	return roles, nil
}

func (r *RoleRepository) List(ctx context.Context) ([]domain.Role, error) {
	return r.listByCategory(ctx, "")
}

func (r *RoleRepository) listByCategory(ctx context.Context, category string) ([]domain.Role, error) {
	query := `
		SELECT id, name, display_name, category, parent_role_id, is_system, description
		FROM roles
	`
	args := []interface{}{}
	if category != "" {
		query += " WHERE category = $1"
		args = append(args, category)
	}
	query += " ORDER BY category, name"

	rows, err := r.db.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list roles: %w", err)
	}
	defer rows.Close()

	var roles []domain.Role
	for rows.Next() {
		var role domain.Role
		if err := rows.Scan(&role.ID, &role.Name, &role.DisplayName, &role.Category, &role.ParentID, &role.IsSystem, &role.Description); err != nil {
			return nil, fmt.Errorf("failed to scan role: %w", err)
		}
		roles = append(roles, role)
	}
	return roles, nil
}

func (r *RoleRepository) ListByCategory(ctx context.Context, category string) ([]domain.Role, error) {
	return r.listByCategory(ctx, category)
}

func (r *RoleRepository) Create(ctx context.Context, role *domain.Role) error {
	query := `
		INSERT INTO roles (id, name, display_name, category, parent_role_id, is_system, description)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := r.db.Pool.Exec(ctx, query,
		role.ID, role.Name, role.DisplayName, role.Category, role.ParentID, role.IsSystem, role.Description,
	)
	if err != nil {
		return fmt.Errorf("failed to create role: %w", err)
	}
	return nil
}

func (r *RoleRepository) Update(ctx context.Context, role *domain.Role) error {
	query := `
		UPDATE roles SET name = $2, display_name = $3, category = $4, parent_role_id = $5, is_system = $6, description = $7
		WHERE id = $1
	`
	_, err := r.db.Pool.Exec(ctx, query,
		role.ID, role.Name, role.DisplayName, role.Category, role.ParentID, role.IsSystem, role.Description,
	)
	if err != nil {
		return fmt.Errorf("failed to update role: %w", err)
	}
	return nil
}

func (r *RoleRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM roles WHERE id = $1`
	_, err := r.db.Pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete role: %w", err)
	}
	return nil
}

func (r *RoleRepository) AssignPermission(ctx context.Context, roleID, permissionID uuid.UUID) error {
	query := `INSERT INTO role_permissions (role_id, permission_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`
	_, err := r.db.Pool.Exec(ctx, query, roleID, permissionID)
	if err != nil {
		return fmt.Errorf("failed to assign permission to role: %w", err)
	}
	return nil
}

func (r *RoleRepository) RevokePermission(ctx context.Context, roleID, permissionID uuid.UUID) error {
	query := `DELETE FROM role_permissions WHERE role_id = $1 AND permission_id = $2`
	_, err := r.db.Pool.Exec(ctx, query, roleID, permissionID)
	if err != nil {
		return fmt.Errorf("failed to revoke permission from role: %w", err)
	}
	return nil
}

func (r *RoleRepository) AddPermission(ctx context.Context, roleID, permID uuid.UUID) error {
	return r.AssignPermission(ctx, roleID, permID)
}

func (r *RoleRepository) RemovePermission(ctx context.Context, roleID, permID uuid.UUID) error {
	return r.RevokePermission(ctx, roleID, permID)
}

func (r *RoleRepository) GetPermissions(ctx context.Context, roleID uuid.UUID) ([]domain.Permission, error) {
	query := `
		SELECT p.id, p.name, p.resource, p.action, p.scope, p.category, p.description
		FROM permissions p
		JOIN role_permissions rp ON p.id = rp.permission_id
		WHERE rp.role_id = $1
		ORDER BY p.category, p.resource, p.action
	`
	rows, err := r.db.Pool.Query(ctx, query, roleID)
	if err != nil {
		return nil, fmt.Errorf("failed to get role permissions: %w", err)
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