package chat

import (
	"backend/models"
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

type fakeRepository struct {
	createMessageFn            func(context.Context, uint, uint, string) (models.Message, error)
	listConversationsFn        func(context.Context, uint) ([]models.Conversation, error)
	listConversationMessagesFn func(context.Context, uint, uint) ([]models.Message, error)
	markMessageReadFn          func(context.Context, uint, uint) (models.MessageReadReceipt, error)
}

func (repo *fakeRepository) CreateMessage(ctx context.Context, senderID uint, recipientID uint, content string) (models.Message, error) {
	if repo.createMessageFn == nil {
		panic("unexpected CreateMessage call")
	}
	return repo.createMessageFn(ctx, senderID, recipientID, content)
}

func (repo *fakeRepository) ListConversations(ctx context.Context, userID uint) ([]models.Conversation, error) {
	if repo.listConversationsFn == nil {
		panic("unexpected ListConversations call")
	}
	return repo.listConversationsFn(ctx, userID)
}

func (repo *fakeRepository) ListConversationMessages(ctx context.Context, userID uint, conversationID uint) ([]models.Message, error) {
	if repo.listConversationMessagesFn == nil {
		panic("unexpected ListConversationMessages call")
	}
	return repo.listConversationMessagesFn(ctx, userID, conversationID)
}

func (repo *fakeRepository) MarkMessageRead(ctx context.Context, recipientID uint, messageID uint) (models.MessageReadReceipt, error) {
	if repo.markMessageReadFn == nil {
		panic("unexpected MarkMessageRead call")
	}
	return repo.markMessageReadFn(ctx, recipientID, messageID)
}

func TestServiceSaveMessageRejectsInvalidInput(t *testing.T) {
	tests := []struct {
		name        string
		senderID    uint
		recipientID uint
		content     string
		wantErr     error
	}{
		{
			name:        "same sender and recipient",
			senderID:    10,
			recipientID: 10,
			content:     "hello",
			wantErr:     ChatErrors.CannotMessageSelf,
		},
		{
			name:        "blank message",
			senderID:    10,
			recipientID: 20,
			content:     "  \n\t ",
			wantErr:     ChatErrors.MessageBlank,
		},
		{
			name:        "message too long",
			senderID:    10,
			recipientID: 20,
			content:     strings.Repeat("a", maxMessageLength+1),
			wantErr:     ChatErrors.MessageTooLong,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			service := NewService(&fakeRepository{})
			_, err := service.SaveMessage(context.Background(), test.senderID, test.recipientID, test.content)
			if !errors.Is(err, test.wantErr) {
				t.Fatalf("SaveMessage() error = %v, want %v", err, test.wantErr)
			}
		})
	}
}

func TestServiceSaveMessageNormalizesAndCreatesMessage(t *testing.T) {
	const senderID uint = 10
	const recipientID uint = 20

	want := models.Message{
		ID:          30,
		SenderID:    senderID,
		RecipientID: recipientID,
		Content:     "hello",
	}

	repository := &fakeRepository{
		createMessageFn: func(_ context.Context, gotSenderID uint, gotRecipientID uint, gotContent string) (models.Message, error) {
			if gotSenderID != senderID {
				t.Fatalf("senderID = %d, want %d", gotSenderID, senderID)
			}
			if gotRecipientID != recipientID {
				t.Fatalf("recipientID = %d, want %d", gotRecipientID, recipientID)
			}
			if gotContent != want.Content {
				t.Fatalf("content = %q, want %q", gotContent, want.Content)
			}
			return want, nil
		},
	}

	message, err := NewService(repository).SaveMessage(context.Background(), senderID, recipientID, "  hello  ")
	if err != nil {
		t.Fatalf("SaveMessage() unexpected error: %v", err)
	}
	if message != want {
		t.Fatalf("SaveMessage() = %+v, want %+v", message, want)
	}
}

func TestServiceListConversationsDelegatesToRepository(t *testing.T) {
	const userID uint = 10
	want := []models.Conversation{{ID: 40, UserOneID: 10, UserTwoID: 20}}

	repository := &fakeRepository{
		listConversationsFn: func(_ context.Context, gotUserID uint) ([]models.Conversation, error) {
			if gotUserID != userID {
				t.Fatalf("userID = %d, want %d", gotUserID, userID)
			}
			return want, nil
		},
	}

	conversations, err := NewService(repository).ListConversations(context.Background(), userID)
	if err != nil {
		t.Fatalf("ListConversations() unexpected error: %v", err)
	}
	if len(conversations) != 1 || conversations[0] != want[0] {
		t.Fatalf("ListConversations() = %+v, want %+v", conversations, want)
	}
}

func TestServiceListConversationMessagesDelegatesToRepository(t *testing.T) {
	const userID uint = 10
	const conversationID uint = 40
	want := []models.Message{{ID: 50, ConversationID: conversationID}}

	repository := &fakeRepository{
		listConversationMessagesFn: func(_ context.Context, gotUserID uint, gotConversationID uint) ([]models.Message, error) {
			if gotUserID != userID || gotConversationID != conversationID {
				t.Fatalf("arguments = (%d, %d), want (%d, %d)", gotUserID, gotConversationID, userID, conversationID)
			}
			return want, nil
		},
	}

	messages, err := NewService(repository).ListConversationMessages(context.Background(), userID, conversationID)
	if err != nil {
		t.Fatalf("ListConversationMessages() unexpected error: %v", err)
	}
	if len(messages) != 1 || messages[0] != want[0] {
		t.Fatalf("ListConversationMessages() = %+v, want %+v", messages, want)
	}
}

func TestServiceMarkMessageReadDelegatesToRepository(t *testing.T) {
	const recipientID uint = 20
	const messageID uint = 50

	want := models.MessageReadReceipt{
		MessageID:      messageID,
		ConversationID: 40,
		SenderID:       10,
		RecipientID:    recipientID,
		ReadAt:         time.Date(2026, time.August, 2, 12, 0, 0, 0, time.UTC),
	}

	repository := &fakeRepository{
		markMessageReadFn: func(_ context.Context, gotRecipientID uint, gotMessageID uint) (models.MessageReadReceipt, error) {
			if gotRecipientID != recipientID || gotMessageID != messageID {
				t.Fatalf("arguments = (%d, %d), want (%d, %d)", gotRecipientID, gotMessageID, recipientID, messageID)
			}
			return want, nil
		},
	}

	receipt, err := NewService(repository).MarkMessageRead(context.Background(), recipientID, messageID)
	if err != nil {
		t.Fatalf("MarkMessageRead() unexpected error: %v", err)
	}
	if receipt != want {
		t.Fatalf("MarkMessageRead() = %+v, want %+v", receipt, want)
	}
}
