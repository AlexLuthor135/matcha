package main

import (
	"time"

	"gorm.io/gorm"
)

type RegisterRequest struct {
	UserName  string `json:"username"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	Password  string `json:"password"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type User struct {
	gorm.Model
	UserName    string `gorm:"unique;not null"`
	FirstName   string `gorm:"not null"`
	LastName    string `gorm:"not null"`
	Email       string `gorm:"unique;not null"`
	Password    string `gorm:"not null"`
	IsStaff     bool   `gorm:"default:false"`
	IsCompleted bool   `gorm:"default:false"`
	Gender      string
	Preferences string
	Bio         string
	Interests   []string `gorm:"serializer:json"`
	Avatar      string   `gorm:"default:''"`
	Photos      []Photo  `gorm:"constraint:OnDelete:CASCADE;"`
}

type Photo struct {
	gorm.Model
	UserID uint   `gorm:"not null;index"`
	URL    string `gorm:"not null"`
}

type Conversation struct {
	gorm.Model
	UserOneID uint          `gorm:"not null;uniqueIndex:idx_conversation_users"`
	UserTwoID uint          `gorm:"not null;uniqueIndex:idx_conversation_users"`
	Messages  []ChatMessage `gorm:"constraint:OnDelete:CASCADE;"`
}

type ChatMessage struct {
	gorm.Model
	ConversationID uint       `gorm:"not null;index"`
	SenderID       uint       `gorm:"not null;index"`
	RecipientID    uint       `gorm:"not null;index"`
	Content        string     `gorm:"not null"`
	ReadAt         *time.Time `gorm:"index"`
}

type Notification struct {
	gorm.Model
	UserID   uint       `gorm:"not null;index"`
	SenderID *uint      `gorm:"index"`
	Type     string     `gorm:"not null;default:'notification'"`
	Title    string     `gorm:"not null;default:''"`
	Message  string     `gorm:"not null"`
	Data     string     `gorm:"type:jsonb;not null;default:'{}'"`
	ReadAt   *time.Time `gorm:"index"`
}
