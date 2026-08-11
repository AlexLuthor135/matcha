package user

import "context"

func (repo *PostgresRepository) UnblockUser(ctx context.Context, blockerID uint, blockedUserID uint) error {
	const query = `DELETE FROM user_blocks WHERE blocker_id = $1 and blocked_user_id = $2`
	_, err := repo.db.ExecContext(ctx, query, blockerID, blockedUserID)
	return err
}
