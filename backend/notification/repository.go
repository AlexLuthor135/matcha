package notification

import (
	"backend/models"
	"context"
	"database/sql"
)

type Repository interface {
	ListNotifications(ctx context.Context, userID uint, limit uint) ([]models.Notification, error)
	MarkNotificationRead(ctx context.Context, userID uint, notificationID uint) (models.NotificationReadReceipt, error)
	CreateNotification(ctx context.Context, input CreateNotificationInput) (models.Notification, error)
}

type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{
		db: db,
	}
}
