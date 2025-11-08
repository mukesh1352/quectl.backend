package cmd

import (
	"fmt"
	"log"

	"queuectl.backend/internal/config"
	"github.com/spf13/cobra"
)

var cfgKey, cfgValue string

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "View or modify queue configuration",
}

var setCmd = &cobra.Command{
	Use:   "set",
	Short: "Set a configuration value",
	Run: func(cmd *cobra.Command, args []string) {
		CommonInit()
		repoCfg := config.NewRepository(repo.DB())
		if err := repoCfg.Set(cfgKey, cfgValue); err != nil {
			log.Fatalf("Failed to set config: %v", err)
		}
		fmt.Printf("%s set to %s\n", cfgKey, cfgValue)
	},
}

var viewCmd = &cobra.Command{
	Use:   "view",
	Short: "View all configuration values",
	Run: func(cmd *cobra.Command, args []string) {
		CommonInit()
		repoCfg := config.NewRepository(repo.DB())
		items, err := repoCfg.All()
		if err != nil {
			log.Fatalf("Failed to fetch config: %v", err)
		}
		if len(items) == 0 {
			fmt.Println("No configuration values set.")
			return
		}
		fmt.Println("Current Configuration:")
		for _, i := range items {
			fmt.Printf("%s = %s\n", i.Key, i.Value)
		}
	},
}

func init() {
	setCmd.Flags().StringVar(&cfgKey, "key", "", "Configuration key")
	setCmd.Flags().StringVar(&cfgValue, "value", "", "Configuration value")
	setCmd.MarkFlagRequired("key")
	setCmd.MarkFlagRequired("value")

	configCmd.AddCommand(setCmd)
	configCmd.AddCommand(viewCmd)
	rootCmd.AddCommand(configCmd)
}
