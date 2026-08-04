package user

import (
	"context"
	"strings"
)

type CompleteProfileInput struct {
	Gender      string
	Preferences string
	Bio         string
	Interests   []string
}

func (input *CompleteProfileInput) Normalize() {
	input.Gender = strings.TrimSpace(input.Gender)
	input.Preferences = strings.TrimSpace(input.Preferences)
	input.Bio = strings.TrimSpace(input.Bio)
	normalizedInterests := make([]string, 0, len(input.Interests))
	for _, interest := range input.Interests {
		interest = strings.TrimSpace(interest)
		if interest != "" {
			normalizedInterests = append(normalizedInterests, interest)
		}
	}
	input.Interests = normalizedInterests
}

func (input CompleteProfileInput) HasMissingFields() bool {
	return input.Gender == "" || input.Preferences == "" || input.Bio == "" || len(input.Interests) == 0
}

func (s *Service) CompleteProfile(ctx context.Context, userID uint, input CompleteProfileInput) (bool, error) {
	input.Normalize()
	if input.HasMissingFields() {
		return false, UserErrors.ProfileFieldsMissing
	}
	if !isValidGenderPreferences(input.Gender) || !isValidGenderPreferences(input.Preferences) {
		return false, UserErrors.InvalidGenderPreference
	}
	return s.repository.CompleteProfile(ctx, userID, input.Gender, input.Preferences, input.Bio, input.Interests)
}
