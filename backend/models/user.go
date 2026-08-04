package models

import (
	"time"
)

type User struct {
	ID        uint      `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	UserName    string   `json:"user_name"`
	FirstName   string   `json:"first_name"`
	LastName    string   `json:"last_name"`
	Email       string   `json:"email"`
	Password    string   `json:"-"`
	IsCompleted bool     `json:"is_completed"`
	Gender      string   `json:"gender"`
	Preferences string   `json:"preferences"`
	Bio         string   `json:"bio"`
	Interests   []string `json:"interests"`
	Avatar      string   `json:"avatar"`
	Photos      []Photo  `json:"photos,omitempty"`
}
