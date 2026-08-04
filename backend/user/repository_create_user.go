package user

import (
	"backend/models"
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
)

func (repo *PostgresRepository) CreateUser(ctx context.Context, newUser models.User) (models.User, error) {
	const query = `
		INSERT INTO users (
			user_name,
			first_name,
			last_name,
			email,
			password
		)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING
			id,
			created_at,
			updated_at`
	err := repo.db.QueryRowContext(
		ctx,
		query,
		newUser.UserName,
		newUser.FirstName,
		newUser.LastName,
		newUser.Email,
		newUser.Password,
	).Scan(
		&newUser.ID,
		&newUser.CreatedAt,
		&newUser.UpdatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return models.User{}, UserErrors.UserAlreadyExists
		}
		return models.User{}, err
	}
	return newUser, nil
}
