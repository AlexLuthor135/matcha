package user

import (
	"backend/models"
	"context"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

type RegisterInput struct {
	UserName  string
	FirstName string
	LastName  string
	Email     string
	Password  string
}

func (input *RegisterInput) Normalize() {
	input.UserName = strings.TrimSpace(input.UserName)
	input.FirstName = strings.TrimSpace(input.FirstName)
	input.LastName = strings.TrimSpace(input.LastName)
	input.Email = strings.ToLower(strings.TrimSpace(input.Email))
}

func (input RegisterInput) HasMissingFields() bool {
	return input.UserName == "" ||
		input.FirstName == "" ||
		input.LastName == "" ||
		input.Email == "" ||
		strings.TrimSpace(input.Password) == ""
}

func (s *Service) Register(ctx context.Context, input RegisterInput) (models.User, error) {
	input.Normalize()
	if input.HasMissingFields() {
		return models.User{}, UserErrors.RegistrationFieldMissing
	}
	if !isValidPassword(input.Password) {
		return models.User{}, UserErrors.InvalidPassword
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return models.User{}, err
	}
	newUser := models.User{
		UserName:  input.UserName,
		FirstName: input.FirstName,
		LastName:  input.LastName,
		Email:     input.Email,
		Password:  string(hashedPassword),
	}
	return s.repository.CreateUser(ctx, newUser)
}
