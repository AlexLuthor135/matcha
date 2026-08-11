package user

import (
	"backend/models"
	"context"
	"errors"
	"github.com/jackc/pgx/v5/pgconn"
)

func (repo *PostgresRepository) CreateUser(ctx context.Context, newUser models.User, token models.AccountToken) (models.User, error) {
	tx, err := repo.db.BeginTx(ctx, nil)
	if err != nil {
		return models.User{}, err
	}
	defer tx.Rollback()
	const pendingEmailQuery = `SELECT EXISTS (SELECT 1 FROM users WHERE pending_email = $1)`
	var emailReserved bool
	err = tx.QueryRowContext(ctx, pendingEmailQuery, newUser.Email).Scan(&emailReserved)
	if err != nil {
		return models.User{}, err
	}
	if emailReserved {
		return models.User{}, UserErrors.UserAlreadyExists
	}
	const query = `
		INSERT INTO users (user_name,first_name,last_name,email,password)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id,created_at,updated_at`
	err = tx.QueryRowContext(
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
	token.UserID = newUser.ID
	if err := createAccountToken(ctx, tx, token); err != nil {
		return models.User{}, err
	}
	if err := tx.Commit(); err != nil {
		return models.User{}, err
	}
	return newUser, nil
}
