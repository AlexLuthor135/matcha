package profile

import (
	"context"
	"strings"
	"time"
)

type UpdateProfileInput struct {
	Gender      *string
	Preferences *string
	Bio         *string
	Interests   *[]string
	BirthDate   *time.Time
	Location    *LocationInput
}

func (input *UpdateProfileInput) Normalize() error {
	if input.Gender != nil {
		*input.Gender = strings.TrimSpace(*input.Gender)
	}
	if input.Preferences != nil {
		*input.Preferences = normalizeSexualPreference(*input.Preferences)
	}
	if input.Bio != nil {
		*input.Bio = strings.TrimSpace(*input.Bio)
	}
	if input.Interests != nil {
		normalizedInterests, err := normalizeInterests(*input.Interests)
		if err != nil {
			return err
		}
		*input.Interests = normalizedInterests
	}
	return nil
}

func (input UpdateProfileInput) HasNoFields() bool {
	return input.Gender == nil &&
		input.Preferences == nil &&
		input.Bio == nil &&
		input.Interests == nil &&
		input.BirthDate == nil &&
		input.Location == nil
}

func (s *Service) UpdateProfile(ctx context.Context, userID uint, input UpdateProfileInput) error {
	if input.HasNoFields() {
		return ProfileErrors.NoProfileFields
	}
	if err := input.Normalize(); err != nil {
		return err
	}
	if input.Gender != nil && !isValidGender(*input.Gender) {
		return ProfileErrors.InvalidGenderPreference
	}
	if input.Preferences != nil && !isValidSexualPreference(*input.Preferences) {
		return ProfileErrors.InvalidGenderPreference
	}
	if input.Bio != nil && *input.Bio == "" {
		return ProfileErrors.ProfileBioBlank
	}
	if input.Interests != nil && len(*input.Interests) == 0 {
		return ProfileErrors.ProfileInterestsMissing
	}
	if input.BirthDate != nil {
		if !isValidAge(*input.BirthDate) {
			return ProfileErrors.UserUnderage
		}
	}
	if input.Location != nil {
		if err := input.Location.Prepare(); err != nil {
			return err
		}
	}
	return s.repository.UpdateProfile(ctx, userID, input)
}
