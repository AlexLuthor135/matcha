package user

import (
	"backend/models"
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
)

func (s *Service) RequestPasswordReset(ctx context.Context, email string) error {
	email = strings.ToLower(strings.TrimSpace(email))
	if email == "" {
		return UserErrors.EmailBlank
	}
	if s.emailSender == nil {
		return UserErrors.PasswordResetEmailDeliveryFailed
	}
	user, err := s.repository.GetUserByEmail(ctx, email)
	if errors.Is(err, UserErrors.UserNotFound) {
		return nil
	}
	if err != nil {
		return err
	}
	if !user.IsVerified {
		return nil
	}
	rawToken, tokenHash, err := generateAccountToken()
	if err != nil {
		return err
	}
	token := models.AccountToken{
		UserID:    user.ID,
		Hash:      tokenHash,
		Purpose:   models.AccountTokenPurposePasswordReset,
		ExpiresAt: time.Now().UTC().Add(30 * time.Minute),
	}
	if err := s.repository.ReplaceAccountToken(ctx, token); err != nil {
		return err
	}
	if err := s.emailSender.SendPasswordResetEmail(ctx, email, rawToken); err != nil {
		return fmt.Errorf("%w: %v", UserErrors.PasswordResetEmailDeliveryFailed, err)
	}
	return nil
}
