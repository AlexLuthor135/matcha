package notification

import (
	"backend/models"
	"context"
	"encoding/json"
)

const NotificationTypeProfileView = "profile_view"

func (s *Service) NotifyProfileView(ctx context.Context, recipientID uint, viewerID uint) (models.Notification, error) {
	savedNotification, err := s.CreateNotification(ctx, CreateNotificationInput{
		UserID:   recipientID,
		SenderID: &viewerID,
		Type:     NotificationTypeProfileView,
		Title:    "Profile viewed",
		Message:  "Someone viewed your profile",
		Data:     map[string]any{"viewer_id": viewerID},
	})
	if err != nil {
		return models.Notification{}, err
	}
	if s.publisher == nil {
		return savedNotification, nil
	}
	eventData, err := json.Marshal(notificationEvent{Type: "notification", Notification: savedNotification})
	if err != nil {
		return savedNotification, err
	}
	s.publisher.SendToUser(recipientID, eventData)
	return savedNotification, nil
}
