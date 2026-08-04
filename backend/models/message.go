package models

import "time"

type Message struct {
	ID             uint
	ConversationID uint
	SenderID       uint
	RecipientID    uint
	Content        string
	CreatedAt      time.Time
	UpdatedAt      time.Time
	ReadAt         *time.Time
}

type MessageReadReceipt struct {
	MessageID      uint
	ConversationID uint
	SenderID       uint
	RecipientID    uint
	ReadAt         time.Time
}
