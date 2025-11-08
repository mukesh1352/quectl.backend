package queue

import (
	"bytes"
	"context"
	"errors"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"time"

	"queuectl.backend/internal/store"
)

// WorkerConfig defines the behavior of a worker.
type WorkerConfig struct {
	ID           string
	PollInterval time.Duration
	MaxSleepTime time.Duration
	RetryDelay   time.Duration
	ExecTimeout  time.Duration
}

// Worker handles jobs fetched from the repository.
type Worker struct {
	repo *store.JobRepo
	cfg  WorkerConfig
}

// NewWorker creates and initializes a new worker with default values.
func NewWorker(repo *store.JobRepo, cfg WorkerConfig) *Worker {
	if cfg.ID == "" {
		cfg.ID = "worker-" + time.Now().Format("150405")
	}
	if cfg.PollInterval == 0 {
		cfg.PollInterval = 2 * time.Second
	}
	if cfg.MaxSleepTime == 0 {
		cfg.MaxSleepTime = 30 * time.Second
	}
	if cfg.RetryDelay == 0 {
		cfg.RetryDelay = 5 * time.Second
	}
	if cfg.ExecTimeout == 0 {
		cfg.ExecTimeout = 1 * time.Minute
	}
	return &Worker{repo: repo, cfg: cfg}
}

// Run starts the worker loop, which runs until Ctrl+C is pressed or context is canceled.
func (w *Worker) Run(ctx context.Context) error {
	log.Printf("[%s] started", w.cfg.ID)

	// Handle graceful shutdown
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt)
	defer stop()

	idleCount := 0 // adaptive backoff counter

	for {
		select {
		case <-ctx.Done():
			log.Printf("[%s] stopped successfully", w.cfg.ID)
			return nil
		default:
		}

		// STEP 1: Try to claim a pending job safely
		j, err := w.repo.PreventRaceCondition(w.cfg.ID)
		if err != nil {
			log.Printf("[%s] claim error: %v", w.cfg.ID, err)
			time.Sleep(w.cfg.PollInterval)
			continue
		}

		// STEP 2: If no job found → adaptive backoff sleep
		if j == nil {
			sleep := w.cfg.PollInterval * time.Duration(1<<idleCount)
			if sleep > w.cfg.MaxSleepTime {
				sleep = w.cfg.MaxSleepTime
			}
			log.Printf("[%s] idle (no jobs). Sleeping for %v...", w.cfg.ID, sleep)

			// Wait for either sleep timeout OR interrupt
			select {
			case <-time.After(sleep):
				// wake up normally
			case <-ctx.Done():
				log.Printf("[%s] shutdown received during sleep", w.cfg.ID)
				return nil
			}

			if idleCount < 5 {
				idleCount++
			}
			continue

		}

		// Reset idle count when a job is found
		idleCount = 0

		log.Printf("[%s] processing job %s (%s)", w.cfg.ID, j.ID, j.Command)

		// ✅ STEP 3: Execute the command with timeout
		result := w.ExecCommand(j.Command, w.cfg.ExecTimeout)

		// ✅ STEP 4: Handle success or failure
		if result.ExitCode == 0 && result.Err == nil {
			j.Output = result.Stdout + "\n" + result.Stderr
			j.Duration = result.Duration.Seconds()

			if err := w.repo.MarkCompleted(j); err != nil {
				log.Printf("[%s] error marking job complete: %v", w.cfg.ID, err)
			} else {
				log.Printf("[%s] job %s completed successfully in %.2fs", w.cfg.ID, j.ID, j.Duration)
			}
		} else {
			errMsg := result.Stderr
			if errMsg == "" && result.Err != nil {
				errMsg = result.Err.Error()
			}
			if err := w.repo.Failed(j, errMsg, w.cfg.RetryDelay); err != nil {
				log.Printf("[%s] error marking job failed: %v", w.cfg.ID, err)
			} else {
				log.Printf("[%s] job %s failed (retry or DLQ): %s", w.cfg.ID, j.ID, errMsg)
			}
		}
	}
}

// ✅ Timeout-aware command executor
func (w *Worker) ExecCommand(command string, timeout time.Duration) ExecResult {
	start := time.Now()

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "bash", "-c", command)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	duration := time.Since(start)

	result := ExecResult{
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		Err:      err,
		Duration: duration,
	}

	// Determine exit code
	if exitErr, ok := err.(*exec.ExitError); ok {
		result.ExitCode = exitErr.ExitCode()
	} else if err == nil {
		result.ExitCode = 0
	} else if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		result.Err = errors.New("job timeout exceeded")
		result.ExitCode = -1
		log.Printf("[%s] job timed out after %v", w.cfg.ID, timeout)
	}

	return result
}
