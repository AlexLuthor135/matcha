package user

import (
	"backend/models"
	"context"
	"database/sql"
	"errors"
)

func (repo *PostgresRepository) CreateAuthSession(ctx context.Context, session models.AuthSession) error {
	const query = `INSERT INTO auth_sessions (id, user_id, expires_at) VALUES ($1, $2, $3)`
	_, err := repo.db.ExecContext(ctx, query, session.ID, session.UserID, session.ExpiresAt)
	return err
}

func (repo *PostgresRepository) UseAuthSession(ctx context.Context, sessionID string, userID uint) error {
	const query = `UPDATE auth_sessions SET last_used_at = now() WHERE id = $1 AND user_id = $2 AND revoked_at IS NULL AND expires_at > now() RETURNING id`
	var usedSessionID string
	err := repo.db.QueryRowContext(ctx, query, sessionID, userID).Scan(&usedSessionID)
	if errors.Is(err, sql.ErrNoRows) {
		return UserErrors.InvalidAuthSession
	}
	return err
}

func (repo *PostgresRepository) RevokeAuthSession(ctx context.Context, sessionID string, userID uint) error {
	const query = `UPDATE auth_sessions SET revoked_at = COALESCE(revoked_at, now()) WHERE id = $1 AND user_id = $2`
	_, err := repo.db.ExecContext(ctx, query, sessionID, userID)
	return err
}

func (repo *PostgresRepository) RevokeAllAuthSessions(ctx context.Context, userID uint) error {
	const query = `UPDATE auth_sessions SET revoked_at = COALESCE(revoked_at, now()) WHERE user_id = $1 AND revoked_at IS NULL`
	_, err := repo.db.ExecContext(ctx, query, userID)
	return err
}
