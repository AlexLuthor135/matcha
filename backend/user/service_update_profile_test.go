package user

import (
	"backend/models"
	"context"
	"errors"
	"testing"
	"time"
)

func TestUpdateProfileRejectsInvalidNewFields(t *testing.T) {
	latitude := 52.52
	longitude := 13.405
	tests := []struct {
		name    string
		input   UpdateProfileInput
		wantErr error
	}{
		{
			name:    "no fields",
			input:   UpdateProfileInput{},
			wantErr: UserErrors.NoProfileFields,
		},
		{
			name: "underage birth date",
			input: UpdateProfileInput{
				BirthDate: pointerToUpdateProfileTime(time.Now().UTC().AddDate(-17, 0, 0)),
			},
			wantErr: UserErrors.UserUnderage,
		},
		{
			name: "missing latitude",
			input: UpdateProfileInput{Location: &LocationInput{
				Source:    models.LocationSourceGPS,
				Longitude: &longitude,
				Consent:   true,
			}},
			wantErr: UserErrors.InvalidLocation,
		},
		{
			name: "missing longitude",
			input: UpdateProfileInput{Location: &LocationInput{
				Source:   models.LocationSourceGPS,
				Latitude: &latitude,
				Consent:  true,
			}},
			wantErr: UserErrors.InvalidLocation,
		},
		{
			name: "latitude outside range",
			input: UpdateProfileInput{
				Location: &LocationInput{
					Source:    models.LocationSourceGPS,
					Latitude:  pointerToUpdateProfileFloat(91),
					Longitude: &longitude,
					Consent:   true,
				},
			},
			wantErr: UserErrors.InvalidLocation,
		},
		{
			name: "longitude outside range",
			input: UpdateProfileInput{
				Location: &LocationInput{
					Source:    models.LocationSourceGPS,
					Latitude:  &latitude,
					Longitude: pointerToUpdateProfileFloat(-181),
					Consent:   true,
				},
			},
			wantErr: UserErrors.InvalidLocation,
		},
		{
			name: "invalid source",
			input: UpdateProfileInput{Location: &LocationInput{
				Source:    "browser",
				Latitude:  &latitude,
				Longitude: &longitude,
			}},
			wantErr: UserErrors.InvalidLocationSource,
		},
		{
			name: "GPS without consent",
			input: UpdateProfileInput{Location: &LocationInput{
				Source:    models.LocationSourceGPS,
				Latitude:  &latitude,
				Longitude: &longitude,
			}},
			wantErr: UserErrors.LocationConsentRequired,
		},
		{
			name: "manual location without name",
			input: UpdateProfileInput{Location: &LocationInput{
				Source:    models.LocationSourceManual,
				Name:      "  ",
				Latitude:  &latitude,
				Longitude: &longitude,
			}},
			wantErr: UserErrors.ManualLocationNameMissing,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repositoryCalled := false
			repository := &fakeUserRepository{
				updateProfileFn: func(context.Context, uint, UpdateProfileInput) error {
					repositoryCalled = true
					return nil
				},
			}

			err := NewService(repository, &fakeImageStorage{}).UpdateProfile(
				context.Background(),
				10,
				test.input,
			)
			if !errors.Is(err, test.wantErr) {
				t.Fatalf("UpdateProfile() error = %v, want %v", err, test.wantErr)
			}
			if repositoryCalled {
				t.Fatal("UpdateProfile() called repository for invalid input")
			}
		})
	}
}

func TestUpdateProfileNormalizesAndDelegatesNewFields(t *testing.T) {
	const userID uint = 10
	bio := "  Updated bio  "
	interests := []string{" music ", "", " travel "}
	birthDate := time.Date(1999, time.June, 15, 0, 0, 0, 0, time.UTC)
	latitude := 48.8566
	longitude := 2.3522
	input := UpdateProfileInput{
		Bio:       &bio,
		Interests: &interests,
		BirthDate: &birthDate,
		Location: &LocationInput{
			Source:    models.LocationSourceManual,
			Name:      "  Paris  ",
			Latitude:  &latitude,
			Longitude: &longitude,
		},
	}

	repository := &fakeUserRepository{
		updateProfileFn: func(_ context.Context, gotUserID uint, gotInput UpdateProfileInput) error {
			if gotUserID != userID {
				t.Fatalf("UpdateProfile() userID = %d, want %d", gotUserID, userID)
			}
			if gotInput.Bio == nil || *gotInput.Bio != "Updated bio" {
				t.Fatalf("UpdateProfile() bio = %v, want Updated bio", gotInput.Bio)
			}
			if gotInput.Interests == nil || len(*gotInput.Interests) != 2 ||
				(*gotInput.Interests)[0] != "music" || (*gotInput.Interests)[1] != "travel" {
				t.Fatalf("UpdateProfile() interests = %v, want [music travel]", gotInput.Interests)
			}
			if gotInput.BirthDate == nil || !gotInput.BirthDate.Equal(birthDate) {
				t.Fatalf("UpdateProfile() birth date = %v, want %v", gotInput.BirthDate, birthDate)
			}
			if gotInput.Location == nil ||
				gotInput.Location.Latitude == nil || *gotInput.Location.Latitude != latitude ||
				gotInput.Location.Longitude == nil || *gotInput.Location.Longitude != longitude {
				t.Fatalf("UpdateProfile() location = %+v", gotInput.Location)
			}
			if gotInput.Location.Source != models.LocationSourceManual ||
				gotInput.Location.Name != "Paris" ||
				gotInput.Location.Consent ||
				gotInput.Location.ConsentAt != nil {
				t.Fatalf("UpdateProfile() prepared manual location = %+v", gotInput.Location)
			}
			return nil
		},
	}

	if err := NewService(repository, &fakeImageStorage{}).UpdateProfile(
		context.Background(),
		userID,
		input,
	); err != nil {
		t.Fatalf("UpdateProfile() unexpected error: %v", err)
	}
}

func TestUpdateProfileAllowsExistingFieldsWithoutLocation(t *testing.T) {
	bio := "Updated bio"
	repositoryCalls := 0
	repository := &fakeUserRepository{
		updateProfileFn: func(context.Context, uint, UpdateProfileInput) error {
			repositoryCalls++
			return nil
		},
	}

	if err := NewService(repository, &fakeImageStorage{}).UpdateProfile(
		context.Background(),
		10,
		UpdateProfileInput{Bio: &bio},
	); err != nil {
		t.Fatalf("UpdateProfile() bio-only error: %v", err)
	}
	if repositoryCalls != 1 {
		t.Fatalf("UpdateProfile() repository calls = %d, want 1", repositoryCalls)
	}
}

func pointerToUpdateProfileFloat(value float64) *float64 {
	return &value
}

func pointerToUpdateProfileTime(value time.Time) *time.Time {
	return &value
}
