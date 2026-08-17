package cmd

import (
	"fmt"
	"git-ghost/internal/config"
	"log/slog"

	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the git-ghost daemon",
	Run: func(cmd *cobra.Command, args []string) {
		_, err := config.InitConfig()
		if err != nil {
			fmt.Println("Failed to initialize configuration:", err)
			slog.Error("Failed to initialize configuration", "error", err)
			return
		}
		fmt.Println("Configuration initialized successfully.")
	},
}

func init() {
	rootCmd.AddCommand(startCmd)
}
