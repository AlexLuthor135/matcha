package user

import (
	"backend/models"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func TestPostgresUserRepositoryIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping PostgreSQL integration test in short mode")
	}

	dsn, ok := userIntegrationDatabaseDSN()
	if !ok {
		t.Skip("set USER_TEST_DATABASE_DSN or SQL_* variables to run PostgreSQL integration tests")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatalf("open PostgreSQL: %v", err)
	}
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		t.Fatalf("ping PostgreSQL: %v", err)
	}

	schemaName := fmt.Sprintf("user_repository_test_%d", time.Now().UnixNano())
	if _, err := db.ExecContext(ctx, "CREATE SCHEMA "+schemaName); err != nil {
		_ = db.Close()
		t.Fatalf("create test schema: %v", err)
	}

	t.Cleanup(func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cleanupCancel()
		if _, err := db.ExecContext(cleanupCtx, "DROP SCHEMA "+schemaName+" CASCADE"); err != nil {
			t.Errorf("drop test schema: %v", err)
		}
		if err := db.Close(); err != nil {
			t.Errorf("close test database: %v", err)
		}
	})

	if _, err := db.ExecContext(ctx, "SET search_path TO "+schemaName); err != nil {
		t.Fatalf("set test search path: %v", err)
	}

	applyUserInitialMigration(t, ctx, db)

	ownerID := insertUserIntegrationUser(t, ctx, db, "photo-owner", "photo-owner@example.invalid")
	otherUserID := insertUserIntegrationUser(t, ctx, db, "other-user", "other-user@example.invalid")
	repository := NewPostgresRepository(db)
	t.Run("pending email lifecycle", func(t *testing.T) {
		testUserIntegrationPendingEmailLifecycle(t, ctx, db, repository)
	})
	t.Run("pending email reserves registration address", func(t *testing.T) {
		testUserIntegrationPendingEmailReservation(t, ctx, db, repository)
	})
	t.Run("verification reports occupied email", func(t *testing.T) {
		testUserIntegrationPendingEmailConflict(t, ctx, db, repository)
	})

	relationshipViewerID := insertUserIntegrationUser(
		t,
		ctx,
		db,
		"relationship-viewer",
		"relationship-viewer@example.invalid",
	)
	relationshipTargetID := insertUserIntegrationUser(
		t,
		ctx,
		db,
		"relationship-target",
		"relationship-target@example.invalid",
	)

	relationship, err := repository.GetProfileRelationship(
		ctx,
		relationshipViewerID,
		relationshipTargetID,
	)
	if err != nil {
		t.Fatalf("GetProfileRelationship() without likes: %v", err)
	}
	if relationship.LikedByMe || relationship.LikedMe || relationship.IsConnected() {
		t.Fatalf("relationship without likes = %+v, want all false", relationship)
	}

	if _, err := db.ExecContext(
		ctx,
		`INSERT INTO profile_decisions (user_id, target_user_id, decision) VALUES ($1, $2, 'like')`,
		relationshipViewerID,
		relationshipTargetID,
	); err != nil {
		t.Fatalf("insert viewer like for relationship test: %v", err)
	}

	relationship, err = repository.GetProfileRelationship(
		ctx,
		relationshipViewerID,
		relationshipTargetID,
	)
	if err != nil {
		t.Fatalf("GetProfileRelationship() with outgoing like: %v", err)
	}
	if !relationship.LikedByMe || relationship.LikedMe || relationship.IsConnected() {
		t.Fatalf("one-sided relationship = %+v, want liked_by_me only", relationship)
	}

	if _, err := db.ExecContext(
		ctx,
		`INSERT INTO profile_decisions (user_id, target_user_id, decision) VALUES ($1, $2, 'like')`,
		relationshipTargetID,
		relationshipViewerID,
	); err != nil {
		t.Fatalf("insert reverse like for relationship test: %v", err)
	}

	relationship, err = repository.GetProfileRelationship(
		ctx,
		relationshipViewerID,
		relationshipTargetID,
	)
	if err != nil {
		t.Fatalf("GetProfileRelationship() with mutual likes: %v", err)
	}
	if !relationship.LikedByMe || !relationship.LikedMe || !relationship.IsConnected() {
		t.Fatalf("mutual relationship = %+v, want connected", relationship)
	}

	passwordResetRawToken := "integration-password-reset-token"
	passwordResetTokenHash := hashAccountToken(passwordResetRawToken)
	if err := repository.ReplaceAccountToken(ctx, models.AccountToken{
		UserID:    ownerID,
		Hash:      passwordResetTokenHash,
		Purpose:   models.AccountTokenPurposePasswordReset,
		ExpiresAt: time.Now().UTC().Add(10 * time.Minute),
	}); err != nil {
		t.Fatalf("ReplaceAccountToken() password reset token: %v", err)
	}

	const newPasswordHash = "integration-new-password-hash"
	if err := repository.ResetPasswordWithToken(
		ctx,
		passwordResetTokenHash,
		newPasswordHash,
	); err != nil {
		t.Fatalf("ResetPasswordWithToken() first call: %v", err)
	}

	var savedPasswordHash string
	if err := db.QueryRowContext(
		ctx,
		`SELECT password FROM users WHERE id = $1`,
		ownerID,
	).Scan(&savedPasswordHash); err != nil {
		t.Fatalf("read password after ResetPasswordWithToken(): %v", err)
	}
	if savedPasswordHash != newPasswordHash {
		t.Fatalf("saved password hash = %q, want %q", savedPasswordHash, newPasswordHash)
	}

	var resetTokenUsedAt sql.NullTime
	if err := db.QueryRowContext(
		ctx,
		`SELECT used_at FROM account_tokens WHERE token_hash = $1`,
		passwordResetTokenHash,
	).Scan(&resetTokenUsedAt); err != nil {
		t.Fatalf("read password reset token used_at: %v", err)
	}
	if !resetTokenUsedAt.Valid {
		t.Fatal("password reset token used_at is NULL after password reset")
	}

	if err := repository.ResetPasswordWithToken(
		ctx,
		passwordResetTokenHash,
		"second-password-hash",
	); !errors.Is(err, UserErrors.InvalidPasswordResetToken) {
		t.Fatalf(
			"ResetPasswordWithToken() reused token error = %v, want %v",
			err,
			UserErrors.InvalidPasswordResetToken,
		)
	}

	latitude := 52.52
	longitude := 13.405
	locationConsentAt := time.Now().UTC()
	birthDate := time.Date(2000, time.May, 10, 0, 0, 0, 0, time.UTC)
	completed, err := repository.CompleteProfile(ctx, ownerID, CompleteProfileInput{
		Gender:      "Male",
		Preferences: "Female",
		Bio:         "Integration profile",
		Interests:   []string{"music", "travel"},
		BirthDate:   birthDate,
		Location: &LocationInput{
			Source:    models.LocationSourceGPS,
			Latitude:  &latitude,
			Longitude: &longitude,
			Consent:   true,
			ConsentAt: &locationConsentAt,
		},
	})
	if err != nil {
		t.Fatalf("CompleteProfile(): %v", err)
	}
	if !completed {
		t.Fatal("CompleteProfile() completed = false, want true")
	}
	var (
		savedBirthDate time.Time
		savedLatitude  float64
		savedLongitude float64
		savedSource    models.LocationSource
		savedName      string
		savedConsentAt sql.NullTime
	)
	if err := db.QueryRowContext(
		ctx,
		`SELECT birth_date, latitude, longitude, location_source, location_name, location_consent_at FROM users WHERE id = $1`,
		ownerID,
	).Scan(
		&savedBirthDate,
		&savedLatitude,
		&savedLongitude,
		&savedSource,
		&savedName,
		&savedConsentAt,
	); err != nil {
		t.Fatalf("read completed profile location: %v", err)
	}
	if !savedBirthDate.Equal(birthDate) {
		t.Fatalf("saved birth date = %v, want %v", savedBirthDate, birthDate)
	}
	if savedLatitude != latitude || savedLongitude != longitude {
		t.Fatalf(
			"saved location = (%v, %v), want (%v, %v)",
			savedLatitude,
			savedLongitude,
			latitude,
			longitude,
		)
	}
	if savedSource != models.LocationSourceGPS || savedName != "" || !savedConsentAt.Valid {
		t.Fatalf(
			"saved GPS metadata = (source=%q, name=%q, consent=%v)",
			savedSource,
			savedName,
			savedConsentAt,
		)
	}
	updatedBio := "Updated integration profile"
	if err := repository.UpdateProfile(ctx, ownerID, UpdateProfileInput{Bio: &updatedBio}); err != nil {
		t.Fatalf("UpdateProfile() bio only: %v", err)
	}
	profileAfterBioUpdate, err := repository.GetProfile(ctx, ownerID)
	if err != nil {
		t.Fatalf("GetProfile() after bio-only update: %v", err)
	}
	if profileAfterBioUpdate.Bio != updatedBio {
		t.Fatalf("bio after partial update = %q, want %q", profileAfterBioUpdate.Bio, updatedBio)
	}
	if profileAfterBioUpdate.BirthDate == nil || !profileAfterBioUpdate.BirthDate.Equal(birthDate) ||
		profileAfterBioUpdate.Latitude == nil || *profileAfterBioUpdate.Latitude != latitude ||
		profileAfterBioUpdate.Longitude == nil || *profileAfterBioUpdate.Longitude != longitude {
		t.Fatalf("bio-only update changed birth date or location: %+v", profileAfterBioUpdate)
	}
	if profileAfterBioUpdate.LocationSource != models.LocationSourceGPS ||
		profileAfterBioUpdate.LocationConsentAt == nil {
		t.Fatalf("bio-only update changed GPS metadata: %+v", profileAfterBioUpdate)
	}

	updatedBirthDate := time.Date(1999, time.June, 15, 0, 0, 0, 0, time.UTC)
	updatedLatitude := 48.8566
	updatedLongitude := 2.3522
	if err := repository.UpdateProfile(ctx, ownerID, UpdateProfileInput{
		BirthDate: &updatedBirthDate,
		Location: &LocationInput{
			Source:    models.LocationSourceManual,
			Name:      "Paris",
			Latitude:  &updatedLatitude,
			Longitude: &updatedLongitude,
		},
	}); err != nil {
		t.Fatalf("UpdateProfile() birth date and location: %v", err)
	}
	profileAfterLocationUpdate, err := repository.GetProfile(ctx, ownerID)
	if err != nil {
		t.Fatalf("GetProfile() after location update: %v", err)
	}
	if profileAfterLocationUpdate.BirthDate == nil || !profileAfterLocationUpdate.BirthDate.Equal(updatedBirthDate) {
		t.Fatalf("updated birth date = %v, want %v", profileAfterLocationUpdate.BirthDate, updatedBirthDate)
	}
	if profileAfterLocationUpdate.Latitude == nil || *profileAfterLocationUpdate.Latitude != updatedLatitude ||
		profileAfterLocationUpdate.Longitude == nil || *profileAfterLocationUpdate.Longitude != updatedLongitude {
		t.Fatalf(
			"updated location = (%v, %v), want (%v, %v)",
			profileAfterLocationUpdate.Latitude,
			profileAfterLocationUpdate.Longitude,
			updatedLatitude,
			updatedLongitude,
		)
	}
	if profileAfterLocationUpdate.LocationSource != models.LocationSourceManual ||
		profileAfterLocationUpdate.LocationName != "Paris" ||
		profileAfterLocationUpdate.LocationConsentAt != nil {
		t.Fatalf("updated manual location metadata = %+v", profileAfterLocationUpdate)
	}
	storedLatitude, storedLongitude, err := repository.GetUserLocation(ctx, ownerID)
	if err != nil {
		t.Fatalf("GetUserLocation(): %v", err)
	}
	if *storedLatitude != updatedLatitude || *storedLongitude != updatedLongitude {
		t.Fatalf(
			"GetUserLocation() = (%v, %v), want (%v, %v)",
			*storedLatitude,
			*storedLongitude,
			updatedLatitude,
			updatedLongitude,
		)
	}
	if _, _, err := repository.GetUserLocation(ctx, otherUserID); !errors.Is(err, UserErrors.InvalidLocation) {
		t.Fatalf("GetUserLocation() without coordinates error = %v, want %v", err, UserErrors.InvalidLocation)
	}
	if _, _, err := repository.GetUserLocation(ctx, ownerID+otherUserID+1_000_000); !errors.Is(err, UserErrors.UserNotFound) {
		t.Fatalf("missing user GetUserLocation() error = %v, want %v", err, UserErrors.UserNotFound)
	}

	wantLastSeenAt := time.Date(2026, time.August, 7, 13, 30, 0, 0, time.UTC)
	if err := repository.UpdateLastSeen(ctx, ownerID, wantLastSeenAt); err != nil {
		t.Fatalf("UpdateLastSeen(): %v", err)
	}
	ownerProfile, err := repository.GetProfile(ctx, ownerID)
	if err != nil {
		t.Fatalf("GetProfile() after UpdateLastSeen: %v", err)
	}
	if !ownerProfile.LastSeenAt.Equal(wantLastSeenAt) {
		t.Fatalf("last_seen_at = %v, want %v", ownerProfile.LastSeenAt, wantLastSeenAt)
	}
	if ownerProfile.BirthDate == nil || !ownerProfile.BirthDate.Equal(updatedBirthDate) {
		t.Fatalf("GetProfile() birth date = %v, want %v", ownerProfile.BirthDate, updatedBirthDate)
	}
	if ownerProfile.Latitude == nil || *ownerProfile.Latitude != updatedLatitude {
		t.Fatalf("GetProfile() latitude = %v, want %v", ownerProfile.Latitude, updatedLatitude)
	}
	if ownerProfile.Longitude == nil || *ownerProfile.Longitude != updatedLongitude {
		t.Fatalf("GetProfile() longitude = %v, want %v", ownerProfile.Longitude, updatedLongitude)
	}
	if err := repository.UpdateLastSeen(ctx, ownerID+otherUserID+1_000_000, wantLastSeenAt); !errors.Is(err, UserErrors.UserNotFound) {
		t.Fatalf("missing user UpdateLastSeen() error = %v, want %v", err, UserErrors.UserNotFound)
	}

	firstView, err := repository.SaveProfileView(ctx, ownerID, otherUserID)
	if err != nil {
		t.Fatalf("SaveProfileView() first call: %v", err)
	}
	if firstView.ID == 0 {
		t.Fatal("SaveProfileView() ID = 0")
	}
	if firstView.ViewerID != ownerID {
		t.Fatalf("SaveProfileView() ViewerID = %d, want %d", firstView.ViewerID, ownerID)
	}
	if firstView.ViewedUserID != otherUserID {
		t.Fatalf("SaveProfileView() ViewedUserID = %d, want %d", firstView.ViewedUserID, otherUserID)
	}

	time.Sleep(10 * time.Millisecond)

	secondView, err := repository.SaveProfileView(ctx, ownerID, otherUserID)
	if err != nil {
		t.Fatalf("SaveProfileView() second call: %v", err)
	}
	if secondView.ID != firstView.ID {
		t.Fatalf("repeated view ID = %d, want %d", secondView.ID, firstView.ID)
	}
	if !secondView.CreatedAt.Equal(firstView.CreatedAt) {
		t.Fatalf("repeated view created_at = %v, want %v", secondView.CreatedAt, firstView.CreatedAt)
	}
	if !secondView.UpdatedAt.After(firstView.UpdatedAt) {
		t.Fatalf("repeated view updated_at = %v, want after %v", secondView.UpdatedAt, firstView.UpdatedAt)
	}

	var viewCount int
	err = db.QueryRowContext(
		ctx,
		`SELECT count(*) FROM profile_views WHERE viewer_id = $1 AND viewed_user_id = $2`,
		ownerID,
		otherUserID,
	).Scan(&viewCount)
	if err != nil {
		t.Fatalf("count profile views: %v", err)
	}
	if viewCount != 1 {
		t.Fatalf("profile view count = %d, want 1", viewCount)
	}

	time.Sleep(10 * time.Millisecond)
	newestViewerID := insertUserIntegrationUser(
		t,
		ctx,
		db,
		"newest-viewer",
		"newest-viewer@example.invalid",
	)
	newestView, err := repository.SaveProfileView(ctx, newestViewerID, otherUserID)
	if err != nil {
		t.Fatalf("SaveProfileView() newest viewer: %v", err)
	}

	viewers, err := repository.ListProfileViewers(ctx, otherUserID)
	if err != nil {
		t.Fatalf("ListProfileViewers() for viewed user: %v", err)
	}
	if len(viewers) != 2 {
		t.Fatalf("ListProfileViewers() returned %d viewers, want 2", len(viewers))
	}
	if viewers[0].ID != newestViewerID || viewers[1].ID != ownerID {
		t.Fatalf(
			"ListProfileViewers() order = [%d, %d], want [%d, %d]",
			viewers[0].ID,
			viewers[1].ID,
			newestViewerID,
			ownerID,
		)
	}
	if viewers[0].UserName != "newest-viewer" {
		t.Fatalf("newest viewer username = %q, want %q", viewers[0].UserName, "newest-viewer")
	}
	if viewers[0].FirstName != "Integration" || viewers[0].LastName != "Test" {
		t.Fatalf(
			"newest viewer name = (%q, %q), want (%q, %q)",
			viewers[0].FirstName,
			viewers[0].LastName,
			"Integration",
			"Test",
		)
	}
	if !viewers[0].LastViewedAt.Equal(newestView.UpdatedAt) {
		t.Fatalf(
			"newest viewer last_viewed_at = %v, want %v",
			viewers[0].LastViewedAt,
			newestView.UpdatedAt,
		)
	}
	if !viewers[1].LastViewedAt.Equal(secondView.UpdatedAt) {
		t.Fatalf(
			"repeated viewer last_viewed_at = %v, want %v",
			viewers[1].LastViewedAt,
			secondView.UpdatedAt,
		)
	}

	if err := repository.BlockUser(ctx, otherUserID, ownerID); err != nil {
		t.Fatalf("BlockUser() viewed user to viewer: %v", err)
	}
	viewersAfterOwnerBlock, err := repository.ListProfileViewers(ctx, otherUserID)
	if err != nil {
		t.Fatalf("ListProfileViewers() after viewed-user block: %v", err)
	}
	if len(viewersAfterOwnerBlock) != 1 || viewersAfterOwnerBlock[0].ID != newestViewerID {
		t.Fatalf(
			"ListProfileViewers() after viewed-user block = %+v, want only user %d",
			viewersAfterOwnerBlock,
			newestViewerID,
		)
	}
	if err := repository.UnblockUser(ctx, otherUserID, ownerID); err != nil {
		t.Fatalf("UnblockUser() viewed user to viewer: %v", err)
	}

	if err := repository.BlockUser(ctx, ownerID, otherUserID); err != nil {
		t.Fatalf("BlockUser() viewer to viewed user: %v", err)
	}
	viewersAfterViewerBlock, err := repository.ListProfileViewers(ctx, otherUserID)
	if err != nil {
		t.Fatalf("ListProfileViewers() after viewer block: %v", err)
	}
	if len(viewersAfterViewerBlock) != 1 || viewersAfterViewerBlock[0].ID != newestViewerID {
		t.Fatalf(
			"ListProfileViewers() after viewer block = %+v, want only user %d",
			viewersAfterViewerBlock,
			newestViewerID,
		)
	}
	if err := repository.UnblockUser(ctx, ownerID, otherUserID); err != nil {
		t.Fatalf("UnblockUser() viewer to viewed user: %v", err)
	}

	emptyViewers, err := repository.ListProfileViewers(ctx, ownerID)
	if err != nil {
		t.Fatalf("ListProfileViewers() empty result: %v", err)
	}
	if emptyViewers == nil || len(emptyViewers) != 0 {
		t.Fatalf("ListProfileViewers() empty result = %#v, want non-nil empty slice", emptyViewers)
	}

	_, err = repository.SaveProfileView(ctx, ownerID, ownerID)
	if err == nil {
		t.Fatal("SaveProfileView() own profile error = nil")
	}

	firstURLs := []string{
		"/uploads/photos/first.png",
		"/uploads/photos/second.png",
	}
	photos, err := repository.CreatePhotos(ctx, ownerID, firstURLs, 5)
	if err != nil {
		t.Fatalf("CreatePhotos() first call: %v", err)
	}
	if len(photos) != len(firstURLs) {
		t.Fatalf("CreatePhotos() returned %d photos, want %d", len(photos), len(firstURLs))
	}
	for index, photo := range photos {
		if photo.ID == 0 || photo.UserID != ownerID || photo.URL != firstURLs[index] {
			t.Fatalf("CreatePhotos() photo %d = %+v", index, photo)
		}
	}

	if _, err := db.ExecContext(
		ctx,
		`UPDATE users SET is_completed = true, interests = '["Go", "PostgreSQL"]'::jsonb WHERE id = $1`,
		ownerID,
	); err != nil {
		t.Fatalf("prepare completed profile: %v", err)
	}
	profile, err := repository.GetProfile(ctx, ownerID)
	if err != nil {
		t.Fatalf("GetProfile() completed user: %v", err)
	}
	if profile.ID != ownerID || !profile.IsCompleted {
		t.Fatalf("GetProfile() identity/completion = (%d, %t), want (%d, true)", profile.ID, profile.IsCompleted, ownerID)
	}
	if len(profile.Interests) != 2 || profile.Interests[0] != "Go" || profile.Interests[1] != "PostgreSQL" {
		t.Fatalf("GetProfile() interests = %v", profile.Interests)
	}
	if len(profile.Photos) != len(firstURLs) {
		t.Fatalf("GetProfile() photo count = %d, want %d", len(profile.Photos), len(firstURLs))
	}
	for index, photo := range profile.Photos {
		if photo.ID != photos[index].ID || photo.UserID != ownerID || photo.URL != firstURLs[index] {
			t.Fatalf("GetProfile() photo %d = %+v", index, photo)
		}
	}

	_, err = repository.CreatePhotos(ctx, ownerID, []string{
		"/uploads/photos/third.png",
		"/uploads/photos/fourth.png",
		"/uploads/photos/fifth.png",
	}, 5)
	if err != nil {
		t.Fatalf("CreatePhotos() fill to limit: %v", err)
	}

	_, err = repository.CreatePhotos(ctx, ownerID, []string{"/uploads/photos/sixth.png"}, 5)
	if !errors.Is(err, UserErrors.PhotoLimitExceeded) {
		t.Fatalf("CreatePhotos() over limit error = %v, want %v", err, UserErrors.PhotoLimitExceeded)
	}
	if got := countUserIntegrationPhotos(t, ctx, db, ownerID); got != 5 {
		t.Fatalf("photo count after rejected insert = %d, want 5", got)
	}

	_, err = repository.DeletePhoto(ctx, otherUserID, photos[0].ID)
	if !errors.Is(err, UserErrors.PhotoNotFound) {
		t.Fatalf("other user DeletePhoto() error = %v, want %v", err, UserErrors.PhotoNotFound)
	}

	deletedURL, err := repository.DeletePhoto(ctx, ownerID, photos[0].ID)
	if err != nil {
		t.Fatalf("owner DeletePhoto() unexpected error: %v", err)
	}
	if deletedURL != firstURLs[0] {
		t.Fatalf("DeletePhoto() URL = %q, want %q", deletedURL, firstURLs[0])
	}

	_, err = repository.DeletePhoto(ctx, ownerID, photos[0].ID)
	if !errors.Is(err, UserErrors.PhotoNotFound) {
		t.Fatalf("second DeletePhoto() error = %v, want %v", err, UserErrors.PhotoNotFound)
	}

	missingUserID := ownerID + otherUserID + 1_000_000
	_, err = repository.CreatePhotos(ctx, missingUserID, []string{"/uploads/photos/missing.png"}, 5)
	if !errors.Is(err, UserErrors.UserNotFound) {
		t.Fatalf("missing user CreatePhotos() error = %v, want %v", err, UserErrors.UserNotFound)
	}
	_, err = repository.DeletePhoto(ctx, missingUserID, photos[1].ID)
	if !errors.Is(err, UserErrors.UserNotFound) {
		t.Fatalf("missing user DeletePhoto() error = %v, want %v", err, UserErrors.UserNotFound)
	}

	rollbackUserID := insertUserIntegrationUser(t, ctx, db, "rollback-user", "rollback-user@example.invalid")
	_, err = repository.CreatePhotos(ctx, rollbackUserID, []string{"/uploads/photos/valid.png", ""}, 5)
	if err == nil {
		t.Fatal("CreatePhotos() with blank URL error = nil, want database constraint error")
	}
	if got := countUserIntegrationPhotos(t, ctx, db, rollbackUserID); got != 0 {
		t.Fatalf("photo count after rolled-back insert = %d, want 0", got)
	}

	decisionUserID := insertUserIntegrationUser(t, ctx, db, "decision-user", "decision-user@example.invalid")
	decisionTargetID := insertUserIntegrationUser(t, ctx, db, "decision-target", "decision-target@example.invalid")
	decisionTargetBirthDate := time.Date(1998, time.September, 12, 0, 0, 0, 0, time.UTC)
	if _, err := db.ExecContext(
		ctx,
		`UPDATE users SET gender = 'Male', preferences = 'Female', avatar = '/uploads/avatars/decision-user.jpg' WHERE id = $1`,
		decisionUserID,
	); err != nil {
		t.Fatalf("prepare decision user: %v", err)
	}
	if _, err := db.ExecContext(
		ctx,
		`UPDATE users SET is_completed = true, gender = 'Female', preferences = 'Male', bio = 'Test profile', avatar = '/uploads/avatars/decision-target.jpg', birth_date = $2 WHERE id = $1`,
		decisionTargetID,
		decisionTargetBirthDate,
	); err != nil {
		t.Fatalf("prepare decision target: %v", err)
	}

	feedBeforeDecision, err := listUserIntegrationCandidates(ctx, repository, decisionUserID)
	if err != nil {
		t.Fatalf("GetProfileFeed() before decision: %v", err)
	}
	feedDecisionTarget, exists := findUserIntegrationProfile(feedBeforeDecision, decisionTargetID)
	if !exists {
		t.Fatalf("GetProfileFeed() before decision does not contain target user %d", decisionTargetID)
	}
	if feedDecisionTarget.BirthDate == nil || !feedDecisionTarget.BirthDate.Equal(decisionTargetBirthDate) {
		t.Fatalf(
			"GetProfileFeed() birth date = %v, want %v",
			feedDecisionTarget.BirthDate,
			decisionTargetBirthDate,
		)
	}

	dislikeResult, err := repository.SaveProfileDecision(
		ctx,
		decisionUserID,
		decisionTargetID,
		models.ProfileDecisionDislike,
	)
	if err != nil {
		t.Fatalf("SaveProfileDecision() dislike: %v", err)
	}
	if dislikeResult.IsMatch {
		t.Fatal("SaveProfileDecision() dislike isMatch = true, want false")
	}
	if !dislikeResult.DecisionChanged {
		t.Fatal("SaveProfileDecision() first dislike decisionChanged = false, want true")
	}
	if dislikeResult.MatchEnded {
		t.Fatal("SaveProfileDecision() first dislike matchEnded = true, want false")
	}
	dislike := dislikeResult.ProfileDecision
	if dislike.ID == 0 || dislike.Decision != models.ProfileDecisionDislike {
		t.Fatalf("SaveProfileDecision() dislike = %+v", dislike)
	}
	dislikeLikers, err := repository.ListProfileLikers(ctx, decisionTargetID)
	if err != nil {
		t.Fatalf("ListProfileLikers() after dislike: %v", err)
	}
	if len(dislikeLikers) != 0 {
		t.Fatalf("ListProfileLikers() after dislike = %+v, want empty", dislikeLikers)
	}
	feedAfterDecision, err := listUserIntegrationCandidates(ctx, repository, decisionUserID)
	if err != nil {
		t.Fatalf("GetProfileFeed() after decision: %v", err)
	}
	if containsUserIntegrationProfile(feedAfterDecision, decisionTargetID) {
		t.Fatalf("GetProfileFeed() after decision still contains target user %d", decisionTargetID)
	}
	searchAfterDecision, err := listUserIntegrationSearchCandidates(ctx, repository, decisionUserID)
	if err != nil {
		t.Fatalf("SearchProfiles() after decision: %v", err)
	}
	if !containsUserIntegrationProfile(searchAfterDecision, decisionTargetID) {
		t.Fatalf("SearchProfiles() after decision does not contain target user %d", decisionTargetID)
	}

	likeResult, err := repository.SaveProfileDecision(
		ctx,
		decisionUserID,
		decisionTargetID,
		models.ProfileDecisionLike,
	)
	if err != nil {
		t.Fatalf("SaveProfileDecision() update to like: %v", err)
	}
	if likeResult.IsMatch {
		t.Fatal("SaveProfileDecision() one-sided like isMatch = true, want false")
	}
	if !likeResult.DecisionChanged {
		t.Fatal("SaveProfileDecision() dislike-to-like decisionChanged = false, want true")
	}
	if likeResult.MatchEnded {
		t.Fatal("SaveProfileDecision() dislike-to-like matchEnded = true, want false")
	}
	like := likeResult.ProfileDecision
	if like.ID != dislike.ID {
		t.Fatalf("updated decision ID = %d, want original %d", like.ID, dislike.ID)
	}
	if like.Decision != models.ProfileDecisionLike {
		t.Fatalf("updated decision = %q, want %q", like.Decision, models.ProfileDecisionLike)
	}
	likersAfterLike, err := repository.ListProfileLikers(ctx, decisionTargetID)
	if err != nil {
		t.Fatalf("ListProfileLikers() after like: %v", err)
	}
	if len(likersAfterLike) != 1 || likersAfterLike[0].ID != decisionUserID {
		t.Fatalf(
			"ListProfileLikers() after like = %+v, want user %d",
			likersAfterLike,
			decisionUserID,
		)
	}
	if likersAfterLike[0].UserName != "decision-user" {
		t.Fatalf(
			"ListProfileLikers() username = %q, want %q",
			likersAfterLike[0].UserName,
			"decision-user",
		)
	}
	if !likersAfterLike[0].LikedAt.Equal(like.UpdatedAt) {
		t.Fatalf(
			"ListProfileLikers() liked_at = %v, want %v",
			likersAfterLike[0].LikedAt,
			like.UpdatedAt,
		)
	}
	profileAfterFirstLike, err := repository.GetProfile(ctx, decisionTargetID)
	if err != nil {
		t.Fatalf("GetProfile() after first incoming like: %v", err)
	}
	if profileAfterFirstLike.FameRating != 1 {
		t.Fatalf("fame rating after first incoming like = %d, want 1", profileAfterFirstLike.FameRating)
	}

	secondFameLikerID := insertUserIntegrationUser(
		t,
		ctx,
		db,
		"second-fame-liker",
		"second-fame-liker@example.invalid",
	)
	if _, err := repository.SaveProfileDecision(
		ctx,
		secondFameLikerID,
		decisionTargetID,
		models.ProfileDecisionLike,
	); err != nil {
		t.Fatalf("SaveProfileDecision() second fame like: %v", err)
	}
	profileAfterSecondLike, err := repository.GetProfile(ctx, decisionTargetID)
	if err != nil {
		t.Fatalf("GetProfile() after second incoming like: %v", err)
	}
	if profileAfterSecondLike.FameRating != 2 {
		t.Fatalf("fame rating after second incoming like = %d, want 2", profileAfterSecondLike.FameRating)
	}
	feedWithTwoLikes, err := listUserIntegrationCandidates(ctx, repository, ownerID)
	if err != nil {
		t.Fatalf("GetProfileFeed() with two incoming likes: %v", err)
	}
	feedTarget, exists := findUserIntegrationProfile(feedWithTwoLikes, decisionTargetID)
	if !exists {
		t.Fatalf("GetProfileFeed() does not contain fame target user %d", decisionTargetID)
	}
	if feedTarget.FameRating != 2 {
		t.Fatalf("feed fame rating after second incoming like = %d, want 2", feedTarget.FameRating)
	}
	blockDirections := []struct {
		name          string
		blockerID     uint
		blockedUserID uint
	}{
		{
			name:          "profile owner blocks liker",
			blockerID:     decisionTargetID,
			blockedUserID: decisionUserID,
		},
		{
			name:          "liker blocks profile owner",
			blockerID:     decisionUserID,
			blockedUserID: decisionTargetID,
		},
	}
	for _, direction := range blockDirections {
		if _, err := db.ExecContext(
			ctx,
			`INSERT INTO user_blocks (blocker_id, blocked_user_id) VALUES ($1, $2)`,
			direction.blockerID,
			direction.blockedUserID,
		); err != nil {
			t.Fatalf("insert block for %s: %v", direction.name, err)
		}
		blockedLikers, err := repository.ListProfileLikers(ctx, decisionTargetID)
		if err != nil {
			t.Fatalf("ListProfileLikers() for %s: %v", direction.name, err)
		}
		if len(blockedLikers) != 1 || blockedLikers[0].ID != secondFameLikerID {
			t.Fatalf(
				"ListProfileLikers() for %s = %+v, want only user %d",
				direction.name,
				blockedLikers,
				secondFameLikerID,
			)
		}
		if _, err := db.ExecContext(
			ctx,
			`DELETE FROM user_blocks WHERE blocker_id = $1 AND blocked_user_id = $2`,
			direction.blockerID,
			direction.blockedUserID,
		); err != nil {
			t.Fatalf("delete block for %s: %v", direction.name, err)
		}
	}
	restoredLikers, err := repository.ListProfileLikers(ctx, decisionTargetID)
	if err != nil {
		t.Fatalf("ListProfileLikers() after block removal: %v", err)
	}
	if len(restoredLikers) != 2 ||
		!containsUserIntegrationLiker(restoredLikers, decisionUserID) ||
		!containsUserIntegrationLiker(restoredLikers, secondFameLikerID) {
		t.Fatalf(
			"ListProfileLikers() after block removal = %+v, want users %d and %d",
			restoredLikers,
			decisionUserID,
			secondFameLikerID,
		)
	}
	repeatedLikeResult, err := repository.SaveProfileDecision(
		ctx,
		decisionUserID,
		decisionTargetID,
		models.ProfileDecisionLike,
	)
	if err != nil {
		t.Fatalf("SaveProfileDecision() repeated like: %v", err)
	}
	repeatedLike := repeatedLikeResult.ProfileDecision
	if repeatedLike.ID != like.ID {
		t.Fatalf("repeated like ID = %d, want %d", repeatedLike.ID, like.ID)
	}
	if repeatedLikeResult.IsMatch {
		t.Fatal("SaveProfileDecision() repeated one-sided like isMatch = true, want false")
	}
	if repeatedLikeResult.DecisionChanged {
		t.Fatal("SaveProfileDecision() repeated like decisionChanged = true, want false")
	}
	if repeatedLikeResult.MatchEnded {
		t.Fatal("SaveProfileDecision() repeated like matchEnded = true, want false")
	}
	if got := countUserIntegrationDecisions(t, ctx, db, decisionUserID, decisionTargetID); got != 1 {
		t.Fatalf("decision row count after upsert = %d, want 1", got)
	}
	oneSidedMatches, err := repository.ListMatches(ctx, decisionUserID)
	if err != nil {
		t.Fatalf("ListMatches() after one-sided like: %v", err)
	}
	if _, exists := findUserIntegrationMatch(oneSidedMatches, decisionTargetID); exists {
		t.Fatalf("ListMatches() after one-sided like contains user %d", decisionTargetID)
	}

	reverseLikeResult, err := repository.SaveProfileDecision(
		ctx,
		decisionTargetID,
		decisionUserID,
		models.ProfileDecisionLike,
	)
	if err != nil {
		t.Fatalf("SaveProfileDecision() reverse like: %v", err)
	}
	if !reverseLikeResult.IsMatch {
		t.Fatal("SaveProfileDecision() reverse like isMatch = false, want true")
	}
	if !reverseLikeResult.DecisionChanged {
		t.Fatal("SaveProfileDecision() reverse like decisionChanged = false, want true")
	}
	if reverseLikeResult.MatchEnded {
		t.Fatal("SaveProfileDecision() reverse like matchEnded = true, want false")
	}
	reverseLike := reverseLikeResult.ProfileDecision
	if reverseLike.Decision != models.ProfileDecisionLike {
		t.Fatalf("reverse decision = %q, want %q", reverseLike.Decision, models.ProfileDecisionLike)
	}
	reverseLikers, err := repository.ListProfileLikers(ctx, decisionUserID)
	if err != nil {
		t.Fatalf("ListProfileLikers() after reverse like: %v", err)
	}
	if len(reverseLikers) != 1 || reverseLikers[0].ID != decisionTargetID {
		t.Fatalf(
			"ListProfileLikers() after reverse like = %+v, want user %d",
			reverseLikers,
			decisionTargetID,
		)
	}

	userMatches, err := repository.ListMatches(ctx, decisionUserID)
	if err != nil {
		t.Fatalf("ListMatches() for first matched user: %v", err)
	}
	targetMatch, exists := findUserIntegrationMatch(userMatches, decisionTargetID)
	if !exists {
		t.Fatalf("ListMatches() does not contain matched user %d", decisionTargetID)
	}
	if targetMatch.UserName != "decision-target" {
		t.Fatalf("matched username = %q, want %q", targetMatch.UserName, "decision-target")
	}
	if targetMatch.FameRating != 2 {
		t.Fatalf("matched fame rating = %d, want 2", targetMatch.FameRating)
	}

	targetMatches, err := repository.ListMatches(ctx, decisionTargetID)
	if err != nil {
		t.Fatalf("ListMatches() for second matched user: %v", err)
	}
	if _, exists := findUserIntegrationMatch(targetMatches, decisionUserID); !exists {
		t.Fatalf("ListMatches() does not contain reverse matched user %d", decisionUserID)
	}

	if _, err := repository.SaveProfileDecision(
		ctx,
		secondFameLikerID,
		decisionTargetID,
		models.ProfileDecisionDislike,
	); err != nil {
		t.Fatalf("SaveProfileDecision() second liker unlike: %v", err)
	}
	profileAfterUnlike, err := repository.GetProfile(ctx, decisionTargetID)
	if err != nil {
		t.Fatalf("GetProfile() after unlike: %v", err)
	}
	if profileAfterUnlike.FameRating != 1 {
		t.Fatalf("fame rating after unlike = %d, want 1", profileAfterUnlike.FameRating)
	}
	feedAfterUnlike, err := listUserIntegrationCandidates(ctx, repository, ownerID)
	if err != nil {
		t.Fatalf("GetProfileFeed() after unlike: %v", err)
	}
	feedTarget, exists = findUserIntegrationProfile(feedAfterUnlike, decisionTargetID)
	if !exists {
		t.Fatalf("GetProfileFeed() after unlike does not contain user %d", decisionTargetID)
	}
	if feedTarget.FameRating != 1 {
		t.Fatalf("feed fame rating after unlike = %d, want 1", feedTarget.FameRating)
	}
	userMatches, err = repository.ListMatches(ctx, decisionUserID)
	if err != nil {
		t.Fatalf("ListMatches() after unrelated unlike: %v", err)
	}
	targetMatch, exists = findUserIntegrationMatch(userMatches, decisionTargetID)
	if !exists {
		t.Fatalf("ListMatches() after unrelated unlike does not contain user %d", decisionTargetID)
	}
	if targetMatch.FameRating != 1 {
		t.Fatalf("matched fame rating after unlike = %d, want 1", targetMatch.FameRating)
	}

	reverseDislikeResult, err := repository.SaveProfileDecision(
		ctx,
		decisionTargetID,
		decisionUserID,
		models.ProfileDecisionDislike,
	)
	if err != nil {
		t.Fatalf("SaveProfileDecision() reverse dislike: %v", err)
	}
	if reverseDislikeResult.IsMatch {
		t.Fatal("SaveProfileDecision() reverse dislike isMatch = true, want false")
	}
	if !reverseDislikeResult.DecisionChanged {
		t.Fatal("SaveProfileDecision() like-to-dislike decisionChanged = false, want true")
	}
	if !reverseDislikeResult.MatchEnded {
		t.Fatal("SaveProfileDecision() like-to-dislike matchEnded = false, want true")
	}
	reverseLikers, err = repository.ListProfileLikers(ctx, decisionUserID)
	if err != nil {
		t.Fatalf("ListProfileLikers() after reverse dislike: %v", err)
	}
	if len(reverseLikers) != 0 {
		t.Fatalf(
			"ListProfileLikers() after reverse dislike = %+v, want empty",
			reverseLikers,
		)
	}
	userMatches, err = repository.ListMatches(ctx, decisionUserID)
	if err != nil {
		t.Fatalf("ListMatches() after reverse dislike: %v", err)
	}
	if _, exists := findUserIntegrationMatch(userMatches, decisionTargetID); exists {
		t.Fatalf("ListMatches() after reverse dislike still contains user %d", decisionTargetID)
	}

	_, err = repository.SaveProfileDecision(
		ctx,
		decisionUserID,
		decisionTargetID+1_000_000,
		models.ProfileDecisionLike,
	)
	if !errors.Is(err, UserErrors.TargetUserNotFound) {
		t.Fatalf("missing target SaveProfileDecision() error = %v, want %v", err, UserErrors.TargetUserNotFound)
	}

	blockerID := insertUserIntegrationUser(
		t,
		ctx,
		db,
		"blocker-user",
		"blocker-user@example.invalid",
	)
	blockedUserID := insertUserIntegrationUser(
		t,
		ctx,
		db,
		"blocked-user",
		"blocked-user@example.invalid",
	)
	if _, err := db.ExecContext(
		ctx,
		`UPDATE users SET gender = 'Male', preferences = 'Female' WHERE id = $1`,
		blockerID,
	); err != nil {
		t.Fatalf("prepare blocker profile: %v", err)
	}
	if _, err := db.ExecContext(
		ctx,
		`UPDATE users SET is_completed = true, gender = 'Female', preferences = 'Male', bio = 'Block target' WHERE id = $1`,
		blockedUserID,
	); err != nil {
		t.Fatalf("prepare blocked profile: %v", err)
	}

	feedBeforeBlock, err := listUserIntegrationCandidates(ctx, repository, blockerID)
	if err != nil {
		t.Fatalf("GetProfileFeed() before block: %v", err)
	}
	if !containsUserIntegrationProfile(feedBeforeBlock, blockedUserID) {
		t.Fatalf("GetProfileFeed() before block does not contain user %d", blockedUserID)
	}
	blockExists, err := repository.HasBlockBetweenUsers(ctx, blockerID, blockedUserID)
	if err != nil {
		t.Fatalf("HasBlockBetweenUsers() before block: %v", err)
	}
	if blockExists {
		t.Fatal("HasBlockBetweenUsers() before block = true, want false")
	}

	if err := repository.BlockUser(ctx, blockerID, blockedUserID); err != nil {
		t.Fatalf("BlockUser() before forward feed check: %v", err)
	}
	blockExists, err = repository.HasBlockBetweenUsers(ctx, blockerID, blockedUserID)
	if err != nil {
		t.Fatalf("HasBlockBetweenUsers() after forward block: %v", err)
	}
	if !blockExists {
		t.Fatal("HasBlockBetweenUsers() after forward block = false, want true")
	}
	reverseBlockExists, err := repository.HasBlockBetweenUsers(ctx, blockedUserID, blockerID)
	if err != nil {
		t.Fatalf("HasBlockBetweenUsers() with reversed arguments: %v", err)
	}
	if !reverseBlockExists {
		t.Fatal("HasBlockBetweenUsers() with reversed arguments = false, want true")
	}
	blockedDecisionAttempts := []struct {
		name         string
		userID       uint
		targetUserID uint
		decision     models.ProfileDecisionValue
	}{
		{
			name:         "blocker like",
			userID:       blockerID,
			targetUserID: blockedUserID,
			decision:     models.ProfileDecisionLike,
		},
		{
			name:         "blocker dislike",
			userID:       blockerID,
			targetUserID: blockedUserID,
			decision:     models.ProfileDecisionDislike,
		},
		{
			name:         "blocked user like",
			userID:       blockedUserID,
			targetUserID: blockerID,
			decision:     models.ProfileDecisionLike,
		},
		{
			name:         "blocked user dislike",
			userID:       blockedUserID,
			targetUserID: blockerID,
			decision:     models.ProfileDecisionDislike,
		},
	}
	for _, attempt := range blockedDecisionAttempts {
		_, err := repository.SaveProfileDecision(
			ctx,
			attempt.userID,
			attempt.targetUserID,
			attempt.decision,
		)
		if !errors.Is(err, UserErrors.TargetUserNotFound) {
			t.Fatalf(
				"SaveProfileDecision() %s error = %v, want %v",
				attempt.name,
				err,
				UserErrors.TargetUserNotFound,
			)
		}
	}
	if got := countUserIntegrationDecisions(t, ctx, db, blockerID, blockedUserID); got != 0 {
		t.Fatalf("blocked forward decision count = %d, want 0", got)
	}
	if got := countUserIntegrationDecisions(t, ctx, db, blockedUserID, blockerID); got != 0 {
		t.Fatalf("blocked reverse decision count = %d, want 0", got)
	}
	feedAfterForwardBlock, err := listUserIntegrationCandidates(ctx, repository, blockerID)
	if err != nil {
		t.Fatalf("GetProfileFeed() after forward block: %v", err)
	}
	if containsUserIntegrationProfile(feedAfterForwardBlock, blockedUserID) {
		t.Fatalf("GetProfileFeed() after forward block contains user %d", blockedUserID)
	}

	if err := repository.UnblockUser(ctx, blockerID, blockedUserID); err != nil {
		t.Fatalf("UnblockUser() before reverse feed check: %v", err)
	}
	blockExists, err = repository.HasBlockBetweenUsers(ctx, blockerID, blockedUserID)
	if err != nil {
		t.Fatalf("HasBlockBetweenUsers() after unblock: %v", err)
	}
	if blockExists {
		t.Fatal("HasBlockBetweenUsers() after unblock = true, want false")
	}
	feedAfterUnblock, err := listUserIntegrationCandidates(ctx, repository, blockerID)
	if err != nil {
		t.Fatalf("GetProfileFeed() after unblock: %v", err)
	}
	if !containsUserIntegrationProfile(feedAfterUnblock, blockedUserID) {
		t.Fatalf("GetProfileFeed() after unblock does not contain user %d", blockedUserID)
	}

	if err := repository.BlockUser(ctx, blockedUserID, blockerID); err != nil {
		t.Fatalf("reverse BlockUser() before feed check: %v", err)
	}
	feedAfterReverseBlock, err := listUserIntegrationCandidates(ctx, repository, blockerID)
	if err != nil {
		t.Fatalf("GetProfileFeed() after reverse block: %v", err)
	}
	if containsUserIntegrationProfile(feedAfterReverseBlock, blockedUserID) {
		t.Fatalf("GetProfileFeed() after reverse block contains user %d", blockedUserID)
	}
	if err := repository.UnblockUser(ctx, blockedUserID, blockerID); err != nil {
		t.Fatalf("reverse UnblockUser() after feed check: %v", err)
	}

	if _, err := repository.SaveProfileDecision(
		ctx,
		blockerID,
		blockedUserID,
		models.ProfileDecisionLike,
	); err != nil {
		t.Fatalf("prepare blocker decision: %v", err)
	}
	if _, err := repository.SaveProfileDecision(
		ctx,
		blockedUserID,
		blockerID,
		models.ProfileDecisionLike,
	); err != nil {
		t.Fatalf("prepare blocked user decision: %v", err)
	}

	if err := repository.BlockUser(ctx, blockerID, blockedUserID); err != nil {
		t.Fatalf("BlockUser() unexpected error: %v", err)
	}
	if got := countUserIntegrationBlocks(t, ctx, db, blockerID, blockedUserID); got != 1 {
		t.Fatalf("forward block count = %d, want 1", got)
	}
	if got := countUserIntegrationBlocks(t, ctx, db, blockedUserID, blockerID); got != 0 {
		t.Fatalf("reverse block count = %d, want 0", got)
	}
	if got := countUserIntegrationDecisions(t, ctx, db, blockerID, blockedUserID); got != 0 {
		t.Fatalf("blocker decision count after block = %d, want 0", got)
	}
	if got := countUserIntegrationDecisions(t, ctx, db, blockedUserID, blockerID); got != 0 {
		t.Fatalf("blocked user decision count after block = %d, want 0", got)
	}

	if err := repository.BlockUser(ctx, blockerID, blockedUserID); err != nil {
		t.Fatalf("repeated BlockUser() unexpected error: %v", err)
	}
	if got := countUserIntegrationBlocks(t, ctx, db, blockerID, blockedUserID); got != 1 {
		t.Fatalf("block count after repeated block = %d, want 1", got)
	}

	missingBlockUserID := blockedUserID + 1_000_000
	if err := repository.BlockUser(ctx, missingBlockUserID, blockedUserID); !errors.Is(err, UserErrors.UserNotFound) {
		t.Fatalf("missing blocker BlockUser() error = %v, want %v", err, UserErrors.UserNotFound)
	}
	if err := repository.BlockUser(ctx, blockerID, missingBlockUserID); !errors.Is(err, UserErrors.TargetUserNotFound) {
		t.Fatalf("missing blocked user BlockUser() error = %v, want %v", err, UserErrors.TargetUserNotFound)
	}
	if err := repository.BlockUser(ctx, blockerID, blockerID); err == nil {
		t.Fatal("self BlockUser() error = nil, want database constraint error")
	}

	if err := repository.UnblockUser(ctx, blockerID, blockedUserID); err != nil {
		t.Fatalf("UnblockUser() unexpected error: %v", err)
	}
	if got := countUserIntegrationBlocks(t, ctx, db, blockerID, blockedUserID); got != 0 {
		t.Fatalf("block count after unblock = %d, want 0", got)
	}
	if err := repository.UnblockUser(ctx, blockerID, blockedUserID); err != nil {
		t.Fatalf("repeated UnblockUser() unexpected error: %v", err)
	}

	if err := repository.ReportUser(ctx, blockerID, blockedUserID); err != nil {
		t.Fatalf("ReportUser() unexpected error: %v", err)
	}
	if got := countUserIntegrationReports(t, ctx, db, blockerID, blockedUserID); got != 1 {
		t.Fatalf("report count = %d, want 1", got)
	}
	if err := repository.ReportUser(ctx, blockerID, blockedUserID); err != nil {
		t.Fatalf("repeated ReportUser() unexpected error: %v", err)
	}
	if got := countUserIntegrationReports(t, ctx, db, blockerID, blockedUserID); got != 1 {
		t.Fatalf("report count after repeated report = %d, want 1", got)
	}
	if err := repository.ReportUser(ctx, blockedUserID, blockerID); err != nil {
		t.Fatalf("reverse ReportUser() unexpected error: %v", err)
	}
	if got := countUserIntegrationReports(t, ctx, db, blockedUserID, blockerID); got != 1 {
		t.Fatalf("reverse report count = %d, want 1", got)
	}

	missingReportedUserID := blockedUserID + 2_000_000
	if err := repository.ReportUser(ctx, missingReportedUserID, blockedUserID); !errors.Is(err, UserErrors.UserNotFound) {
		t.Fatalf("missing reporter ReportUser() error = %v, want %v", err, UserErrors.UserNotFound)
	}
	if err := repository.ReportUser(ctx, blockerID, missingReportedUserID); !errors.Is(err, UserErrors.TargetUserNotFound) {
		t.Fatalf("missing reported user ReportUser() error = %v, want %v", err, UserErrors.TargetUserNotFound)
	}
	if err := repository.ReportUser(ctx, blockerID, blockerID); err == nil {
		t.Fatal("self ReportUser() error = nil, want database constraint error")
	}
}

