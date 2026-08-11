package user

import "context"

func (repo *PostgresRepository) ReportUser(ctx context.Context, reporterID uint, reportedUserID uint) error {
	tx, err := repo.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	const usersExistQuery = `SELECT EXISTS (SELECT 1 FROM users WHERE id = $1), EXISTS (SELECT 1 FROM users WHERE id = $2)`
	var reporterExists bool
	var reportedUserExists bool
	err = tx.QueryRowContext(ctx, usersExistQuery, reporterID, reportedUserID).Scan(&reporterExists, &reportedUserExists)
	if err != nil {
		return err
	}
	if !reporterExists {
		return UserErrors.UserNotFound
	}
	if !reportedUserExists {
		return UserErrors.TargetUserNotFound
	}
	const reportQuery = `INSERT INTO user_reports (reporter_id, reported_user_id) VALUES ($1, $2) ON CONFLICT (reporter_id, reported_user_id) DO NOTHING`
	if _, err = tx.ExecContext(ctx, reportQuery, reporterID, reportedUserID); err != nil {
		return err
	}
	return tx.Commit()
}
