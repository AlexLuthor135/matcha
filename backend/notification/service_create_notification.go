package notification

import (
	"backend/models"
	"context"
	"strings"
)

type CreateNotificationInput struct {
	UserID   uint
	SenderID *uint
	Type     string
	Title    string
	Message  string
	Data     map[string]any
}

func (input *CreateNotificationInput) Normalize() {
	input.Type = strings.TrimSpace(input.Type)
	input.Title = strings.TrimSpace(input.Title)
	input.Message = strings.TrimSpace(input.Message)
}

func (s *Service) CreateNotification(ctx context.Context, input CreateNotificationInput) (models.Notification, error) {
	input.Normalize()
	if input.UserID == 0 {
		return models.Notification{}, NotificationErrors.InvalidUserID
	}
	if input.SenderID != nil && *input.SenderID == 0 {
		return models.Notification{}, NotificationErrors.InvalidSenderID
	}
	if input.Type == "" {
		return models.Notification{}, NotificationErrors.TypeBlank
	}
	if input.Message == "" {
		return models.Notification{}, NotificationErrors.MessageBlank
	}
	if input.Data == nil {
		input.Data = make(map[string]any)
	}
	return s.repository.CreateNotification(ctx, input)
}