func testUserIntegrationPendingEmailLifecycle(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	repository *PostgresRepository,
) {
	t.Helper()
	const oldEmail = "pending-lifecycle-old@example.invalid"
	const newEmail = "pending-lifecycle-new@example.invalid"
	const tokenHash = "pending-lifecycle-token-hash"
	userID := insertUserIntegrationUser(t, ctx, db, "pending-lifecycle-user", oldEmail)
	if _, err := db.ExecContext(
		ctx,
		`UPDATE users SET is_verified = true WHERE id = $1`,
		userID,
	); err != nil {
		t.Fatalf("mark current email verified: %v", err)
	}
	token := models.AccountToken{
		UserID:    userID,
		Hash:      tokenHash,
		Purpose:   models.AccountTokenPurposeEmailVerification,
		ExpiresAt: time.Now().UTC().Add(time.Hour),
	}

	result, err := repository.UpdateUser(
		ctx,
		userID,
		UserUpdateInput{Email: stringPointer(newEmail)},
		&token,
	)
	if err != nil {
		t.Fatalf("UpdateUser() email change: %v", err)
	}
	if !result.EmailChanged || result.PendingEmail != newEmail {
		t.Fatalf("UpdateUser() result = %+v, want pending email %q", result, newEmail)
	}

	var currentEmail string
	var pendingEmail sql.NullString
	var isVerified bool
	if err := db.QueryRowContext(
		ctx,
		`SELECT email, pending_email, is_verified FROM users WHERE id = $1`,
		userID,
	).Scan(&currentEmail, &pendingEmail, &isVerified); err != nil {
		t.Fatalf("read pending email before verification: %v", err)
	}
	if currentEmail != oldEmail {
		t.Fatalf("email before verification = %q, want old email %q", currentEmail, oldEmail)
	}
	if !pendingEmail.Valid || pendingEmail.String != newEmail {
		t.Fatalf("pending email before verification = %+v, want %q", pendingEmail, newEmail)
	}
	if !isVerified {
		t.Fatal("current email lost its verified state while new email is pending")
	}

	var activeTokenCount int
	if err := db.QueryRowContext(
		ctx,
		`SELECT COUNT(*) FROM account_tokens WHERE user_id = $1 AND token_hash = $2 AND used_at IS NULL`,
		userID,
		tokenHash,
	).Scan(&activeTokenCount); err != nil {
		t.Fatalf("count active verification token: %v", err)
	}
	if activeTokenCount != 1 {
		t.Fatalf("active verification token count = %d, want 1", activeTokenCount)
	}

	if err := repository.VerifyEmail(ctx, tokenHash); err != nil {
		t.Fatalf("VerifyEmail() pending address: %v", err)
	}
	if err := db.QueryRowContext(
		ctx,
		`SELECT email, pending_email, is_verified FROM users WHERE id = $1`,
		userID,
	).Scan(&currentEmail, &pendingEmail, &isVerified); err != nil {
		t.Fatalf("read email after verification: %v", err)
	}
	if currentEmail != newEmail {
		t.Fatalf("email after verification = %q, want %q", currentEmail, newEmail)
	}
	if pendingEmail.Valid {
		t.Fatalf("pending email after verification = %+v, want NULL", pendingEmail)
	}
	if !isVerified {
		t.Fatal("is_verified = false after verification")
	}

	var usedAt sql.NullTime
	if err := db.QueryRowContext(
		ctx,
		`SELECT used_at FROM account_tokens WHERE token_hash = $1`,
		tokenHash,
	).Scan(&usedAt); err != nil {
		t.Fatalf("read consumed verification token: %v", err)
	}
	if !usedAt.Valid {
		t.Fatal("verification token remains active after verification")
	}
}

