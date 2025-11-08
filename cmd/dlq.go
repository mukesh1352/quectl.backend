package cmd

import (
	"fmt"
	"log"

	"queuectl.backend/internal/job"
	"github.com/spf13/cobra"
)

var dlqCmd = &cobra.Command{
	Use:   "dlq",
	Short: "View dead letter queue (failed jobs)",
	Long: `Inspect or retry jobs that have permanently failed.

Examples:
  queuectl dlq                  # View DLQ jobs
  queuectl dlq --retry <job-id> # Retry a specific DLQ job`,
	Run: func(cmd *cobra.Command, args []string) {
		CommonInit()

		deadState := job.StateDead
		jobs, err := repo.ListJobs([]job.JobState{deadState}, 100, 0, true)
		if err != nil {
			log.Fatal("Failed to list DLQ jobs:", err)
		}

		if len(jobs) == 0 {
			fmt.Println("✅ DLQ is empty!")
			return
		}

		fmt.Println("🪦 Dead Letter Queue:")
		for _, j := range jobs {
			errMsg := ""
			if j.LastError != nil {
				errMsg = *j.LastError
			}
			fmt.Printf("- [%s] %s | Last Error: %s\n", j.ID, j.Command, errMsg)
		}
	},
}

func init() {
	rootCmd.AddCommand(dlqCmd)
}
