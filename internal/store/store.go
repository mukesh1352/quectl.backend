package store

import (
	"fmt"
	"log"

	"queuectl.backend/internal/config"
	"queuectl.backend/internal/job"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func InitDB() (*gorm.DB, error) {
	dsn := "queue.db?_journal_mode=WAL&_busy_timeout=5000"
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("database opening failure: %w", err)
	}

	if err := db.AutoMigrate(&job.Job{}, &config.Config{}); err != nil {
		return nil, fmt.Errorf("migration failure: %w", err)
	}

	if err := db.Exec("PRAGMA journal_mode=WAL;").Error; err != nil {
		log.Printf("warning: failed to enable WAL mode: %v", err)
	}
	if err := db.Exec("PRAGMA busy_timeout=5000;").Error; err != nil {
		log.Printf("warning: failed to set busy timeout: %v", err)
	}

	log.Printf("Database connected and ready (WAL mode, 5s busy timeout)...")
	return db, nil
}
