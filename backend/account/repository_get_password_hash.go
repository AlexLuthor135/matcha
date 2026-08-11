package account

import (
	"context"
	"database/sql"
	"errors"
)

func (repo *PostgresRepository) GetPasswordHash(ctx context.Context, userID uint) (string, error) {
	const query = `SELECT password FROM users WHERE id = $1`
	var passwordHash string
	err := repo.db.QueryRowContext(ctx, query, userID).Scan(&passwordHash)
	if errors.Is(err, sql.ErrNoRows) {
		return "", AccountErrors.UserNotFound
	}
	if err != nil {
		return "", err
	}
	return passwordHash, nil
}
