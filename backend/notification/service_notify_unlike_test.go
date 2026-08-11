package notification

import (
	"backend/models"
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"testing"
)

func TestServiceNotifyUnlikeRejectsInvalidRecipient(t *testing.T) {
	service := NewService(&fakeNotificationRepository{})

	_, err := service.NotifyUnlike(context.Background(), 0, 20)
	if !errors.Is(err, NotificationErrors.InvalidUserID) {
		t.Fatalf("NotifyUnlike() error = %v, want %v", err, NotificationErrors.InvalidUserID)
	}
}

func TestServiceNotifyUnlikeReturnsRepositoryError(t *testing.T) {
	repositoryError := errors.New("database unavailable")
	repository := &fakeNotificationRepository{
		createNotificationFn: func(
			context.Context,
			CreateNotificationInput,
		) (models.Notification, error) {
			return models.Notification{}, repositoryError
		},
	}

	_, err := NewService(repository).NotifyUnlike(context.Background(), 10, 20)
	if !errors.Is(err, repositoryError) {
		t.Fatalf("NotifyUnlike() error = %v, want %v", err, repositoryError)
	}
}

func TestServiceNotifyUnlikeSavesWithoutPublisher(t *testing.T) {
	const recipientID uint = 10
	const unlikerID uint = 20
	wantNotification := models.Notification{
		ID:       30,
		UserID:   recipientID,
		SenderID: func() *uint { value := unlikerID; return &value }(),
		Type:     NotificationTypeUnlike,
		Title:    "Match ended",
		Message:  "Someone unmatched with you",
		Data: map[string]any{
			"unliker_id": unlikerID,
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
			if input.SenderID == nil || *input.SenderID != unlikerID {
				t.Fatalf("SenderID = %v, want %d", input.SenderID, unlikerID)
			}
			if input.Type != NotificationTypeUnlike {
				t.Fatalf("Type = %q, want %q", input.Type, NotificationTypeUnlike)
			}
			if input.Title != "Match ended" || input.Message != "Someone unmatched with you" {
				t.Fatalf("notification text = (%q, %q)", input.Title, input.Message)
			}
			gotUnlikerID, ok := input.Data["unliker_id"].(uint)
			if !ok || gotUnlikerID != unlikerID {
				t.Fatalf("Data = %#v, want unliker_id %d", input.Data, unlikerID)
			}
			return wantNotification, nil
		},
	}

	got, err := NewService(repository).NotifyUnlike(context.Background(), recipientID, unlikerID)
	if err != nil {
		t.Fatalf("NotifyUnlike() unexpected error: %v", err)
	}
	if !reflect.DeepEqual(got, wantNotification) {
		t.Fatalf("NotifyUnlike() = %+v, want %+v", got, wantNotification)
	}
}

func TestServiceNotifyUnlikePublishesSavedNotification(t *testing.T) {
	const recipientID uint = 10
	const unlikerID uint = 20
	wantNotification := models.Notification{
		ID:       30,
		UserID:   recipientID,
		SenderID: func() *uint { value := unlikerID; return &value }(),
		Type:     NotificationTypeUnlike,
		Title:    "Match ended",
		Message:  "Someone unmatched with you",
		Data: map[string]any{
			"unliker_id": unlikerID,
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
			gotUnlikerID, ok := event.Notification.Data["unliker_id"].(float64)
			if !ok || gotUnlikerID != float64(unlikerID) {
				t.Fatalf("published data = %#v, want unliker_id %d", event.Notification.Data, unlikerID)
			}
		},
	}
	service := NewService(repository)
	service.SetPublisher(publisher)

	got, err := service.NotifyUnlike(context.Background(), recipientID, unlikerID)
	if err != nil {
		t.Fatalf("NotifyUnlike() unexpected error: %v", err)
	}
	if !reflect.DeepEqual(got, wantNotification) {
		t.Fatalf("NotifyUnlike() = %+v, want %+v", got, wantNotification)
	}
	if sendCalls != 1 {
		t.Fatalf("SendToUser() calls = %d, want 1", sendCalls)
	}
}

func TestServiceNotifyUnlikeReturnsEncodingError(t *testing.T) {
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

	got, err := service.NotifyUnlike(context.Background(), 10, 20)
	if err == nil {
		t.Fatal("NotifyUnlike() error = nil, want JSON encoding error")
	}
	if got.ID != wantNotification.ID {
		t.Fatalf("NotifyUnlike() notification ID = %d, want %d", got.ID, wantNotification.ID)
	}
}
