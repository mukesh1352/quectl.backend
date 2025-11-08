package cmd

import (
	"fmt"
	"log"

	"queuectl.backend/internal/job"
	"github.com/spf13/cobra"
)

var stateFilter string

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List jobs by state (pending, processing, completed, failed, dead)",
	Long:  "Display jobs on the queue",
	Run: func(cmd *cobra.Command, args []string) {
		CommonInit()

		var states []job.JobState
		if stateFilter != "" {
			states = append(states, job.JobState(stateFilter))
		}

		jobs, err := repo.ListJobs(states, 100, 0, true)
		if err != nil {
			log.Fatal("Failed to list jobs:", err)
		}

		fmt.Printf("📋 Listing jobs (state=%v):\n", stateFilter)
		for _, j := range jobs {
			fmt.Printf("- [%s] %s | Attempts: %d/%d | State: %s\n",
				j.ID, j.Command, j.Attempts, j.MaxRetries, j.State)
		}
	},
}

func init() {
	listCmd.Flags().StringVarP(&stateFilter, "state", "s", "", "filter by job state")
	rootCmd.AddCommand(listCmd)
}
