package user

import (
	"context"
	"database/sql"
	"errors"
)

func (repo *PostgresRepository) UpdatePasswordHash(ctx context.Context, userID uint, passwordHash string) error {
	const query = `UPDATE users SET password = $1, updated_at = now() WHERE id = $2 RETURNING id`
	var updateUserID uint
	err := repo.db.QueryRowContext(ctx, query, passwordHash, userID).Scan(&updateUserID)
	if errors.Is(err, sql.ErrNoRows) {
		return UserErrors.UserNotFound
	}
	return err
}
