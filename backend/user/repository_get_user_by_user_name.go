package user

import (
	"backend/models"
	"context"
	"database/sql"
	"errors"
)

func (repo *PostgresRepository) GetUserByUserName(ctx context.Context, userName string) (models.User, error) {
	const query = `SELECT id, password, is_completed, is_verified FROM users WHERE user_name = $1`
	var user models.User
	err := repo.db.QueryRowContext(ctx, query, userName).Scan(&user.ID, &user.Password, &user.IsCompleted, &user.IsVerified)
	if errors.Is(err, sql.ErrNoRows) {
		return models.User{}, UserErrors.UserNotFound
	}
	if err != nil {
		return models.User{}, err
	}
	return user, nil
}
