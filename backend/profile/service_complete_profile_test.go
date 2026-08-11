package profile

import (
	"backend/models"
	"context"
	"errors"
	"testing"
	"time"
)

func completeProfileTestInput() CompleteProfileInput {
	latitude := 52.52
	longitude := 13.405
	return CompleteProfileInput{
		Gender:      "Male",
		Preferences: "Female",
		Bio:         "Hello",
		Interests:   []string{"music"},
		BirthDate:   time.Date(2000, time.May, 10, 0, 0, 0, 0, time.UTC),
		Location: &LocationInput{
			Source:    models.LocationSourceGPS,
			Latitude:  &latitude,
			Longitude: &longitude,
			Consent:   true,
		},
	}
}

func TestCompleteProfileRejectsInvalidInput(t *testing.T) {
	tests := []struct {
		name    string
		change  func(*CompleteProfileInput)
		wantErr error
	}{
		{
			name: "missing birth date",
			change: func(input *CompleteProfileInput) {
				input.BirthDate = time.Time{}
			},
			wantErr: ProfileErrors.ProfileFieldsMissing,
		},
		{
			name: "missing location",
			change: func(input *CompleteProfileInput) {
				input.Location = nil
			},
			wantErr: ProfileErrors.ProfileFieldsMissing,
		},
		{
			name: "missing latitude",
			change: func(input *CompleteProfileInput) {
				input.Location.Latitude = nil
			},
			wantErr: ProfileErrors.InvalidLocation,
		},
		{
			name: "missing longitude",
			change: func(input *CompleteProfileInput) {
				input.Location.Longitude = nil
			},
			wantErr: ProfileErrors.InvalidLocation,
		},
		{
			name: "underage",
			change: func(input *CompleteProfileInput) {
				input.BirthDate = time.Now().UTC().AddDate(-17, 0, 0)
			},
			wantErr: ProfileErrors.UserUnderage,
		},
		{
			name: "latitude below range",
			change: func(input *CompleteProfileInput) {
				value := -90.01
				input.Location.Latitude = &value
			},
			wantErr: ProfileErrors.InvalidLocation,
		},
		{
			name: "latitude above range",
			change: func(input *CompleteProfileInput) {
				value := 90.01
				input.Location.Latitude = &value
			},
			wantErr: ProfileErrors.InvalidLocation,
		},
		{
			name: "longitude below range",
			change: func(input *CompleteProfileInput) {
				value := -180.01
				input.Location.Longitude = &value
			},
			wantErr: ProfileErrors.InvalidLocation,
		},
		{
			name: "longitude above range",
			change: func(input *CompleteProfileInput) {
				value := 180.01
				input.Location.Longitude = &value
			},
			wantErr: ProfileErrors.InvalidLocation,
		},
		{
			name: "invalid location source",
			change: func(input *CompleteProfileInput) {
				input.Location.Source = "browser"
			},
			wantErr: ProfileErrors.InvalidLocationSource,
		},
		{
			name: "GPS without consent",
			change: func(input *CompleteProfileInput) {
				input.Location.Consent = false
			},
			wantErr: ProfileErrors.LocationConsentRequired,
		},
		{
			name: "manual location without name",
			change: func(input *CompleteProfileInput) {
				input.Location.Source = models.LocationSourceManual
				input.Location.Name = "  "
				input.Location.Consent = false
			},
			wantErr: ProfileErrors.ManualLocationNameMissing,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			input := completeProfileTestInput()
			test.change(&input)
			repositoryCalled := false
			repository := &fakeUserRepository{
				completeProfileFn: func(context.Context, uint, CompleteProfileInput) (bool, error) {
					repositoryCalled = true
					return true, nil
				},
			}

			_, err := NewService(repository, &fakeImageStorage{}).CompleteProfile(
				context.Background(),
				10,
				input,
			)
			if !errors.Is(err, test.wantErr) {
				t.Fatalf("CompleteProfile() error = %v, want %v", err, test.wantErr)
			}
			if repositoryCalled {
				t.Fatal("CompleteProfile() called repository for invalid input")
			}
		})
	}
}

func TestCompleteProfileNormalizesAndDelegates(t *testing.T) {
	const userID uint = 10
	input := completeProfileTestInput()
	input.Gender = "  Male  "
	input.Preferences = "  Female  "
	input.Bio = "  Hello  "
	input.Interests = []string{" music ", "", "  travel  "}

	repository := &fakeUserRepository{
		getAvatarURLFn: func(_ context.Context, gotUserID uint) (string, error) {
			if gotUserID != userID {
				t.Fatalf("GetAvatarURL() userID = %d, want %d", gotUserID, userID)
			}
			return "/uploads/avatars/profile.png", nil
		},
		completeProfileFn: func(_ context.Context, gotUserID uint, gotInput CompleteProfileInput) (bool, error) {
			if gotUserID != userID {
				t.Fatalf("CompleteProfile() userID = %d, want %d", gotUserID, userID)
			}
			if gotInput.Gender != "Male" || gotInput.Preferences != "Female" || gotInput.Bio != "Hello" {
				t.Fatalf("CompleteProfile() normalized strings = %+v", gotInput)
			}
			if len(gotInput.Interests) != 2 || gotInput.Interests[0] != "music" || gotInput.Interests[1] != "travel" {
				t.Fatalf("CompleteProfile() interests = %#v, want [music travel]", gotInput.Interests)
			}
			if !gotInput.BirthDate.Equal(input.BirthDate) {
				t.Fatalf("CompleteProfile() birth date = %v, want %v", gotInput.BirthDate, input.BirthDate)
			}
			if gotInput.Location == nil ||
				gotInput.Location.Latitude == nil ||
				*gotInput.Location.Latitude != *input.Location.Latitude ||
				gotInput.Location.Longitude == nil ||
				*gotInput.Location.Longitude != *input.Location.Longitude {
				t.Fatalf("CompleteProfile() location = %+v", gotInput.Location)
			}
			if gotInput.Location.Source != models.LocationSourceGPS ||
				!gotInput.Location.Consent ||
				gotInput.Location.ConsentAt == nil ||
				gotInput.Location.Name != "" {
				t.Fatalf("CompleteProfile() prepared GPS location = %+v", gotInput.Location)
			}
			return true, nil
		},
	}

	completed, err := NewService(repository, &fakeImageStorage{}).CompleteProfile(
		context.Background(),
		userID,
		input,
	)
	if err != nil {
		t.Fatalf("CompleteProfile() unexpected error: %v", err)
	}
	if !completed {
		t.Fatal("CompleteProfile() completed = false, want true")
	}
}

func TestCompleteProfileRequiresProfilePicture(t *testing.T) {
	repositoryCalled := false
	repository := &fakeUserRepository{
		getAvatarURLFn: func(context.Context, uint) (string, error) {
			return "  ", nil
		},
		completeProfileFn: func(context.Context, uint, CompleteProfileInput) (bool, error) {
			repositoryCalled = true
			return true, nil
		},
	}

	_, err := NewService(repository, &fakeImageStorage{}).CompleteProfile(
		context.Background(),
		10,
		completeProfileTestInput(),
	)
	if !errors.Is(err, ProfileErrors.ProfilePictureRequired) {
		t.Fatalf("CompleteProfile() error = %v, want %v", err, ProfileErrors.ProfilePictureRequired)
	}
	if repositoryCalled {
		t.Fatal("CompleteProfile() saved a profile without a profile picture")
	}
}
