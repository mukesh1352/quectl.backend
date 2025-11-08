package cmd

import (
	"context"
	"fmt"

	"queuectl.backend/internal/queue"
	"github.com/spf13/cobra"
)

var workerCount int

var workerCmd = &cobra.Command{
	Use:   "worker start",
	Short: "Start one or more background workers",
	Run: func(cmd *cobra.Command, args []string) {
		CommonInit()

		if workerCount <= 0 {
			workerCount = 1
		}

		ctx := context.Background()
		for i := 1; i <= workerCount; i++ {
			id := fmt.Sprintf("worker-%d", i)
			w := queue.NewWorker(repo, queue.WorkerConfig{ID: id})
			go w.Run(ctx)
		}

		select {} // Keep running
	},
}

func init() {
	workerCmd.Flags().IntVarP(&workerCount, "count", "c", 1, "number of workers to start")
	rootCmd.AddCommand(workerCmd)
}
