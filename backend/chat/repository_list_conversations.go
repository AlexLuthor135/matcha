package chat

import (
	"backend/models"
	"context"
)

func (repo *PostgresRepository) ListConversations(ctx context.Context, userID uint) ([]models.Conversation, error) {
	const query = `
	SELECT
		id,
		user_one_id,
		user_two_id,
		created_at,
		updated_at
	FROM conversations
	WHERE
		user_one_id = $1
		OR user_two_id = $1
	ORDER BY
		updated_at DESC,
		id DESC
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
