package cmd

import (
	"fmt"
	"log"

	"queuectl.backend/internal/job"
	"github.com/spf13/cobra"
)

// statusCmd gives a quick summary of queue state.
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show summary of all job states and active workers",
	Long:  "Displays the number of jobs in each state to monitor queue health.",
	Run: func(cmd *cobra.Command, args []string) {
		CommonInit()

		states := []job.JobState{
			job.StatePending,
			job.StateProcessing,
			job.StateCompleted,
			job.StateFailed,
			job.StateDead,
		}

		counts := make(map[job.JobState]int64)
		var total int64

		for _, state := range states {
			var count int64
			if err := repo.DB().Model(&job.Job{}).
				Where("state = ?", state).
				Count(&count).Error; err != nil {
				log.Fatalf("Failed to count jobs for state %s: %v", state, err)
			}
			counts[state] = count
			total += count
		}

		fmt.Println("Job Queue Status:")
		fmt.Printf("Total Jobs: %d\n", total)
		fmt.Printf("Pending: %d\n", counts[job.StatePending])
		fmt.Printf("Processing: %d\n", counts[job.StateProcessing])
		fmt.Printf("Completed: %d\n", counts[job.StateCompleted])
		fmt.Printf("Failed: %d\n", counts[job.StateFailed])
		fmt.Printf("Dead (DLQ): %d\n", counts[job.StateDead])
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
