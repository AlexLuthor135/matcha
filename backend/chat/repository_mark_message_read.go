package chat

import (
	"backend/models"
	"context"
	"database/sql"
	"errors"
)

func (repo *PostgresRepository) MarkMessageRead(ctx context.Context, recipientID uint, messageID uint) (models.MessageReadReceipt, error) {
	const query = `
	UPDATE chat_messages AS message
	SET read_at = COALESCE(message.read_at, now()), updated_at = now()
	WHERE message.id = $1 AND message.recipient_id = $2
	AND EXISTS (
		SELECT 1 FROM profile_decisions AS sender_decision
		WHERE sender_decision.user_id = message.sender_id
		AND sender_decision.target_user_id = message.recipient_id
		AND sender_decision.decision = 'like')
	AND EXISTS (
			SELECT 1 FROM profile_decisions AS recipient_decision
			WHERE recipient_decision.user_id = message.recipient_id
			AND recipient_decision.target_user_id = message.sender_id
			AND recipient_decision.decision = 'like')
	AND NOT EXISTS (
			SELECT 1
			FROM user_blocks AS ub
			WHERE (ub.blocker_id = message.sender_id AND ub.blocked_user_id = message.recipient_id)
			OR (ub.blocker_id = message.recipient_id AND ub.blocked_user_id = message.sender_id))
	RETURNING
		message.id,
		message.conversation_id,
		message.sender_id,
		message.recipient_id,
		message.read_at
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
