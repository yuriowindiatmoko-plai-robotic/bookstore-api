package main

import (
	"fmt"
	"os"
	"time"

	"github.com/yourusername/bookstore-api/config"
)

func main() {
	db, err := config.Connect()
	if err != nil {
		fmt.Fprintf(os.Stderr, "FAIL: %v\n", err)
		os.Exit(1)
	}

	sqlDB, err := db.DB()
	if err != nil {
		fmt.Fprintf(os.Stderr, "FAIL: obtain *sql.DB: %v\n", err)
		os.Exit(1)
	}
	defer sqlDB.Close()

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
