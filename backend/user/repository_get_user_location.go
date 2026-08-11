package user

import (
	"context"
	"database/sql"
	"errors"
)

func (repo *PostgresRepository) GetUserLocation(ctx context.Context, userID uint) (*float64, *float64, error) {
	const query = `SELECT latitude, longitude FROM users WHERE id = $1`
	var latitude *float64
	var longitude *float64
	err := repo.db.QueryRowContext(ctx, query, userID).Scan(&latitude, &longitude)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil, UserErrors.UserNotFound
	}
	if err != nil {
		return nil, nil, err
	}
	if latitude == nil || longitude == nil || !isLocationValid(latitude, longitude) {
		return nil, nil, UserErrors.InvalidLocation
	}
	return latitude, longitude, nil
}
