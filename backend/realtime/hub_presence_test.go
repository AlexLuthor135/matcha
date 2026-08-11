package realtime

import (
	"context"
	"testing"
	"time"
)

type recordedLastSeen struct {
	userID     uint
	lastSeenAt time.Time
}

type fakePresenceRecorder struct {
	records chan recordedLastSeen
}

func (recorder *fakePresenceRecorder) RecordLastSeen(
	_ context.Context,
	userID uint,
	lastSeenAt time.Time,
) error {
	recorder.records <- recordedLastSeen{
		userID:     userID,
		lastSeenAt: lastSeenAt,
	}
	return nil
}

func TestHubTracksMultipleConnectionsAndRecordsLastDisconnect(t *testing.T) {
	const userID uint = 10
	recorder := &fakePresenceRecorder{
		records: make(chan recordedLastSeen, 1),
	}
	hub := NewHub()
	hub.SetPresenceRecorder(recorder)
	go hub.Run()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	if hub.IsUserOnline(ctx, userID) {
		t.Fatal("IsUserOnline() before register = true, want false")
	}

	firstClient := newClient(userID, nil)
	secondClient := newClient(userID, nil)
	hub.Register(firstClient)
	hub.Register(secondClient)

	if !hub.IsUserOnline(ctx, userID) {
		t.Fatal("IsUserOnline() after register = false, want true")
	}

	hub.Unregister(firstClient)
	if !hub.IsUserOnline(ctx, userID) {
		t.Fatal("IsUserOnline() after first disconnect = false, want true")
	}
	select {
	case record := <-recorder.records:
		t.Fatalf("RecordLastSeen() after first disconnect = %+v, want no call", record)
	default:
	}

	disconnectedAfter := time.Now().UTC()
	hub.Unregister(secondClient)
	if hub.IsUserOnline(ctx, userID) {
		t.Fatal("IsUserOnline() after last disconnect = true, want false")
	}

	select {
	case record := <-recorder.records:
		if record.userID != userID {
			t.Fatalf("RecordLastSeen() userID = %d, want %d", record.userID, userID)
		}
		if record.lastSeenAt.Before(disconnectedAfter) {
			t.Fatalf("RecordLastSeen() time = %v, want at or after %v", record.lastSeenAt, disconnectedAfter)
		}
	case <-ctx.Done():
		t.Fatal("RecordLastSeen() was not called after last disconnect")
	}
}
