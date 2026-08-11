package user

import (
	"backend/models"
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
)

func (s *Service) ResendVerificationEmail(ctx context.Context, email string) error {
	email = strings.ToLower(strings.TrimSpace(email))
	if email == "" {
		return UserErrors.EmailBlank
	}
	if s.emailSender == nil {
		return UserErrors.EmailDeliveryFailed
	}
	user, err := s.repository.GetUserByEmail(ctx, email)
	if errors.Is(err, UserErrors.UserNotFound) {
		return nil
	}
	if err != nil {
		return err
	}
	if user.IsVerified {
		return nil
	}
	rawToken, tokenHash, err := generateAccountToken()
	if err != nil {
		return err
	}
	token := models.AccountToken{
		UserID:    user.ID,
		Hash:      tokenHash,
		Purpose:   models.AccountTokenPurposeEmailVerification,
		ExpiresAt: time.Now().UTC().Add(24 * time.Hour),
	}
	if err := s.repository.ReplaceAccountToken(ctx, token); err != nil {
		return err
	}
	if err := s.emailSender.SendVerificationEmail(ctx, email, rawToken); err != nil {
		return fmt.Errorf("%w: %v", UserErrors.EmailDeliveryFailed, err)
	}
	return nil
}
