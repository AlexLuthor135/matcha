package notification

import (
	"backend/models"
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"testing"
	"time"
)

func TestServiceNotifyProfileViewRejectsInvalidRecipient(t *testing.T) {
	service := NewService(&fakeNotificationRepository{})

	_, err := service.NotifyProfileView(context.Background(), 0, 20)
	if !errors.Is(err, NotificationErrors.InvalidUserID) {
		t.Fatalf("NotifyProfileView() error = %v, want %v", err, NotificationErrors.InvalidUserID)
	}
}

func TestServiceNotifyProfileViewReturnsRepositoryError(t *testing.T) {
	repositoryError := errors.New("database unavailable")
	repository := &fakeNotificationRepository{
		createNotificationFn: func(
			context.Context,
			CreateNotificationInput,
		) (models.Notification, error) {
			return models.Notification{}, repositoryError
		},
	}

	_, err := NewService(repository).NotifyProfileView(
		context.Background(),
		10,
		20,
	)
	if !errors.Is(err, repositoryError) {
		t.Fatalf("NotifyProfileView() error = %v, want %v", err, repositoryError)
	}
}

func TestServiceNotifyProfileViewSavesWithoutPublisher(t *testing.T) {
	const recipientID uint = 10
	const viewerID uint = 20
	createdAt := time.Date(2026, time.August, 6, 12, 0, 0, 0, time.UTC)
	wantNotification := models.Notification{
		ID:        30,
		CreatedAt: createdAt,
		UpdatedAt: createdAt,
		UserID:    recipientID,
		SenderID:  func() *uint { value := viewerID; return &value }(),
		Type:      NotificationTypeProfileView,
		Title:     "Profile viewed",
		Message:   "Someone viewed your profile",
		Data: map[string]any{
			"viewer_id": viewerID,
		},
	}
	repository := &fakeNotificationRepository{
		createNotificationFn: func(
			_ context.Context,
			input CreateNotificationInput,
		) (models.Notification, error) {
			if input.UserID != recipientID {
				t.Fatalf("UserID = %d, want %d", input.UserID, recipientID)
			}
			if input.SenderID == nil || *input.SenderID != viewerID {
				t.Fatalf("SenderID = %v, want %d", input.SenderID, viewerID)
			}
			if input.Type != NotificationTypeProfileView {
				t.Fatalf("Type = %q, want %q", input.Type, NotificationTypeProfileView)
			}
			if input.Title != "Profile viewed" || input.Message != "Someone viewed your profile" {
				t.Fatalf("notification text = (%q, %q)", input.Title, input.Message)
			}
			gotViewerID, ok := input.Data["viewer_id"].(uint)
			if !ok || gotViewerID != viewerID {
				t.Fatalf("Data = %#v, want viewer_id %d", input.Data, viewerID)
			}
			return wantNotification, nil
		},
	}

	got, err := NewService(repository).NotifyProfileView(
		context.Background(),
		recipientID,
		viewerID,
	)
	if err != nil {
		t.Fatalf("NotifyProfileView() unexpected error: %v", err)
	}
	if !reflect.DeepEqual(got, wantNotification) {
		t.Fatalf("NotifyProfileView() = %+v, want %+v", got, wantNotification)
	}
}

func TestServiceNotifyProfileViewPublishesSavedNotification(t *testing.T) {
	const recipientID uint = 10
	const viewerID uint = 20
	wantNotification := models.Notification{
		ID:       30,
		UserID:   recipientID,
		SenderID: func() *uint { value := viewerID; return &value }(),
		Type:     NotificationTypeProfileView,
		Title:    "Profile viewed",
		Message:  "Someone viewed your profile",
		Data: map[string]any{
			"viewer_id": viewerID,
		},
	}
	repository := &fakeNotificationRepository{
		createNotificationFn: func(
			context.Context,
			CreateNotificationInput,
		) (models.Notification, error) {
			return wantNotification, nil
		},
	}

	sendCalls := 0
	publisher := &fakePublisher{
		sendToUserFn: func(userID uint, message []byte) {
			sendCalls++
			if userID != recipientID {
				t.Fatalf("SendToUser() userID = %d, want %d", userID, recipientID)
			}

			var event notificationEvent
			if err := json.Unmarshal(message, &event); err != nil {
				t.Fatalf("decode published event: %v", err)
			}
			if event.Type != "notification" {
				t.Fatalf("event type = %q, want notification", event.Type)
			}
			if event.Notification.ID != wantNotification.ID || event.Notification.UserID != recipientID {
				t.Fatalf("published notification = %+v, want %+v", event.Notification, wantNotification)
			}
			if event.Notification.SenderID == nil || *event.Notification.SenderID != viewerID {
				t.Fatalf("published sender = %v, want %d", event.Notification.SenderID, viewerID)
			}
			gotViewerID, ok := event.Notification.Data["viewer_id"].(float64)
			if !ok || gotViewerID != float64(viewerID) {
				t.Fatalf("published data = %#v, want viewer_id %d", event.Notification.Data, viewerID)
			}
		},
	}
	service := NewService(repository)
	service.SetPublisher(publisher)

	got, err := service.NotifyProfileView(
		context.Background(),
		recipientID,
		viewerID,
	)
	if err != nil {
		t.Fatalf("NotifyProfileView() unexpected error: %v", err)
	}
	if !reflect.DeepEqual(got, wantNotification) {
		t.Fatalf("NotifyProfileView() = %+v, want %+v", got, wantNotification)
	}
	if sendCalls != 1 {
		t.Fatalf("SendToUser() calls = %d, want 1", sendCalls)
	}
}

func TestServiceNotifyProfileViewReturnsEncodingError(t *testing.T) {
	wantNotification := models.Notification{
		ID:     30,
		UserID: 10,
		Data: map[string]any{
			"unsupported": make(chan int),
		},
	}
	repository := &fakeNotificationRepository{
		createNotificationFn: func(
			context.Context,
			CreateNotificationInput,
		) (models.Notification, error) {
			return wantNotification, nil
		},
	}
	publisher := &fakePublisher{
		sendToUserFn: func(uint, []byte) {
			t.Fatal("SendToUser() called after JSON encoding error")
		},
	}
	service := NewService(repository)
	service.SetPublisher(publisher)

	got, err := service.NotifyProfileView(context.Background(), 10, 20)
	if err == nil {
		t.Fatal("NotifyProfileView() error = nil, want JSON encoding error")
	}
	if got.ID != wantNotification.ID {
		t.Fatalf("NotifyProfileView() notification ID = %d, want %d", got.ID, wantNotification.ID)
	}
}
