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

const roleColumns = "id, name, description, tenant_id, is_system_role, created_at, updated_at"

func scanRole(row interface {
	Scan(dest ...interface{}) error
}) (*domain.Role, error) {
	var role domain.Role
	if err := row.Scan(&role.ID, &role.Name, &role.Description, &role.TenantID, &role.IsSystemRole, &role.CreatedAt, &role.UpdatedAt); err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *RoleRepository) Create(ctx context.Context, role *domain.Role) error {
	query := `
		INSERT INTO roles (id, name, description, tenant_id, is_system_role, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $6)
	`
	_, err := r.db.Pool.Exec(ctx, query,
		role.ID, role.Name, role.Description, role.TenantID, role.IsSystemRole, role.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create role: %w", err)
	}
	return nil
}

func (r *RoleRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Role, error) {
	query := `SELECT ` + roleColumns + ` FROM roles WHERE id = $1`
	role, err := scanRole(r.db.Pool.QueryRow(ctx, query, id))
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get role by ID: %w", err)
	}
	return role, nil
}

func (r *RoleRepository) GetByNameAndTenant(ctx context.Context, name string, tenantID uuid.UUID) (*domain.Role, error) {
	query := `SELECT ` + roleColumns + ` FROM roles WHERE name = $1 AND tenant_id = $2`
	role, err := scanRole(r.db.Pool.QueryRow(ctx, query, name, tenantID))
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get role by name and tenant: %w", err)
	}
	return role, nil
}

func (r *RoleRepository) GetByIDs(ctx context.Context, ids []uuid.UUID) ([]*domain.Role, error) {
	if len(ids) == 0 {
		return []*domain.Role{}, nil
	}

	query := `SELECT ` + roleColumns + ` FROM roles WHERE id = ANY($1)`
	rows, err := r.db.Pool.Query(ctx, query, ids)
	if err != nil {
		return nil, fmt.Errorf("failed to get roles by IDs: %w", err)
	}
	defer rows.Close()
	return scanRoleRows(rows)
}

func (r *RoleRepository) ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]*domain.Role, error) {
	query := `SELECT ` + roleColumns + ` FROM roles WHERE tenant_id = $1 ORDER BY name`
	rows, err := r.db.Pool.Query(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to list roles by tenant: %w", err)
	}
	defer rows.Close()
	return scanRoleRows(rows)
}

func scanRoleRows(rows pgx.Rows) ([]*domain.Role, error) {
	var roles []*domain.Role
	for rows.Next() {
		var role domain.Role
		if err := rows.Scan(&role.ID, &role.Name, &role.Description, &role.TenantID, &role.IsSystemRole, &role.CreatedAt, &role.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan role: %w", err)
		}
		roles = append(roles, &role)
	}
	return roles, nil
}
