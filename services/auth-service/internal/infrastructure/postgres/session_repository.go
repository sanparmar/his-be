package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/deloitte-us-consulting/his-be/services/auth-service/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type SessionRepository struct {
	db *DB
}

func NewSessionRepository(db *DB) *SessionRepository {
	return &SessionRepository{db: db}
}

func (r *SessionRepository) Create(ctx context.Context, session *domain.Session) error {
	query := `
		INSERT INTO sessions (id, user_id, token, expires_at, ip_address, user_agent)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := r.db.Pool.Exec(ctx, query,
		session.ID, session.UserID, session.Token, session.ExpiresAt,
		session.IPAddress, session.UserAgent,
	)
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}
	return nil
}

func (r *SessionRepository) GetByToken(ctx context.Context, token string) (*domain.Session, error) {
	query := `
		SELECT id, user_id, token, expires_at, ip_address, user_agent
		FROM sessions WHERE token = $1
	`
	var session domain.Session
	err := r.db.Pool.QueryRow(ctx, query, token).Scan(
		&session.ID, &session.UserID, &session.Token, &session.ExpiresAt,
		&session.IPAddress, &session.UserAgent,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get session by token: %w", err)
	}
	return &session, nil
}

func (r *SessionRepository) Delete(ctx context.Context, token string) error {
	query := `DELETE FROM sessions WHERE token = $1`
	_, err := r.db.Pool.Exec(ctx, query, token)
	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}
	return nil
}

func (r *SessionRepository) UpdateLastActivity(ctx context.Context, token string) error {
	// For simplicity, in this implementation, we don't have a 'last_activity' column,
	// but we would typically update a timestamp here.
	return nil
}

func (r *SessionRepository) UpdateToken(ctx context.Context, oldToken, newToken string, expiresAt time.Time) error {
	query := `UPDATE sessions SET token = $1, expires_at = $2 WHERE token = $3`
	result, err := r.db.Pool.Exec(ctx, query, newToken, expiresAt, oldToken)
	if err != nil {
		return fmt.Errorf("failed to update session token: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("session not found")
	}
	return nil
}

func (r *SessionRepository) DeleteByUserID(ctx context.Context, userID uuid.UUID) error {
	query := `DELETE FROM sessions WHERE user_id = $1`
	_, err := r.db.Pool.Exec(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to delete sessions by user ID: %w", err)
	}
	return nil
}
