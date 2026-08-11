package user

import "context"

func (repo *PostgresRepository) HasBlockBetweenUsers(ctx context.Context, firstUserID uint, secondUserID uint) (bool, error) {
	const query = `SELECT EXISTS (SELECT 1 FROM user_blocks as ub WHERE (ub.blocker_id = $1 AND ub.blocked_user_id = $2) OR (ub.blocker_id = $2 AND ub.blocked_user_id = $1))`
	var blockExists bool
	err := repo.db.QueryRowContext(ctx, query, firstUserID, secondUserID).Scan(&blockExists)
	if err != nil {
		return false, err
	}
	return blockExists, nil

}
