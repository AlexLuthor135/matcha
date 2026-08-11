package notification

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

func TestPostgresNotificationRepositoryIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping PostgreSQL integration test in short mode")
	}

	dsn, ok := notificationIntegrationDatabaseDSN()
	if !ok {
		t.Skip("set NOTIFICATION_TEST_DATABASE_DSN or SQL_* variables to run PostgreSQL integration tests")
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

	schemaName := fmt.Sprintf("notification_repository_test_%d", time.Now().UnixNano())
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

	applyNotificationInitialMigration(t, ctx, db)

	ownerID := insertNotificationIntegrationUser(t, ctx, db, "notification-owner", "notification-owner@example.invalid")
	senderID := insertNotificationIntegrationUser(t, ctx, db, "notification-sender", "notification-sender@example.invalid")
	otherUserID := insertNotificationIntegrationUser(t, ctx, db, "notification-other", "notification-other@example.invalid")

	olderAt := time.Date(2026, time.August, 5, 10, 0, 0, 0, time.UTC)
	newerAt := olderAt.Add(time.Hour)
	oldestID := insertIntegrationNotification(
		t,
		ctx,
		db,
		ownerID,
		nil,
		"system",
		"System",
		"Old notification",
		`{"kind":"system"}`,
		olderAt,
		nil,
	)
	readAt := newerAt.Add(time.Hour)
	firstNewerID := insertIntegrationNotification(
		t,
		ctx,
		db,
		ownerID,
		&senderID,
		"match",
		"New match",
		"First notification at the same time",
		`{"matched_user_id":2}`,
		newerAt,
		&readAt,
	)
	secondNewerID := insertIntegrationNotification(
		t,
		ctx,
		db,
		ownerID,
		nil,
		"system",
		"System",
		"Second notification at the same time",
		`{"position":2}`,
		newerAt,
		nil,
	)
	insertIntegrationNotification(
		t,
		ctx,
		db,
		otherUserID,
		&senderID,
		"match",
		"Other user",
		"Must not be returned",
		`{"private":true}`,
		newerAt.Add(time.Hour),
		nil,
	)

	repository := NewPostgresRepository(db)
	notifications, err := repository.ListNotifications(ctx, ownerID, 50)
	if err != nil {
		t.Fatalf("ListNotifications(): %v", err)
	}
	if len(notifications) != 3 {
		t.Fatalf("ListNotifications() length = %d, want 3", len(notifications))
	}
	wantOrder := []uint{secondNewerID, firstNewerID, oldestID}
	for index, wantID := range wantOrder {
		if notifications[index].ID != wantID {
			t.Fatalf("notification %d ID = %d, want %d", index, notifications[index].ID, wantID)
		}
		if notifications[index].UserID != ownerID {
			t.Fatalf("notification %d userID = %d, want %d", index, notifications[index].UserID, ownerID)
		}
	}
	if notifications[0].SenderID != nil || notifications[0].ReadAt != nil {
		t.Fatalf("unread system notification nullable fields = sender %v, readAt %v", notifications[0].SenderID, notifications[0].ReadAt)
	}
	if position, ok := notifications[0].Data["position"].(float64); !ok || position != 2 {
		t.Fatalf("decoded data = %#v, want position 2", notifications[0].Data)
	}
	if notifications[1].SenderID == nil || *notifications[1].SenderID != senderID {
		t.Fatalf("senderID = %v, want %d", notifications[1].SenderID, senderID)
	}
	if notifications[1].ReadAt == nil || !notifications[1].ReadAt.Equal(readAt) {
		t.Fatalf("readAt = %v, want %v", notifications[1].ReadAt, readAt)
	}

	limited, err := repository.ListNotifications(ctx, ownerID, 2)
	if err != nil {
		t.Fatalf("ListNotifications() with limit: %v", err)
	}
	if len(limited) != 2 || limited[0].ID != secondNewerID || limited[1].ID != firstNewerID {
		t.Fatalf("limited notifications = %+v, want first two newest notifications", limited)
	}

	_, err = repository.MarkNotificationRead(ctx, otherUserID, secondNewerID)
	if !errors.Is(err, NotificationErrors.NotificationNotFound) {
		t.Fatalf("other user MarkNotificationRead() error = %v, want %v", err, NotificationErrors.NotificationNotFound)
	}

	firstReceipt, err := repository.MarkNotificationRead(ctx, ownerID, secondNewerID)
	if err != nil {
		t.Fatalf("owner MarkNotificationRead() first call: %v", err)
	}
	if firstReceipt.NotificationID != secondNewerID || firstReceipt.ReadAt.IsZero() {
		t.Fatalf("first receipt = %+v", firstReceipt)
	}

	time.Sleep(5 * time.Millisecond)
	secondReceipt, err := repository.MarkNotificationRead(ctx, ownerID, secondNewerID)
	if err != nil {
		t.Fatalf("owner MarkNotificationRead() second call: %v", err)
	}
	if !secondReceipt.ReadAt.Equal(firstReceipt.ReadAt) {
		t.Fatalf("second readAt = %s, want original %s", secondReceipt.ReadAt, firstReceipt.ReadAt)
	}

	_, err = repository.MarkNotificationRead(ctx, ownerID, secondNewerID+1_000_000)
	if !errors.Is(err, NotificationErrors.NotificationNotFound) {
		t.Fatalf("missing notification error = %v, want %v", err, NotificationErrors.NotificationNotFound)
	}

	created, err := repository.CreateNotification(
		ctx,
		CreateNotificationInput{
			UserID:   ownerID,
			SenderID: &senderID,
			Type:     "match",
			Title:    "Created by repository",
			Message:  "A persisted notification",
			Data: map[string]any{
				"matched_user_id": senderID,
			},
		},
	)
	if err != nil {
		t.Fatalf("CreateNotification(): %v", err)
	}
	if created.ID == 0 || created.CreatedAt.IsZero() || created.UpdatedAt.IsZero() {
		t.Fatalf("CreateNotification() generated fields = %+v", created)
	}
	if created.UserID != ownerID || created.SenderID == nil || *created.SenderID != senderID {
		t.Fatalf("CreateNotification() ownership fields = %+v", created)
	}
	if created.Type != "match" || created.Title != "Created by repository" || created.Message != "A persisted notification" {
		t.Fatalf("CreateNotification() content fields = %+v", created)
	}
	if matchedUserID, ok := created.Data["matched_user_id"].(uint); !ok || matchedUserID != senderID {
		t.Fatalf("CreateNotification() data = %#v, want matched user %d", created.Data, senderID)
	}

	stored, err := repository.ListNotifications(ctx, ownerID, 50)
	if err != nil {
		t.Fatalf("ListNotifications() after create: %v", err)
	}
	createdFromDatabase, found := findIntegrationNotification(stored, created.ID)
	if !found {
		t.Fatalf("created notification %d was not persisted", created.ID)
	}
	if createdFromDatabase.SenderID == nil || *createdFromDatabase.SenderID != senderID {
		t.Fatalf("persisted senderID = %v, want %d", createdFromDatabase.SenderID, senderID)
	}
	if matchedUserID, ok := createdFromDatabase.Data["matched_user_id"].(float64); !ok || matchedUserID != float64(senderID) {
		t.Fatalf("persisted data = %#v, want matched user %d", createdFromDatabase.Data, senderID)
	}

	withoutOptionalFields, err := repository.CreateNotification(
		ctx,
		CreateNotificationInput{
			UserID:  ownerID,
			Type:    "system",
			Message: "No sender and no data",
		},
	)
	if err != nil {
		t.Fatalf("CreateNotification() without optional fields: %v", err)
	}
	if withoutOptionalFields.SenderID != nil || withoutOptionalFields.Data == nil || len(withoutOptionalFields.Data) != 0 {
		t.Fatalf("CreateNotification() optional fields = sender %v, data %#v", withoutOptionalFields.SenderID, withoutOptionalFields.Data)
	}

	_, err = repository.CreateNotification(
		ctx,
		CreateNotificationInput{
			UserID:  ownerID + senderID + otherUserID + 1_000_000,
			Type:    "system",
			Message: "Missing recipient",
		},
	)
	if err == nil {
		t.Fatal("CreateNotification() for missing user error = nil, want foreign key error")
	}

	_, err = repository.CreateNotification(
		ctx,
		CreateNotificationInput{
			UserID:  ownerID,
			Type:    "system",
			Message: "Invalid JSON data",
			Data: map[string]any{
				"unsupported": make(chan int),
			},
		},
	)
	if err == nil {
		t.Fatal("CreateNotification() with unsupported data error = nil, want JSON encoding error")
	}

	var countBeforeBlock int
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM notifications`).Scan(&countBeforeBlock); err != nil {
		t.Fatalf("count notifications before block: %v", err)
	}
	if _, err := db.ExecContext(
		ctx,
		`INSERT INTO user_blocks (blocker_id, blocked_user_id) VALUES ($1, $2)`,
		ownerID,
		senderID,
	); err != nil {
		t.Fatalf("block notification users: %v", err)
	}

	blockedNotifications := []struct {
		name     string
		userID   uint
		senderID uint
	}{
		{
			name:     "recipient blocked sender",
			userID:   ownerID,
			senderID: senderID,
		},
		{
			name:     "sender blocked recipient",
			userID:   senderID,
			senderID: ownerID,
		},
	}
	for _, test := range blockedNotifications {
		t.Run(test.name, func(t *testing.T) {
			_, err := repository.CreateNotification(
				ctx,
				CreateNotificationInput{
					UserID:   test.userID,
					SenderID: &test.senderID,
					Type:     "like",
					Message:  "Must not be persisted",
				},
			)
			if !errors.Is(err, NotificationErrors.UserBlocked) {
				t.Fatalf(
					"CreateNotification() error = %v, want %v",
					err,
					NotificationErrors.UserBlocked,
				)
			}
		})
	}

	var countAfterBlockedAttempts int
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM notifications`).Scan(&countAfterBlockedAttempts); err != nil {
		t.Fatalf("count notifications after blocked attempts: %v", err)
	}
	if countAfterBlockedAttempts != countBeforeBlock {
		t.Fatalf(
			"notification count after blocked attempts = %d, want %d",
			countAfterBlockedAttempts,
			countBeforeBlock,
		)
	}

	systemNotification, err := repository.CreateNotification(
		ctx,
		CreateNotificationInput{
			UserID:  ownerID,
			Type:    "system",
			Message: "System notifications have no sender",
		},
	)
	if err != nil {
		t.Fatalf("CreateNotification() system notification after block: %v", err)
	}
	if systemNotification.ID == 0 || systemNotification.SenderID != nil {
		t.Fatalf("system notification = %+v, want persisted notification without sender", systemNotification)
	}
}

