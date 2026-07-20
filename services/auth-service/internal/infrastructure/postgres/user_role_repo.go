package postgres

import (
	"context"
	"fmt"

	"github.com/deloitte-us-consulting/his-be/services/auth-service/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type UserRoleRepository struct {
	db *DB
}

func NewUserRoleRepository(db *DB) *UserRoleRepository {
	return &UserRoleRepository{db: db}
}

func (r *UserRoleRepository) Assign(ctx context.Context, ur *domain.UserRole) error {
	query := `
		INSERT INTO user_roles (user_id, role_id, tenant_id, organization_id, hospital_id, department_id, assigned_by, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (user_id, role_id, tenant_id) DO UPDATE SET
			organization_id = EXCLUDED.organization_id,
			hospital_id = EXCLUDED.hospital_id,
			department_id = EXCLUDED.department_id,
			assigned_by = EXCLUDED.assigned_by,
			expires_at = EXCLUDED.expires_at,
			assigned_at = NOW()
	`
	_, err := r.db.Pool.Exec(ctx, query,
		ur.UserID, ur.RoleID, ur.TenantID, ur.OrganizationID, ur.HospitalID, ur.DepartmentID, ur.AssignedBy, ur.ExpiresAt,
	)
	if err != nil {
		return fmt.Errorf("failed to assign role: %w", err)
	}
	return nil
}

func (r *UserRoleRepository) Revoke(ctx context.Context, userID, roleID, tenantID uuid.UUID) error {
	query := `DELETE FROM user_roles WHERE user_id = $1 AND role_id = $2 AND tenant_id = $3`
	result, err := r.db.Pool.Exec(ctx, query, userID, roleID, tenantID)
	if err != nil {
		return fmt.Errorf("failed to revoke role: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("role assignment not found")
	}
	return nil
}

func (r *UserRoleRepository) GetByUser(ctx context.Context, userID uuid.UUID) ([]domain.UserRole, error) {
	query := `
		SELECT user_id, role_id, tenant_id, organization_id, hospital_id, department_id, assigned_by, assigned_at, expires_at
		FROM user_roles WHERE user_id = $1
	`
	rows, err := r.db.Pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user roles: %w", err)
	}
	defer rows.Close()

	return r.scanUserRoles(rows)
}

func (r *UserRoleRepository) GetByUserAndTenant(ctx context.Context, userID, tenantID uuid.UUID) ([]domain.UserRole, error) {
	query := `
		SELECT user_id, role_id, tenant_id, organization_id, hospital_id, department_id, assigned_by, assigned_at, expires_at
		FROM user_roles WHERE user_id = $1 AND tenant_id = $2
	`
	rows, err := r.db.Pool.Query(ctx, query, userID, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user roles by tenant: %w", err)
	}
	defer rows.Close()

	return r.scanUserRoles(rows)
}

func (r *UserRoleRepository) ListByRole(ctx context.Context, roleID uuid.UUID) ([]domain.UserRole, error) {
	query := `
		SELECT user_id, role_id, tenant_id, organization_id, hospital_id, department_id, assigned_by, assigned_at, expires_at
		FROM user_roles WHERE role_id = $1
	`
	rows, err := r.db.Pool.Query(ctx, query, roleID)
	if err != nil {
		return nil, fmt.Errorf("failed to list user roles by role: %w", err)
	}
	defer rows.Close()

	return r.scanUserRoles(rows)
}

func (r *UserRoleRepository) GetActiveByUserAndTenant(ctx context.Context, userID, tenantID uuid.UUID) ([]domain.UserRole, error) {
	query := `
		SELECT user_id, role_id, tenant_id, organization_id, hospital_id, department_id, assigned_by, assigned_at, expires_at
		FROM user_roles 
		WHERE user_id = $1 AND tenant_id = $2 
		AND (expires_at IS NULL OR expires_at > NOW())
	`
	rows, err := r.db.Pool.Query(ctx, query, userID, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get active user roles: %w", err)
	}
	defer rows.Close()

	return r.scanUserRoles(rows)
}

func (r *UserRoleRepository) scanUserRoles(rows pgx.Rows) ([]domain.UserRole, error) {
	var roles []domain.UserRole
	for rows.Next() {
		var ur domain.UserRole
		if err := rows.Scan(
			&ur.UserID, &ur.RoleID, &ur.TenantID,
			&ur.OrganizationID, &ur.HospitalID, &ur.DepartmentID,
			&ur.AssignedBy, &ur.AssignedAt, &ur.ExpiresAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan user role: %w", err)
		}
		roles = append(roles, ur)
	}
	return roles, nil
}

func (r *UserRoleRepository) CountByUserAndTenant(ctx context.Context, userID, tenantID uuid.UUID) (int, error) {
	query := `SELECT COUNT(*) FROM user_roles WHERE user_id = $1 AND tenant_id = $2`
	var count int
	err := r.db.Pool.QueryRow(ctx, query, userID, tenantID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count user roles: %w", err)
	}
	return count, nil
}