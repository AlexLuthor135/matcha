package notification

import (
	"backend/middleware"
	"backend/models"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func authenticatedNotificationRequest(request *http.Request, userID uint) *http.Request {
	ctx := context.WithValue(request.Context(), middleware.UserIDKey, userID)
	return request.WithContext(ctx)
}

func markNotificationReadRequest(userID uint, notificationID string) *http.Request {
	request := httptest.NewRequest(
		http.MethodPatch,
		"/notifications/"+notificationID+"/read",
		nil,
	)
	request.SetPathValue("notificationID", notificationID)
	return authenticatedNotificationRequest(request, userID)
}

func TestListNotificationsHandler(t *testing.T) {
	t.Run("unauthorized", func(t *testing.T) {
		handler := NewNotificationHandler(NewService(&fakeNotificationRepository{}))
		response := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodGet, "/notifications", nil)

		handler.ListNotifications(response, request)

		if response.Code != http.StatusUnauthorized {
			t.Fatalf("status = %d, want %d", response.Code, http.StatusUnauthorized)
		}
	})

	t.Run("repository error", func(t *testing.T) {
		repository := &fakeNotificationRepository{
			listNotificationsFn: func(context.Context, uint, uint) ([]models.Notification, error) {
				return nil, errors.New("database unavailable")
			},
		}
		handler := NewNotificationHandler(NewService(repository))
		response := httptest.NewRecorder()
		request := authenticatedNotificationRequest(
			httptest.NewRequest(http.MethodGet, "/notifications", nil),
			10,
		)

		handler.ListNotifications(response, request)

		if response.Code != http.StatusInternalServerError {
			t.Fatalf("status = %d, want %d", response.Code, http.StatusInternalServerError)
		}
		if response.Body.String() != "Failed to get notifications\n" {
			t.Fatalf("body = %q, want repository error response", response.Body.String())
		}
	})

	t.Run("empty list", func(t *testing.T) {
		repository := &fakeNotificationRepository{
			listNotificationsFn: func(_ context.Context, userID uint, limit uint) ([]models.Notification, error) {
				if userID != 10 || limit != defaultNotificationLimit {
					t.Fatalf("arguments = (%d, %d), want (10, %d)", userID, limit, defaultNotificationLimit)
				}
				return make([]models.Notification, 0), nil
			},
		}
		handler := NewNotificationHandler(NewService(repository))
		response := httptest.NewRecorder()
		request := authenticatedNotificationRequest(
			httptest.NewRequest(http.MethodGet, "/notifications", nil),
			10,
		)

		handler.ListNotifications(response, request)

		if response.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
		}
		if contentType := response.Header().Get("Content-Type"); contentType != "application/json" {
			t.Fatalf("Content-Type = %q, want application/json", contentType)
		}

		var body ListNotificationsResponse
		if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
			t.Fatalf("decode response: %v", err)
		}
		if body.Notifications == nil || len(body.Notifications) != 0 {
			t.Fatalf("notifications = %#v, want non-nil empty slice", body.Notifications)
		}
	})

	t.Run("notifications", func(t *testing.T) {
		createdAt := time.Date(2026, time.August, 5, 12, 0, 0, 0, time.UTC)
		senderID := uint(20)
		want := []models.Notification{
			{
				ID:        30,
				CreatedAt: createdAt,
				UpdatedAt: createdAt,
				UserID:    10,
				SenderID:  &senderID,
				Type:      "match",
				Title:     "New match",
				Message:   "You have a new match",
				Data:      map[string]any{"matched_user_id": float64(20)},
			},
		}
		repository := &fakeNotificationRepository{
			listNotificationsFn: func(context.Context, uint, uint) ([]models.Notification, error) {
				return want, nil
			},
		}
		handler := NewNotificationHandler(NewService(repository))
		response := httptest.NewRecorder()
		request := authenticatedNotificationRequest(
			httptest.NewRequest(http.MethodGet, "/notifications", nil),
			10,
		)

		handler.ListNotifications(response, request)

		if response.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
		}
		var body ListNotificationsResponse
		if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
			t.Fatalf("decode response: %v", err)
		}
		if len(body.Notifications) != 1 {
			t.Fatalf("notifications length = %d, want 1", len(body.Notifications))
		}
		got := body.Notifications[0]
		if got.ID != want[0].ID || got.Type != want[0].Type || got.Message != want[0].Message {
			t.Fatalf("notification = %+v, want %+v", got, want[0])
		}
	})
}

