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
)

var workerCmd = &cobra.Command{
	Use:   "worker start",
	Short: "Start one or more background workers to process queued jobs",
	Long: `Start one or more queue workers.
Each worker continuously polls for new jobs and executes them.
Example:
  queuectl worker start --count 3 --timeout 30s`,
	Run: func(cmd *cobra.Command, args []string) {
		CommonInit()

		if workerCount <= 0 {
			workerCount = 1
		}

		db, err := store.InitDB()
		if err != nil {
			log.Fatalf("DB init failed: %v", err)
		}
		repo := store.NewJobRepo(db)

		// Create a cancelable context for graceful shutdown
		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
		defer stop()

		log.Printf("🚀 Starting %d worker(s) with timeout=%v", workerCount, timeoutFlag)

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
					RetryDelay:   5 * time.Second,
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
	rootCmd.AddCommand(workerCmd)
}
