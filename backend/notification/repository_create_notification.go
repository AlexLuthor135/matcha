package notification

import (
	"backend/models"
	"context"
	"encoding/json"
)

func (repo *PostgresRepository) CreateNotification(ctx context.Context, input CreateNotificationInput) (models.Notification, error) {
	if input.Data == nil {
		input.Data = make(map[string]any)
	}
	rawData, err := json.Marshal(input.Data)
	if err != nil {
		return models.Notification{}, err
	}
	var senderValue any
	if input.SenderID != nil {
		const blockQuery = `SELECT EXISTS (SELECT 1 FROM user_blocks WHERE (blocker_id = $1 AND blocked_user_id = $2) OR (blocker_id = $2 AND blocked_user_id = $1))`
		var usersBlocked bool
		err := repo.db.QueryRowContext(ctx, blockQuery, input.UserID, *input.SenderID).Scan(&usersBlocked)
		if err != nil {
			return models.Notification{}, err
		}
		if usersBlocked {
			return models.Notification{}, NotificationErrors.UserBlocked
		}
		senderValue = *input.SenderID
	}
	const query = `
		INSERT INTO notifications (
			user_id,
			sender_id,
			type,
			title,
			message,
			data
		)
		VALUES ($1, $2, $3, $4, $5, $6::jsonb)
		ON CONFLICT (user_id, sender_id)
		WHERE type = 'match'
		DO UPDATE SET updated_at = notifications.updated_at
		RETURNING id, created_at, updated_at
	`
	notification := models.Notification{
		UserID:   input.UserID,
		SenderID: input.SenderID,
		Type:     input.Type,
		Title:    input.Title,
		Message:  input.Message,
		Data:     input.Data,
	}
	err = repo.db.QueryRowContext(
		ctx,
		query,
		input.UserID,
		senderValue,
		input.Type,
		input.Title,
		input.Message,
		string(rawData)).Scan(
		&notification.ID,
		&notification.CreatedAt,
		&notification.UpdatedAt,
	)
	if err != nil {
		return models.Notification{}, err
	}
	return notification, nil
}
