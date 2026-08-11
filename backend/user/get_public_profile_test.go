package user

import (
	"backend/models"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"
)

func publicProfileRequest(viewerID uint, targetUserID string) *http.Request {
	request := httptest.NewRequest(
		http.MethodGet,
		"/profiles/"+targetUserID,
		nil,
	)
	request.SetPathValue("targetUserID", targetUserID)
	return authenticatedUserRequest(request, viewerID)
}

func TestServiceGetPublicProfileRejectsInvalidTargetID(t *testing.T) {
	service := NewService(&fakeUserRepository{}, &fakeImageStorage{})

	_, err := service.GetPublicProfile(context.Background(), 10, 0)
	if !errors.Is(err, UserErrors.InvalidTargetUserID) {
		t.Fatalf("GetPublicProfile() error = %v, want %v", err, UserErrors.InvalidTargetUserID)
	}
}

func TestServiceGetPublicProfileMapsMissingUser(t *testing.T) {
	repository := &fakeUserRepository{
		hasBlockBetweenUsersFn: func(context.Context, uint, uint) (bool, error) {
			return false, nil
		},
		getProfileFn: func(context.Context, uint) (models.User, error) {
			return models.User{}, UserErrors.UserNotFound
		},
	}

	_, err := NewService(repository, &fakeImageStorage{}).GetPublicProfile(
		context.Background(),
		10,
		20,
	)
	if !errors.Is(err, UserErrors.TargetUserNotFound) {
		t.Fatalf("GetPublicProfile() error = %v, want %v", err, UserErrors.TargetUserNotFound)
	}
}

func TestServiceGetPublicProfileRejectsIncompleteProfile(t *testing.T) {
	repository := &fakeUserRepository{
		hasBlockBetweenUsersFn: func(context.Context, uint, uint) (bool, error) {
			return false, nil
		},
		getProfileFn: func(context.Context, uint) (models.User, error) {
			return models.User{ID: 20, IsCompleted: false}, nil
		},
	}

	_, err := NewService(repository, &fakeImageStorage{}).GetPublicProfile(
		context.Background(),
		10,
		20,
	)
	if !errors.Is(err, UserErrors.TargetUserNotFound) {
		t.Fatalf("GetPublicProfile() error = %v, want %v", err, UserErrors.TargetUserNotFound)
	}
}

func TestServiceGetPublicProfilePropagatesRepositoryError(t *testing.T) {
	repositoryError := errors.New("database unavailable")
	repository := &fakeUserRepository{
		hasBlockBetweenUsersFn: func(context.Context, uint, uint) (bool, error) {
			return false, nil
		},
		getProfileFn: func(context.Context, uint) (models.User, error) {
			return models.User{}, repositoryError
		},
	}

	_, err := NewService(repository, &fakeImageStorage{}).GetPublicProfile(
		context.Background(),
		10,
		20,
	)
	if !errors.Is(err, repositoryError) {
		t.Fatalf("GetPublicProfile() error = %v, want %v", err, repositoryError)
	}
}

func TestServiceGetPublicProfileReturnsCompletedProfile(t *testing.T) {
	distance := 0.0
	wantProfile := models.User{
		ID:          20,
		UserName:    "public-user",
		IsCompleted: true,
		Interests:   []string{"Go", "PostgreSQL"},
		Distance:    &distance,
	}
	repository := &fakeUserRepository{
		getProfileFn: func(_ context.Context, userID uint) (models.User, error) {
			if userID != wantProfile.ID {
				t.Fatalf("GetProfile() userID = %d, want %d", userID, wantProfile.ID)
			}
			return wantProfile, nil
		},
	}

	result, err := NewService(repository, &fakeImageStorage{}).GetPublicProfile(
		context.Background(),
		wantProfile.ID,
		wantProfile.ID,
	)
	if err != nil {
		t.Fatalf("GetPublicProfile() unexpected error: %v", err)
	}
	if !reflect.DeepEqual(result.Profile, wantProfile) {
		t.Fatalf("GetPublicProfile().Profile = %+v, want %+v", result.Profile, wantProfile)
	}
	if result.Relationship != (models.ProfileRelationship{}) {
		t.Fatalf("GetPublicProfile().Relationship = %+v, want empty relationship for own profile", result.Relationship)
	}
}

