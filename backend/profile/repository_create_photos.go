package profile

import (
	"backend/models"
	"context"
	"database/sql"
	"errors"
)

func (repo *PostgresRepository) CreatePhotos(ctx context.Context, userID uint, photoURLs []string, maxAllowed int) ([]models.Photo, error) {
	tx, err := repo.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	const lockUserQuery = `SELECT id FROM users WHERE id = $1 FOR UPDATE`
	var lockedUserID uint
	err = tx.QueryRowContext(ctx, lockUserQuery, userID).Scan(&lockedUserID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ProfileErrors.UserNotFound
	}
	if err != nil {
		return nil, err
	}
	const countPhotoQuery = `SELECT COUNT(*) FROM photos WHERE user_id = $1`
	var existingPhotoCount int
	err = tx.QueryRowContext(ctx, countPhotoQuery, userID).Scan(&existingPhotoCount)
	if err != nil {
		return nil, err
	}
	if existingPhotoCount+len(photoURLs) > maxAllowed {
		return nil, ProfileErrors.PhotoLimitExceeded
	}
	const insertPhotoQuery = `INSERT INTO photos (user_id, url) VALUES ($1, $2) RETURNING id, user_id, url`
	photos := make([]models.Photo, 0, len(photoURLs))
	for _, photoURL := range photoURLs {
		var photo models.Photo
		err = tx.QueryRowContext(ctx, insertPhotoQuery, userID, photoURL).Scan(&photo.ID, &photo.UserID, &photo.URL)
		if err != nil {
			return nil, err
		}
		photos = append(photos, photo)
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return photos, nil
}
