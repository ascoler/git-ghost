package cmd

import (
	"fmt"
	
	"os"
	"time"

	"git-ghost/internal/backup"
	"git-ghost/internal/config"

	"github.com/spf13/cobra"
)

var restoreCmd = &cobra.Command{
	Use:   "restore",
	Short: "Restore a specific backup version",
	Run: func(cmd *cobra.Command, args []string) {
		repoName, _ := cmd.Flags().GetString("repo")
		dateStr, _ := cmd.Flags().GetString("date")
		branch, _ := cmd.Flags().GetString("branch")

		if repoName == "" || dateStr == "" || branch == "" {
			config.PrintError("--repo, --date, --branch are required")
			cmd.Help()
			return
		}

		date, err := time.Parse("2006-01-02_15-04", dateStr)
		if err != nil {
			config.PrintError(fmt.Sprintf("Invalid date format, use YYYY-MM-DD_HH-MM: %v", err))
			return
		}

		configPath, _, err := config.GetDefaultPaths()
		if err != nil {
			config.PrintError(fmt.Sprintf("Failed to get paths: %v", err))
			return
		}

		cfg, err := config.LoadConfig(configPath)
		if err != nil {
			config.PrintError("Config not found. Run 'git-ghost init' first.")
			return
		}

		token := os.Getenv("GIT_GHOST_TOKEN")

		path, err := backup.GetVersion(&cfg, repoName, date, branch, token)
		if err != nil {
			config.PrintError(fmt.Sprintf("Restore failed: %v", err))
			return
		}

		config.PrintSuccess(fmt.Sprintf("Backup restored to: %s", path))
	},
}

func init() {
	restoreCmd.Flags().String("repo", "", "Repository name to restore")
	restoreCmd.Flags().String("date", "", "Backup date (format: YYYY-MM-DD_HH-MM)")
	restoreCmd.Flags().String("branch", "main", "Branch to restore")
	rootCmd.AddCommand(restoreCmd)
}