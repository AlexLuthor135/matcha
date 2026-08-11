package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"golang.org/x/crypto/bcrypt"
)

const (
	defaultSeedProfileCount = 500
	minimumSeedProfileCount = 500
	maximumSeedProfileCount = 5000
	defaultSeedAssetDir     = "./uploads/seed"
)

type seedConfig struct {
	DSN          string
	Password     string
	ProfileCount int
	AssetDir     string
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := run(ctx); err != nil {
		log.Fatalf("Seed database: %v", err)
	}
}

func run(parent context.Context) error {
	config, err := loadSeedConfig()
	if err != nil {
		return err
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(config.Password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash seed password: %w", err)
	}
	if err := generateSeedAssets(config.AssetDir); err != nil {
		return fmt.Errorf("generate seed image assets: %w", err)
	}
	database, err := sql.Open("pgx", config.DSN)
	if err != nil {
		return fmt.Errorf("open PostgreSQL: %w", err)
	}
	defer database.Close()
	database.SetMaxOpenConns(4)
	database.SetMaxIdleConns(4)
	database.SetConnMaxLifetime(5 * time.Minute)
	ctx, cancel := context.WithTimeout(parent, 3*time.Minute)
	defer cancel()
	if err := database.PingContext(ctx); err != nil {
		return fmt.Errorf("ping PostgreSQL: %w", err)
	}
	tx, err := database.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin seed transaction: %w", err)
	}
	defer tx.Rollback()
	stats, err := seedDatabase(ctx, tx, config.ProfileCount, string(passwordHash), time.Now().UTC())
	if err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit seed transaction: %w", err)
	}
	log.Printf("Seed complete: profiles=%d photos=%d decisions=%d", stats.Profiles, stats.Photos, stats.Decisions)
	return nil
}

func loadSeedConfig() (seedConfig, error) {
	password := os.Getenv("SEED_PASSWORD")
	if strings.TrimSpace(password) == "" {
		return seedConfig{}, errors.New("SEED_PASSWORD is required")
	}
	if len([]byte(password)) > 72 {
		return seedConfig{}, errors.New("SEED_PASSWORD must not exceed 72 bytes")
	}
	profileCount := defaultSeedProfileCount
	if rawCount := strings.TrimSpace(os.Getenv("SEED_PROFILE_COUNT")); rawCount != "" {
		parsedCount, err := strconv.Atoi(rawCount)
		if err != nil {
			return seedConfig{}, errors.New("SEED_PROFILE_COUNT must be an integer")
		}
		profileCount = parsedCount
	}
	if profileCount < minimumSeedProfileCount || profileCount > maximumSeedProfileCount {
		return seedConfig{}, fmt.Errorf("SEED_PROFILE_COUNT must be between %d and %d", minimumSeedProfileCount, maximumSeedProfileCount)
	}
	assetDir := strings.TrimSpace(os.Getenv("SEED_ASSET_DIRECTORY"))
	if assetDir == "" {
		assetDir = defaultSeedAssetDir
	}
	dsn, err := seedDatabaseDSN()
	if err != nil {
		return seedConfig{}, err
	}
	return seedConfig{
		DSN:          dsn,
		Password:     password,
		ProfileCount: profileCount,
		AssetDir:     assetDir,
	}, nil
}

func seedDatabaseDSN() (string, error) {
	if dsn := strings.TrimSpace(os.Getenv("DATABASE_URL")); dsn != "" {
		return dsn, nil
	}

	required := []string{"SQL_USER", "SQL_PASSWORD", "SQL_HOST", "SQL_PORT", "SQL_DATABASE"}
	values := make(map[string]string, len(required))
	for _, name := range required {
		value := strings.TrimSpace(os.Getenv(name))
		if value == "" {
			return "", fmt.Errorf("%s is required", name)
		}
		values[name] = value
	}

	databaseURL := &url.URL{
		Scheme: "postgres",
		User: url.UserPassword(
			values["SQL_USER"],
			values["SQL_PASSWORD"],
		),
		Host: net.JoinHostPort(
			values["SQL_HOST"],
			values["SQL_PORT"],
		),
		Path: values["SQL_DATABASE"],
	}
	query := databaseURL.Query()
	query.Set("sslmode", "disable")
	databaseURL.RawQuery = query.Encode()
	return databaseURL.String(), nil
}
