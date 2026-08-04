package user

import (
	"context"
	"strings"
)

type UpdateProfileInput struct {
	Gender      *string
	Preferences *string
	Bio         *string
	Interests   *[]string
}

func (input UpdateProfileInput) Normalize() {
	if input.Gender != nil {
		*input.Gender = strings.TrimSpace(*input.Gender)
	}
	if input.Preferences != nil {
		*input.Preferences = strings.TrimSpace(*input.Preferences)
	}
	if input.Bio != nil {
		*input.Bio = strings.TrimSpace(*input.Bio)
	}
	if input.Interests != nil {
		normalizedInterests := make([]string, 0, len(*input.Interests))
		for _, interest := range *input.Interests {
			interest = strings.TrimSpace(interest)
			if interest != "" {
				normalizedInterests = append(normalizedInterests, interest)
			}
		}
		*input.Interests = normalizedInterests
	}
}

func (input UpdateProfileInput) HasNoFields() bool {
	return input.Gender == nil && input.Preferences == nil && input.Bio == nil && input.Interests == nil
}

func (s *Service) UpdateProfile(ctx context.Context, userID uint, input UpdateProfileInput) error {
	if input.HasNoFields() {
		return UserErrors.NoProfileFields
	}
	input.Normalize()
	if input.Gender != nil && !isValidGenderPreferences(*input.Gender) {
		return UserErrors.InvalidGenderPreference
	}
	if input.Preferences != nil && !isValidGenderPreferences(*input.Preferences) {
		return UserErrors.InvalidGenderPreference
	}
	if input.Bio != nil && *input.Bio == "" {
		return UserErrors.ProfileBioBlank
	}
	if input.Interests != nil && len(*input.Interests) == 0 {
		return UserErrors.ProfileInterestsMissing
	}
	return s.repository.UpdateProfile(ctx, userID, input.Gender, input.Preferences, input.Bio, input.Interests)
}
