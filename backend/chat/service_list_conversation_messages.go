package chat

import (
	"backend/models"
	"context"
)

func (s *Service) ListConversationMessages(ctx context.Context, userID uint, conversationID uint) ([]models.Message, error) {
	return s.repository.ListConversationMessages(ctx, userID, conversationID)
}
