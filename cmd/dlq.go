// cmd/dlq.go
package cmd

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"queuectl.backend/internal/job"
	"queuectl.backend/internal/store"
)

var retryID string

var dlqCmd = &cobra.Command{
	Use:   "dlq",
	Short: "View or retry jobs in the Dead Letter Queue",
	Long: `Examples:
  queuectl dlq
  queuectl dlq --retry <job-id>`,
	Run: func(cmd *cobra.Command, args []string) {
		CommonInit()
		db, err := store.InitDB()
		if err != nil {
			log.Fatal("DB init failed:", err)
		}
		repo := store.NewJobRepo(db)

		if retryID != "" {
			// fetch dead job by id
			var j job.Job
			if err := repo.DB().Where("id = ? AND state = ?", retryID, job.StateDead).First(&j).Error; err != nil {
				log.Fatalf("DLQ retry failed: %v", err)
			}
			// reset and move back to pending
			j.State = job.StatePending
			j.Attempts = 0
			j.RunAt = nil
			j.LastError = nil
			if err := repo.Update(&j); err != nil {
				log.Fatalf("DLQ retry update failed: %v", err)
			}
			fmt.Printf("DLQ: job %s moved back to pending\n", j.ID)
			return
		}

		// list DLQ
		dead := job.StateDead
		jobs, err := repo.ListJobs([]job.JobState{dead}, 200, 0, true)
		if err != nil {
			log.Fatal("Failed to list DLQ:", err)
		}
		if len(jobs) == 0 {
			fmt.Println("DLQ is empty")
			return
		}
		fmt.Println("Dead Letter Queue:")
		for _, j := range jobs {
			msg := ""
			if j.LastError != nil {
				msg = *j.LastError
			}
			fmt.Printf("- %s | %s | attempts %d/%d | last_error: %s\n",
				j.ID, j.Command, j.Attempts, j.MaxRetries, msg)
		}
	},
}

func init() {
	dlqCmd.Flags().StringVar(&retryID, "retry", "", "retry a specific DLQ job id (moves it back to pending)")
	rootCmd.AddCommand(dlqCmd)
}