func testUserIntegrationPendingEmailReservation(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	repository *PostgresRepository,
) {
	t.Helper()
	const reservedEmail = "reserved-pending@example.invalid"
	ownerID := insertUserIntegrationUser(
		t,
		ctx,
		db,
		"pending-reservation-owner",
		"pending-reservation-owner@example.invalid",
	)
	verificationToken := models.AccountToken{
		Hash:      "pending-reservation-token-hash",
		Purpose:   models.AccountTokenPurposeEmailVerification,
		ExpiresAt: time.Now().UTC().Add(time.Hour),
	}
	if _, err := repository.UpdateUser(
		ctx,
		ownerID,
		UserUpdateInput{Email: stringPointer(reservedEmail)},
		&verificationToken,
	); err != nil {
		t.Fatalf("reserve pending email: %v", err)
	}

	_, err := repository.CreateUser(
		ctx,
		models.User{
			UserName:  "pending-reservation-new-user",
			FirstName: "Integration",
			LastName:  "Test",
			Email:     reservedEmail,
			Password:  "not-a-real-password-hash",
		},
		models.AccountToken{
			Hash:      "pending-reservation-registration-token-hash",
			Purpose:   models.AccountTokenPurposeEmailVerification,
			ExpiresAt: time.Now().UTC().Add(time.Hour),
		},
	)
	if !errors.Is(err, UserErrors.UserAlreadyExists) {
		t.Fatalf("CreateUser() reserved pending email error = %v, want %v", err, UserErrors.UserAlreadyExists)
	}

	var createdCount int
	if err := db.QueryRowContext(
		ctx,
		`SELECT COUNT(*) FROM users WHERE user_name = 'pending-reservation-new-user'`,
	).Scan(&createdCount); err != nil {
		t.Fatalf("count user registered with reserved address: %v", err)
	}
	if createdCount != 0 {
		t.Fatalf("users registered with reserved pending address = %d, want 0", createdCount)
	}
}

