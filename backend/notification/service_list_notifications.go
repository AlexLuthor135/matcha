package notification

import (
	"backend/models"
	"context"
)

func (s *Service) ListNotifications(ctx context.Context, userID uint) ([]models.Notification, error) {
	return s.repository.ListNotifications(ctx, userID, defaultNotificationLimit)
}
