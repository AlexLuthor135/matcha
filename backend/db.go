package main

import (
	"fmt"
	"log"
	"os"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func initDB() {
	host := os.Getenv("SQL_HOST")
	user := os.Getenv("SQL_USER")
	password := os.Getenv("SQL_PASSWORD")
	dbname := os.Getenv("SQL_DATABASE")
	port := os.Getenv("SQL_PORT")

	if host == "" {
		host = "db"
		port = "5432"
	}

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=UTC", 
		host, user, password, dbname, port)

	var err error

	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	err = DB.AutoMigrate(&User{})
	if err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}
	fmt.Println("Database connection established and migrated successfully")
}
