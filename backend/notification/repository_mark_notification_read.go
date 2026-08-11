package notification

import (
	"backend/models"
	"context"
	"database/sql"
	"errors"
)

func (repo *PostgresRepository) MarkNotificationRead(ctx context.Context, userID uint, notificationID uint) (models.NotificationReadReceipt, error) {
	const query = `UPDATE notifications SET read_at = COALESCE(read_at, now()), updated_at = now() WHERE id = $1 AND user_id = $2 RETURNING id, read_at`
	var notificationReceipt models.NotificationReadReceipt
	err := repo.db.QueryRowContext(ctx, query, notificationID, userID).Scan(&notificationReceipt.NotificationID, &notificationReceipt.ReadAt)
	if errors.Is(err, sql.ErrNoRows) {
		return models.NotificationReadReceipt{}, NotificationErrors.NotificationNotFound
	}
	if err != nil {
		return models.NotificationReadReceipt{}, err
	}
	return notificationReceipt, nil

}
