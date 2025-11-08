package job_test

import (
	"testing"
	"time"

	"queuectl.backend/internal/job"
	"queuectl.backend/internal/store"
)

func TestCreateAndFetchJob(t *testing.T) {
	db, err := store.InitDB()
	if err != nil {
		t.Fatal(err)
	}
	repo := store.NewJobRepo(db)

	// Clean up before test
	db.Exec("DELETE FROM jobs")

	// Create a new job
	j := &job.Job{
		ID:         "test-job",
		Command:    "echo hello",
		State:      job.StatePending,
		MaxRetries: 3,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	if err := repo.Create(j); err != nil {
		t.Fatalf("failed to create job: %v", err)
	}

	fetched, err := repo.FindPending()
	if err != nil {
		t.Fatalf("failed to find pending job: %v", err)
	}

	if fetched == nil || fetched.ID != j.ID {
		t.Fatalf("expected job ID %s, got %+v", j.ID, fetched)
	}
}