func testUserIntegrationPendingEmailConflict(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	repository *PostgresRepository,
) {
	t.Helper()
	const pendingEmail = "verification-conflict@example.invalid"
	const tokenHash = "verification-conflict-token-hash"
	userID := insertUserIntegrationUser(
		t,
		ctx,
		db,
		"verification-conflict-owner",
		"verification-conflict-owner@example.invalid",
	)
	verificationToken := models.AccountToken{
		Hash:      tokenHash,
		Purpose:   models.AccountTokenPurposeEmailVerification,
		ExpiresAt: time.Now().UTC().Add(time.Hour),
	}
	if _, err := repository.UpdateUser(
		ctx,
		userID,
		UserUpdateInput{Email: stringPointer(pendingEmail)},
		&verificationToken,
	); err != nil {
		t.Fatalf("create conflicting pending email: %v", err)
	}
	insertUserIntegrationUser(t, ctx, db, "verification-conflict-other", pendingEmail)

	err := repository.VerifyEmail(ctx, tokenHash)
	if !errors.Is(err, UserErrors.UserAlreadyExists) {
		t.Fatalf("VerifyEmail() occupied address error = %v, want %v", err, UserErrors.UserAlreadyExists)
	}

	var savedPendingEmail sql.NullString
	if err := db.QueryRowContext(
		ctx,
		`SELECT pending_email FROM users WHERE id = $1`,
		userID,
	).Scan(&savedPendingEmail); err != nil {
		t.Fatalf("read pending email after conflict: %v", err)
	}
	if !savedPendingEmail.Valid || savedPendingEmail.String != pendingEmail {
		t.Fatalf("pending email after conflict = %+v, want %q", savedPendingEmail, pendingEmail)
	}

	var usedAt sql.NullTime
	if err := db.QueryRowContext(
		ctx,
		`SELECT used_at FROM account_tokens WHERE token_hash = $1`,
		tokenHash,
	).Scan(&usedAt); err != nil {
		t.Fatalf("read token after verification conflict: %v", err)
	}
	if usedAt.Valid {
		t.Fatal("verification token was consumed after address conflict")
	}
}

