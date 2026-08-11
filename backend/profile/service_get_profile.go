package profile

import (
	"backend/models"
	"context"
)

func (s *Service) GetProfile(ctx context.Context, userID uint) (models.User, error) {
	return s.repository.GetProfile(ctx, userID)
}
