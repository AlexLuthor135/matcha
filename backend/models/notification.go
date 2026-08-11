package models

import "time"

type Notification struct {
	ID        uint           `json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	UserID    uint           `json:"user_id"`
	SenderID  *uint          `json:"sender_id"`
	Type      string         `json:"type"`
	Title     string         `json:"title"`
	Message   string         `json:"message"`
	Data      map[string]any `json:"data"`
	ReadAt    *time.Time     `json:"read_at"`
}

type NotificationReadReceipt struct {
	NotificationID uint      `json:"notification_id"`
	ReadAt         time.Time `json:"read_at"`
}
