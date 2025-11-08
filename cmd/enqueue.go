package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/spf13/cobra"
	"queuectl.backend/internal/job"
)

// enqueueCmd represents the "enqueue" command
var enqueueCmd = &cobra.Command{
	Use:   "enqueue [job-json]",
	Short: "Add a new job to the queue (supports scheduling and priority)",
	Long: `Add a new job to the background queue.

Examples:
  queuectl enqueue '{"command":"echo Hello World"}'
  queuectl enqueue '{"command":"echo High Priority"}' --priority 10
  queuectl enqueue '{"command":"echo Run Later"}' --delay 30s
  queuectl enqueue '{"command":"echo Scheduled"}' --run-at "2025-11-09T01:00:00Z"`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		CommonInit()

		var j job.Job
		if err := json.Unmarshal([]byte(args[0]), &j); err != nil {
			log.Fatalf("Invalid job JSON: %v", err)
		}

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

		priority, _ := cmd.Flags().GetInt("priority")
		j.Priority = priority

		delay, _ := cmd.Flags().GetDuration("delay")
		if delay > 0 {
			runAt := time.Now().Add(delay).UTC()
			j.RunAt = &runAt
		}

		runAtStr, _ := cmd.Flags().GetString("run-at")
		if runAtStr != "" {
			parsedTime, err := time.Parse(time.RFC3339, runAtStr)
			if err != nil {
				log.Fatalf("Invalid --run-at value, must use RFC3339 format (e.g., 2025-11-09T01:00:00Z): %v", err)
			}
			j.RunAt = &parsedTime
		}

		// Save to DB
		if err := repo.Create(&j); err != nil {
			log.Fatalf("Failed to enqueue job: %v", err)
		}

		// Friendly output
		fmt.Printf("Job %s enqueued successfully", j.ID)
		if j.Priority > 0 {
			fmt.Printf(" (priority=%d)", j.Priority)
		}
		if j.RunAt != nil {
			fmt.Printf(" (scheduled for %s)", j.RunAt.Format(time.RFC3339))
		} else {
			fmt.Print(" (immediate execution)")
		}
		fmt.Println()
	},
}

func init() {
	enqueueCmd.Flags().IntP("priority", "p", 0, "set job priority (higher = more important)")
	enqueueCmd.Flags().Duration("delay", 0, "schedule job to run after a delay (e.g., 10s, 1m, 2h)")
	enqueueCmd.Flags().String("run-at", "", "specific time to run the job (RFC3339 format, e.g., 2025-11-09T01:00:00Z)")
	rootCmd.AddCommand(enqueueCmd)
}
