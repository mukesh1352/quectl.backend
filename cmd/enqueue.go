package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"queuectl.backend/internal/job"
	"github.com/spf13/cobra"
)

// enqueueCmd represents the "enqueue" command
var enqueueCmd = &cobra.Command{
	Use:   "enqueue [job-json]",
	Short: "Add a new job to the queue",
	Long: `Add a new job to the background queue.
Example:
  queuectl enqueue '{"command":"echo Hello World"}'`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		CommonInit()
		var j job.Job
		if err := json.Unmarshal([]byte(args[0]), &j); err != nil {
			log.Fatalf("Invalid job JSON: %v", err)
		}

		// Apply defaults if not provided
		if j.ID == "" {
			j.ID = fmt.Sprintf("job-%d", time.Now().UnixNano())
		}
		if j.State == "" {
			j.State = job.StatePending
		}
		if j.MaxRetries == 0 {
			j.MaxRetries = 3
		}
		j.CreatedAt = time.Now().UTC()
		j.UpdatedAt = j.CreatedAt

		if err := repo.Create(&j); err != nil {
			log.Fatalf("Failed to enqueue job: %v", err)
		}

		fmt.Printf("Job %s enqueued successfully.\n", j.ID)
	},
}

func init() {
	rootCmd.AddCommand(enqueueCmd)
}
