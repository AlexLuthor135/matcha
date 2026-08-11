package notification

import (
	"backend/models"
	"context"
	"encoding/json"
)

const NotificationTypeLike = "like"

func (s *Service) NotifyLike(ctx context.Context, recipientID uint, likerID uint) (models.Notification, error) {
	savedNotification, err := s.CreateNotification(
		ctx,
		CreateNotificationInput{
			UserID:   recipientID,
			SenderID: &likerID,
			Type:     NotificationTypeLike,
			Title:    "New like",
			Message:  "Someone liked your profile",
			Data: map[string]any{
				"liker_id": likerID,
			},
		},
	)
	if err != nil {
		return models.Notification{}, err
	}
	if s.publisher == nil {
		return savedNotification, nil
	}
	eventData, err := json.Marshal(notificationEvent{
		Type:         "notification",
		Notification: savedNotification,
	})
	if err != nil {
		return savedNotification, err
	}
	s.publisher.SendToUser(recipientID, eventData)
	return savedNotification, nil
}
