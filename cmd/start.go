package cmd

import (
	"context"
	"fmt"
	
	"os"
	"os/signal"
	"path/filepath"
	"time"

	database "git-ghost/internal/DataBase"
	"git-ghost/internal/backup"
	"git-ghost/internal/config"
	"git-ghost/internal/scanner"

	"github.com/go-git/go-git/v5"
	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the git-ghost daemon",
	Run: func(cmd *cobra.Command, args []string) {
		config.PrintLogo()

		configPath,_, err := config.GetDefaultPaths()
		if err != nil {
			config.PrintError(fmt.Sprintf("Failed to get config path: %v", err))
			return
		}

		cfg, err := config.LoadConfig(configPath)
		if err != nil {
			config.PrintInfo("Config not found, creating...")
			newCfg, newErr := config.InitConfig()
			if newErr != nil {
				config.PrintError(fmt.Sprintf("Failed to create config: %v", newErr))
				return
			}
			cfg = *newCfg
		} else {
			config.PrintSuccess("Config loaded successfully")
		}

		token := os.Getenv("GIT_GHOST_TOKEN")
		if token == "" {
			fmt.Println()
			config.PrintWarning("GIT_GHOST_TOKEN not set")
			fmt.Println()
			fmt.Printf("  Create token: %shttps://github.com/settings/tokens/new%s\n", config.GetColorCyan(), config.GetColorReset())
			fmt.Printf("  Then run: %sexport GIT_GHOST_TOKEN=\"ghp_...\"%s\n\n", config.GetColorCyan(), config.GetColorReset())
			return
		}

		config.PrintSuccess("GitHub token found")
		fmt.Println()
		config.PrintHeader("Daemon Starting")
		config.PrintInfo(fmt.Sprintf("Backup repo: %s", cfg.BackupRepo))
		config.PrintInfo(fmt.Sprintf("Scan interval: %d seconds", cfg.ScanInterval))
		config.PrintInfo(fmt.Sprintf("Watching: %d directories", len(cfg.WatchDirs)))
		fmt.Println()
		config.PrintInfo("Press Ctrl+C to stop")
		fmt.Println()

		ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
		defer cancel()

		ticker := time.NewTicker(time.Duration(cfg.ScanInterval) * time.Second)
		defer ticker.Stop()

	
		backupAll(cfg, token)

		for {
			select {
			case <-ctx.Done():
				fmt.Println()
				config.PrintWarning("Shutting down gracefully...")
				return
			case <-ticker.C:
				backupAll(cfg, token)
			}
		}
	},
}

func backupAll(cfg config.Config, token string) {
    fmt.Printf("\n[%s] Scanning repositories...\n", time.Now().Format("15:04:05"))

    repos, err := scanner.ScanRepositories(cfg.WatchDirs)
    if err != nil {
        config.PrintError(fmt.Sprintf("Failed to scan repositories: %v", err))
        return
    }

    if len(repos) == 0 {
        config.PrintWarning("No repositories found")
        return
    }

    config.PrintInfo(fmt.Sprintf("Found %d repositories", len(repos)))

    db, err := database.Open(cfg.DBpath)
    if err != nil {
        config.PrintError(fmt.Sprintf("Failed to open database: %v", err))
        return
    }
    defer db.Close()

    successCount := 0
    skippedCount := 0

    for _, repo := range repos {
        
		
        needsBackup, err := db.GetHashRepo(repo.Path, repo.CurrentCommit.Hash)
        if err != nil {
            config.PrintError(fmt.Sprintf("DB error for %s: %v", repo.Path, err))
            continue
        }

        if !needsBackup {
            config.PrintInfo(fmt.Sprintf("Skipped (no changes): %s", repo.Path))
            skippedCount++
            continue
        }

        
        gitRepo, err := git.PlainOpen(repo.Path)
        if err != nil {
            config.PrintError(fmt.Sprintf("Failed to open %s: %v", repo.Path, err))
            db.UpdateRepoState(repo.Path, "error", err.Error())
            continue
        }

        
        repoName := filepath.Base(repo.Path)
        err = backup.BackupRepository(gitRepo, repoName, cfg.BackupRepo, token)
        if err != nil {
            config.PrintError(fmt.Sprintf("Failed to backup %s: %v", repo.Path, err))
            db.UpdateRepoState(repo.Path, "error", err.Error())
            continue
        }

        
        db.UpdateRepoState(repo.Path, "ok", "", repo.CurrentCommit.Hash)
        config.PrintSuccess(fmt.Sprintf("Backed up: %s", repo.Path))
        successCount++
    }

    if skippedCount > 0 {
        config.PrintInfo(fmt.Sprintf("Skipped %d repositories (no changes)", skippedCount))
    }

    if successCount > 0 {
        fmt.Println()
        config.PrintSuccess(fmt.Sprintf("Backup complete: %d/%d repositories", successCount, len(repos)))
    }
}

func init() {
	rootCmd.AddCommand(startCmd)
}