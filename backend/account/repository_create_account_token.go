package account

import (
	"backend/models"
	"context"
	"database/sql"
)

func createAccountToken(ctx context.Context, tx *sql.Tx, token models.AccountToken) error {
	if !token.Purpose.IsValid() {
		return AccountErrors.InvalidAccountTokenPurpose
	}
	const query = `
		INSERT INTO account_tokens(user_id, token_hash, purpose, expires_at)
		VALUES ($1, $2, $3, $4)`
	_, err := tx.ExecContext(ctx, query, token.UserID, token.Hash, token.Purpose, token.ExpiresAt)
	return err
}
