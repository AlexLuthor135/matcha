package user

import (
	"backend/models"
	"context"
	"database/sql"
	"errors"
)

type SaveProfileDecisionResult struct {
	ProfileDecision models.ProfileDecision
	IsMatch         bool
	DecisionChanged bool
	MatchEnded      bool
}

func (repo *PostgresRepository) SaveProfileDecision(ctx context.Context, userID uint, targetUserID uint, decision models.ProfileDecisionValue) (SaveProfileDecisionResult, error) {
	tx, err := repo.db.BeginTx(ctx, nil)
	if err != nil {
		return SaveProfileDecisionResult{}, err
	}
	defer tx.Rollback()
	const lockUserQuery = `SELECT id FROM users WHERE id = $1 FOR UPDATE`
	var lockedUserID uint
	err = tx.QueryRowContext(ctx, lockUserQuery, userID).Scan(&lockedUserID)
	if errors.Is(err, sql.ErrNoRows) {
		return SaveProfileDecisionResult{}, UserErrors.UserNotFound
	}
	if err != nil {
		return SaveProfileDecisionResult{}, err
	}
	const userExistQuery = `SELECT EXISTS (SELECT 1 FROM users WHERE id = $1), EXISTS (SELECT 1 FROM users WHERE id = $2)`
	var userExists bool
	var targetUserExists bool
	err = tx.QueryRowContext(ctx, userExistQuery, userID, targetUserID).Scan(&userExists, &targetUserExists)
	if err != nil {
		return SaveProfileDecisionResult{}, err
	}
	if !userExists {
		return SaveProfileDecisionResult{}, UserErrors.UserNotFound
	}
	if !targetUserExists {
		return SaveProfileDecisionResult{}, UserErrors.TargetUserNotFound
	}
	const blockExistsQuery = `SELECT EXISTS (SELECT 1 FROM user_blocks AS ub WHERE (ub.blocker_id = $1 AND ub.blocked_user_id = $2) OR (ub.blocker_id = $2 AND ub.blocked_user_id = $1))`
	var blockExists bool
	err = tx.QueryRowContext(ctx, blockExistsQuery, userID, targetUserID).Scan(&blockExists)
	if err != nil {
		return SaveProfileDecisionResult{}, err
	}
	if blockExists {
		return SaveProfileDecisionResult{}, UserErrors.TargetUserNotFound
	}
	const previousDecisionQuery = `SELECT decision FROM profile_decisions WHERE user_id = $1 AND target_user_id = $2 FOR UPDATE`
	var previousDecision models.ProfileDecisionValue
	hasPreviousDecision := true
	err = tx.QueryRowContext(ctx, previousDecisionQuery, userID, targetUserID).Scan(&previousDecision)
	if errors.Is(err, sql.ErrNoRows) {
		hasPreviousDecision = false
	} else if err != nil {
		return SaveProfileDecisionResult{}, err
	}
	const reverseLikeQuery = `SELECT EXISTS (SELECT 1 FROM profile_decisions WHERE user_id = $1 AND target_user_id = $2 AND decision = 'like')`
	var reverseLikeExists bool
	err = tx.QueryRowContext(ctx, reverseLikeQuery, targetUserID, userID).Scan(&reverseLikeExists)
	if err != nil {
		return SaveProfileDecisionResult{}, err
	}
	wasMatch := hasPreviousDecision && previousDecision == models.ProfileDecisionLike && reverseLikeExists
	decisionChanged := !hasPreviousDecision || previousDecision != decision
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
		return SaveProfileDecisionResult{}, err
	}
	isMatch := savedDecision.Decision == models.ProfileDecisionLike && reverseLikeExists
	matchEnded := wasMatch && decisionChanged && !isMatch
	if err := tx.Commit(); err != nil {
		return SaveProfileDecisionResult{}, err
	}
	return SaveProfileDecisionResult{ProfileDecision: savedDecision, IsMatch: isMatch, DecisionChanged: decisionChanged, MatchEnded: matchEnded}, nil

}
