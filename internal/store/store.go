package store

import (
	"fmt"
	"log"

	"queuectl.backend/internal/config" // ✅ added import
	"queuectl.backend/internal/job"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func InitDB() (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open("queue.db"), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("database opening failure: %w", err)
	}

	if err := db.AutoMigrate(&job.Job{}, &config.Config{}); err != nil {
		return nil, fmt.Errorf("migration failure: %w", err)
	}

	log.Printf("Database connected and ready...")
	return db, nil
}
