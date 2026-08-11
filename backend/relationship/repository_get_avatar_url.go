package relationship

import (
	"context"
	"database/sql"
	"errors"
)

func (repo *PostgresRepository) GetAvatarURL(ctx context.Context, userID uint) (string, error) {
	const query = `SELECT avatar FROM users WHERE id = $1`
	var avatarURL string
	err := repo.db.QueryRowContext(ctx, query, userID).Scan(&avatarURL)
	if errors.Is(err, sql.ErrNoRows) {
		return "", RelationshipErrors.UserNotFound
	}
	if err != nil {
		return "", err
	}
	return avatarURL, nil
}
