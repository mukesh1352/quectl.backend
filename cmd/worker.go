package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/spf13/cobra"
	"queuectl.backend/internal/queue"
	"queuectl.backend/internal/store"
)

var (
	workerCount int
	timeoutFlag time.Duration
	backoffBase time.Duration
)

var workerCmd = &cobra.Command{
	Use:   "worker start",
	Short: "Start one or more background workers to process queued jobs",
	Long: `Start one or more queue workers.
Each worker continuously polls for new jobs and executes them.

Examples:
  queuectl worker start --count 3 --timeout 30s
  queuectl worker start --count 2 --backoff-base 2s`,
	Run: func(cmd *cobra.Command, args []string) {
		CommonInit()

		if workerCount <= 0 {
			workerCount = 1
		}
		if backoffBase <= 0 {
			backoffBase = 5 * time.Second
		}

		db, err := store.InitDB()
		if err != nil {
			log.Fatalf("DB init failed: %v", err)
		}
		repo := store.NewJobRepo(db)

		// Create cancelable context for graceful shutdown
		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
		defer stop()

		log.Printf("🚀 Starting %d worker(s) | timeout=%v | backoff-base=%v", workerCount, timeoutFlag, backoffBase)

		var wg sync.WaitGroup
		for i := 1; i <= workerCount; i++ {
			wg.Add(1)
			id := fmt.Sprintf("worker-%d", i)

			go func(workerID string) {
				defer wg.Done()
				worker := queue.NewWorker(repo, queue.WorkerConfig{
					ID:           workerID,
					PollInterval: 2 * time.Second,
					MaxSleepTime: 30 * time.Second,
					RetryDelay:   backoffBase, // ✅ configurable base delay
					ExecTimeout:  timeoutFlag,
				})
				if err := worker.Run(ctx); err != nil {
					log.Printf("[%s] exited with error: %v", workerID, err)
				}
			}(id)
		}

		// Wait for all workers to finish gracefully
		wg.Wait()
		log.Println("All workers stopped gracefully.")
	},
}

func init() {
	workerCmd.Flags().IntVarP(&workerCount, "count", "c", 1, "number of workers to start")
	workerCmd.Flags().DurationVar(&timeoutFlag, "timeout", time.Minute, "maximum execution time per job (e.g., 30s, 2m)")
	workerCmd.Flags().DurationVar(&backoffBase, "backoff-base", 5*time.Second, "base retry backoff duration (e.g., 2s, 5s, 10s)")
	rootCmd.AddCommand(workerCmd)
}
