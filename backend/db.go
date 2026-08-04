package main

import (
	"context"
	"database/sql"
	"fmt"
	_ "github.com/jackc/pgx/v5/stdlib"
	"log"
	"os"
	"time"
)

func env(str string) string {
	return os.Getenv(str)
}

func initDB() *sql.DB {

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		env("SQL_USER"), env("SQL_PASSWORD"), env("SQL_HOST"), env("SQL_PORT"), env("SQL_DATABASE"))
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Fatalf("Failed to connect to DB: %v", err)
	}
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		log.Fatalf("Failed to ping to DB: %v", err)
	}
	fmt.Println("Database initialization successful")
	return db
}
