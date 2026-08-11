package models

import "time"

type AccountTokenPurpose string

const (
	AccountTokenPurposeEmailVerification AccountTokenPurpose = "email_verification"
	AccountTokenPurposePasswordReset     AccountTokenPurpose = "password_reset"
)

type AccountToken struct {
	UserID    uint
	Hash      string
	Purpose   AccountTokenPurpose
	ExpiresAt time.Time
}

func (purpose AccountTokenPurpose) IsValid() bool {
	switch purpose {
	case AccountTokenPurposeEmailVerification, AccountTokenPurposePasswordReset:
		return true
	default:
		return false
	}
}