func userIntegrationDatabaseDSN() (string, bool) {
	if dsn := os.Getenv("USER_TEST_DATABASE_DSN"); dsn != "" {
		return dsn, true
	}

	host := os.Getenv("SQL_HOST")
	port := os.Getenv("SQL_PORT")
	userName := os.Getenv("SQL_USER")
	password := os.Getenv("SQL_PASSWORD")
	databaseName := os.Getenv("SQL_DATABASE")
	if host == "" || port == "" || userName == "" || password == "" || databaseName == "" {
		return "", false
	}

	dsn := &url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(userName, password),
		Host:   net.JoinHostPort(host, port),
		Path:   databaseName,
	}
	query := dsn.Query()
	query.Set("sslmode", "disable")
	dsn.RawQuery = query.Encode()
	return dsn.String(), true
}

func applyUserInitialMigration(t *testing.T, ctx context.Context, db *sql.DB) {
	t.Helper()

	migrationPath := filepath.Join("..", "database", "migrations", "00001_initial_schema.sql")
	contents, err := os.ReadFile(migrationPath)
	if err != nil {
		t.Fatalf("read initial migration: %v", err)
	}

	upSection, _, found := strings.Cut(string(contents), "-- +goose Down")
	if !found {
		t.Fatal("initial migration does not contain -- +goose Down")
	}
	upSection = strings.Replace(upSection, "-- +goose Up", "", 1)

	for _, statement := range strings.Split(upSection, ";") {
		statement = strings.TrimSpace(statement)
		if statement == "" {
			continue
		}
		if _, err := db.ExecContext(ctx, statement); err != nil {
			t.Fatalf("apply initial migration statement: %v\nSQL: %s", err, statement)
		}
	}
}

