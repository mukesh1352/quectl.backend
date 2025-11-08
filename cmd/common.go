package cmd

import (
	"log"

	"queuectl.backend/internal/store"
)

var repo *store.JobRepo

//initializing the db once

func CommonInit() {
	if repo != nil {
		return
	}
	db, err := store.InitDB()
	if err != nil {
		log.Fatalf("Initialization of the database failed")
	}
	repo = store.NewJobRepo(db)
	log.Println("Database initialized and repository ready")
}
