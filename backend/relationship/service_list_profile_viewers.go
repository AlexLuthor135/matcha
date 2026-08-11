package relationship

import (
	"backend/models"
	"context"
)

func (s *Service) ListProfileViewers(ctx context.Context, userID uint) ([]models.ProfileViewer, error) {
	return s.repository.ListProfileViewers(ctx, userID)
}
