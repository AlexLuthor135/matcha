package notification

import (
	"backend/models"
	"context"
	"errors"
	"testing"
	"time"
)

func TestServiceListNotificationsUsesDefaultLimit(t *testing.T) {
	const userID uint = 10
	want := []models.Notification{{ID: 20, UserID: userID}}

	repository := &fakeNotificationRepository{
		listNotificationsFn: func(
			_ context.Context,
			gotUserID uint,
			gotLimit uint,
		) ([]models.Notification, error) {
			if gotUserID != userID {
				t.Fatalf("userID = %d, want %d", gotUserID, userID)
			}
			if gotLimit != defaultNotificationLimit {
				t.Fatalf("limit = %d, want %d", gotLimit, defaultNotificationLimit)
			}
			return want, nil
		},
	}

	notifications, err := NewService(repository).ListNotifications(context.Background(), userID)
	if err != nil {
		t.Fatalf("ListNotifications() unexpected error: %v", err)
	}
	if len(notifications) != 1 || notifications[0].ID != want[0].ID {
		t.Fatalf("ListNotifications() = %+v, want %+v", notifications, want)
	}
}

func TestServiceMarkNotificationReadDelegatesToRepository(t *testing.T) {
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

	receipt, err := NewService(repository).MarkNotificationRead(
		context.Background(),
		userID,
		notificationID,
	)
	if err != nil {
		t.Fatalf("MarkNotificationRead() unexpected error: %v", err)
	}
	if receipt != want {
		t.Fatalf("MarkNotificationRead() = %+v, want %+v", receipt, want)
	}
}

func TestServiceCreateNotificationRejectsInvalidInput(t *testing.T) {
	zeroSenderID := uint(0)
	tests := []struct {
		name    string
		input   CreateNotificationInput
		wantErr error
	}{
		{
			name: "invalid user ID",
			input: CreateNotificationInput{
				Type:    "match",
				Message: "New match",
			},
			wantErr: NotificationErrors.InvalidUserID,
		},
		{
			name: "invalid sender ID",
			input: CreateNotificationInput{
				UserID:   10,
				SenderID: &zeroSenderID,
				Type:     "match",
				Message:  "New match",
			},
			wantErr: NotificationErrors.InvalidSenderID,
		},
		{
			name: "blank type",
			input: CreateNotificationInput{
				UserID:  10,
				Type:    "  \n\t ",
				Message: "New match",
			},
			wantErr: NotificationErrors.TypeBlank,
		},
		{
			name: "blank message",
			input: CreateNotificationInput{
				UserID:  10,
				Type:    "match",
				Message: "  \n\t ",
			},
			wantErr: NotificationErrors.MessageBlank,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			service := NewService(&fakeNotificationRepository{})
			_, err := service.CreateNotification(context.Background(), test.input)
			if !errors.Is(err, test.wantErr) {
				t.Fatalf("CreateNotification() error = %v, want %v", err, test.wantErr)
			}
		})
	}
}

func TestServiceCreateNotificationNormalizesAndDelegates(t *testing.T) {
	const userID uint = 10
	senderID := uint(20)
	want := models.Notification{
		ID:       30,
		UserID:   userID,
		SenderID: &senderID,
		Type:     "match",
		Title:    "New match",
		Message:  "You have a new match",
		Data:     map[string]any{},
	}

	repository := &fakeNotificationRepository{
		createNotificationFn: func(
			_ context.Context,
			input CreateNotificationInput,
		) (models.Notification, error) {
			if input.UserID != userID {
				t.Fatalf("userID = %d, want %d", input.UserID, userID)
			}
			if input.SenderID == nil || *input.SenderID != senderID {
				t.Fatalf("senderID = %v, want %d", input.SenderID, senderID)
			}
			if input.Type != want.Type || input.Title != want.Title || input.Message != want.Message {
				t.Fatalf("normalized input = %+v", input)
			}
			if input.Data == nil || len(input.Data) != 0 {
				t.Fatalf("data = %#v, want non-nil empty map", input.Data)
			}
			return want, nil
		},
	}

	got, err := NewService(repository).CreateNotification(
		context.Background(),
		CreateNotificationInput{
			UserID:   userID,
			SenderID: &senderID,
			Type:     "  match  ",
			Title:    "  New match  ",
			Message:  "  You have a new match  ",
		},
	)
	if err != nil {
		t.Fatalf("CreateNotification() unexpected error: %v", err)
	}
	if got.ID != want.ID || got.UserID != want.UserID || got.Type != want.Type || got.Message != want.Message {
		t.Fatalf("CreateNotification() = %+v, want %+v", got, want)
	}
}

func TestServiceCreateNotificationReturnsRepositoryError(t *testing.T) {
	wantErr := errors.New("database unavailable")
	repository := &fakeNotificationRepository{
		createNotificationFn: func(
			context.Context,
			CreateNotificationInput,
		) (models.Notification, error) {
			return models.Notification{}, wantErr
		},
	}

	_, err := NewService(repository).CreateNotification(
		context.Background(),
		CreateNotificationInput{
			UserID:  10,
			Type:    "match",
			Message: "New match",
		},
	)
	if !errors.Is(err, wantErr) {
		t.Fatalf("CreateNotification() error = %v, want %v", err, wantErr)
	}
}