func TestServiceGetPublicProfileReturnsRelationship(t *testing.T) {
	viewerLatitude := 52.52
	viewerLongitude := 13.405
	targetLatitude := 52.50
	targetLongitude := 13.40
	wantRelationship := models.ProfileRelationship{
		LikedByMe: true,
		LikedMe:   true,
	}
	repository := &fakeUserRepository{
		hasBlockBetweenUsersFn: func(context.Context, uint, uint) (bool, error) {
			return false, nil
		},
		getProfileFn: func(context.Context, uint) (models.User, error) {
			return models.User{
				ID:          20,
				IsCompleted: true,
				Latitude:    &targetLatitude,
				Longitude:   &targetLongitude,
			}, nil
		},
		getUserLocationFn: func(context.Context, uint) (*float64, *float64, error) {
			return &viewerLatitude, &viewerLongitude, nil
		},
		getProfileRelationshipFn: func(
			_ context.Context,
			viewerID uint,
			targetUserID uint,
		) (models.ProfileRelationship, error) {
			if viewerID != 10 || targetUserID != 20 {
				t.Fatalf("GetProfileRelationship() arguments = (%d, %d), want (10, 20)", viewerID, targetUserID)
			}
			return wantRelationship, nil
		},
	}

	result, err := NewService(repository, &fakeImageStorage{}).GetPublicProfile(
		context.Background(),
		10,
		20,
	)
	if err != nil {
		t.Fatalf("GetPublicProfile() unexpected error: %v", err)
	}
	if !reflect.DeepEqual(result.Relationship, wantRelationship) {
		t.Fatalf("relationship = %+v, want %+v", result.Relationship, wantRelationship)
	}
	if !result.Relationship.IsConnected() {
		t.Fatal("IsConnected() = false, want true")
	}
}

func TestServiceGetPublicProfileRejectsBlockedRelationship(t *testing.T) {
	getProfileCalls := 0
	repository := &fakeUserRepository{
		hasBlockBetweenUsersFn: func(
			_ context.Context,
			firstUserID uint,
			secondUserID uint,
		) (bool, error) {
			if firstUserID != 10 || secondUserID != 20 {
				t.Fatalf(
					"HasBlockBetweenUsers() arguments = (%d, %d), want (10, 20)",
					firstUserID,
					secondUserID,
				)
			}
			return true, nil
		},
		getProfileFn: func(context.Context, uint) (models.User, error) {
			getProfileCalls++
			return models.User{}, nil
		},
	}

	_, err := NewService(repository, &fakeImageStorage{}).GetPublicProfile(
		context.Background(),
		10,
		20,
	)
	if !errors.Is(err, UserErrors.TargetUserNotFound) {
		t.Fatalf("GetPublicProfile() error = %v, want %v", err, UserErrors.TargetUserNotFound)
	}
	if getProfileCalls != 0 {
		t.Fatalf("GetProfile() calls = %d, want 0", getProfileCalls)
	}
}

func TestServiceGetPublicProfilePropagatesBlockLookupError(t *testing.T) {
	repositoryError := errors.New("block lookup failed")
	repository := &fakeUserRepository{
		hasBlockBetweenUsersFn: func(context.Context, uint, uint) (bool, error) {
			return false, repositoryError
		},
	}

	_, err := NewService(repository, &fakeImageStorage{}).GetPublicProfile(
		context.Background(),
		10,
		20,
	)
	if !errors.Is(err, repositoryError) {
		t.Fatalf("GetPublicProfile() error = %v, want %v", err, repositoryError)
	}
}

