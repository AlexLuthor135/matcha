package relationship

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

func TestServiceListProfileViewersReturnsRepositoryViewers(t *testing.T) {
	const userID uint = 10
	wantViewers := []models.ProfileViewer{
		{
			ID:           20,
			UserName:     "recent-viewer",
			FirstName:    "Recent",
			LastName:     "Viewer",
			Avatar:       "/uploads/avatars/recent.jpg",
			LastViewedAt: time.Date(2026, time.August, 6, 12, 0, 0, 0, time.UTC),
		},
	}
	repository := &fakeUserRepository{
		listProfileViewersFn: func(_ context.Context, gotUserID uint) ([]models.ProfileViewer, error) {
			if gotUserID != userID {
				t.Fatalf("ListProfileViewers() userID = %d, want %d", gotUserID, userID)
			}
			return wantViewers, nil
		},
	}

	viewers, err := NewService(repository).ListProfileViewers(
		context.Background(),
		userID,
	)
	if err != nil {
		t.Fatalf("ListProfileViewers() unexpected error: %v", err)
	}
	if !reflect.DeepEqual(viewers, wantViewers) {
		t.Fatalf("ListProfileViewers() = %+v, want %+v", viewers, wantViewers)
	}
}

func TestServiceListProfileViewersPropagatesRepositoryError(t *testing.T) {
	repositoryError := errors.New("database unavailable")
	repository := &fakeUserRepository{
		listProfileViewersFn: func(context.Context, uint) ([]models.ProfileViewer, error) {
			return nil, repositoryError
		},
	}

	_, err := NewService(repository).ListProfileViewers(
		context.Background(),
		10,
	)
	if !errors.Is(err, repositoryError) {
		t.Fatalf("ListProfileViewers() error = %v, want %v", err, repositoryError)
	}
}

func TestListProfileViewersHandlerRejectsUnauthorizedRequest(t *testing.T) {
	handler := NewHandler(NewService(&fakeUserRepository{}))
	request := httptest.NewRequest(http.MethodGet, "/profile/views", nil)
	response := httptest.NewRecorder()

	handler.ListProfileViewers(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusUnauthorized)
	}
	if response.Body.String() != "Unauthorized\n" {
		t.Fatalf("body = %q, want %q", response.Body.String(), "Unauthorized\n")
	}
}

func TestListProfileViewersHandlerMapsServiceError(t *testing.T) {
	repository := &fakeUserRepository{
		listProfileViewersFn: func(context.Context, uint) ([]models.ProfileViewer, error) {
			return nil, errors.New("database unavailable")
		},
	}
	handler := NewHandler(NewService(repository))
	request := authenticatedUserRequest(
		httptest.NewRequest(http.MethodGet, "/profile/views", nil),
		10,
	)
	response := httptest.NewRecorder()

	handler.ListProfileViewers(response, request)

	if response.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusInternalServerError)
	}
	if response.Body.String() != "Failed to get profile viewers\n" {
		t.Fatalf(
			"body = %q, want %q",
			response.Body.String(),
			"Failed to get profile viewers\n",
		)
	}
}

func TestListProfileViewersHandlerReturnsViewers(t *testing.T) {
	const userID uint = 10
	wantViewers := []models.ProfileViewer{
		{
			ID:           20,
			UserName:     "recent-viewer",
			FirstName:    "Recent",
			LastName:     "Viewer",
			Avatar:       "/uploads/avatars/recent.jpg",
			LastViewedAt: time.Date(2026, time.August, 6, 12, 0, 0, 0, time.UTC),
		},
	}
	repository := &fakeUserRepository{
		listProfileViewersFn: func(_ context.Context, gotUserID uint) ([]models.ProfileViewer, error) {
			if gotUserID != userID {
				t.Fatalf("ListProfileViewers() userID = %d, want %d", gotUserID, userID)
			}
			return wantViewers, nil
		},
	}
	handler := NewHandler(NewService(repository))
	request := authenticatedUserRequest(
		httptest.NewRequest(http.MethodGet, "/profile/views", nil),
		userID,
	)
	response := httptest.NewRecorder()

	handler.ListProfileViewers(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
	if contentType := response.Header().Get("Content-Type"); contentType != "application/json" {
		t.Fatalf("Content-Type = %q, want application/json", contentType)
	}

	var body ListProfileViewers
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode profile viewers response: %v", err)
	}
	if !reflect.DeepEqual(body.Viewers, wantViewers) {
		t.Fatalf("viewers = %+v, want %+v", body.Viewers, wantViewers)
	}
}

func TestListProfileViewersHandlerReturnsEmptyArray(t *testing.T) {
	repository := &fakeUserRepository{
		listProfileViewersFn: func(context.Context, uint) ([]models.ProfileViewer, error) {
			return make([]models.ProfileViewer, 0), nil
		},
	}
	handler := NewHandler(NewService(repository))
	request := authenticatedUserRequest(
		httptest.NewRequest(http.MethodGet, "/profile/views", nil),
		10,
	)
	response := httptest.NewRecorder()

	handler.ListProfileViewers(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
	if response.Body.String() != "{\"viewers\":[]}\n" {
		t.Fatalf("body = %q, want empty viewers array", response.Body.String())
	}
}