func findIntegrationNotification(notifications []models.Notification, notificationID uint) (models.Notification, bool) {
	for _, notification := range notifications {
		if notification.ID == notificationID {
			return notification, true
		}
	}
	return models.Notification{}, false
}

func notificationIntegrationDatabaseDSN() (string, bool) {
	if dsn := os.Getenv("NOTIFICATION_TEST_DATABASE_DSN"); dsn != "" {
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

func applyNotificationInitialMigration(t *testing.T, ctx context.Context, db *sql.DB) {
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

func insertNotificationIntegrationUser(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	name string,
	email string,
) uint {
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

func insertIntegrationNotification(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	userID uint,
	senderID *uint,
	notificationType string,
	title string,
	message string,
	data string,
	createdAt time.Time,
	readAt *time.Time,
) uint {
	t.Helper()
	const query = `
		INSERT INTO notifications (
			created_at,
			updated_at,
			user_id,
			sender_id,
			type,
			title,
			message,
			data,
			read_at
		)
		VALUES ($1, $1, $2, $3, $4, $5, $6, $7::jsonb, $8)
		RETURNING id
	`

	var senderValue any
	if senderID != nil {
		senderValue = *senderID
	}
	var readAtValue any
	if readAt != nil {
		readAtValue = *readAt
	}

	var notificationID uint
	if err := db.QueryRowContext(
		ctx,
		query,
		createdAt,
		userID,
		senderValue,
		notificationType,
		title,
		message,
		data,
		readAtValue,
	).Scan(&notificationID); err != nil {
		t.Fatalf("insert notification for user %d: %v", userID, err)
	}
	return notificationID
}
