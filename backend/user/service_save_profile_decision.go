package user

import (
	"backend/models"
	"context"
	"strings"
)

func isValidProfileDecision(decision models.ProfileDecisionValue) bool {
	switch decision {
	case models.ProfileDecisionLike, models.ProfileDecisionDislike:
		return true
	default:
		return false
	}
}

func (s *Service) SaveProfileDecision(ctx context.Context, userID uint, targetUserID uint, decision models.ProfileDecisionValue) (SaveProfileDecisionResult, error) {
	if targetUserID == 0 {
		return SaveProfileDecisionResult{}, UserErrors.InvalidTargetUserID
	}
	if userID == targetUserID {
		return SaveProfileDecisionResult{}, UserErrors.CannotDecideOwnProfile
	}
	if !isValidProfileDecision(decision) {
		return SaveProfileDecisionResult{}, UserErrors.InvalidProfileDecision
	}
	avatarURL, err := s.repository.GetAvatarURL(ctx, userID)
	if err != nil {
		return SaveProfileDecisionResult{}, err
	}
	if strings.TrimSpace(avatarURL) == "" {
		return SaveProfileDecisionResult{}, UserErrors.ProfilePictureRequired
	}
	return s.repository.SaveProfileDecision(ctx, userID, targetUserID, decision)
}
