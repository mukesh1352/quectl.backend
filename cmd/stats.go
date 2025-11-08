package cmd

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"queuectl.backend/internal/store"
)

// statsCmd displays queue metrics and performance stats.
var statsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Show aggregated job metrics and performance stats",
	Long:  "Displays total jobs, per-state counts, average duration, and retry statistics.",
	Run: func(cmd *cobra.Command, args []string) {
		CommonInit()
		db, err := store.InitDB()
		if err != nil {
			log.Fatal("DB init failed:", err)
		}
		repo := store.NewJobRepo(db)

		summary, err := repo.JobMetrics()
		if err != nil {
			log.Fatalf("Failed to get metrics: %v", err)
		}

		fmt.Println("\n📊 Queue Metrics Summary")
		fmt.Println("----------------------------")
		fmt.Printf("Total Jobs:       %d\n", summary.Total)
		fmt.Printf("Pending:          %d\n", summary.Pending)
		fmt.Printf("Processing:       %d\n", summary.Processing)
		fmt.Printf("Completed:        %d\n", summary.Completed)
		fmt.Printf("Failed:           %d\n", summary.Failed)
		fmt.Printf("Dead (DLQ):       %d\n", summary.Dead)
		fmt.Printf("Avg Duration:     %.2fs\n", summary.AvgDuration)
		fmt.Printf("Avg Retries/job:  %.2f\n", summary.AvgRetries)
		fmt.Println("----------------------------")
	},
}

func init() {
	rootCmd.AddCommand(statsCmd)
}
