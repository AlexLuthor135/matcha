package user

import (
	"context"
	"database/sql"
	"errors"
)

func (repo *PostgresRepository) GetCompletionStatus(ctx context.Context, userID uint) (bool, error) {
	const query = `SELECT is_completed FROM users WHERE id = $1`
	var isCompleted bool
	err := repo.db.QueryRowContext(ctx, query, userID).Scan(&isCompleted)
	if errors.Is(err, sql.ErrNoRows) {
		return false, UserErrors.UserNotFound
	}
	if err != nil {
		return false, err
	}
	return isCompleted, nil
}
