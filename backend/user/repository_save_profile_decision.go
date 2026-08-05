package user

import (
	"backend/models"
	"context"
)

func (repo *PostgresRepository) SaveProfileDecision(ctx context.Context, userID uint, targetUserID uint, decision models.ProfileDecisionValue) (models.ProfileDecision, bool, error) {
	tx, err := repo.db.BeginTx(ctx, nil)
	if err != nil {
		return models.ProfileDecision{}, false, err
	}
	defer tx.Rollback()
	const userExistQuery = `SELECT EXISTS (SELECT 1 FROM users WHERE id = $1), EXISTS (SELECT 1 FROM users WHERE id = $2)`
	var userExists bool
	var targetUserExists bool
	err = tx.QueryRowContext(ctx, userExistQuery, userID, targetUserID).Scan(&userExists, &targetUserExists)
	if err != nil {
		return models.ProfileDecision{}, false, err
	}
	if !userExists {
		return models.ProfileDecision{}, false, UserErrors.UserNotFound
	}
	if !targetUserExists {
		return models.ProfileDecision{}, false, UserErrors.TargetUserNotFound
	}
	const saveDecisionQuery = `
		INSERT INTO profile_decisions (user_id, target_user_id, decision)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id, target_user_id)
		DO UPDATE SET decision = EXCLUDED.decision, updated_at = now()
		RETURNING id, created_at, updated_at, user_id, target_user_id, decision
	`
	var savedDecision models.ProfileDecision
	err = tx.QueryRowContext(
		ctx,
		saveDecisionQuery,
		userID,
		targetUserID,
		decision).Scan(
		&savedDecision.ID,
		&savedDecision.CreatedAt,
		&savedDecision.UpdatedAt,
		&savedDecision.UserID,
		&savedDecision.TargetUserID,
		&savedDecision.Decision)
	if err != nil {
		return models.ProfileDecision{}, false, err
	}
	var isMatch bool
	if savedDecision.Decision == models.ProfileDecisionLike {
		const reverseLikeQuery = `SELECT EXISTS (SELECT 1 FROM profile_decisions WHERE user_id = $1 AND target_user_id = $2 AND decision = 'like')`
		err = tx.QueryRowContext(ctx, reverseLikeQuery, targetUserID, userID).Scan(&isMatch)
		if err != nil {
			return models.ProfileDecision{}, false, err
		}
	}
	if err := tx.Commit(); err != nil {
		return models.ProfileDecision{}, false, err
	}
	return savedDecision, isMatch, nil

}
