package account

import (
	"backend/models"
	"context"
	"fmt"
	"strings"
	"time"
)

type UserUpdateInput struct {
	UserName  *string
	FirstName *string
	LastName  *string
	Email     *string
}

type UpdateUserResult struct {
	EmailChanged bool
	PendingEmail string
}

func (input UserUpdateInput) HasNoFields() bool {
	return input.UserName == nil && input.FirstName == nil && input.LastName == nil && input.Email == nil
}

func (input *UserUpdateInput) Normalize() {
	if input.UserName != nil {
		*input.UserName = strings.TrimSpace(*input.UserName)
	}
	if input.FirstName != nil {
		*input.FirstName = strings.TrimSpace(*input.FirstName)
	}
	if input.LastName != nil {
		*input.LastName = strings.TrimSpace(*input.LastName)
	}
	if input.Email != nil {
		*input.Email = strings.ToLower(strings.TrimSpace(*input.Email))
	}

}

func (s *Service) UpdateUser(ctx context.Context, userID uint, input UserUpdateInput) (UpdateUserResult, error) {
	if input.HasNoFields() {
		return UpdateUserResult{}, AccountErrors.NoUserFields
	}
	input.Normalize()
	if input.UserName != nil {
		if *input.UserName == "" {
			return UpdateUserResult{}, AccountErrors.UserNameBlank
		}
		if !isValidUserName(*input.UserName) {
			return UpdateUserResult{}, AccountErrors.InvalidUserName
		}
	}
	if input.FirstName != nil && *input.FirstName == "" {
		return UpdateUserResult{}, AccountErrors.FirstNameBlank
	}
	if input.LastName != nil && *input.LastName == "" {
		return UpdateUserResult{}, AccountErrors.LastNameBlank
	}
	if input.Email != nil {
		if !isValidEmail(*input.Email) {
			return UpdateUserResult{}, AccountErrors.InvalidEmail
		}
		if s.emailSender == nil {
			return UpdateUserResult{}, AccountErrors.EmailDeliveryFailed
		}
	}
	var rawToken string
	var verificationToken *models.AccountToken
	if input.Email != nil {
		generatedToken, tokenHash, err := generateAccountToken()
		if err != nil {
			return UpdateUserResult{}, err
		}
		rawToken = generatedToken
		token := models.AccountToken{
			UserID:    userID,
			Hash:      tokenHash,
			Purpose:   models.AccountTokenPurposeEmailVerification,
			ExpiresAt: time.Now().UTC().Add(24 * time.Hour),
		}
		verificationToken = &token
	}
	result, err := s.repository.UpdateUser(ctx, userID, input, verificationToken)
	if err != nil {
		return UpdateUserResult{}, err
	}
	if !result.EmailChanged {
		return result, nil
	}
	if err := s.emailSender.SendVerificationEmail(ctx, result.PendingEmail, rawToken); err != nil {
		return result, fmt.Errorf("%w: %v", AccountErrors.EmailDeliveryFailed, err)
	}
	return result, nil
}
