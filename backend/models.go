package main

import "gorm.io/gorm"

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
