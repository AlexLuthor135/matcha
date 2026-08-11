package user

import (
	"backend/models"
	"context"
)

func (s *Service) ListProfileLikers(ctx context.Context, userID uint) ([]models.ProfileLiker, error) {
	return s.repository.ListProfileLikers(ctx, userID)
}
