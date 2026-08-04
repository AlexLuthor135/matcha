package user

import (
	"backend/models"
	"context"
)

const (
	defaultProfileFeedLimit = 20
	maxProfileFeedLimit     = 50
)

func (s *Service) GetProfileFeed(ctx context.Context, userID uint, limit int) ([]models.User, error) {
	if limit < 1 {
		return nil, UserErrors.InvalidProfileFeedLimit
	}
	if limit > maxProfileFeedLimit {
		limit = maxProfileFeedLimit
	}
	return s.repository.GetProfileFeed(ctx, userID, limit)
}
