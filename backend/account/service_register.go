package account

import (
	"backend/models"
	"context"
	"fmt"
	"strings"
	"time"

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
		return models.User{}, AccountErrors.RegistrationFieldMissing
	}
	if !isValidEmail(input.Email) {
		return models.User{}, AccountErrors.InvalidEmail
	}
	if !isValidUserName(input.UserName) {
		return models.User{}, AccountErrors.InvalidUserName
	}
	if !isValidPassword(input.Password) {
		return models.User{}, AccountErrors.InvalidPassword
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return models.User{}, err
	}
	if s.emailSender == nil {
		return models.User{}, AccountErrors.EmailDeliveryFailed
	}
	rawToken, tokenHash, err := generateAccountToken()
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
	token := models.AccountToken{
		Hash:      tokenHash,
		Purpose:   models.AccountTokenPurposeEmailVerification,
		ExpiresAt: time.Now().UTC().Add(24 * time.Hour),
	}
	createdUser, err := s.repository.CreateUser(ctx, newUser, token)
	if err != nil {
		return models.User{}, err
	}
	if err := s.emailSender.SendVerificationEmail(ctx, createdUser.Email, rawToken); err != nil {
		return models.User{}, fmt.Errorf("%w: %v", AccountErrors.EmailDeliveryFailed, err)
	}
	return createdUser, nil
}
