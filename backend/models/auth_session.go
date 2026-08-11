package models

import "time"

type AuthSession struct {
	ID         string
	UserID     uint
	CreatedAt  time.Time
	LastUsedAt time.Time
	ExpiresAt  time.Time
	RevokedAt  *time.Time
}
