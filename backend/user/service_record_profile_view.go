package user

import (
	"backend/models"
	"context"
	"errors"
)

func (s *Service) RecordProfileView(ctx context.Context, viewerID uint, viewedUserID uint) (models.ProfileView, error) {
	if viewedUserID == 0 {
		return models.ProfileView{}, UserErrors.InvalidTargetUserID
	}
	if viewerID == viewedUserID {
		return models.ProfileView{}, UserErrors.CannotViewOwnProfile
	}
	isCompleted, err := s.repository.GetCompletionStatus(ctx, viewedUserID)
	if errors.Is(err, UserErrors.UserNotFound) {
		return models.ProfileView{}, UserErrors.TargetUserNotFound
	}
	if err != nil {
		return models.ProfileView{}, err
	}
	if !isCompleted {
		return models.ProfileView{}, UserErrors.TargetUserNotFound
	}
	return s.repository.SaveProfileView(ctx, viewerID, viewedUserID)
}
