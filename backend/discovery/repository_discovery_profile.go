package discovery

import (
	"backend/models"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
)

func (repo *PostgresRepository) GetDiscoveryProfile(ctx context.Context, userID uint) (models.User, error) {
	const query = `
	SELECT
		u.id,
		u.gender,
		u.preferences,
		u.interests,
		u.latitude,
		u.longitude
	FROM users AS u WHERE u.id = $1
	`
	var profile models.User
	var rawInterests []byte

	err := repo.db.QueryRowContext(ctx, query, userID).Scan(
		&profile.ID,
		&profile.Gender,
		&profile.Preferences,
		&rawInterests,
		&profile.Latitude,
		&profile.Longitude,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return models.User{}, DiscoveryErrors.UserNotFound
	}
	if err != nil {
		return models.User{}, err
	}
	if err := json.Unmarshal(rawInterests, &profile.Interests); err != nil {
		return models.User{}, err
	}
	return profile, nil
}
