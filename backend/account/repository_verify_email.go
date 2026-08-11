package account

import (
	"backend/models"
	"context"
	"database/sql"
	"errors"
	"github.com/jackc/pgx/v5/pgconn"
)

func (repo *PostgresRepository) VerifyEmail(ctx context.Context, tokenHash string) error {
	tx, err := repo.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	const findTokenQuery = `SELECT user_id FROM account_tokens WHERE token_hash = $1 AND purpose = $2 AND used_at IS NULL AND expires_at > now() FOR UPDATE`
	var userID uint
	err = tx.QueryRowContext(ctx, findTokenQuery, tokenHash, string(models.AccountTokenPurposeEmailVerification)).Scan(&userID)
	if errors.Is(err, sql.ErrNoRows) {
		return AccountErrors.InvalidVerificationToken
	}
	if err != nil {
		return err
	}
	const verifyUserQuery = `UPDATE users SET email = COALESCE(pending_email, email), pending_email = NULL, is_verified = true, updated_at = now() WHERE id = $1`
	if _, err := tx.ExecContext(ctx, verifyUserQuery, userID); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return AccountErrors.UserAlreadyExists
		}
		return err
	}

	const consumeTokenQuery = `UPDATE account_tokens SET used_at = now() WHERE user_id = $1 AND purpose = $2 AND used_at IS NULL`
	if _, err := tx.ExecContext(ctx, consumeTokenQuery, userID, models.AccountTokenPurposeEmailVerification); err != nil {
		return err
	}
	return tx.Commit()
}
