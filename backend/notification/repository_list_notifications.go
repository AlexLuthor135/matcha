package notification

import (
	"backend/models"
	"context"
	"database/sql"
	"encoding/json"
)

func (repo *PostgresRepository) ListNotifications(ctx context.Context, userID uint, limit uint) ([]models.Notification, error) {
	const query = `
		SELECT
			id,
			created_at,
			updated_at,
			user_id,
			sender_id,
			type,
			title,
			message,
			data,
			read_at
		FROM notifications
		WHERE user_id = $1
		ORDER BY
			created_at DESC,
			id DESC
		LIMIT $2
	`
	rows, err := repo.db.QueryContext(ctx, query, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	notifications := make([]models.Notification, 0)
	for rows.Next() {
		var n models.Notification
		var senderID sql.NullInt64
		var readAt sql.NullTime
		var rawData []byte
		if err := rows.Scan(&n.ID, &n.CreatedAt, &n.UpdatedAt, &n.UserID, &senderID, &n.Type, &n.Title, &n.Message, &rawData, &readAt); err != nil {
			return nil, err
		}
		if senderID.Valid {
			value := uint(senderID.Int64)
			n.SenderID = &value
		}
		if readAt.Valid {
			value := readAt.Time
			n.ReadAt = &value
		}
		if err := json.Unmarshal(rawData, &n.Data); err != nil {
			return nil, err
		}
		notifications = append(notifications, n)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return notifications, nil
}
