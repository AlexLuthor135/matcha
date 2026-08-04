package user

import (
	"backend/models"
	"context"
	"errors"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

func (s *Service) Login(ctx context.Context, email string, password string) (models.User, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	if email == "" || strings.TrimSpace(password) == "" {
		return models.User{}, UserErrors.LoginFieldsMissing
	}
	user, err := s.repository.GetUserByEmail(ctx, email)
	if errors.Is(err, UserErrors.UserNotFound) {
		return models.User{}, UserErrors.InvalidCredentials
	}
	if err != nil {
		return models.User{}, err
	}
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return models.User{}, UserErrors.InvalidCredentials
	}
	return user, nil
}
