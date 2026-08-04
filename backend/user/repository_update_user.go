package user

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
)

func (repo *PostgresRepository) UpdateUser(ctx context.Context, userID uint, userName *string, firstName *string, lastName *string, email *string) error {
	const query = `
		UPDATE users
		SET
			user_name = COALESCE($1, user_name),
			first_name = COALESCE($2, first_name),
			last_name = COALESCE($3, last_name),
			email = COALESCE($4, email),
			updated_at = now()
		WHERE id = $5
		RETURNING id
		`
	var updatedUserID uint
	err := repo.db.QueryRowContext(ctx, query, optionalString(userName), optionalString(firstName), optionalString(lastName), optionalString(email), userID).Scan(&updatedUserID)
	if errors.Is(err, sql.ErrNoRows) {
		return UserErrors.UserNotFound
	}
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return UserErrors.UserAlreadyExists
		}
		return err
	}
	return nil
}