func TestMarkNotificationReadHandlerRejectsInvalidRequests(t *testing.T) {
	tests := []struct {
		name         string
		request      func() *http.Request
		wantStatus   int
		wantResponse string
	}{
		{
			name: "unauthorized",
			request: func() *http.Request {
				request := httptest.NewRequest(http.MethodPatch, "/notifications/20/read", nil)
				request.SetPathValue("notificationID", "20")
				return request
			},
			wantStatus:   http.StatusUnauthorized,
			wantResponse: "Unauthorized\n",
		},
		{
			name: "missing notification ID",
			request: func() *http.Request {
				return markNotificationReadRequest(10, "")
			},
			wantStatus:   http.StatusBadRequest,
			wantResponse: "Invalid notification ID\n",
		},
		{
			name: "zero notification ID",
			request: func() *http.Request {
				return markNotificationReadRequest(10, "0")
			},
			wantStatus:   http.StatusBadRequest,
			wantResponse: "Invalid notification ID\n",
		},
		{
			name: "non-numeric notification ID",
			request: func() *http.Request {
				return markNotificationReadRequest(10, "invalid")
			},
			wantStatus:   http.StatusBadRequest,
			wantResponse: "Invalid notification ID\n",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			handler := NewNotificationHandler(NewService(&fakeNotificationRepository{}))
			response := httptest.NewRecorder()

			handler.MarkNotificationRead(response, test.request())

			if response.Code != test.wantStatus {
				t.Fatalf("status = %d, want %d", response.Code, test.wantStatus)
			}
			if response.Body.String() != test.wantResponse {
				t.Fatalf("body = %q, want %q", response.Body.String(), test.wantResponse)
			}
		})
	}
}

func TestMarkNotificationReadHandlerMapsRepositoryErrors(t *testing.T) {
	tests := []struct {
		name          string
		repositoryErr error
		wantStatus    int
		wantResponse  string
	}{
		{
			name:          "notification not found",
			repositoryErr: NotificationErrors.NotificationNotFound,
			wantStatus:    http.StatusNotFound,
			wantResponse:  "Notification not found\n",
		},
		{
			name:          "unexpected database error",
			repositoryErr: errors.New("database unavailable"),
			wantStatus:    http.StatusInternalServerError,
			wantResponse:  "Failed to mark notification read\n",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repository := &fakeNotificationRepository{
				markNotificationReadFn: func(
					context.Context,
					uint,
					uint,
				) (models.NotificationReadReceipt, error) {
					return models.NotificationReadReceipt{}, test.repositoryErr
				},
			}
			handler := NewNotificationHandler(NewService(repository))
			response := httptest.NewRecorder()

			handler.MarkNotificationRead(
				response,
				markNotificationReadRequest(10, "20"),
			)

			if response.Code != test.wantStatus {
				t.Fatalf("status = %d, want %d", response.Code, test.wantStatus)
			}
			if response.Body.String() != test.wantResponse {
				t.Fatalf("body = %q, want %q", response.Body.String(), test.wantResponse)
			}
		})
	}
}

func TestMarkNotificationReadHandlerReturnsReceipt(t *testing.T) {
	const userID uint = 10
	const notificationID uint = 20
	readAt := time.Date(2026, time.August, 5, 12, 0, 0, 0, time.UTC)
	want := models.NotificationReadReceipt{
		NotificationID: notificationID,
		ReadAt:         readAt,
	}
	repository := &fakeNotificationRepository{
		markNotificationReadFn: func(
			_ context.Context,
			gotUserID uint,
			gotNotificationID uint,
		) (models.NotificationReadReceipt, error) {
			if gotUserID != userID || gotNotificationID != notificationID {
				t.Fatalf(
					"arguments = (%d, %d), want (%d, %d)",
					gotUserID,
					gotNotificationID,
					userID,
					notificationID,
				)
			}
			return want, nil
		},
	}
	handler := NewNotificationHandler(NewService(repository))
	response := httptest.NewRecorder()

	handler.MarkNotificationRead(
		response,
		markNotificationReadRequest(userID, "20"),
	)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
	if contentType := response.Header().Get("Content-Type"); contentType != "application/json" {
		t.Fatalf("Content-Type = %q, want application/json", contentType)
	}
	var body MarkNotificationResponse
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.NotificationID != want.NotificationID || !body.ReadAt.Equal(want.ReadAt) {
		t.Fatalf("response = %+v, want %+v", body, want)
	}
}
