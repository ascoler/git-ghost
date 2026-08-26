package cmd

import (
	"fmt"
	"log/slog"

	"git-ghost/internal/config"
	"git-ghost/internal/scanner"

	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Scan a Git repository for branches and commits",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Git-Ghost status")
		fmt.Println("--------------------------")
		path,_, err := config.GetDefaultPaths()
		if err != nil {
			slog.Error("Failed to get config path", "error", err)
			return
		}
		config, err := config.LoadConfig(path)
		if err != nil {
			slog.Error("Failed to load configuration", "error", err)
			return
		}
		repositories, err := scanner.ScanRepositories(config.WatchDirs)
		if err != nil {
			slog.Error("Failed to scan repositories", "error", err)
			return
		}
		for i := 0; i < len(repositories); i++ {
			fmt.Println(i+1)
			fmt.Println("--------------------------")
			repo := repositories[i]
			fmt.Printf("Repository: %s\n", repo.Path)
			fmt.Println("Branches:")
			for _, branch := range repo.Branches {
				fmt.Printf("  - %s (%s)\n", branch.Name, branch.Hash[:8])
			}
			fmt.Println("Current Commit:")
			fmt.Printf("  - %s: %s by %s on %s\n", repo.CurrentCommit.Hash[:8], repo.CurrentCommit.Message, repo.CurrentCommit.Author, repo.CurrentCommit.Date)
			fmt.Println("--------------------------")
			
		}
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}