func insertUserIntegrationUser(t *testing.T, ctx context.Context, db *sql.DB, name string, email string) uint {
	t.Helper()
	const query = `
		INSERT INTO users (
			user_name,
			first_name,
			last_name,
			email,
			password
		)
		VALUES ($1, 'Integration', 'Test', $2, 'not-a-real-password-hash')
		RETURNING id
	`

	var userID uint
	if err := db.QueryRowContext(ctx, query, name, email).Scan(&userID); err != nil {
		t.Fatalf("insert integration user %q: %v", name, err)
	}
	return userID
}

func countUserIntegrationPhotos(t *testing.T, ctx context.Context, db *sql.DB, userID uint) int {
	t.Helper()
	const query = `SELECT COUNT(*) FROM photos WHERE user_id = $1`

	var count int
	if err := db.QueryRowContext(ctx, query, userID).Scan(&count); err != nil {
		t.Fatalf("count photos for user %d: %v", userID, err)
	}
	return count
}

func countUserIntegrationDecisions(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	userID uint,
	targetUserID uint,
) int {
	t.Helper()
	const query = `
		SELECT COUNT(*)
		FROM profile_decisions
		WHERE user_id = $1
		  AND target_user_id = $2
	`

	var count int
	if err := db.QueryRowContext(ctx, query, userID, targetUserID).Scan(&count); err != nil {
		t.Fatalf("count profile decisions %d -> %d: %v", userID, targetUserID, err)
	}
	return count
}

