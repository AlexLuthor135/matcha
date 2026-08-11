package user

import (
	"backend/models"
	"context"
)

func (repo *PostgresRepository) ListProfileLikers(ctx context.Context, userID uint) ([]models.ProfileLiker, error) {
	const query = `
		SELECT
			users.id,
			users.user_name,
			users.first_name,
			users.last_name,
			users.avatar,
			profile_decisions.updated_at
		FROM profile_decisions
		JOIN users
			ON users.id = profile_decisions.user_id
		WHERE profile_decisions.target_user_id = $1
			AND profile_decisions.decision = 'like'
		AND NOT EXISTS (SELECT 1 FROM user_blocks AS ub 
		WHERE
        	(ub.blocker_id = $1 AND ub.blocked_user_id = users.id)
        OR (ub.blocker_id = users.id AND ub.blocked_user_id = $1))
		ORDER BY
			profile_decisions.updated_at DESC,
			profile_decisions.id DESC
	`
	rows, err := repo.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	likers := make([]models.ProfileLiker, 0)
	for rows.Next() {
		var l models.ProfileLiker
		if err := rows.Scan(&l.ID, &l.UserName, &l.FirstName, &l.LastName, &l.Avatar, &l.LikedAt); err != nil {
			return nil, err
		}
		likers = append(likers, l)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return likers, nil
}
