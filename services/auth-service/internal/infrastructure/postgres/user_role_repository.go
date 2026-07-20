package postgres

import (
	"context"
	"fmt"

	"github.com/deloitte-us-consulting/his-be/services/auth-service/internal/domain"
	"github.com/google/uuid"
)

type UserRoleRepository struct {
	db *DB
}

func NewUserRoleRepository(db *DB) *UserRoleRepository {
	return &UserRoleRepository{db: db}
}

func (r *UserRoleRepository) Add(ctx context.Context, userID, roleID uuid.UUID) error {
	query := `
		INSERT INTO user_roles (user_id, role_id)
		VALUES ($1, $2)
		ON CONFLICT (user_id, role_id) DO NOTHING
	`
	_, err := r.db.Pool.Exec(ctx, query, userID, roleID)
	if err != nil {
		return fmt.Errorf("failed to add user role: %w", err)
	}
	return nil
}

func (r *UserRoleRepository) Remove(ctx context.Context, userID, roleID uuid.UUID) error {
	query := `DELETE FROM user_roles WHERE user_id = $1 AND role_id = $2`
	_, err := r.db.Pool.Exec(ctx, query, userID, roleID)
	if err != nil {
		return fmt.Errorf("failed to remove user role: %w", err)
	}
	return nil
}

func (r *UserRoleRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.UserRole, error) {
	query := `
		SELECT user_id, role_id, assigned_at, assigned_by, expires_at
		FROM user_roles WHERE user_id = $1
	`
	rows, err := r.db.Pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user roles: %w", err)
	}
	defer rows.Close()

	var userRoles []*domain.UserRole
	for rows.Next() {
		var ur domain.UserRole
		if err := rows.Scan(&ur.UserID, &ur.RoleID, &ur.AssignedAt, &ur.AssignedBy, &ur.ExpiresAt); err != nil {
			return nil, fmt.Errorf("failed to scan user role: %w", err)
		}
		userRoles = append(userRoles, &ur)
	}
	return userRoles, nil
}

// GetActiveByUserAndTenant joins user_roles -> roles to scope by tenant
// (user_roles carries no tenant_id column in the deployed schema — see
// migrations/000004_add_rbac_tables.up.sql) and excludes expired
// assignments.
func (r *UserRoleRepository) GetActiveByUserAndTenant(ctx context.Context, userID, tenantID uuid.UUID) ([]*domain.UserRole, error) {
	query := `
		SELECT ur.user_id, ur.role_id, ur.assigned_at, ur.assigned_by, ur.expires_at
		FROM user_roles ur
		JOIN roles r ON r.id = ur.role_id
		WHERE ur.user_id = $1 AND r.tenant_id = $2
		  AND (ur.expires_at IS NULL OR ur.expires_at > NOW())
	`
	rows, err := r.db.Pool.Query(ctx, query, userID, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get active user roles: %w", err)
	}
	defer rows.Close()

	var userRoles []*domain.UserRole
	for rows.Next() {
		var ur domain.UserRole
		if err := rows.Scan(&ur.UserID, &ur.RoleID, &ur.AssignedAt, &ur.AssignedBy, &ur.ExpiresAt); err != nil {
			return nil, fmt.Errorf("failed to scan user role: %w", err)
		}
		userRoles = append(userRoles, &ur)
	}
	return userRoles, nil
}

func (r *UserRoleRepository) GetByUserIDs(ctx context.Context, userIDs []uuid.UUID) ([]*domain.UserRole, error) {
	if len(userIDs) == 0 {
		return []*domain.UserRole{}, nil
	}

	query := `
		SELECT user_id, role_id
		FROM user_roles WHERE user_id = ANY($1)
	`
	rows, err := r.db.Pool.Query(ctx, query, userIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to get user roles by IDs: %w", err)
	}
	defer rows.Close()

	var userRoles []*domain.UserRole
	for rows.Next() {
		var ur domain.UserRole
		if err := rows.Scan(&ur.UserID, &ur.RoleID); err != nil {
			return nil, fmt.Errorf("failed to scan user role: %w", err)
		}
		userRoles = append(userRoles, &ur)
	}
	return userRoles, nil
}