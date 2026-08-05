package user

import (
	"backend/models"
	"context"
)

func (s *Service) ListMatches(ctx context.Context, userID uint) ([]models.Match, error) {
	return s.repository.ListMatches(ctx, userID)
}
