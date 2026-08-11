package relationship

import (
	"backend/models"
	"context"
)

func (repo *PostgresRepository) ListMatches(ctx context.Context, userID uint) ([]models.Match, error) {
	const query = `
		SELECT
			u.id,
			u.user_name,
			u.first_name,
			u.last_name,
			u.avatar,
			(SELECT COUNT(*) FROM profile_decisions AS pd WHERE pd.target_user_id = u.id AND pd.decision = 'like') AS fame_rating
		FROM users AS u
		WHERE u.id IN (
			SELECT target_user_id
			FROM profile_decisions
			WHERE user_id = $1
			  AND decision = 'like'
		)
		AND u.id IN (
			SELECT user_id
			FROM profile_decisions
			WHERE target_user_id = $1
			  AND decision = 'like'
		)
		ORDER BY u.user_name ASC
	`
	rows, err := repo.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	matches := make([]models.Match, 0)
	for rows.Next() {
		var match models.Match
		if err := rows.Scan(&match.ID, &match.UserName, &match.FirstName, &match.LastName, &match.Avatar, &match.FameRating); err != nil {
			return nil, err
		}
		matches = append(matches, match)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return matches, nil
}
