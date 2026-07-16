package postgres

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/deloitte-us-consulting/his-be/services/auth-service/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type MFARepository struct {
	db *DB
}

func NewMFARepository(db *DB) *MFARepository {
	return &MFARepository{db: db}
}

func (r *MFARepository) Create(ctx context.Context, mfa *domain.MFACredential) error {
	backupCodesJSON, _ := json.Marshal(mfa.BackupCodes)

	query := `
		INSERT INTO mfa_credentials (id, user_id, type, secret, enabled, backup_codes, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err := r.db.Pool.Exec(ctx, query,
		mfa.ID, mfa.UserID, mfa.Type, mfa.Secret,
		mfa.Enabled, backupCodesJSON,
		mfa.CreatedAt, mfa.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create MFA credential: %w", err)
	}
	return nil
}

func (r *MFARepository) GetByUserID(ctx context.Context, userID uuid.UUID, mfaType domain.MFAType) (*domain.MFACredential, error) {
	query := `
		SELECT id, user_id, type, secret, enabled, backup_codes, created_at, updated_at
		FROM mfa_credentials WHERE user_id = $1 AND type = $2
	`
	var mfa domain.MFACredential
	var backupCodesJSON []byte

	err := r.db.Pool.QueryRow(ctx, query, userID, mfaType).Scan(
		&mfa.ID, &mfa.UserID, &mfa.Type, &mfa.Secret,
		&mfa.Enabled, &backupCodesJSON,
		&mfa.CreatedAt, &mfa.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get MFA credential: %w", err)
	}

	if len(backupCodesJSON) > 0 {
		_ = json.Unmarshal(backupCodesJSON, &mfa.BackupCodes)
	}

	return &mfa, nil
}

func (r *MFARepository) Update(ctx context.Context, mfa *domain.MFACredential) error {
	backupCodesJSON, _ := json.Marshal(mfa.BackupCodes)

	query := `
		UPDATE mfa_credentials
		SET secret = $2, enabled = $3, backup_codes = $4, updated_at = $5
		WHERE id = $1
	`
	_, err := r.db.Pool.Exec(ctx, query,
		mfa.ID, mfa.Secret, mfa.Enabled, backupCodesJSON, mfa.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to update MFA credential: %w", err)
	}
	return nil
}

func (r *MFARepository) Delete(ctx context.Context, userID uuid.UUID, mfaType domain.MFAType) error {
	query := `DELETE FROM mfa_credentials WHERE user_id = $1 AND type = $2`
	_, err := r.db.Pool.Exec(ctx, query, userID, mfaType)
	if err != nil {
		return fmt.Errorf("failed to delete MFA credential: %w", err)
	}
	return nil
}