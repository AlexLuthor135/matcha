package user

import (
	"backend/models"
	"context"
)

func isValidProfileDecision(decision models.ProfileDecisionValue) bool {
	switch decision {
	case models.ProfileDecisionLike, models.ProfileDecisionDislike:
		return true
	default:
		return false
	}
}

func (s *Service) SaveProfileDecision(ctx context.Context, userID uint, targetUserID uint, decision models.ProfileDecisionValue) (models.ProfileDecision, bool, error) {
	if targetUserID == 0 {
		return models.ProfileDecision{}, false, UserErrors.InvalidTargetUserID
	}
	if userID == targetUserID {
		return models.ProfileDecision{}, false, UserErrors.CannotDecideOwnProfile
	}
	if !isValidProfileDecision(decision) {
		return models.ProfileDecision{}, false, UserErrors.InvalidProfileDecision
	}
	return s.repository.SaveProfileDecision(ctx, userID, targetUserID, decision)
}
