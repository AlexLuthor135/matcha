package models

import "time"

type ProfileDecisionValue string

const (
	ProfileDecisionLike    ProfileDecisionValue = "like"
	ProfileDecisionDislike ProfileDecisionValue = "dislike"
)

type ProfileDecision struct {
	ID           uint                 `json:"id"`
	CreatedAt    time.Time            `json:"created_at"`
	UpdatedAt    time.Time            `json:"updated_at"`
	UserID       uint                 `json:"user_id"`
	TargetUserID uint                 `json:"target_user_id"`
	Decision     ProfileDecisionValue `json:"decision"`
}

type ProfileLiker struct {
	ID        uint      `json:"id"`
	UserName  string    `json:"user_name"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Avatar    string    `json:"avatar"`
	LikedAt   time.Time `json:"liked_at"`
}

type ProfileRelationship struct {
	LikedByMe bool
	LikedMe   bool
}

func (relationship ProfileRelationship) IsConnected() bool {
	return relationship.LikedByMe && relationship.LikedMe
}
