package profile

import (
	"context"
	"time"
)

func (s *Service) RecordLastSeen(ctx context.Context, userID uint, lastSeenAt time.Time) error {
	return s.repository.UpdateLastSeen(ctx, userID, lastSeenAt)
}
