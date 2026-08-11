package user

import (
	"backend/models"
	"context"
)

func (repo *PostgresRepository) SaveProfileView(ctx context.Context, viewerID uint, viewedUserID uint) (models.ProfileView, error) {
	const query = `
		INSERT INTO profile_views (
			viewer_id,
			viewed_user_id
		)
		VALUES ($1, $2)
		ON CONFLICT ON CONSTRAINT
			profile_views_viewer_viewed_unique
		DO UPDATE SET
			updated_at = now()
		RETURNING
			id,
			created_at,
			updated_at,
			viewer_id,
			viewed_user_id
	`
	var profileView models.ProfileView
	err := repo.db.QueryRowContext(ctx, query, viewerID, viewedUserID).Scan(&profileView.ID, &profileView.CreatedAt, &profileView.UpdatedAt, &profileView.ViewerID, &profileView.ViewedUserID)
	if err != nil {
		return models.ProfileView{}, err
	}
	return profileView, nil
}
