package main

import (
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	if err := godotenv.Load(); err != nil {
		fmt.Printf("warning: no .env loaded: %v\n", err)
	}

	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s TimeZone=UTC",
		env("DB_HOST", "localhost"),
		env("DB_PORT", "5432"),
		env("DB_USER", "bookstore"),
		env("DB_PASSWORD", "bookstore_secret"),
		env("DB_NAME", "bookstore_db"),
		env("DB_SSLMODE", "disable"),
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		fmt.Fprintf(os.Stderr, "FAIL: connect to postgres: %v\n", err)
		os.Exit(1)
	}

	sqlDB, err := db.DB()
	if err != nil {
		fmt.Fprintf(os.Stderr, "FAIL: obtain *sql.DB: %v\n", err)
		os.Exit(1)
	}
	sqlDB.SetMaxOpenConns(5)
	sqlDB.SetConnMaxLifetime(time.Minute)
	defer sqlDB.Close()

	if err := sqlDB.Ping(); err != nil {
		fmt.Fprintf(os.Stderr, "FAIL: ping postgres: %v\n", err)
		os.Exit(1)
	}

	var version string
	if err := db.Raw("SELECT version()").Scan(&version).Error; err != nil {
		fmt.Fprintf(os.Stderr, "FAIL: query version: %v\n", err)
		os.Exit(1)
	}

	var now time.Time
	if err := db.Raw("SELECT now()").Scan(&now).Error; err != nil {
		fmt.Fprintf(os.Stderr, "FAIL: query now(): %v\n", err)
		os.Exit(1)
	}

	fmt.Println("OK: connected to PostgreSQL")
	fmt.Printf("  server : %s\n", version)
	fmt.Printf("  db time: %s\n", now.UTC().Format(time.RFC3339))
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
