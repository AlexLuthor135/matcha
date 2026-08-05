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
