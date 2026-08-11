package account

import (
	"context"
	"database/sql"
	"errors"
)

func (repo *PostgresRepository) UpdatePasswordHashAndRevokeSessions(ctx context.Context, userID uint, passwordHash string) error {
	tx, err := repo.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	const updatePasswordQuery = `UPDATE users SET password = $1, updated_at = now() WHERE id = $2 RETURNING id`
	var updateUserID uint
	err = tx.QueryRowContext(ctx, updatePasswordQuery, passwordHash, userID).Scan(&updateUserID)
	if errors.Is(err, sql.ErrNoRows) {
		return AccountErrors.UserNotFound
	}
	if err != nil {
		return err
	}
	const revokeSessionQuery = `UPDATE auth_sessions SET revoked_at = now() WHERE user_id = $1 AND revoked_at IS NULL`
	if _, err := tx.ExecContext(ctx, revokeSessionQuery, userID); err != nil {
		return err
	}
	return tx.Commit()
}
