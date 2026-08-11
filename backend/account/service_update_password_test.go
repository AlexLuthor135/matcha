package account

import (
	"context"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func TestUpdatePasswordUpdatesHashAndRevokesSessions(t *testing.T) {
	const (
		userID          uint = 42
		currentPassword      = "Current1!Password"
		newPassword          = "NewValid1!Password"
	)
	currentHash, err := bcrypt.GenerateFromPassword([]byte(currentPassword), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("hash current password: %v", err)
	}
	updated := false
	repository := &fakeUserRepository{
		getPasswordHashFn: func(_ context.Context, gotUserID uint) (string, error) {
			if gotUserID != userID {
				t.Fatalf("GetPasswordHash() user ID = %d, want %d", gotUserID, userID)
			}
			return string(currentHash), nil
		},
		updatePasswordHashAndRevokeSessionsFn: func(_ context.Context, gotUserID uint, passwordHash string) error {
			if gotUserID != userID {
				t.Fatalf("UpdatePasswordHashAndRevokeSessions() user ID = %d, want %d", gotUserID, userID)
			}
			if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(newPassword)); err != nil {
				t.Fatalf("new password hash does not match: %v", err)
			}
			updated = true
			return nil
		},
	}

	err = NewService(repository, nil).UpdatePassword(
		context.Background(),
		userID,
		currentPassword,
		newPassword,
	)
	if err != nil {
		t.Fatalf("UpdatePassword() unexpected error: %v", err)
	}
	if !updated {
		t.Fatal("UpdatePasswordHashAndRevokeSessions() was not called")
	}
}
