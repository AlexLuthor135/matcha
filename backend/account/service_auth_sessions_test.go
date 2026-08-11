package account

import (
	"backend/models"
	"context"
	"errors"
	"testing"
	"time"
)

func TestCreateAuthSessionPersistsGeneratedSession(t *testing.T) {
	const userID uint = 42
	var savedSession models.AuthSession
	repository := &fakeUserRepository{
		createAuthSessionFn: func(_ context.Context, session models.AuthSession) error {
			savedSession = session
			return nil
		},
	}

	session, err := NewService(repository, nil).CreateAuthSession(
		context.Background(),
		userID,
	)
	if err != nil {
		t.Fatalf("CreateAuthSession() unexpected error: %v", err)
	}
	if session.ID == "" || savedSession.ID != session.ID {
		t.Fatalf("saved session ID = %q, returned ID = %q", savedSession.ID, session.ID)
	}
	if session.UserID != userID || savedSession.UserID != userID {
		t.Fatalf("session user ID = %d, saved user ID = %d", session.UserID, savedSession.UserID)
	}
	wantMinimumExpiry := time.Now().UTC().Add(authSessionLifeTime - time.Minute)
	if session.ExpiresAt.Before(wantMinimumExpiry) {
		t.Fatalf("session expiration = %v, want about %v from now", session.ExpiresAt, authSessionLifeTime)
	}
}

func TestValidateAuthSessionMapsInvalidSession(t *testing.T) {
	repository := &fakeUserRepository{
		useAuthSessionFn: func(context.Context, string, uint) error {
			return AccountErrors.InvalidAuthSession
		},
	}
	valid, err := NewService(repository, nil).ValidateAuthSession(
		context.Background(),
		"session-42",
		42,
	)
	if err != nil || valid {
		t.Fatalf("ValidateAuthSession() = (%v, %v), want (false, nil)", valid, err)
	}
}

func TestValidateAuthSessionPropagatesRepositoryError(t *testing.T) {
	wantErr := errors.New("database unavailable")
	repository := &fakeUserRepository{
		useAuthSessionFn: func(context.Context, string, uint) error {
			return wantErr
		},
	}
	valid, err := NewService(repository, nil).ValidateAuthSession(
		context.Background(),
		"session-42",
		42,
	)
	if valid || !errors.Is(err, wantErr) {
		t.Fatalf("ValidateAuthSession() = (%v, %v), want (false, %v)", valid, err, wantErr)
	}
}
