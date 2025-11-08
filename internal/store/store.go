package store

import (
	"fmt"
	"log"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func InitDB() (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open("queue.db"), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("database opening failure: %w", err)
	}
	if err := db.AutoMigrate(); err != nil {
		return nil, fmt.Errorf("migration failure: %w", err)
	}

	log.Printf("Database connected and ready...")
	return db, nil
}