func TestGetPublicProfileHandlerRejectsInvalidRequests(t *testing.T) {
	tests := []struct {
		name         string
		request      *http.Request
		wantStatus   int
		wantResponse string
	}{
		{
			name: "unauthorized",
			request: func() *http.Request {
				request := httptest.NewRequest(http.MethodGet, "/profiles/20", nil)
				request.SetPathValue("targetUserID", "20")
				return request
			}(),
			wantStatus:   http.StatusUnauthorized,
			wantResponse: "Unauthorized\n",
		},
		{
			name:         "invalid target user ID",
			request:      publicProfileRequest(10, "invalid"),
			wantStatus:   http.StatusBadRequest,
			wantResponse: "Invalid target user id\n",
		},
		{
			name:         "zero target user ID",
			request:      publicProfileRequest(10, "0"),
			wantStatus:   http.StatusBadRequest,
			wantResponse: "Invalid target user id\n",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			handler := NewUserHandler(NewService(&fakeUserRepository{}, &fakeImageStorage{}))
			response := httptest.NewRecorder()

			handler.GetPublicProfile(response, test.request)

			if response.Code != test.wantStatus {
				t.Fatalf("status = %d, want %d", response.Code, test.wantStatus)
			}
			if response.Body.String() != test.wantResponse {
				t.Fatalf("body = %q, want %q", response.Body.String(), test.wantResponse)
			}
		})
	}
}

func TestGetPublicProfileHandlerMapsServiceErrors(t *testing.T) {
	tests := []struct {
		name          string
		repositoryErr error
		wantStatus    int
		wantResponse  string
	}{
		{
			name:          "profile not found",
			repositoryErr: UserErrors.UserNotFound,
			wantStatus:    http.StatusNotFound,
			wantResponse:  "Profile not found\n",
		},
		{
			name:          "unexpected database error",
			repositoryErr: errors.New("database unavailable"),
			wantStatus:    http.StatusInternalServerError,
			wantResponse:  "Failed to get profile\n",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repository := &fakeUserRepository{
				hasBlockBetweenUsersFn: func(context.Context, uint, uint) (bool, error) {
					return false, nil
				},
				getProfileFn: func(context.Context, uint) (models.User, error) {
					return models.User{}, test.repositoryErr
				},
			}
			handler := NewUserHandler(NewService(repository, &fakeImageStorage{}))
			response := httptest.NewRecorder()

			handler.GetPublicProfile(response, publicProfileRequest(10, "20"))

			if response.Code != test.wantStatus {
				t.Fatalf("status = %d, want %d", response.Code, test.wantStatus)
			}
			if response.Body.String() != test.wantResponse {
				t.Fatalf("body = %q, want %q", response.Body.String(), test.wantResponse)
			}
		})
	}
}

