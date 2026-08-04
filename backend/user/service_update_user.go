package user

import (
	"context"
	"strings"
)

type UserUpdateInput struct {
	UserName  *string
	FirstName *string
	LastName  *string
	Email     *string
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

func (s *Service) UpdateUser(ctx context.Context, userID uint, input UserUpdateInput) error {
	if input.HasNoFields() {
		return UserErrors.NoUserFields
	}
	input.Normalize()
	if input.UserName != nil && *input.UserName == "" {
		return UserErrors.UserNameBlank
	}
	if input.FirstName != nil && *input.FirstName == "" {
		return UserErrors.FirstNameBlank
	}
	if input.LastName != nil && *input.LastName == "" {
		return UserErrors.LastNameBlank
	}
	if input.Email != nil && *input.Email == "" {
		return UserErrors.EmailBlank
	}
	return s.repository.UpdateUser(ctx, userID, input.UserName, input.FirstName, input.LastName, input.Email)
}
