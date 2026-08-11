package profile

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
)

func (repo *PostgresRepository) CompleteProfile(ctx context.Context, userID uint, input CompleteProfileInput) (bool, error) {
	rawInterests, err := json.Marshal(input.Interests)
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
			birth_date = $5,
			location_source = $6,
			location_name = $7,
			location_consent_at = $8,
			latitude = $9,
			longitude = $10,
			is_completed = true,
			updated_at = now()
		WHERE id = $11
		RETURNING is_completed
	`
	var isCompleted bool
	err = repo.db.QueryRowContext(
		ctx,
		query,
		input.Gender,
		input.Preferences,
		input.Bio,
		string(rawInterests),
		input.BirthDate,
		string(input.Location.Source),
		input.Location.Name,
		input.Location.ConsentAt,
		input.Location.Latitude,
		input.Location.Longitude,
		userID).Scan(
		&isCompleted,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return false, ProfileErrors.UserNotFound
	}
	if err != nil {
		return false, err
	}
	return isCompleted, nil
}
