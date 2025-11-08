package cmd

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"queuectl.backend/internal/job"
	"queuectl.backend/internal/store"
)

var (
	stateFilter string
	showOutput  bool
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List jobs by state (pending, processing, completed, failed, dead)",
	Long:  "Display jobs in the queue with optional output display.",
	Run: func(cmd *cobra.Command, args []string) {
		CommonInit()
		db, err := store.InitDB()
		if err != nil {
			log.Fatal("DB init failed:", err)
		}
		repo := store.NewJobRepo(db)

		var states []job.JobState
		if stateFilter != "" {
			states = append(states, job.JobState(stateFilter))
		}

		jobs, err := repo.ListJobs(states, 100, 0, true)
		if err != nil {
			log.Fatal("Failed to list jobs:", err)
		}

		fmt.Printf("Listing jobs (state=%v):\n", stateFilter)
		for _, j := range jobs {
			fmt.Printf("- [%s] %s | Attempts: %d/%d | State: %s\n",
				j.ID, j.Command, j.Attempts, j.MaxRetries, j.State)
			if showOutput && j.Output != "" {
				fmt.Printf("  Output:\n%s\n", j.Output)
			}
		}
	},
}

func init() {
	listCmd.Flags().StringVarP(&stateFilter, "state", "s", "", "filter by job state")
	listCmd.Flags().BoolVarP(&showOutput, "show-output", "o", false, "display job output") // ✅ add this line
	rootCmd.AddCommand(listCmd)
}
