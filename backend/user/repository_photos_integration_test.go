package user

import (
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

func TestPostgresPhotoRepositoryIntegration(t *testing.T) {
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
