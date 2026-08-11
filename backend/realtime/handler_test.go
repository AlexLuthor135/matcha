package realtime

import (
	"backend/chat"
	"backend/models"
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"
)

type fakeRealtimeChatRepository struct {
	chat.Repository
	createMessageFn func(context.Context, uint, uint, string) (models.Message, error)
}

func (repo *fakeRealtimeChatRepository) CreateMessage(
	ctx context.Context,
	senderID uint,
	recipientID uint,
	content string,
) (models.Message, error) {
	if repo.createMessageFn == nil {
		panic("unexpected CreateMessage call")
	}
	return repo.createMessageFn(ctx, senderID, recipientID, content)
}

type fakeRealtimeMessageNotifier struct {
	notifyMessageFn func(context.Context, models.Message) (models.Notification, error)
}

func (notifier *fakeRealtimeMessageNotifier) NotifyMessage(
	ctx context.Context,
	message models.Message,
) (models.Notification, error) {
	if notifier.notifyMessageFn == nil {
		panic("unexpected NotifyMessage call")
	}
	return notifier.notifyMessageFn(ctx, message)
}

func TestHandleChatMessageDeliversMessageAndNotifiesRecipient(t *testing.T) {
	wantMessage := models.Message{
		ID:             40,
		ConversationID: 30,
		SenderID:       10,
		RecipientID:    20,
		Content:        "hello",
		CreatedAt:      time.Date(2026, time.August, 7, 12, 0, 0, 0, time.UTC),
	}
	repositoryCalls := 0
	repository := &fakeRealtimeChatRepository{
		createMessageFn: func(
			_ context.Context,
			senderID uint,
			recipientID uint,
			content string,
		) (models.Message, error) {
			repositoryCalls++
			if senderID != wantMessage.SenderID || recipientID != wantMessage.RecipientID || content != wantMessage.Content {
				t.Fatalf("CreateMessage() arguments = (%d, %d, %q)", senderID, recipientID, content)
			}
			return wantMessage, nil
		},
	}
	notifierCalls := 0
	notifier := &fakeRealtimeMessageNotifier{
		notifyMessageFn: func(_ context.Context, message models.Message) (models.Notification, error) {
			notifierCalls++
			if message.ID != wantMessage.ID || message.ConversationID != wantMessage.ConversationID {
				t.Fatalf("NotifyMessage() message = %+v, want %+v", message, wantMessage)
			}
			return models.Notification{ID: 50}, nil
		},
	}
	hub := NewHub()
	hub.deliver = make(chan delivery, 4)
	handler := NewWebsocketHandler(hub, chat.NewService(repository), notifier)
	client := newClient(wantMessage.SenderID, nil)

	handler.handleChatMessage(context.Background(), client, incomingEvent{
		Type:        "chat_message",
		RecipientID: wantMessage.RecipientID,
		Message:     "  hello  ",
	})

	if repositoryCalls != 1 || notifierCalls != 1 {
		t.Fatalf("calls = repository %d, notifier %d; want 1 and 1", repositoryCalls, notifierCalls)
	}
	if len(hub.deliver) != 2 {
		t.Fatalf("chat deliveries = %d, want 2", len(hub.deliver))
	}
	wantRecipients := []uint{wantMessage.SenderID, wantMessage.RecipientID}
	for _, wantUserID := range wantRecipients {
		outgoing := <-hub.deliver
		if outgoing.userID != wantUserID {
			t.Fatalf("delivery userID = %d, want %d", outgoing.userID, wantUserID)
		}
		var event chatMessageEvent
		if err := json.Unmarshal(outgoing.message, &event); err != nil {
			t.Fatalf("decode chat event: %v", err)
		}
		if event.ID != wantMessage.ID || event.Message != wantMessage.Content {
			t.Fatalf("chat event = %+v", event)
		}
	}
}

func TestHandleChatMessageKeepsDeliveriesWhenNotificationFails(t *testing.T) {
	wantMessage := models.Message{
		ID:             40,
		ConversationID: 30,
		SenderID:       10,
		RecipientID:    20,
		Content:        "hello",
	}
	repository := &fakeRealtimeChatRepository{
		createMessageFn: func(context.Context, uint, uint, string) (models.Message, error) {
			return wantMessage, nil
		},
	}
	wantErr := errors.New("notification unavailable")
	notifier := &fakeRealtimeMessageNotifier{
		notifyMessageFn: func(context.Context, models.Message) (models.Notification, error) {
			return models.Notification{}, wantErr
		},
	}
	hub := NewHub()
	hub.deliver = make(chan delivery, 4)
	handler := NewWebsocketHandler(hub, chat.NewService(repository), notifier)

	handler.handleChatMessage(
		context.Background(),
		newClient(wantMessage.SenderID, nil),
		incomingEvent{
			Type:        "chat_message",
			RecipientID: wantMessage.RecipientID,
			Message:     wantMessage.Content,
		},
	)

	if len(hub.deliver) != 2 {
		t.Fatalf("chat deliveries after notification error = %d, want 2", len(hub.deliver))
	}
}

func TestHandleChatMessageDoesNotNotifyWhenSaveFails(t *testing.T) {
	repository := &fakeRealtimeChatRepository{
		createMessageFn: func(context.Context, uint, uint, string) (models.Message, error) {
			return models.Message{}, chat.ChatErrors.UsersNotMatched
		},
	}
	notifierCalls := 0
	notifier := &fakeRealtimeMessageNotifier{
		notifyMessageFn: func(context.Context, models.Message) (models.Notification, error) {
			notifierCalls++
			return models.Notification{}, nil
		},
	}
	hub := NewHub()
	hub.deliver = make(chan delivery, 4)
	handler := NewWebsocketHandler(hub, chat.NewService(repository), notifier)

	handler.handleChatMessage(
		context.Background(),
		newClient(10, nil),
		incomingEvent{
			Type:        "chat_message",
			RecipientID: 20,
			Message:     "hello",
		},
	)

	if notifierCalls != 0 {
		t.Fatalf("NotifyMessage() calls = %d, want 0", notifierCalls)
	}
	if len(hub.deliver) != 1 {
		t.Fatalf("deliveries = %d, want one sender error event", len(hub.deliver))
	}
	outgoing := <-hub.deliver
	if outgoing.userID != 10 {
		t.Fatalf("error delivery userID = %d, want 10", outgoing.userID)
	}
	var event errorEvent
	if err := json.Unmarshal(outgoing.message, &event); err != nil {
		t.Fatalf("decode error event: %v", err)
	}
	if event.Type != "error" || event.Message != chat.ChatErrors.UsersNotMatched.Error() {
		t.Fatalf("error event = %+v", event)
	}
}
