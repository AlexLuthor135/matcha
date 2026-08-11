package chat

import (
	"backend/models"
	"context"
)

func (repo *PostgresRepository) CreateMessage(ctx context.Context, senderID uint, recipientID uint, content string) (models.Message, error) {
	tx, err := repo.db.BeginTx(ctx, nil)
	if err != nil {
		return models.Message{}, err
	}
	defer tx.Rollback()
	const userQuery = `SELECT EXISTS (SELECT 1 FROM users WHERE id = $1), EXISTS (SELECT 1 FROM users WHERE id = $2)`
	var senderExists bool
	var recipientExists bool
	err = tx.QueryRowContext(ctx, userQuery, senderID, recipientID).Scan(&senderExists, &recipientExists)
	if err != nil {
		return models.Message{}, err
	}
	if !senderExists {
		return models.Message{}, ChatErrors.SenderNotFound
	}
	if !recipientExists {
		return models.Message{}, ChatErrors.RecipientNotFound
	}
	const matchQuery = `
	SELECT
		EXISTS (
			SELECT 1 FROM profile_decisions
			WHERE user_id = $1
			AND target_user_id = $2
			AND decision = 'like')
		AND EXISTS (
			SELECT 1 FROM profile_decisions
			WHERE user_id = $2
			AND target_user_id = $1
			AND decision = 'like')
		AND NOT EXISTS (
			SELECT 1 FROM user_blocks AS ub
			WHERE (ub.blocker_id = $1 AND ub.blocked_user_id = $2)
			OR (ub.blocker_id = $2 AND ub.blocked_user_id = $1))
	`
	var usersAreMatched bool
	err = tx.QueryRowContext(ctx, matchQuery, senderID, recipientID).Scan(&usersAreMatched)
	if err != nil {
		return models.Message{}, err
	}
	if !usersAreMatched {
		return models.Message{}, ChatErrors.UsersNotMatched
	}
	userOneID := senderID
	userTwoID := recipientID
	if userOneID > userTwoID {
		userOneID, userTwoID = userTwoID, userOneID
	}
	const conversationQuery = `
		INSERT INTO conversations
			(user_one_id, user_two_id)
		VALUES
			($1, $2)
		ON CONFLICT ON CONSTRAINT 
			conversations_user_pair_unique
		DO UPDATE SET 
			updated_at = now()
		RETURNING
			id
		`
	var conversationID uint
	err = tx.QueryRowContext(ctx, conversationQuery, userOneID, userTwoID).Scan(&conversationID)
	if err != nil {
		return models.Message{}, err
	}
	const messageQuery = `INSERT INTO chat_messages(conversation_id, sender_id, recipient_id, content) VALUES ($1, $2, $3, $4) RETURNING id, created_at, updated_at`
	message := models.Message{
		ConversationID: conversationID,
		SenderID:       senderID,
		RecipientID:    recipientID,
		Content:        content,
	}
	err = tx.QueryRowContext(ctx, messageQuery, message.ConversationID, message.SenderID, message.RecipientID, message.Content).Scan(&message.ID, &message.CreatedAt, &message.UpdatedAt)
	if err != nil {
		return models.Message{}, err
	}
	if err := tx.Commit(); err != nil {
		return models.Message{}, err
	}
	return message, nil
}
