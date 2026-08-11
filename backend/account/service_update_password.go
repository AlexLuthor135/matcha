package account

import (
	"context"
	"errors"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

func (s *Service) UpdatePassword(ctx context.Context, userID uint, currentPassword string, newPassword string) error {
	if strings.TrimSpace(currentPassword) == "" || strings.TrimSpace(newPassword) == "" {
		return AccountErrors.PasswordFieldsMissing
	}
	if !isValidPassword(newPassword) {
		return AccountErrors.InvalidPassword
	}
	if currentPassword == newPassword {
		return AccountErrors.NewPasswordUnchanged
	}
	currentPasswordHash, err := s.repository.GetPasswordHash(ctx, userID)
	if err != nil {
		return err
	}
	err = bcrypt.CompareHashAndPassword([]byte(currentPasswordHash), []byte(currentPassword))
	if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
		return AccountErrors.CurrentPasswordIncorrect
	}
	if err != nil {
		return err
	}
	newPasswordHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	return s.repository.UpdatePasswordHashAndRevokeSessions(ctx, userID, string(newPasswordHash))
}
