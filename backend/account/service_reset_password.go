package account

import (
	"context"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

type ResetPasswordInput struct {
	Token       string
	NewPassword string
}

func (s *Service) ResetPassword(ctx context.Context, input ResetPasswordInput) error {
	input.Token = strings.TrimSpace(input.Token)
	if input.Token == "" || strings.TrimSpace(input.NewPassword) == "" {
		return AccountErrors.PasswordResetFieldsMissing
	}
	if !isValidPassword(input.NewPassword) {
		return AccountErrors.InvalidPassword
	}
	newPasswordHash, err := bcrypt.GenerateFromPassword([]byte(input.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	tokenHash := hashAccountToken(input.Token)
	return s.repository.ResetPasswordWithToken(ctx, tokenHash, string(newPasswordHash))
}
