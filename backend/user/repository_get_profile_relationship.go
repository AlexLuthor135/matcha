package user

import (
	"backend/models"
	"context"
)

func (repo *PostgresRepository) GetProfileRelationship(ctx context.Context, viewerID uint, targetUserID uint) (models.ProfileRelationship, error) {
	const query = `
		SELECT 
		EXISTS (SELECT 1 FROM profile_decisions WHERE user_id = $1 AND target_user_id = $2 AND decision = 'like'),
		EXISTS (SELECT 1 FROM profile_decisions WHERE user_id = $2 AND target_user_id = $1 AND decision = 'like')`
	var relationship models.ProfileRelationship
	err := repo.db.QueryRowContext(ctx, query, viewerID, targetUserID).Scan(&relationship.LikedByMe, &relationship.LikedMe)
	if err != nil {
		return models.ProfileRelationship{}, err
	}
	return relationship, nil
}
