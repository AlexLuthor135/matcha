package user

import "context"

func validateUserBlock(blockerID uint, blockedUserID uint) error {
	if blockedUserID == 0 {
		return UserErrors.InvalidTargetUserID
	}
	if blockerID == blockedUserID {
		return UserErrors.CannotBlockSelf
	}
	return nil
}

func (s *Service) BlockUser(ctx context.Context, blockerID uint, blockedUserID uint) error {
	if err := validateUserBlock(blockerID, blockedUserID); err != nil {
		return err
	}
	return s.repository.BlockUser(ctx, blockerID, blockedUserID)
}

func (s *Service) UnblockUser(ctx context.Context, blockerID uint, blockedUserID uint) error {
	if err := validateUserBlock(blockerID, blockedUserID); err != nil {
		return err
	}
	return s.repository.UnblockUser(ctx, blockerID, blockedUserID)
}
