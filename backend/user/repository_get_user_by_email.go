package user

import (
	"backend/models"
	"context"
	"database/sql"
	"errors"
)

func (repo *PostgresRepository) GetUserByEmail(ctx context.Context, email string) (models.User, error) {
	const query = `SELECT id, password, is_completed FROM users WHERE email = $1`
	var user models.User
	err := repo.db.QueryRowContext(ctx, query, email).Scan(&user.ID, &user.Password, &user.IsCompleted)
	if errors.Is(err, sql.ErrNoRows) {
		return models.User{}, UserErrors.UserNotFound
	}
	if err != nil {
		return models.User{}, err
	}
	return user, nil
}
