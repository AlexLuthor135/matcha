package chat

import (
	"backend/models"
	"context"
	"strings"
	"unicode/utf8"
)

func (s *Service) SaveMessage(ctx context.Context, senderID uint, recipientID uint, content string) (models.Message, error) {
	content = strings.TrimSpace(content)
	if senderID == recipientID {
		return models.Message{}, ChatErrors.CannotMessageSelf
	}
	if content == "" {
		return models.Message{}, ChatErrors.MessageBlank
	}
	if utf8.RuneCountInString(content) > maxMessageLength {
		return models.Message{}, ChatErrors.MessageTooLong
	}
	return s.repository.CreateMessage(ctx, senderID, recipientID, content)
}