func countUserIntegrationBlocks(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	blockerID uint,
	blockedUserID uint,
) int {
	t.Helper()
	const query = `
		SELECT COUNT(*)
		FROM user_blocks
		WHERE blocker_id = $1
		  AND blocked_user_id = $2
	`

	var count int
	if err := db.QueryRowContext(ctx, query, blockerID, blockedUserID).Scan(&count); err != nil {
		t.Fatalf("count user blocks %d -> %d: %v", blockerID, blockedUserID, err)
	}
	return count
}

func countUserIntegrationReports(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	reporterID uint,
	reportedUserID uint,
) int {
	t.Helper()
	const query = `
		SELECT COUNT(*)
		FROM user_reports
		WHERE reporter_id = $1
		  AND reported_user_id = $2
	`

	var count int
	if err := db.QueryRowContext(ctx, query, reporterID, reportedUserID).Scan(&count); err != nil {
		t.Fatalf("count user reports %d -> %d: %v", reporterID, reportedUserID, err)
	}
	return count
}

func containsUserIntegrationProfile(profiles []models.User, userID uint) bool {
	for _, profile := range profiles {
		if profile.ID == userID {
			return true
		}
	}
	return false
}

func findUserIntegrationProfile(profiles []models.User, userID uint) (models.User, bool) {
	for _, profile := range profiles {
		if profile.ID == userID {
			return profile, true
		}
	}
	return models.User{}, false
}

func listUserIntegrationCandidates(
	ctx context.Context,
	repository *PostgresRepository,
	userID uint,
) ([]models.User, error) {
	profile, err := repository.GetProfile(ctx, userID)
	if err != nil {
		return nil, err
	}
	return repository.ListProfileCandidates(
		ctx,
		userID,
		strings.TrimSpace(profile.Preferences),
		strings.TrimSpace(profile.Gender),
		true,
	)
}

func listUserIntegrationSearchCandidates(
	ctx context.Context,
	repository *PostgresRepository,
	userID uint,
) ([]models.User, error) {
	profile, err := repository.GetProfile(ctx, userID)
	if err != nil {
		return nil, err
	}
	return repository.ListProfileCandidates(
		ctx,
		userID,
		strings.TrimSpace(profile.Preferences),
		strings.TrimSpace(profile.Gender),
		false,
	)
}

func containsUserIntegrationLiker(likers []models.ProfileLiker, userID uint) bool {
	for _, liker := range likers {
		if liker.ID == userID {
			return true
		}
	}
	return false
}

func findUserIntegrationMatch(matches []models.Match, userID uint) (models.Match, bool) {
	for _, match := range matches {
		if match.ID == userID {
			return match, true
		}
	}
	return models.Match{}, false
}
