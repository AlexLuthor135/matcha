package account

import (
	"backend/models"
	"context"
	"database/sql"
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
)

func (repo *PostgresRepository) UpdateUser(ctx context.Context, userID uint, input UserUpdateInput, verificationToken *models.AccountToken) (UpdateUserResult, error) {
	tx, err := repo.db.BeginTx(ctx, nil)
	if err != nil {
		return UpdateUserResult{}, err
	}
	defer tx.Rollback()
	const currentEmailQuery = `SELECT email FROM users WHERE id = $1 FOR UPDATE`
	var currentEmail string
	err = tx.QueryRowContext(ctx, currentEmailQuery, userID).Scan(&currentEmail)
	if errors.Is(err, sql.ErrNoRows) {
		return UpdateUserResult{}, AccountErrors.UserNotFound
	}
	if err != nil {
		return UpdateUserResult{}, err
	}
	emailChanged := input.Email != nil && *input.Email != currentEmail
	if emailChanged {
		const emailExistsQuery = `SELECT EXISTS (SELECT 1 FROM users WHERE id <> $1 AND (email = $2 OR pending_email = $2))`
		var emailExists bool
		err = tx.QueryRowContext(ctx, emailExistsQuery, userID, *input.Email).Scan(&emailExists)
		if err != nil {
			return UpdateUserResult{}, err
		}
		if emailExists {
			return UpdateUserResult{}, AccountErrors.UserAlreadyExists
		}
		if verificationToken == nil {
			return UpdateUserResult{}, errors.New("verification token is required for email change")
		}
	}
	const query = `
		UPDATE users
		SET
			user_name = COALESCE($1, user_name),
			first_name = COALESCE($2, first_name),
			last_name = COALESCE($3, last_name),
			pending_email = CASE WHEN $6::boolean THEN $4 ELSE pending_email END,
			updated_at = now()
		WHERE id = $5
		RETURNING id
		`
	var updatedUserID uint
	err = tx.QueryRowContext(
		ctx,
		query,
		optionalString(input.UserName),
		optionalString(input.FirstName),
		optionalString(input.LastName),
		optionalString(input.Email),
		userID,
		emailChanged).Scan(&updatedUserID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return UpdateUserResult{}, AccountErrors.UserAlreadyExists
		}
		return UpdateUserResult{}, err
	}
	result := UpdateUserResult{EmailChanged: emailChanged}
	if emailChanged {
		result.PendingEmail = *input.Email
		const invalidateQuery = `UPDATE account_tokens SET used_at = now() WHERE user_id = $1 AND purpose = $2 AND used_at IS NULL`
		_, err = tx.ExecContext(ctx, invalidateQuery, userID, models.AccountTokenPurposeEmailVerification)
		if err != nil {
			return UpdateUserResult{}, err
		}
		verificationToken.UserID = userID
		if err := createAccountToken(ctx, tx, *verificationToken); err != nil {
			return UpdateUserResult{}, err
		}
	}
	if err = tx.Commit(); err != nil {
		return UpdateUserResult{}, err
	}
	return result, nil
}
