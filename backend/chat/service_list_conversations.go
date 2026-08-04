package chat

import (
	"backend/models"
	"context"
)

func (s *Service) ListConversations(ctx context.Context, userID uint) ([]models.Conversation, error) {
	return s.repository.ListConversations(ctx, userID)
}
