package main

import (
	"backend/models"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"time"
)

const seedAssetCount = 12

var seedCities = []struct {
	Name      string
	Latitude  float64
	Longitude float64
}{
	{Name: "Berlin", Latitude: 52.5200, Longitude: 13.4050},
	{Name: "Paris", Latitude: 48.8566, Longitude: 2.3522},
	{Name: "London", Latitude: 51.5074, Longitude: -0.1278},
	{Name: "Madrid", Latitude: 40.4168, Longitude: -3.7038},
	{Name: "Rome", Latitude: 41.9028, Longitude: 12.4964},
}

var seedFirstNames = []string{
	"Alex", "Sam", "Jordan", "Taylor", "Morgan", "Casey",
	"Jamie", "Robin", "Cameron", "Drew", "Riley", "Avery",
}

var seedLastNames = []string{
	"Anderson", "Bennett", "Carter", "Diaz", "Evans", "Fischer",
	"Garcia", "Hughes", "Ivanov", "Johnson", "Klein", "Lopez",
}

var seedInterestPool = []string{
	"art", "books", "coding", "cooking", "cycling", "fitness",
	"gaming", "hiking", "movies", "music", "photography", "travel",
}

type seedProfile struct {
	UserName          string
	FirstName         string
	LastName          string
	Email             string
	BirthDate         time.Time
	LastSeenAt        time.Time
	Gender            string
	Preferences       string
	Bio               string
	Interests         []string
	LocationSource    string
	LocationName      string
	LocationConsentAt *time.Time
	Latitude          float64
	Longitude         float64
	AvatarURL         string
	PhotoURLs         []string
}

type seedDecision struct {
	UserID       int64
	TargetUserID int64
	Decision     string
}

type seedStats struct {
	Profiles  int
	Photos    int
	Decisions int
}