func TestGetPublicProfileHandlerReturnsOnlyPublicFields(t *testing.T) {
	lastSeenAt := time.Date(2026, time.August, 7, 14, 0, 0, 0, time.UTC)
	viewerLatitude := 0.0
	viewerLongitude := 0.0
	targetLatitude := 1.0
	targetLongitude := 0.0
	wantDistance := 111.2
	wantProfile := models.User{
		ID:          20,
		LastSeenAt:  lastSeenAt,
		FameRating:  17,
		UserName:    "public-user",
		FirstName:   "Public",
		LastName:    "User",
		Email:       "private@example.invalid",
		Password:    "private-password-hash",
		IsCompleted: true,
		Latitude:    &targetLatitude,
		Longitude:   &targetLongitude,
		Gender:      "Female",
		Preferences: "Male",
		Bio:         "Public bio",
		Interests:   []string{"Go", "PostgreSQL"},
		Avatar:      "/uploads/avatars/public.jpg",
		Photos: []models.Photo{
			{ID: 30, UserID: 20, URL: "/uploads/photos/public.jpg"},
		},
		Distance: &wantDistance,
	}
	repository := &fakeUserRepository{
		hasBlockBetweenUsersFn: func(context.Context, uint, uint) (bool, error) {
			return false, nil
		},
		getProfileFn: func(
			_ context.Context,
			userID uint,
		) (models.User, error) {
			return wantProfile, nil
		},
		getUserLocationFn: func(_ context.Context, userID uint) (*float64, *float64, error) {
			if userID != 10 {
				t.Fatalf("GetUserLocation() userID = %d, want 10", userID)
			}
			return &viewerLatitude, &viewerLongitude, nil
		},
		getProfileRelationshipFn: func(
			_ context.Context,
			viewerID uint,
			targetUserID uint,
		) (models.ProfileRelationship, error) {
			if viewerID != 10 || targetUserID != 20 {
				t.Fatalf("GetProfileRelationship() arguments = (%d, %d), want (10, 20)", viewerID, targetUserID)
			}
			return models.ProfileRelationship{
				LikedByMe: true,
				LikedMe:   true,
			}, nil
		},

		getCompletionStatusFn: func(
			_ context.Context,
			userID uint,
		) (bool, error) {
			return true, nil
		},

		saveProfileViewFn: func(
			_ context.Context,
			viewerID uint,
			viewedUserID uint,
		) (models.ProfileView, error) {
			return models.ProfileView{
				ViewerID:     viewerID,
				ViewedUserID: viewedUserID,
			}, nil
		},
	}
	handler := NewUserHandler(NewService(repository, &fakeImageStorage{}))
	presenceCalls := 0
	handler.SetUserPresence(&fakeUserPresence{
		isUserOnlineFn: func(_ context.Context, userID uint) bool {
			presenceCalls++
			if userID != wantProfile.ID {
				t.Fatalf("IsUserOnline() userID = %d, want %d", userID, wantProfile.ID)
			}
			return true
		},
	})
	response := httptest.NewRecorder()

	handler.GetPublicProfile(response, publicProfileRequest(10, "20"))

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
	if contentType := response.Header().Get("Content-Type"); contentType != "application/json" {
		t.Fatalf("Content-Type = %q, want application/json", contentType)
	}

	responseData := response.Body.Bytes()
	var body PublicProfileDetailsResponse
	if err := json.Unmarshal(responseData, &body); err != nil {
		t.Fatalf("decode public profile response: %v", err)
	}
	wantResponse := publicProfileResponse(wantProfile)
	if !reflect.DeepEqual(body.PublicProfileResponse, wantResponse) {
		t.Fatalf("public profile response = %+v, want %+v", body.PublicProfileResponse, wantResponse)
	}
	if !body.IsOnline {
		t.Fatal("is_online = false, want true")
	}
	if !body.LastSeenAt.Equal(lastSeenAt) {
		t.Fatalf("last_seen_at = %v, want %v", body.LastSeenAt, lastSeenAt)
	}
	if !body.LikedByMe {
		t.Fatal("liked_by_me = false, want true")
	}
	if !body.LikedMe {
		t.Fatal("liked_me = false, want true")
	}
	if !body.IsConnected {
		t.Fatal("is_connected = false, want true")
	}
	if presenceCalls != 1 {
		t.Fatalf("IsUserOnline() calls = %d, want 1", presenceCalls)
	}

	var fields map[string]json.RawMessage
	if err := json.Unmarshal(responseData, &fields); err != nil {
		t.Fatalf("decode public profile fields: %v", err)
	}
	for _, privateField := range []string{"email", "password", "is_completed", "birth_date", "latitude", "longitude"} {
		if _, exists := fields[privateField]; exists {
			t.Errorf("public response contains private field %q", privateField)
		}
	}
}

func TestGetPublicProfileHandlerHidesBlockedProfile(t *testing.T) {
	profileViewCalls := 0
	repository := &fakeUserRepository{
		hasBlockBetweenUsersFn: func(context.Context, uint, uint) (bool, error) {
			return true, nil
		},
		saveProfileViewFn: func(context.Context, uint, uint) (models.ProfileView, error) {
			profileViewCalls++
			return models.ProfileView{}, nil
		},
	}
	handler := NewUserHandler(NewService(repository, &fakeImageStorage{}))
	response := httptest.NewRecorder()

	handler.GetPublicProfile(response, publicProfileRequest(10, "20"))

	if response.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusNotFound)
	}
	if response.Body.String() != "Profile not found\n" {
		t.Fatalf("body = %q, want %q", response.Body.String(), "Profile not found\n")
	}
	if profileViewCalls != 0 {
		t.Fatalf("SaveProfileView() calls = %d, want 0", profileViewCalls)
	}
}
