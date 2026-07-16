package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/deloitte-us-consulting/his-be/services/auth-service/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository struct {
	db *DB
}

func NewUserRepository(db *DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, user *domain.User) error {
	query := `
		INSERT INTO users (id, username, email, password_hash, tenant_id, organization_id, hospital_id, department_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW(), NOW())
	`
	_, err := r.db.Pool.Exec(ctx, query,
		user.ID, user.Username, user.Email, user.PasswordHash,
		user.TenantID, user.OrganizationID, user.HospitalID, user.DepartmentID,
	)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}
	return nil
}

func (r *UserRepository) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	query := `
		SELECT id, username, email, password_hash, tenant_id, organization_id, hospital_id, department_id, created_at, updated_at
		FROM users WHERE username = $1
	`
	var user domain.User
	err := r.db.Pool.QueryRow(ctx, query, username).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash,
		&user.TenantID, &user.OrganizationID, &user.HospitalID, &user.DepartmentID,
		&user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user by username: %w", err)
	}
	return &user, nil
}

func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	query := `
		SELECT id, username, email, password_hash, tenant_id, organization_id, hospital_id, department_id, created_at, updated_at
		FROM users WHERE id = $1
	`
	var user domain.User
	err := r.db.Pool.QueryRow(ctx, query, id).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash,
		&user.TenantID, &user.OrganizationID, &user.HospitalID, &user.DepartmentID,
		&user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}
	return &user, nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `
		SELECT id, username, email, password_hash, tenant_id, organization_id, hospital_id, created_at, updated_at
		FROM users WHERE email = $1
	`
	var user domain.User
	err := r.db.Pool.QueryRow(ctx, query, email).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash,
		&user.TenantID, &user.OrganizationID, &user.HospitalID,
		&user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}
	return &user, nil
}

func (r *UserRepository) Update(ctx context.Context, user *domain.User) error {
	query := `
		UPDATE users SET username = $2, email = $3, password_hash = $4, tenant_id = $5, organization_id = $6, hospital_id = $7, department_id = $8, updated_at = NOW()
		WHERE id = $1
	`
	_, err := r.db.Pool.Exec(ctx, query,
		user.ID, user.Username, user.Email, user.PasswordHash,
		user.TenantID, user.OrganizationID, user.HospitalID, user.DepartmentID,
	)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}
	return nil
}