package profile

import (
	"context"
	"time"
)

func (repo *PostgresRepository) UpdateLastSeen(ctx context.Context, userID uint, lastSeenAt time.Time) error {
	const query = `UPDATE users SET last_seen_at = $1 WHERE id = $2`
	result, err := repo.db.ExecContext(ctx, query, lastSeenAt, userID)
	if err != nil {
		return err
	}
	affectedRows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affectedRows == 0 {
		return ProfileErrors.UserNotFound
	}
	return nil
}
