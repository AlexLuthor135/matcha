package chat

import (
	"backend/models"
	"context"
)

func (repo *PostgresRepository) ListConversations(ctx context.Context, userID uint) ([]models.Conversation, error) {
	const query = `
	SELECT c.id, c.user_one_id, c.user_two_id, c.created_at, c.updated_at
	FROM conversations AS c
	WHERE (c.user_one_id = $1 OR c.user_two_id = $1)
		AND EXISTS (
			SELECT 1
			FROM profile_decisions AS first_decision
			WHERE first_decision.user_id = c.user_one_id
				AND first_decision.target_user_id = c.user_two_id
				AND first_decision.decision = 'like'
		)
		AND EXISTS (
			SELECT 1
			FROM profile_decisions AS second_decision
			WHERE second_decision.user_id = c.user_two_id
				AND second_decision.target_user_id = c.user_one_id
				AND second_decision.decision = 'like'
		)
		AND NOT EXISTS (
			SELECT 1
			FROM user_blocks AS ub
			WHERE
				(ub.blocker_id = c.user_one_id AND ub.blocked_user_id = c.user_two_id)
				OR (ub.blocker_id = c.user_two_id AND ub.blocked_user_id = c.user_one_id))
	ORDER BY
		c.updated_at DESC,
		c.id DESC
`
	rows, err := repo.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	conversations := make([]models.Conversation, 0)
	for rows.Next() {
		var c models.Conversation
		if err := rows.Scan(&c.ID, &c.UserOneID, &c.UserTwoID, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		conversations = append(conversations, c)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return conversations, nil
}
