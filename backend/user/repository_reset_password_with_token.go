package user

import (
	"backend/models"
	"context"
	"database/sql"
	"errors"
)

func (repo *PostgresRepository) ResetPasswordWithToken(ctx context.Context, tokenHash string, newPasswordHash string) error {
	tx, err := repo.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	const findTokenQuery = `SELECT user_id FROM account_tokens WHERE token_hash = $1 AND purpose = $2 AND used_at IS NULL AND expires_at > now() FOR UPDATE`
	var userID uint
	err = tx.QueryRowContext(ctx, findTokenQuery, tokenHash, string(models.AccountTokenPurposePasswordReset)).Scan(&userID)
	if errors.Is(err, sql.ErrNoRows) {
		return UserErrors.InvalidPasswordResetToken
	}
	if err != nil {
		return err
	}
	const updatePasswordQuery = `UPDATE users SET password = $1, updated_at = now() WHERE id = $2`
	if _, err := tx.ExecContext(ctx, updatePasswordQuery, newPasswordHash, userID); err != nil {
		return err
	}
	const revokeSessionsQuery = `UPDATE auth_sessions SET revoked_at = now() WHERE user_id = $1 AND revoked_at IS NULL`
	if _, err := tx.ExecContext(ctx, revokeSessionsQuery, userID); err != nil {
		return err
	}
	const consumeTokenQuery = `UPDATE account_tokens SET used_at = now() WHERE user_id = $1 AND purpose = $2 AND used_at IS NULL`
	if _, err := tx.ExecContext(ctx, consumeTokenQuery, userID, string(models.AccountTokenPurposePasswordReset)); err != nil {
		return err
	}
	return tx.Commit()
}
