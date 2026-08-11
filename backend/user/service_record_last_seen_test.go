package user

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestServiceRecordLastSeenDelegatesToRepository(t *testing.T) {
	const userID uint = 10
	wantLastSeenAt := time.Date(2026, time.August, 7, 15, 0, 0, 0, time.UTC)
	repositoryCalls := 0
	repository := &fakeUserRepository{
		updateLastSeenFn: func(_ context.Context, gotUserID uint, gotLastSeenAt time.Time) error {
			repositoryCalls++
			if gotUserID != userID {
				t.Fatalf("userID = %d, want %d", gotUserID, userID)
			}
			if !gotLastSeenAt.Equal(wantLastSeenAt) {
				t.Fatalf("lastSeenAt = %v, want %v", gotLastSeenAt, wantLastSeenAt)
			}
			return nil
		},
	}

	err := NewService(repository, &fakeImageStorage{}).RecordLastSeen(
		context.Background(),
		userID,
		wantLastSeenAt,
	)
	if err != nil {
		t.Fatalf("RecordLastSeen() unexpected error: %v", err)
	}
	if repositoryCalls != 1 {
		t.Fatalf("UpdateLastSeen() calls = %d, want 1", repositoryCalls)
	}
}

func TestServiceRecordLastSeenReturnsRepositoryError(t *testing.T) {
	wantErr := errors.New("database unavailable")
	repository := &fakeUserRepository{
		updateLastSeenFn: func(context.Context, uint, time.Time) error {
			return wantErr
		},
	}

	err := NewService(repository, &fakeImageStorage{}).RecordLastSeen(
		context.Background(),
		10,
		time.Now(),
	)
	if !errors.Is(err, wantErr) {
		t.Fatalf("RecordLastSeen() error = %v, want %v", err, wantErr)
	}
}
