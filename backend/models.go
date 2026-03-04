package main
import "gorm.io/gorm"

type RegisterRequest struct {
	Username string `json:"username"`
	FirstName string `json:"first_name"`
	LastName string `json:"last_name"`
	Email string `json:"email"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type User struct {
	gorm.Model
	Username string `gorm:"unique;not null"`
	FirstName string `gorm:"not null"`
	LastName string `gorm:"not null"`
	Email string `gorm:"unique;not null"`
	Password string `gorm:"not null"`
	IsStaff bool `gorm:"default:false"`
	Gender    string
    Bio       string
    Interests []string `gorm:"serializer:json"`
}