func seedDatabase(
	ctx context.Context,
	tx *sql.Tx,
	profileCount int,
	passwordHash string,
	now time.Time,
) (seedStats, error) {
	const userQuery = `
		INSERT INTO users (
			is_verified,
			last_seen_at,
			birth_date,
			location_source,
			location_name,
			location_consent_at,
			latitude,
			longitude,
			user_name,
			first_name,
			last_name,
			email,
			password,
			is_completed,
			gender,
			preferences,
			bio,
			interests,
			avatar
		)
		VALUES (
			true, $1, $2, $3, $4, $5, $6, $7, $8, $9,
			$10, $11, $12, true, $13, $14, $15, $16::jsonb, $17
		)
		ON CONFLICT (user_name)
		DO UPDATE SET
			is_verified = true,
			last_seen_at = EXCLUDED.last_seen_at,
			birth_date = EXCLUDED.birth_date,
			location_source = EXCLUDED.location_source,
			location_name = EXCLUDED.location_name,
			location_consent_at = EXCLUDED.location_consent_at,
			latitude = EXCLUDED.latitude,
			longitude = EXCLUDED.longitude,
			first_name = EXCLUDED.first_name,
			last_name = EXCLUDED.last_name,
			email = EXCLUDED.email,
			password = EXCLUDED.password,
			is_completed = true,
			gender = EXCLUDED.gender,
			preferences = EXCLUDED.preferences,
			bio = EXCLUDED.bio,
			interests = EXCLUDED.interests,
			avatar = EXCLUDED.avatar,
			updated_at = now()
		RETURNING id
	`
	userStatement, err := tx.PrepareContext(ctx, userQuery)
	if err != nil {
		return seedStats{}, fmt.Errorf("prepare seed user query: %w", err)
	}
	defer userStatement.Close()

	const deletePhotosQuery = `DELETE FROM photos WHERE user_id = $1`
	deletePhotosStatement, err := tx.PrepareContext(ctx, deletePhotosQuery)
	if err != nil {
		return seedStats{}, fmt.Errorf("prepare seed photo cleanup: %w", err)
	}
	defer deletePhotosStatement.Close()

	const photoQuery = `INSERT INTO photos (user_id, url) VALUES ($1, $2)`
	photoStatement, err := tx.PrepareContext(ctx, photoQuery)
	if err != nil {
		return seedStats{}, fmt.Errorf("prepare seed photo query: %w", err)
	}
	defer photoStatement.Close()

	userIDs := make([]int64, 0, profileCount)
	stats := seedStats{}
	for index := 0; index < profileCount; index++ {
		profile := buildSeedProfile(index, now)
		rawInterests, err := json.Marshal(profile.Interests)
		if err != nil {
			return seedStats{}, fmt.Errorf("encode interests for %s: %w", profile.UserName, err)
		}

		var userID int64
		err = userStatement.QueryRowContext(
			ctx,
			profile.LastSeenAt,
			profile.BirthDate,
			profile.LocationSource,
			profile.LocationName,
			profile.LocationConsentAt,
			profile.Latitude,
			profile.Longitude,
			profile.UserName,
			profile.FirstName,
			profile.LastName,
			profile.Email,
			passwordHash,
			profile.Gender,
			profile.Preferences,
			profile.Bio,
			string(rawInterests),
			profile.AvatarURL,
		).Scan(&userID)
		if err != nil {
			return seedStats{}, fmt.Errorf("save seed user %s: %w", profile.UserName, err)
		}
		userIDs = append(userIDs, userID)
		stats.Profiles++

		if _, err := deletePhotosStatement.ExecContext(ctx, userID); err != nil {
			return seedStats{}, fmt.Errorf("clear photos for %s: %w", profile.UserName, err)
		}
		for _, photoURL := range profile.PhotoURLs {
			if _, err := photoStatement.ExecContext(ctx, userID, photoURL); err != nil {
				return seedStats{}, fmt.Errorf("save photo for %s: %w", profile.UserName, err)
			}
			stats.Photos++
		}
	}

	const clearDecisionsQuery = `
		DELETE FROM profile_decisions
		WHERE user_id IN (
			SELECT id FROM users WHERE LEFT(user_name, 5) = 'seed_'
		)
		AND target_user_id IN (
			SELECT id FROM users WHERE LEFT(user_name, 5) = 'seed_'
		)
	`
	if _, err := tx.ExecContext(ctx, clearDecisionsQuery); err != nil {
		return seedStats{}, fmt.Errorf("clear seed decisions: %w", err)
	}

	const decisionQuery = `
		INSERT INTO profile_decisions (user_id, target_user_id, decision)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id, target_user_id)
		DO UPDATE SET decision = EXCLUDED.decision, updated_at = now()
	`
	decisionStatement, err := tx.PrepareContext(ctx, decisionQuery)
	if err != nil {
		return seedStats{}, fmt.Errorf("prepare seed decision query: %w", err)
	}
	defer decisionStatement.Close()

	for _, decision := range buildSeedDecisions(userIDs) {
		if _, err := decisionStatement.ExecContext(
			ctx,
			decision.UserID,
			decision.TargetUserID,
			decision.Decision,
		); err != nil {
			return seedStats{}, fmt.Errorf("save seed decision: %w", err)
		}
		stats.Decisions++
	}

	return stats, nil
}

func buildSeedProfile(index int, now time.Time) seedProfile {
	city := seedCities[index%len(seedCities)]
	latitudeOffset := float64(index%11-5) * 0.004
	longitudeOffset := float64((index*3)%11-5) * 0.004
	latitude := city.Latitude + latitudeOffset
	longitude := city.Longitude + longitudeOffset

	gender := models.GenderMale
	preferences := models.PreferenceFemale
	if index%2 == 1 {
		gender = models.GenderFemale
		preferences = models.PreferenceMale
	}
	if index%4 == 0 {
		preferences = models.PreferenceEveryone
	}
	if index%10 == 9 {
		gender = models.GenderOther
		preferences = models.PreferenceEveryone
	}

	age := 20 + index%31
	birthDate := time.Date(
		now.Year()-age,
		time.Month(index%12+1),
		index%27+1,
		0,
		0,
		0,
		0,
		time.UTC,
	)
	lastSeenAt := now.Add(-time.Duration(index%240) * time.Minute)

	locationSource := "gps"
	locationName := ""
	consentAt := now.Add(-time.Duration(index%365) * 24 * time.Hour)
	locationConsentAt := &consentAt
	if index%3 == 0 {
		locationSource = "manual"
		locationName = city.Name
		locationConsentAt = nil
	}

	interests := make([]string, 0, 4)
	for offset := 0; offset < 4; offset++ {
		interestIndex := (index + offset*3) % len(seedInterestPool)
		interests = append(interests, seedInterestPool[interestIndex])
	}

	assetNumber := index%seedAssetCount + 1
	secondAssetNumber := (index+5)%seedAssetCount + 1
	return seedProfile{
		UserName:          fmt.Sprintf("seed_%04d", index+1),
		FirstName:         seedFirstNames[index%len(seedFirstNames)],
		LastName:          seedLastNames[(index*5)%len(seedLastNames)],
		Email:             fmt.Sprintf("seed_%04d@example.invalid", index+1),
		BirthDate:         birthDate,
		LastSeenAt:        lastSeenAt,
		Gender:            gender,
		Preferences:       preferences,
		Bio:               fmt.Sprintf("Seed profile %d from %s", index+1, city.Name),
		Interests:         interests,
		LocationSource:    locationSource,
		LocationName:      locationName,
		LocationConsentAt: locationConsentAt,
		Latitude:          latitude,
		Longitude:         longitude,
		AvatarURL:         fmt.Sprintf("/uploads/seed/avatar-%02d.png", assetNumber),
		PhotoURLs: []string{
			fmt.Sprintf("/uploads/seed/photo-%02d.png", assetNumber),
			fmt.Sprintf("/uploads/seed/photo-%02d.png", secondAssetNumber),
		},
	}
}

