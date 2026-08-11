package models

import "time"

type ProfileView struct {
	ID           uint      `json:"id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	ViewerID     uint      `json:"viewer_id"`
	ViewedUserID uint      `json:"viewed_user_id"`
}

type ProfileViewer struct {
	ID           uint      `json:"id"`
	UserName     string    `json:"user_name"`
	FirstName    string    `json:"first_name"`
	LastName     string    `json:"last_name"`
	Avatar       string    `json:"avatar"`
	LastViewedAt time.Time `json:"last_viewed_at"`
}
