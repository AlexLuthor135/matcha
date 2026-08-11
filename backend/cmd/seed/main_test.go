package main

import (
	"backend/models"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestBuildSeedProfileProducesValidLocationStates(t *testing.T) {
	now := time.Date(2026, time.August, 10, 12, 0, 0, 0, time.UTC)

	manual := buildSeedProfile(0, now)
	if manual.LocationSource != "manual" || manual.LocationName == "" || manual.LocationConsentAt != nil {
		t.Fatalf("manual location state = %+v", manual)
	}
	if manual.Latitude < -90 || manual.Latitude > 90 || manual.Longitude < -180 || manual.Longitude > 180 {
		t.Fatalf("manual coordinates are invalid: (%v, %v)", manual.Latitude, manual.Longitude)
	}
	if manual.Gender != models.GenderMale || manual.Preferences != models.PreferenceEveryone {
		t.Fatalf("first profile orientation = (%q, %q), want (Male, Everyone)", manual.Gender, manual.Preferences)
	}

	gps := buildSeedProfile(1, now)
	if gps.LocationSource != "gps" || gps.LocationName != "" || gps.LocationConsentAt == nil {
		t.Fatalf("GPS location state = %+v", gps)
	}
	if gps.UserName != "seed_0002" || gps.Email != "seed_0002@example.invalid" {
		t.Fatalf("deterministic identity = (%q, %q)", gps.UserName, gps.Email)
	}
	if len(gps.Interests) != 4 || len(gps.PhotoURLs) != 2 || gps.AvatarURL == "" {
		t.Fatalf("profile content is incomplete: %+v", gps)
	}
	if gps.Gender != models.GenderFemale || gps.Preferences != models.PreferenceMale {
		t.Fatalf("second profile orientation = (%q, %q), want (Female, Male)", gps.Gender, gps.Preferences)
	}

	other := buildSeedProfile(9, now)
	if other.Gender != models.GenderOther || other.Preferences != models.PreferenceEveryone {
		t.Fatalf("other profile orientation = (%q, %q), want (Other, Everyone)", other.Gender, other.Preferences)
	}
}

func TestBuildSeedDecisionsHasNoSelfOrDuplicatePairs(t *testing.T) {
	userIDs := make([]int64, 500)
	for index := range userIDs {
		userIDs[index] = int64(index + 1)
	}

	decisions := buildSeedDecisions(userIDs)
	if len(decisions) == 0 {
		t.Fatal("buildSeedDecisions() returned no decisions")
	}
	pairs := make(map[[2]int64]struct{}, len(decisions))
	for _, decision := range decisions {
		if decision.UserID == decision.TargetUserID {
			t.Fatalf("self decision found: %+v", decision)
		}
		if decision.Decision != "like" && decision.Decision != "dislike" {
			t.Fatalf("invalid decision value: %+v", decision)
		}
		pair := [2]int64{decision.UserID, decision.TargetUserID}
		if _, exists := pairs[pair]; exists {
			t.Fatalf("duplicate decision pair: %+v", pair)
		}
		pairs[pair] = struct{}{}
	}
	if _, exists := pairs[[2]int64{1, 2}]; !exists {
		t.Fatal("expected mutual-match seed pair 1 -> 2")
	}
	if _, exists := pairs[[2]int64{2, 1}]; !exists {
		t.Fatal("expected mutual-match seed pair 2 -> 1")
	}
}

func TestGenerateSeedAssetsCreatesReadablePNGs(t *testing.T) {
	directory := t.TempDir()
	if err := generateSeedAssets(directory); err != nil {
		t.Fatalf("generateSeedAssets(): %v", err)
	}

	paths := []string{
		filepath.Join(directory, "avatar-01.png"),
		filepath.Join(directory, "photo-12.png"),
	}
	for _, path := range paths {
		file, err := os.Open(path)
		if err != nil {
			t.Fatalf("open generated asset %q: %v", path, err)
		}
		_, decodeErr := png.Decode(file)
		closeErr := file.Close()
		if decodeErr != nil {
			t.Fatalf("decode generated asset %q: %v", path, decodeErr)
		}
		if closeErr != nil {
			t.Fatalf("close generated asset %q: %v", path, closeErr)
		}
	}
}

func TestLoadSeedConfigRequiresSafeValues(t *testing.T) {
	requiredEnvironment := map[string]string{
		"SQL_USER":     "postgres",
		"SQL_PASSWORD": "password",
		"SQL_HOST":     "db",
		"SQL_PORT":     "5432",
		"SQL_DATABASE": "matcha_db",
	}
	for name, value := range requiredEnvironment {
		t.Setenv(name, value)
	}
	t.Setenv("DATABASE_URL", "")
	t.Setenv("SEED_PROFILE_COUNT", "500")
	t.Setenv("SEED_PASSWORD", "")

	if _, err := loadSeedConfig(); err == nil || !strings.Contains(err.Error(), "SEED_PASSWORD") {
		t.Fatalf("missing password error = %v", err)
	}

	t.Setenv("SEED_PASSWORD", "SeedPassword9!")
	t.Setenv("SEED_PROFILE_COUNT", "499")
	if _, err := loadSeedConfig(); err == nil || !strings.Contains(err.Error(), "between 500 and 5000") {
		t.Fatalf("small profile count error = %v", err)
	}

	t.Setenv("SEED_PROFILE_COUNT", "500")
	config, err := loadSeedConfig()
	if err != nil {
		t.Fatalf("loadSeedConfig() unexpected error: %v", err)
	}
	if config.ProfileCount != 500 || config.Password != "SeedPassword9!" {
		t.Fatalf("config = %+v", config)
	}
	if !strings.Contains(config.DSN, "matcha_db") || !strings.Contains(config.DSN, "sslmode=disable") {
		t.Fatalf("DSN = %q", config.DSN)
	}
}
