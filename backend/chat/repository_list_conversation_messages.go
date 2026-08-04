package chat

import (
	"backend/models"
	"context"
	"database/sql"
	"errors"
)

func (repo *PostgresRepository) ListConversationMessages(ctx context.Context, userID uint, conversationID uint) ([]models.Message, error) {
	const accessQuery = `SELECT id FROM conversations WHERE id = $1 AND (user_one_id = $2 OR user_two_id = $2)`
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
