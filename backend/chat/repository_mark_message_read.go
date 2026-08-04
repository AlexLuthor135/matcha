package chat

import (
	"backend/models"
	"context"
	"database/sql"
	"errors"
)

func (repo *PostgresRepository) MarkMessageRead(ctx context.Context, recipientID uint, messageID uint) (models.MessageReadReceipt, error) {
	const query = `
		UPDATE 
			chat_messages
		SET 
			read_at = COALESCE(read_at, now()), updated_at = now() 
		WHERE 
			id = $1 AND recipient_id = $2 
		RETURNING 
			id, conversation_id, sender_id, recipient_id, read_at
		`
	var receipt models.MessageReadReceipt
	err := repo.db.QueryRowContext(ctx, query, messageID, recipientID).Scan(&receipt.MessageID, &receipt.ConversationID, &receipt.SenderID, &receipt.RecipientID, &receipt.ReadAt)
	if errors.Is(err, sql.ErrNoRows) {
		return models.MessageReadReceipt{}, ChatErrors.MessageNotFound
	}
	if err != nil {
		return models.MessageReadReceipt{}, err
	}
	return receipt, nil
}
