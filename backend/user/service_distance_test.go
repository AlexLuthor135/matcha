package user

import (
	"backend/models"
	"context"
	"testing"
	"time"
)

func TestDistanceInKM(t *testing.T) {
	tests := []struct {
		name                            string
		firstLatitude, firstLongitude   float64
		secondLatitude, secondLongitude float64
		want                            float64
	}{
		{
			name: "same point",
			want: 0,
		},
		{
			name:           "one degree of latitude",
			secondLatitude: 1,
			want:           111.2,
		},
		{
			name:            "one degree of longitude at equator",
			secondLongitude: 1,
			want:            111.2,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := distanceInKM(
				test.firstLatitude,
				test.firstLongitude,
				test.secondLatitude,
				test.secondLongitude,
			)
			if got != test.want {
				t.Fatalf("distanceInKM() = %v, want %v", got, test.want)
			}
		})
	}
}

func TestGetProfileFeedAddsDistanceWithoutExposingCoordinates(t *testing.T) {
	const userID uint = 10
	viewerLatitude := 0.0
	viewerLongitude := 0.0
	targetLatitude := 1.0
	targetLongitude := 0.0
	targetBirthDate := time.Now().UTC().AddDate(-25, 0, 0)
	repository := &fakeUserRepository{
		getProfileFn: func(_ context.Context, gotUserID uint) (models.User, error) {
			if gotUserID != userID {
				t.Fatalf("GetProfile() userID = %d, want %d", gotUserID, userID)
			}
			return models.User{
				ID:          userID,
				Gender:      "Male",
				Preferences: "Female",
				Latitude:    &viewerLatitude,
				Longitude:   &viewerLongitude,
			}, nil
		},
		listProfileCandidatesFn: func(
			_ context.Context,
			gotUserID uint,
			preferredGender string,
			ownGender string,
			excludeDecided bool,
		) ([]models.User, error) {
			if gotUserID != userID || preferredGender != "Female" || ownGender != "Male" || !excludeDecided {
				t.Fatalf(
					"ListProfileCandidates() arguments = (%d, %q, %q, %t), want (%d, Female, Male, true)",
					gotUserID,
					preferredGender,
					ownGender,
					excludeDecided,
					userID,
				)
			}
			return []models.User{
				{ID: 20, BirthDate: &targetBirthDate, Latitude: &targetLatitude, Longitude: &targetLongitude},
				{ID: 30},
			}, nil
		},
	}

	profiles, err := NewService(repository, &fakeImageStorage{}).GetProfileFeed(
		context.Background(),
		userID,
		ProfileFeedOptions{Limit: 20},
	)
	if err != nil {
		t.Fatalf("GetProfileFeed() unexpected error: %v", err)
	}
	if len(profiles) != 1 {
		t.Fatalf("profile count = %d, want 1 valid-location candidate", len(profiles))
	}
	if profiles[0].Distance == nil || *profiles[0].Distance != 111.2 {
		t.Fatalf("first profile distance = %v, want 111.2", profiles[0].Distance)
	}

	response := publicProfileResponses(profiles)
	if response[0].Distance == nil || *response[0].Distance != 111.2 {
		t.Fatalf("public response distance = %v, want 111.2", response[0].Distance)
	}
}

func TestGetProfileFeedRanksByDistanceThenInterestsThenFame(t *testing.T) {
	const userID uint = 10
	viewerLatitude := 0.0
	viewerLongitude := 0.0
	oneDegreeLatitude := 1.0
	halfDegreeLatitude := 0.5
	longitude := 0.0
	targetBirthDate := time.Now().UTC().AddDate(-25, 0, 0)
	repository := &fakeUserRepository{
		getProfileFn: func(context.Context, uint) (models.User, error) {
			return models.User{
				ID:          userID,
				Gender:      "Male",
				Preferences: "Female",
				Interests:   []string{"Go", "Music"},
				Latitude:    &viewerLatitude,
				Longitude:   &viewerLongitude,
			}, nil
		},
		listProfileCandidatesFn: func(context.Context, uint, string, string, bool) ([]models.User, error) {
			return []models.User{
				{
					ID:         20,
					BirthDate:  &targetBirthDate,
					Interests:  []string{"Go"},
					FameRating: 100,
					Latitude:   &oneDegreeLatitude,
					Longitude:  &longitude,
				},
				{
					ID:         30,
					BirthDate:  &targetBirthDate,
					Interests:  []string{"go", "music"},
					FameRating: 1,
					Latitude:   &oneDegreeLatitude,
					Longitude:  &longitude,
				},
				{
					ID:         40,
					BirthDate:  &targetBirthDate,
					Interests:  nil,
					FameRating: 0,
					Latitude:   &halfDegreeLatitude,
					Longitude:  &longitude,
				},
				{
					ID:         50,
					BirthDate:  &targetBirthDate,
					Interests:  []string{"Go", "Music"},
					FameRating: 5,
					Latitude:   &oneDegreeLatitude,
					Longitude:  &longitude,
				},
			}, nil
		},
	}

	profiles, err := NewService(repository, &fakeImageStorage{}).GetProfileFeed(
		context.Background(),
		userID,
		ProfileFeedOptions{Limit: 4},
	)
	if err != nil {
		t.Fatalf("GetProfileFeed() unexpected error: %v", err)
	}
	wantOrder := []uint{40, 50, 30, 20}
	if len(profiles) != len(wantOrder) {
		t.Fatalf("profile count = %d, want %d; profiles = %+v", len(profiles), len(wantOrder), profiles)
	}
	for index, wantID := range wantOrder {
		if profiles[index].ID != wantID {
			t.Fatalf("profile order at %d = %d, want %d; profiles = %+v", index, profiles[index].ID, wantID, profiles)
		}
	}
}

func TestSearchProfilesIncludesDecidedCandidates(t *testing.T) {
	const userID uint = 10
	viewerLatitude := 0.0
	viewerLongitude := 0.0
	targetLatitude := 0.25
	targetLongitude := 0.0
	targetBirthDate := time.Now().UTC().AddDate(-25, 0, 0)

	repository := &fakeUserRepository{
		getProfileFn: func(context.Context, uint) (models.User, error) {
			return models.User{
				ID:          userID,
				Gender:      "Male",
				Preferences: "Female",
				Latitude:    &viewerLatitude,
				Longitude:   &viewerLongitude,
			}, nil
		},
		listProfileCandidatesFn: func(
			_ context.Context,
			gotUserID uint,
			preferredGender string,
			ownGender string,
			excludeDecided bool,
		) ([]models.User, error) {
			if gotUserID != userID || preferredGender != "Female" || ownGender != "Male" {
				t.Fatalf("unexpected candidate arguments: (%d, %q, %q)", gotUserID, preferredGender, ownGender)
			}
			if excludeDecided {
				t.Fatal("SearchProfiles() excludeDecided = true, want false")
			}
			return []models.User{{
				ID:         20,
				BirthDate:  &targetBirthDate,
				Latitude:   &targetLatitude,
				Longitude:  &targetLongitude,
				FameRating: 5,
			}}, nil
		},
	}

	maxFame := int64(5)
	profiles, err := NewService(repository, &fakeImageStorage{}).SearchProfiles(
		context.Background(),
		userID,
		ProfileFeedOptions{Limit: 20, MaxFame: &maxFame},
	)
	if err != nil {
		t.Fatalf("SearchProfiles() unexpected error: %v", err)
	}
	if len(profiles) != 1 || profiles[0].ID != 20 {
		t.Fatalf("SearchProfiles() = %+v, want profile 20", profiles)
	}
}
