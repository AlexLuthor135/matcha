package user

import (
	"backend/models"
	"context"
	"errors"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

func (s *Service) Login(ctx context.Context, login string, password string) (models.User, error) {
	login = strings.TrimSpace(login)
	if login == "" || strings.TrimSpace(password) == "" {
		return models.User{}, UserErrors.LoginFieldsMissing
	}
	var (
		user models.User
		err  error
	)
	if strings.Contains(login, "@") {
		login = strings.ToLower(login)
		user, err = s.repository.GetUserByEmail(ctx, login)
	} else {
		user, err = s.repository.GetUserByUserName(ctx, login)
	}
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
	if !user.IsVerified {
		return models.User{}, UserErrors.EmailNotVerified
	}
	return user, nil
}
