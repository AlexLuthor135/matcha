package notification

import (
	"backend/models"
	"context"
	"encoding/json"
)

const NotificationTypeMessage = "message"

func (s *Service) NotifyMessage(ctx context.Context, message models.Message) (models.Notification, error) {
	if message.ID == 0 {
		return models.Notification{}, NotificationErrors.InvalidMessageID
	}
	if message.ConversationID == 0 {
		return models.Notification{}, NotificationErrors.InvalidConversationID
	}
	savedNotification, err := s.CreateNotification(
		ctx,
		CreateNotificationInput{
			UserID:   message.RecipientID,
			SenderID: &message.SenderID,
			Type:     NotificationTypeMessage,
			Title:    "New message",
			Message:  "You received a new message",
			Data: map[string]any{
				"message_id":      message.ID,
				"conversation_id": message.ConversationID,
				"sender_id":       message.SenderID,
			},
		},
	)
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
	s.publisher.SendToUser(message.RecipientID, eventData)
	return savedNotification, nil
}
