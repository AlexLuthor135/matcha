package chat

import (
	"backend/models"
	"context"
	"database/sql"
)

type Repository interface {
	CreateMessage(ctx context.Context, senderID uint, recipientID uint, content string) (models.Message, error)
	ListConversations(ctx context.Context, userID uint) ([]models.Conversation, error)
	ListConversationMessages(ctx context.Context, userID uint, conversationID uint) ([]models.Message, error)
	MarkMessageRead(ctx context.Context, recipientID uint, messageID uint) (models.MessageReadReceipt, error)
}

type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}
