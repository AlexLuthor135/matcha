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

func TestServiceListProfileLikersReturnsRepositoryLikers(t *testing.T) {
	const userID uint = 10
	wantLikers := []models.ProfileLiker{
		{
			ID:        20,
			UserName:  "recent-liker",
			FirstName: "Recent",
			LastName:  "Liker",
			Avatar:    "/uploads/avatars/recent.jpg",
			LikedAt:   time.Date(2026, time.August, 6, 12, 0, 0, 0, time.UTC),
		},
	}
	repository := &fakeUserRepository{
		listProfileLikersFn: func(_ context.Context, gotUserID uint) ([]models.ProfileLiker, error) {
			if gotUserID != userID {
				t.Fatalf("ListProfileLikers() userID = %d, want %d", gotUserID, userID)
			}
			return wantLikers, nil
		},
	}

	likers, err := NewService(repository, &fakeImageStorage{}).ListProfileLikers(
		context.Background(),
		userID,
	)
	if err != nil {
		t.Fatalf("ListProfileLikers() unexpected error: %v", err)
	}
	if !reflect.DeepEqual(likers, wantLikers) {
		t.Fatalf("ListProfileLikers() = %+v, want %+v", likers, wantLikers)
	}
}

func TestServiceListProfileLikersPropagatesRepositoryError(t *testing.T) {
	repositoryError := errors.New("database unavailable")
	repository := &fakeUserRepository{
		listProfileLikersFn: func(context.Context, uint) ([]models.ProfileLiker, error) {
			return nil, repositoryError
		},
	}

	_, err := NewService(repository, &fakeImageStorage{}).ListProfileLikers(
		context.Background(),
		10,
	)
	if !errors.Is(err, repositoryError) {
		t.Fatalf("ListProfileLikers() error = %v, want %v", err, repositoryError)
	}
}

func TestListProfileLikersHandlerRejectsUnauthorizedRequest(t *testing.T) {
	handler := NewUserHandler(NewService(&fakeUserRepository{}, &fakeImageStorage{}))
	request := httptest.NewRequest(http.MethodGet, "/profile/likes", nil)
	response := httptest.NewRecorder()

	handler.ListProfileLikers(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusUnauthorized)
	}
	if response.Body.String() != "Unauthorized\n" {
		t.Fatalf("body = %q, want %q", response.Body.String(), "Unauthorized\n")
	}
}

func TestListProfileLikersHandlerMapsServiceError(t *testing.T) {
	repository := &fakeUserRepository{
		listProfileLikersFn: func(context.Context, uint) ([]models.ProfileLiker, error) {
			return nil, errors.New("database unavailable")
		},
	}
	handler := NewUserHandler(NewService(repository, &fakeImageStorage{}))
	request := authenticatedUserRequest(
		httptest.NewRequest(http.MethodGet, "/profile/likes", nil),
		10,
	)
	response := httptest.NewRecorder()

	handler.ListProfileLikers(response, request)

	if response.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusInternalServerError)
	}
	if response.Body.String() != "Failed to get profile likers\n" {
		t.Fatalf(
			"body = %q, want %q",
			response.Body.String(),
			"Failed to get profile likers\n",
		)
	}
}

func TestListProfileLikersHandlerReturnsLikers(t *testing.T) {
	const userID uint = 10
	wantLikers := []models.ProfileLiker{
		{
			ID:        20,
			UserName:  "recent-liker",
			FirstName: "Recent",
			LastName:  "Liker",
			Avatar:    "/uploads/avatars/recent.jpg",
			LikedAt:   time.Date(2026, time.August, 6, 12, 0, 0, 0, time.UTC),
		},
	}
	repository := &fakeUserRepository{
		listProfileLikersFn: func(_ context.Context, gotUserID uint) ([]models.ProfileLiker, error) {
			if gotUserID != userID {
				t.Fatalf("ListProfileLikers() userID = %d, want %d", gotUserID, userID)
			}
			return wantLikers, nil
		},
	}
	handler := NewUserHandler(NewService(repository, &fakeImageStorage{}))
	request := authenticatedUserRequest(
		httptest.NewRequest(http.MethodGet, "/profile/likes", nil),
		userID,
	)
	response := httptest.NewRecorder()

	handler.ListProfileLikers(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
	if contentType := response.Header().Get("Content-Type"); contentType != "application/json" {
		t.Fatalf("Content-Type = %q, want application/json", contentType)
	}

	var body ListProfileLikersResponse
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode profile likers response: %v", err)
	}
	if !reflect.DeepEqual(body.Likers, wantLikers) {
		t.Fatalf("likers = %+v, want %+v", body.Likers, wantLikers)
	}
}

func TestListProfileLikersHandlerReturnsEmptyArray(t *testing.T) {
	repository := &fakeUserRepository{
		listProfileLikersFn: func(context.Context, uint) ([]models.ProfileLiker, error) {
			return make([]models.ProfileLiker, 0), nil
		},
	}
	handler := NewUserHandler(NewService(repository, &fakeImageStorage{}))
	request := authenticatedUserRequest(
		httptest.NewRequest(http.MethodGet, "/profile/likes", nil),
		10,
	)
	response := httptest.NewRecorder()

	handler.ListProfileLikers(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
	if response.Body.String() != "{\"likers\":[]}\n" {
		t.Fatalf("body = %q, want empty likers array", response.Body.String())
	}
}
