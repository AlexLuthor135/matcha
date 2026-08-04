package user

import (
	"backend/models"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
)

func (repo *PostgresRepository) GetProfileFeed(ctx context.Context, userID uint, limit int) ([]models.User, error) {
	const currentUserQuery = `SELECT gender, preferences FROM users WHERE id = $1`
	var currentGender string
	var currentPreferences string
	err := repo.db.QueryRowContext(ctx, currentUserQuery, userID).Scan(&currentGender, &currentPreferences)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, UserErrors.UserNotFound
	}
	if err != nil {
		return nil, err
	}
	currentGender = strings.TrimSpace(currentGender)
	currentPreferences = strings.TrimSpace(currentPreferences)
	const profileQuery = `
		SELECT
			id,
			user_name,
			first_name,
			last_name,
			gender,
			preferences,
			bio,
			interests,
			avatar 
		FROM users
		WHERE id <> $1
			AND is_completed = true
			AND ($2 = '' OR gender = $2)
			AND ($3 = '' OR preferences = $3)
		ORDER BY random()
		LIMIT $4
	`
	rows, err := repo.db.QueryContext(ctx, profileQuery, userID, currentPreferences, currentGender, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	profiles := make([]models.User, 0, limit)
	profileIDs := make([]int64, 0, limit)
	profilePositions := make(map[uint]int, limit)
	for rows.Next() {
		var p models.User
		var rawInterests []byte
		if err := rows.Scan(&p.ID, &p.UserName, &p.FirstName, &p.LastName, &p.Gender, &p.Preferences, &p.Bio, &rawInterests, &p.Avatar); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(rawInterests, &p.Interests); err != nil {
			return nil, err
		}
		p.Photos = make([]models.Photo, 0, 5)
		profiles = append(profiles, p)
		profileIDs = append(profileIDs, int64(p.ID))
		profilePositions[p.ID] = len(profiles) - 1
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(profileIDs) == 0 {
		return profiles, nil
	}
	const photosQuery = `SELECT user_id, id, url FROM photos WHERE user_id = ANY($1::bigint[]) ORDER BY user_id ASC, created_at ASC, id ASC`
	photoRows, err := repo.db.QueryContext(ctx, photosQuery, profileIDs)
	if err != nil {
		return nil, err
	}
	defer photoRows.Close()
	for photoRows.Next() {
		var photo models.Photo
		if err := photoRows.Scan(&photo.UserID, &photo.ID, &photo.URL); err != nil {
			return nil, err
		}
		position, exists := profilePositions[photo.UserID]
		if !exists {
			continue
		}
		profiles[position].Photos = append(profiles[position].Photos, photo)
	}
	if err := photoRows.Err(); err != nil {
		return nil, err
	}
	return profiles, nil
}
