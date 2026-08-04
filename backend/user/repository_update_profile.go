package user

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
)

func (repo *PostgresRepository) UpdateProfile(ctx context.Context, userID uint, gender *string, preferences *string, bio *string, interests *[]string) error {
	var interestsValue any
	if interests != nil {
		rawInterests, err := json.Marshal(*interests)
		if err != nil {
			return err
		}
		interestsValue = string(rawInterests)
	}
	const query = `
		UPDATE users
		SET
			gender = COALESCE($1, gender),
			preferences = COALESCE($2, preferences),
			bio = COALESCE($3, bio),
			interests = COALESCE($4::jsonb, interests),
			updated_at = now()
		WHERE id = $5
		RETURNING id
		`
	var updatedUserID uint
	err := repo.db.QueryRowContext(ctx, query, optionalString(gender), optionalString(preferences), optionalString(bio), interestsValue, userID).Scan(&updatedUserID)
	if errors.Is(err, sql.ErrNoRows) {
		return UserErrors.UserNotFound
	}
	return err
}
