package user

import (
	"backend/models"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
)

func (repo *PostgresRepository) GetProfile(ctx context.Context, userID uint) (models.User, error) {
	const query = `SELECT id, user_name, first_name, last_name, email, avatar, gender, preferences, bio, interests FROM users WHERE id = $1`
	var profile models.User
	var rawInterests []byte

	err := repo.db.QueryRowContext(ctx, query, userID).Scan(
		&profile.ID,
		&profile.UserName,
		&profile.FirstName,
		&profile.LastName,
		&profile.Email,
		&profile.Avatar,
		&profile.Gender,
		&profile.Preferences,
		&profile.Bio,
		&rawInterests,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return models.User{}, UserErrors.UserNotFound
	}
	if err != nil {
		return models.User{}, err
	}
	if err := json.Unmarshal(rawInterests, &profile.Interests); err != nil {
		return models.User{}, err
	}
	const photoQuery = `SELECT id, url FROM photos WHERE user_id = $1 ORDER BY created_at ASC, id ASC`
	rows, err := repo.db.QueryContext(ctx, photoQuery, userID)
	if err != nil {
		return models.User{}, err
	}
	defer rows.Close()
	profile.Photos = make([]models.Photo, 0, 5)
	for rows.Next() {
		photo := models.Photo{UserID: userID}
		if err := rows.Scan(&photo.ID, &photo.URL); err != nil {
			return models.User{}, err
		}
		profile.Photos = append(profile.Photos, photo)
	}
	if err := rows.Err(); err != nil {
		return models.User{}, err
	}
	return profile, nil
}
