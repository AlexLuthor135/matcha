package chat

import (
	"backend/models"
	"context"
)

func (s *Service) MarkMessageRead(ctx context.Context, recipientID uint, messageID uint) (models.MessageReadReceipt, error) {
	return s.repository.MarkMessageRead(ctx, recipientID, messageID)
}
