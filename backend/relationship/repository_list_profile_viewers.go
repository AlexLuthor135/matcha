package relationship

import (
	"backend/models"
	"context"
)

func (repo *PostgresRepository) ListProfileViewers(ctx context.Context, userID uint) ([]models.ProfileViewer, error) {
	const query = `
		SELECT
			users.id,
			users.user_name,
			users.first_name,
			users.last_name,
			users.avatar,
			profile_views.updated_at
		FROM profile_views
		JOIN users
			ON users.id = profile_views.viewer_id
		WHERE profile_views.viewed_user_id = $1
		AND NOT EXISTS (
    		SELECT 1 FROM user_blocks AS ub WHERE
				(ub.blocker_id = $1 AND ub.blocked_user_id = users.id)
        OR (ub.blocker_id = users.id AND ub.blocked_user_id = $1))
		ORDER BY
			profile_views.updated_at DESC,
			profile_views.id DESC
	`
	rows, err := repo.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	viewers := make([]models.ProfileViewer, 0)
	for rows.Next() {
		var v models.ProfileViewer
		if err := rows.Scan(&v.ID, &v.UserName, &v.FirstName, &v.LastName, &v.Avatar, &v.LastViewedAt); err != nil {
			return nil, err
		}
		viewers = append(viewers, v)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return viewers, nil
}
