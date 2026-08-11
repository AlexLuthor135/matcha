package user

import "context"

func (repo *PostgresRepository) BlockUser(ctx context.Context, blockerID uint, blockedUserID uint) error {
	tx, err := repo.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	const usersExistQuery = `SELECT EXISTS (SELECT 1 FROM users WHERE id = $1), EXISTS (SELECT 1 FROM users WHERE id = $2)`
	var blockerExists bool
	var blockedUserExists bool
	err = tx.QueryRowContext(ctx, usersExistQuery, blockerID, blockedUserID).Scan(&blockerExists, &blockedUserExists)
	if err != nil {
		return err
	}
	if !blockerExists {
		return UserErrors.UserNotFound
	}
	if !blockedUserExists {
		return UserErrors.TargetUserNotFound
	}
	const blockUserQuery = `INSERT INTO user_blocks (blocker_id, blocked_user_id) VALUES ($1, $2) ON CONFLICT (blocker_id, blocked_user_id) DO NOTHING`
	if _, err := tx.ExecContext(ctx, blockUserQuery, blockerID, blockedUserID); err != nil {
		return err
	}
	const deleteDecisionQuery = `DELETE FROM profile_decisions WHERE (user_id = $1 AND target_user_id = $2) OR (user_id = $2 AND target_user_id = $1)`
	if _, err := tx.ExecContext(ctx, deleteDecisionQuery, blockerID, blockedUserID); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}
