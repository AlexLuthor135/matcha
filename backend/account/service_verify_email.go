package account

import (
	"context"
	"strings"
)

func (s *Service) VerifyEmail(ctx context.Context, rawToken string) error {
	rawToken = strings.TrimSpace(rawToken)
	if rawToken == "" {
		return AccountErrors.InvalidVerificationToken
	}
	tokenHash := hashAccountToken(rawToken)
	return s.repository.VerifyEmail(ctx, tokenHash)
}