func buildSeedDecisions(userIDs []int64) []seedDecision {
	decisions := make([]seedDecision, 0, len(userIDs)*4)
	decisionPositions := make(map[[2]int64]int)
	addDecision := func(userID int64, targetUserID int64, decision string) {
		if userID == targetUserID {
			return
		}
		key := [2]int64{userID, targetUserID}
		if position, exists := decisionPositions[key]; exists {
			decisions[position].Decision = decision
			return
		}
		decisionPositions[key] = len(decisions)
		decisions = append(decisions, seedDecision{
			UserID:       userID,
			TargetUserID: targetUserID,
			Decision:     decision,
		})
	}

	if len(userIDs) < 2 {
		return decisions
	}
	for index, userID := range userIDs {
		for _, offset := range []int{1, 7, 31} {
			targetID := userIDs[(index+offset)%len(userIDs)]
			addDecision(userID, targetID, "like")
		}
		if index%5 == 0 {
			targetID := userIDs[(index+17)%len(userIDs)]
			addDecision(userID, targetID, "dislike")
		}
	}
	for index := 0; index+1 < len(userIDs); index += 10 {
		addDecision(userIDs[index], userIDs[index+1], "like")
		addDecision(userIDs[index+1], userIDs[index], "like")
	}
	return decisions
}

func generateSeedAssets(directory string) error {
	if err := os.MkdirAll(directory, 0o755); err != nil {
		return err
	}
	for assetNumber := 1; assetNumber <= seedAssetCount; assetNumber++ {
		avatarPath := filepath.Join(
			directory,
			fmt.Sprintf("avatar-%02d.png", assetNumber),
		)
		if err := writeSeedPNG(avatarPath, assetNumber, 256, 256); err != nil {
			return err
		}

		photoPath := filepath.Join(
			directory,
			fmt.Sprintf("photo-%02d.png", assetNumber),
		)
		if err := writeSeedPNG(photoPath, assetNumber+seedAssetCount, 640, 480); err != nil {
			return err
		}
	}
	return nil
}

func writeSeedPNG(path string, variant int, width int, height int) error {
	if _, err := os.Stat(path); err == nil {
		return nil
	} else if !os.IsNotExist(err) {
		return err
	}

	imageData := image.NewRGBA(image.Rect(0, 0, width, height))
	baseRed := uint8((variant*47)%155 + 60)
	baseGreen := uint8((variant*83)%155 + 60)
	baseBlue := uint8((variant*29)%155 + 60)
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			shade := uint8((x + y + variant*17) % 48)
			imageData.SetRGBA(x, y, color.RGBA{
				R: baseRed - baseRed%48 + shade,
				G: baseGreen - baseGreen%48 + shade,
				B: baseBlue - baseBlue%48 + shade,
				A: 255,
			})
		}
	}

	file, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
	if err != nil {
		if os.IsExist(err) {
			return nil
		}
		return err
	}
	if err := png.Encode(file, imageData); err != nil {
		_ = file.Close()
		_ = os.Remove(path)
		return err
	}
	if err := file.Close(); err != nil {
		_ = os.Remove(path)
		return err
	}
	return nil
}
