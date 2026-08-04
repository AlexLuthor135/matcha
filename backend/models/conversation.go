package models

import "time"

type Conversation struct {
	ID        uint
	UserOneID uint
	UserTwoID uint
	CreatedAt time.Time
	UpdatedAt time.Time
}
