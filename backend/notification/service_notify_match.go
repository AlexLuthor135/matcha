package notification

import (
	"backend/models"
	"context"
	"encoding/json"
)

const NotificationTypeMatch = "match"

type notificationEvent struct {
	Type         string              `json:"type"`
	Notification models.Notification `json:"notification"`
}

func (s *Service) NotifyMatch(ctx context.Context, recipientID uint, matchedUserID uint) (models.Notification, error) {
	savedNotification, err := s.CreateNotification(
		ctx,
		CreateNotificationInput{
			UserID:   recipientID,
			SenderID: &matchedUserID,
			Type:     NotificationTypeMatch,
			Title:    "New match",
			Message:  "You have a new match",
			Data: map[string]any{
				"matched_user_id": matchedUserID,
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
