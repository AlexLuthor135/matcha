package notification

import (
	"backend/models"
	"context"
	"encoding/json"
)

const NotificationTypeUnlike = "unlike"

func (s *Service) NotifyUnlike(ctx context.Context, recipientID uint, unlikerID uint) (models.Notification, error) {
	savedNotification, err := s.CreateNotification(
		ctx,
		CreateNotificationInput{
			UserID:   recipientID,
			SenderID: &unlikerID,
			Type:     NotificationTypeUnlike,
			Title:    "Match ended",
			Message:  "Someone unmatched with you",
			Data: map[string]any{
				"unliker_id": unlikerID,
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
