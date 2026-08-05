package user

import (
	"backend/models"
	"context"
)

func (repo *PostgresRepository) ListMatches(ctx context.Context, userID uint) ([]models.Match, error) {
	const query = `
	SELECT
		id,
		user_name,
		first_name,
		last_name,
		avatar
	FROM users
	WHERE id IN (
		SELECT target_user_id
		FROM profile_decisions
		WHERE user_id = $1
			AND decision = 'like'
	)
	AND id IN (
		SELECT user_id
		FROM profile_decisions
		WHERE target_user_id = $1
			AND decision = 'like'
	)
	ORDER BY user_name ASC
	`
	rows, err := repo.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	matches := make([]models.Match, 0)
	for rows.Next() {
		var match models.Match
		if err := rows.Scan(&match.ID, &match.UserName, &match.FirstName, &match.LastName, &match.Avatar); err != nil {
			return nil, err
		}
		matches = append(matches, match)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return matches, nil
}
