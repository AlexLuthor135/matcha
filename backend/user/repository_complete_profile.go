package user

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
)

func (repo *PostgresRepository) CompleteProfile(ctx context.Context, userID uint, gender string, preferences string, bio string, interests []string) (bool, error) {
	rawInterests, err := json.Marshal(interests)
	if err != nil {
		return false, err
	}
	const query = `
		UPDATE users
		SET
			gender = $1,
			preferences = $2,
			bio = $3,
			interests = $4::jsonb,
			is_completed = true,
			updated_at = now()
		WHERE id = $5
		RETURNING is_completed
	`
	var isCompleted bool
	err = repo.db.QueryRowContext(ctx, query, gender, preferences, bio, string(rawInterests), userID).Scan(&isCompleted)
	if errors.Is(err, sql.ErrNoRows) {
		return false, UserErrors.UserNotFound
	}
	if err != nil {
		return false, err
	}
	return isCompleted, nil
}
