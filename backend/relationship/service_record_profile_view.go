package relationship

import (
	"backend/models"
	"context"
	"errors"
)

func (s *Service) RecordProfileView(ctx context.Context, viewerID uint, viewedUserID uint) (models.ProfileView, error) {
	if viewedUserID == 0 {
		return models.ProfileView{}, RelationshipErrors.InvalidTargetUserID
	}
	if viewerID == viewedUserID {
		return models.ProfileView{}, RelationshipErrors.CannotViewOwnProfile
	}
	isCompleted, err := s.repository.GetCompletionStatus(ctx, viewedUserID)
	if errors.Is(err, RelationshipErrors.UserNotFound) {
		return models.ProfileView{}, RelationshipErrors.TargetUserNotFound
	}
	if err != nil {
		return models.ProfileView{}, err
	}
	if !isCompleted {
		return models.ProfileView{}, RelationshipErrors.TargetUserNotFound
	}
	return s.repository.SaveProfileView(ctx, viewerID, viewedUserID)
}
