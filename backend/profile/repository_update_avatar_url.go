package profile

import (
	"context"
	"database/sql"
	"errors"
)

func (repo *PostgresRepository) UpdateAvatarURL(ctx context.Context, userID uint, avatarURL string) (string, error) {
	const query = `UPDATE users SET avatar = $1, updated_at = now() WHERE id = $2 RETURNING avatar`
	var updatedAvatarURL string
	err := repo.db.QueryRowContext(ctx, query, avatarURL, userID).Scan(&updatedAvatarURL)
	if errors.Is(err, sql.ErrNoRows) {
		return "", ProfileErrors.UserNotFound
	}
	if err != nil {
		return "", err
	}
	return updatedAvatarURL, nil
}
