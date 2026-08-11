package user

import (
	"context"
	"strings"
	"time"
)

type CompleteProfileInput struct {
	Gender      string
	Preferences string
	Bio         string
	Interests   []string
	BirthDate   time.Time
	Location    *LocationInput
}

func (input *CompleteProfileInput) Normalize() error {
	input.Gender = strings.TrimSpace(input.Gender)
	input.Preferences = normalizeSexualPreference(input.Preferences)
	input.Bio = strings.TrimSpace(input.Bio)
	normalizedInterests, err := normalizeInterests(input.Interests)
	if err != nil {
		return err
	}
	input.Interests = normalizedInterests
	return nil
}

func (input CompleteProfileInput) HasMissingFields() bool {
	return input.Gender == "" ||
		input.Bio == "" ||
		len(input.Interests) == 0 ||
		input.BirthDate.IsZero() ||
		input.Location == nil
}

func (s *Service) CompleteProfile(ctx context.Context, userID uint, input CompleteProfileInput) (bool, error) {
	if err := input.Normalize(); err != nil {
		return false, err
	}
	if input.HasMissingFields() {
		return false, UserErrors.ProfileFieldsMissing
	}
	if !isValidGender(input.Gender) {
		return false, UserErrors.InvalidGenderPreference
	}
	if !isValidSexualPreference(input.Preferences) {
		return false, UserErrors.InvalidGenderPreference
	}
	if !isValidAge(input.BirthDate) {
		return false, UserErrors.UserUnderage
	}
	if err := input.Location.Prepare(); err != nil {
		return false, err
	}
	avatarURL, err := s.repository.GetAvatarURL(ctx, userID)
	if err != nil {
		return false, err
	}
	if strings.TrimSpace(avatarURL) == "" {
		return false, UserErrors.ProfilePictureRequired
	}
	return s.repository.CompleteProfile(ctx, userID, input)
}
