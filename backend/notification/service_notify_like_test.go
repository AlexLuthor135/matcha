package notification

import (
	"backend/models"
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"testing"
)

func TestServiceNotifyLikeRejectsInvalidRecipient(t *testing.T) {
	service := NewService(&fakeNotificationRepository{})

	_, err := service.NotifyLike(context.Background(), 0, 20)
	if !errors.Is(err, NotificationErrors.InvalidUserID) {
		t.Fatalf("NotifyLike() error = %v, want %v", err, NotificationErrors.InvalidUserID)
	}
}

func TestServiceNotifyLikeReturnsRepositoryError(t *testing.T) {
	repositoryError := errors.New("database unavailable")
	publishCalls := 0
	repository := &fakeNotificationRepository{
		createNotificationFn: func(
			context.Context,
			CreateNotificationInput,
		) (models.Notification, error) {
			return models.Notification{}, repositoryError
		},
	}
	service := NewService(repository)
	service.SetPublisher(&fakePublisher{
		sendToUserFn: func(uint, []byte) {
			publishCalls++
		},
	})

	_, err := service.NotifyLike(context.Background(), 10, 20)
	if !errors.Is(err, repositoryError) {
		t.Fatalf("NotifyLike() error = %v, want %v", err, repositoryError)
	}
	if publishCalls != 0 {
		t.Fatalf("SendToUser() calls = %d, want 0 after repository error", publishCalls)
	}
}

func TestServiceNotifyLikeSavesWithoutPublisher(t *testing.T) {
	const recipientID uint = 10
	const likerID uint = 20
	wantNotification := models.Notification{
		ID:       30,
		UserID:   recipientID,
		SenderID: func() *uint { value := likerID; return &value }(),
		Type:     NotificationTypeLike,
		Title:    "New like",
		Message:  "Someone liked your profile",
		Data: map[string]any{
			"liker_id": likerID,
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
			if input.SenderID == nil || *input.SenderID != likerID {
				t.Fatalf("SenderID = %v, want %d", input.SenderID, likerID)
			}
			if input.Type != NotificationTypeLike {
				t.Fatalf("Type = %q, want %q", input.Type, NotificationTypeLike)
			}
			if input.Title != "New like" || input.Message != "Someone liked your profile" {
				t.Fatalf("notification text = (%q, %q)", input.Title, input.Message)
			}
			gotLikerID, ok := input.Data["liker_id"].(uint)
			if !ok || gotLikerID != likerID {
				t.Fatalf("Data = %#v, want liker_id %d", input.Data, likerID)
			}
			return wantNotification, nil
		},
	}

	got, err := NewService(repository).NotifyLike(context.Background(), recipientID, likerID)
	if err != nil {
		t.Fatalf("NotifyLike() unexpected error: %v", err)
	}
	if !reflect.DeepEqual(got, wantNotification) {
		t.Fatalf("NotifyLike() = %+v, want %+v", got, wantNotification)
	}
}

func TestServiceNotifyLikePublishesSavedNotification(t *testing.T) {
	const recipientID uint = 10
	const likerID uint = 20
	wantNotification := models.Notification{
		ID:       30,
		UserID:   recipientID,
		SenderID: func() *uint { value := likerID; return &value }(),
		Type:     NotificationTypeLike,
		Title:    "New like",
		Message:  "Someone liked your profile",
		Data: map[string]any{
			"liker_id": likerID,
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
			gotLikerID, ok := event.Notification.Data["liker_id"].(float64)
			if !ok || gotLikerID != float64(likerID) {
				t.Fatalf("published data = %#v, want liker_id %d", event.Notification.Data, likerID)
			}
		},
	}
	service := NewService(repository)
	service.SetPublisher(publisher)

	got, err := service.NotifyLike(context.Background(), recipientID, likerID)
	if err != nil {
		t.Fatalf("NotifyLike() unexpected error: %v", err)
	}
	if !reflect.DeepEqual(got, wantNotification) {
		t.Fatalf("NotifyLike() = %+v, want %+v", got, wantNotification)
	}
	if sendCalls != 1 {
		t.Fatalf("SendToUser() calls = %d, want 1", sendCalls)
	}
}

func TestServiceNotifyLikeReturnsEncodingError(t *testing.T) {
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

	got, err := service.NotifyLike(context.Background(), 10, 20)
	if err == nil {
		t.Fatal("NotifyLike() error = nil, want JSON encoding error")
	}
	if got.ID != wantNotification.ID {
		t.Fatalf("NotifyLike() notification ID = %d, want %d", got.ID, wantNotification.ID)
	}
}
