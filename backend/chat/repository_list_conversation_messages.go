package chat

import (
	"backend/models"
	"context"
	"database/sql"
	"errors"
)

func (repo *PostgresRepository) ListConversationMessages(ctx context.Context, userID uint, conversationID uint) ([]models.Message, error) {
	const accessQuery = `
	SELECT c.id
	FROM conversations AS c
	WHERE c.id = $1
	AND (c.user_one_id = $2 OR c.user_two_id = $2)
	AND EXISTS (
		SELECT 1 FROM profile_decisions AS first_decision
		WHERE first_decision.user_id = c.user_one_id
		AND first_decision.target_user_id = c.user_two_id
		AND first_decision.decision = 'like')
	AND EXISTS (
			SELECT 1 FROM profile_decisions AS second_decision
			WHERE second_decision.user_id = c.user_two_id
			AND second_decision.target_user_id = c.user_one_id
			AND second_decision.decision = 'like')
	AND NOT EXISTS (
			SELECT 1 FROM user_blocks AS ub
			WHERE (ub.blocker_id = c.user_one_id AND ub.blocked_user_id = c.user_two_id)
			OR (ub.blocker_id = c.user_two_id AND ub.blocked_user_id = c.user_one_id))
	`
	var accessibleConversationID uint
	err := repo.db.QueryRowContext(ctx, accessQuery, conversationID, userID).Scan(&accessibleConversationID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ChatErrors.ConversationNotFound
	}
	if err != nil {
		return nil, err
	}
	const messageQuery = `
		SELECT
			id,
			conversation_id,
			sender_id,
			recipient_id,
			content,
			created_at,
			updated_at,
			read_at
		FROM chat_messages
		WHERE conversation_id = $1
		ORDER BY
			created_at ASC,
			id ASC
		`
	rows, err := repo.db.QueryContext(ctx, messageQuery, accessibleConversationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	messages := make([]models.Message, 0)
	for rows.Next() {
		var m models.Message
		var readAt sql.NullTime
		if err := rows.Scan(&m.ID, &m.ConversationID, &m.SenderID, &m.RecipientID, &m.Content, &m.CreatedAt, &m.UpdatedAt, &readAt); err != nil {
			return nil, err
		}
		if readAt.Valid {
			value := readAt.Time
			m.ReadAt = &value
		}
		messages = append(messages, m)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return messages, nil
}
