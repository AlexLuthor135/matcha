package notification

import (
	"backend/models"
	"context"
)

func (s *Service) MarkNotificationRead(ctx context.Context, userID uint, notificationID uint) (models.NotificationReadReceipt, error) {
	return s.repository.MarkNotificationRead(ctx, userID, notificationID)
}
