package user

import (
	"context"
	"database/sql"
	"errors"
)

func (repo *PostgresRepository) DeletePhoto(ctx context.Context, userID uint, photoID uint) (string, error) {
	tx, err := repo.db.BeginTx(ctx, nil)
	if err != nil {
		return "", err
	}
	defer tx.Rollback()
	const lockUserQuery = `SELECT id FROM users WHERE id = $1 FOR UPDATE`
	var lockedUserID uint
	err = tx.QueryRowContext(ctx, lockUserQuery, userID).Scan(&lockedUserID)
	if errors.Is(err, sql.ErrNoRows) {
		return "", UserErrors.UserNotFound
	}
	if err != nil {
		return "", err
	}
	const deletePhotoQuery = `DELETE FROM photos WHERE id = $1 AND user_id = $2 RETURNING url`
	var photoURL string
	err = tx.QueryRowContext(ctx, deletePhotoQuery, photoID, userID).Scan(&photoURL)
	if errors.Is(err, sql.ErrNoRows) {
		return "", UserErrors.PhotoNotFound
	}
	if err != nil {
		return "", err
	}
	if err := tx.Commit(); err != nil {
		return "", err
	}
	return photoURL, nil
}
