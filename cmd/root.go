package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "queuectl",
	Short: "queuectl - a CLI background job manager",
	Long: `queuectl lets you enqueue and manage background jobs with
workers, retries, and a dead-letter queue.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Welcome to queuectl 🎯")
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
