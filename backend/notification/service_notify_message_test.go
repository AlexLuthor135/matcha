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

func TestServiceNotifyMessageRejectsInvalidMessage(t *testing.T) {
	tests := []struct {
		name    string
		message models.Message
		wantErr error
	}{
		{
			name: "invalid message ID",
			message: models.Message{
				ConversationID: 30,
				SenderID:       10,
				RecipientID:    20,
			},
			wantErr: NotificationErrors.InvalidMessageID,
		},
		{
			name: "invalid conversation ID",
			message: models.Message{
				ID:          40,
				SenderID:    10,
				RecipientID: 20,
			},
			wantErr: NotificationErrors.InvalidConversationID,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			service := NewService(&fakeNotificationRepository{})

			_, err := service.NotifyMessage(context.Background(), test.message)
			if !errors.Is(err, test.wantErr) {
				t.Fatalf("NotifyMessage() error = %v, want %v", err, test.wantErr)
			}
		})
	}
}

func TestServiceNotifyMessageSavesAndPublishes(t *testing.T) {
	message := models.Message{
		ID:             40,
		ConversationID: 30,
		SenderID:       10,
		RecipientID:    20,
		Content:        "private message body",
		CreatedAt:      time.Date(2026, time.August, 7, 12, 0, 0, 0, time.UTC),
	}
	wantNotification := models.Notification{
		ID:       50,
		UserID:   message.RecipientID,
		SenderID: &message.SenderID,
		Type:     NotificationTypeMessage,
		Title:    "New message",
		Message:  "You received a new message",
		Data: map[string]any{
			"message_id":      message.ID,
			"conversation_id": message.ConversationID,
			"sender_id":       message.SenderID,
		},
	}
	repositoryCalls := 0
	repository := &fakeNotificationRepository{
		createNotificationFn: func(
			_ context.Context,
			input CreateNotificationInput,
		) (models.Notification, error) {
			repositoryCalls++
			if input.UserID != message.RecipientID {
				t.Fatalf("UserID = %d, want %d", input.UserID, message.RecipientID)
			}
			if input.SenderID == nil || *input.SenderID != message.SenderID {
				t.Fatalf("SenderID = %v, want %d", input.SenderID, message.SenderID)
			}
			if input.Type != NotificationTypeMessage || input.Title != "New message" {
				t.Fatalf("notification metadata = (%q, %q)", input.Type, input.Title)
			}
			if input.Message != "You received a new message" {
				t.Fatalf("Message = %q", input.Message)
			}
			if reflect.DeepEqual(input.Data, wantNotification.Data) == false {
				t.Fatalf("Data = %#v, want %#v", input.Data, wantNotification.Data)
			}
			if encodedData, err := json.Marshal(input.Data); err != nil || string(encodedData) == "" {
				t.Fatalf("notification data is not JSON encodable: %v", err)
			}
			return wantNotification, nil
		},
	}
	publishCalls := 0
	publisher := &fakePublisher{
		sendToUserFn: func(userID uint, data []byte) {
			publishCalls++
			if userID != message.RecipientID {
				t.Fatalf("published userID = %d, want %d", userID, message.RecipientID)
			}
			var event notificationEvent
			if err := json.Unmarshal(data, &event); err != nil {
				t.Fatalf("decode notification event: %v", err)
			}
			if event.Type != "notification" || event.Notification.ID != wantNotification.ID {
				t.Fatalf("published event = %+v", event)
			}
		},
	}
	service := NewService(repository)
	service.SetPublisher(publisher)

	got, err := service.NotifyMessage(context.Background(), message)
	if err != nil {
		t.Fatalf("NotifyMessage() unexpected error: %v", err)
	}
	if got.ID != wantNotification.ID {
		t.Fatalf("NotifyMessage() notification ID = %d, want %d", got.ID, wantNotification.ID)
	}
	if repositoryCalls != 1 || publishCalls != 1 {
		t.Fatalf("calls = repository %d, publisher %d; want 1 and 1", repositoryCalls, publishCalls)
	}
}

func TestServiceNotifyMessageReturnsRepositoryErrorWithoutPublishing(t *testing.T) {
	wantErr := errors.New("database unavailable")
	repository := &fakeNotificationRepository{
		createNotificationFn: func(
			context.Context,
			CreateNotificationInput,
		) (models.Notification, error) {
			return models.Notification{}, wantErr
		},
	}
	publishCalls := 0
	service := NewService(repository)
	service.SetPublisher(&fakePublisher{
		sendToUserFn: func(uint, []byte) {
			publishCalls++
		},
	})

	_, err := service.NotifyMessage(context.Background(), models.Message{
		ID:             40,
		ConversationID: 30,
		SenderID:       10,
		RecipientID:    20,
	})
	if !errors.Is(err, wantErr) {
		t.Fatalf("NotifyMessage() error = %v, want %v", err, wantErr)
	}
	if publishCalls != 0 {
		t.Fatalf("SendToUser() calls = %d, want 0", publishCalls)
	}
}

func TestServiceNotifyMessageReturnsSavedNotificationOnEncodingError(t *testing.T) {
	wantNotification := models.Notification{
		ID:     50,
		UserID: 20,
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
	publishCalls := 0
	service := NewService(repository)
	service.SetPublisher(&fakePublisher{
		sendToUserFn: func(uint, []byte) {
			publishCalls++
		},
	})

	got, err := service.NotifyMessage(context.Background(), models.Message{
		ID:             40,
		ConversationID: 30,
		SenderID:       10,
		RecipientID:    20,
	})
	if err == nil {
		t.Fatal("NotifyMessage() error = nil, want JSON encoding error")
	}
	if got.ID != wantNotification.ID {
		t.Fatalf("NotifyMessage() notification ID = %d, want %d", got.ID, wantNotification.ID)
	}
	if publishCalls != 0 {
		t.Fatalf("SendToUser() calls = %d, want 0", publishCalls)
	}
}
