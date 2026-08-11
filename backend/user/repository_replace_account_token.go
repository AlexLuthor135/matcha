package user

import (
	"backend/models"
	"context"
)

func (repo *PostgresRepository) ReplaceAccountToken(ctx context.Context, token models.AccountToken) error {
	if !token.Purpose.IsValid() {
		return UserErrors.InvalidAccountTokenPurpose
	}
	tx, err := repo.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	const invalidateTokenQuery = `UPDATE account_tokens SET used_at = now() WHERE user_id = $1 AND purpose = $2 AND used_at IS NULL`
	if _, err := tx.ExecContext(ctx, invalidateTokenQuery, token.UserID, string(token.Purpose)); err != nil {
		return err
	}
	if err := createAccountToken(ctx, tx, token); err != nil {
		return err
	}
	return tx.Commit()
}
