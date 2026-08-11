package discovery

import (
	"backend/models"
	"context"
	"encoding/json"
)

func (repo *PostgresRepository) ListProfileCandidates(ctx context.Context, userID uint, preferredGender string, ownGender string, excludeDecided bool) ([]models.User, error) {
	const profileQuery = `
		SELECT
			u.id,
			u.user_name,
			u.first_name,
			u.last_name,
			u.gender,
			u.preferences,
			u.bio,
			u.interests,
			u.avatar,
			u.birth_date,
			u.latitude,
			u.longitude,
			(SELECT COUNT(*) FROM profile_decisions AS pd WHERE pd.target_user_id = u.id AND pd.decision = 'like') AS fame_rating
		FROM users AS u
		WHERE u.id <> $1
			AND u.is_completed = true
			AND ($2 = $5 OR u.gender = $2)
			AND (u.preferences = $5 OR u.preferences = $3)
			AND ($4::boolean = false OR NOT EXISTS (
				SELECT 1
				FROM profile_decisions AS own_decision
				WHERE own_decision.user_id = $1
				  AND own_decision.target_user_id = u.id
			))
			AND NOT EXISTS (
				SELECT 1
				FROM user_blocks AS ub
				WHERE (ub.blocker_id = $1 AND ub.blocked_user_id = u.id)
				OR (ub.blocker_id = u.id AND ub.blocked_user_id = $1))
	`
	rows, err := repo.db.QueryContext(ctx, profileQuery, userID, preferredGender, ownGender, excludeDecided, models.PreferenceEveryone)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	profiles := make([]models.User, 0)
	profileIDs := make([]int64, 0)
	profilePositions := make(map[uint]int)
	for rows.Next() {
		var p models.User
		var rawInterests []byte
		if err := rows.Scan(
			&p.ID,
			&p.UserName,
			&p.FirstName,
			&p.LastName,
			&p.Gender,
			&p.Preferences,
			&p.Bio,
			&rawInterests,
			&p.Avatar,
			&p.BirthDate,
			&p.Latitude,
			&p.Longitude,
			&p.FameRating); err != nil {
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
