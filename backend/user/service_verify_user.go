package user

import "context"

func (s *Service) VerifyUser(ctx context.Context, userID uint) (bool, error) {
	return s.repository.GetCompletionStatus(ctx, userID)
}
