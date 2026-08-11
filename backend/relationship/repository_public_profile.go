package relationship

import (
	"backend/models"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
)

func (repo *PostgresRepository) GetPublicProfile(ctx context.Context, userID uint) (models.User, error) {
	const query = `
	SELECT
		u.id,
		u.user_name,
		u.first_name,
		u.last_name,
		u.is_completed,
		u.avatar,
		u.gender,
		u.preferences,
		u.bio,
		u.interests,
		u.birth_date,
		u.latitude,
		u.longitude,
		u.last_seen_at,
		(SELECT COUNT(*) FROM profile_decisions as pd WHERE pd.target_user_id = u.id AND pd.decision = 'like') as fame_rating
	FROM users AS u WHERE u.id = $1
	`
	var profile models.User
	var rawInterests []byte

	err := repo.db.QueryRowContext(ctx, query, userID).Scan(
		&profile.ID,
		&profile.UserName,
		&profile.FirstName,
		&profile.LastName,
		&profile.IsCompleted,
		&profile.Avatar,
		&profile.Gender,
		&profile.Preferences,
		&profile.Bio,
		&rawInterests,
		&profile.BirthDate,
		&profile.Latitude,
		&profile.Longitude,
		&profile.LastSeenAt,
		&profile.FameRating,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return models.User{}, RelationshipErrors.UserNotFound
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
