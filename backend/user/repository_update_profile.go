package user

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
)

func (repo *PostgresRepository) UpdateProfile(ctx context.Context, userID uint, input UpdateProfileInput) error {
	var interestsValue any
	if input.Interests != nil {
		rawInterests, err := json.Marshal(*input.Interests)
		if err != nil {
			return err
		}
		interestsValue = string(rawInterests)
	}
	locationProvided := input.Location != nil
	var (
		locationSource    any
		locationName      any
		locationConsentAt any
		latitude          any
		longitude         any
	)
	if input.Location != nil {
		locationSource = string(input.Location.Source)
		locationName = input.Location.Name
		locationConsentAt = input.Location.ConsentAt
		latitude = input.Location.Latitude
		longitude = input.Location.Longitude
	}
	const query = `
		UPDATE users
		SET
			gender = COALESCE($1, gender),
			preferences = COALESCE($2, preferences),
			bio = COALESCE($3, bio),
			interests = COALESCE($4::jsonb, interests),
			birth_date = COALESCE($5::date, birth_date),
			location_source = CASE WHEN $6::boolean THEN $7::text ELSE location_source END,
			location_name = CASE WHEN $6::boolean THEN $8::text ELSE location_name END,
			location_consent_at = CASE WHEN $6::boolean THEN $9::timestamptz ELSE location_consent_at END,
			latitude = CASE WHEN $6::boolean THEN $10::double precision ELSE latitude END,
			longitude = CASE WHEN $6::boolean THEN $11::double precision ELSE longitude END,
			updated_at = now()
		WHERE id = $12
		RETURNING id
		`
	var updatedUserID uint
	err := repo.db.QueryRowContext(
		ctx,
		query,
		optionalString(input.Gender),
		optionalString(input.Preferences),
		optionalString(input.Bio),
		interestsValue,
		input.BirthDate,
		locationProvided,
		locationSource,
		locationName,
		locationConsentAt,
		latitude,
		longitude,
		userID).Scan(&updatedUserID)
	if errors.Is(err, sql.ErrNoRows) {
		return UserErrors.UserNotFound
	}
	return err
}
