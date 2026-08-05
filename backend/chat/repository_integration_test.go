package chat

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

func TestPostgresRepositoryIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping PostgreSQL integration test in short mode")
	}

	dsn, ok := integrationDatabaseDSN()
	if !ok {
		t.Skip("set CHAT_TEST_DATABASE_DSN or SQL_* variables to run PostgreSQL integration tests")
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

	schemaName := fmt.Sprintf("chat_repository_test_%d", time.Now().UnixNano())
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

	applyInitialMigration(t, ctx, db)

	senderID := insertIntegrationUser(t, ctx, db, "sender", "sender@example.invalid")
	recipientID := insertIntegrationUser(t, ctx, db, "recipient", "recipient@example.invalid")
	outsiderID := insertIntegrationUser(t, ctx, db, "outsider", "outsider@example.invalid")

	repository := NewPostgresRepository(db)

	_, err = repository.CreateMessage(ctx, senderID, recipientID, "message without match")
	if !errors.Is(err, ChatErrors.UsersNotMatched) {
		t.Fatalf("CreateMessage() without match error = %v, want %v", err, ChatErrors.UsersNotMatched)
	}

	insertIntegrationProfileDecision(t, ctx, db, senderID, recipientID, "like")

	_, err = repository.CreateMessage(ctx, senderID, recipientID, "message with one like")
	if !errors.Is(err, ChatErrors.UsersNotMatched) {
		t.Fatalf("CreateMessage() with one like error = %v, want %v", err, ChatErrors.UsersNotMatched)
	}

	insertIntegrationProfileDecision(t, ctx, db, recipientID, senderID, "like")

	firstMessage, err := repository.CreateMessage(ctx, senderID, recipientID, "first message")
	if err != nil {
		t.Fatalf("CreateMessage() first message: %v", err)
	}
	if firstMessage.ID == 0 || firstMessage.ConversationID == 0 {
		t.Fatalf("CreateMessage() returned invalid IDs: %+v", firstMessage)
	}

	secondMessage, err := repository.CreateMessage(ctx, recipientID, senderID, "second message")
	if err != nil {
		t.Fatalf("CreateMessage() second message: %v", err)
	}
	if secondMessage.ConversationID != firstMessage.ConversationID {
		t.Fatalf("second conversation ID = %d, want %d", secondMessage.ConversationID, firstMessage.ConversationID)
	}

	for _, userID := range []uint{senderID, recipientID} {
		conversations, err := repository.ListConversations(ctx, userID)
		if err != nil {
			t.Fatalf("ListConversations(%d): %v", userID, err)
		}
		if len(conversations) != 1 || conversations[0].ID != firstMessage.ConversationID {
			t.Fatalf("ListConversations(%d) = %+v, want conversation %d", userID, conversations, firstMessage.ConversationID)
		}
	}

	messages, err := repository.ListConversationMessages(ctx, senderID, firstMessage.ConversationID)
	if err != nil {
		t.Fatalf("ListConversationMessages() participant: %v", err)
	}
	if len(messages) != 2 || messages[0].ID != firstMessage.ID || messages[1].ID != secondMessage.ID {
		t.Fatalf("ListConversationMessages() = %+v, want messages in creation order", messages)
	}

	_, err = repository.ListConversationMessages(ctx, outsiderID, firstMessage.ConversationID)
	if !errors.Is(err, ChatErrors.ConversationNotFound) {
		t.Fatalf("outsider ListConversationMessages() error = %v, want %v", err, ChatErrors.ConversationNotFound)
	}

	_, err = repository.MarkMessageRead(ctx, senderID, firstMessage.ID)
	if !errors.Is(err, ChatErrors.MessageNotFound) {
		t.Fatalf("sender MarkMessageRead() error = %v, want %v", err, ChatErrors.MessageNotFound)
	}

	_, err = repository.MarkMessageRead(ctx, outsiderID, firstMessage.ID)
	if !errors.Is(err, ChatErrors.MessageNotFound) {
		t.Fatalf("outsider MarkMessageRead() error = %v, want %v", err, ChatErrors.MessageNotFound)
	}

	firstReceipt, err := repository.MarkMessageRead(ctx, recipientID, firstMessage.ID)
	if err != nil {
		t.Fatalf("recipient MarkMessageRead() first call: %v", err)
	}
	if firstReceipt.MessageID != firstMessage.ID ||
		firstReceipt.ConversationID != firstMessage.ConversationID ||
		firstReceipt.SenderID != senderID ||
		firstReceipt.RecipientID != recipientID ||
		firstReceipt.ReadAt.IsZero() {
		t.Fatalf("MarkMessageRead() receipt = %+v", firstReceipt)
	}

	secondReceipt, err := repository.MarkMessageRead(ctx, recipientID, firstMessage.ID)
	if err != nil {
		t.Fatalf("recipient MarkMessageRead() second call: %v", err)
	}
	if !secondReceipt.ReadAt.Equal(firstReceipt.ReadAt) {
		t.Fatalf("second read_at = %s, want original %s", secondReceipt.ReadAt, firstReceipt.ReadAt)
	}
}

func integrationDatabaseDSN() (string, bool) {
	if dsn := os.Getenv("CHAT_TEST_DATABASE_DSN"); dsn != "" {
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

func applyInitialMigration(t *testing.T, ctx context.Context, db *sql.DB) {
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

func insertIntegrationUser(t *testing.T, ctx context.Context, db *sql.DB, name string, email string) uint {
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

func insertIntegrationProfileDecision(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	userID uint,
	targetUserID uint,
	decision string,
) {
	t.Helper()
	const query = `
		INSERT INTO profile_decisions (
			user_id,
			target_user_id,
			decision
		)
		VALUES ($1, $2, $3)
	`

	if _, err := db.ExecContext(ctx, query, userID, targetUserID, decision); err != nil {
		t.Fatalf("insert profile decision %d -> %d: %v", userID, targetUserID, err)
	}
}
