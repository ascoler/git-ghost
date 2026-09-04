package cmd

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	database "git-ghost/internal/DataBase"
	"git-ghost/internal/backup"
	"git-ghost/internal/config"
	"git-ghost/internal/scanner"

	"github.com/go-git/go-git/v5"
	"github.com/kardianos/service"
	"github.com/spf13/cobra"
)


type program struct {
	cfg    config.Config
	token  string
	exit   chan struct{}
	logger service.Logger
}


func (p *program) Start(s service.Service) error {
	p.exit = make(chan struct{})
	go p.run()
	return nil
}


func (p *program) Stop(s service.Service) error {
	close(p.exit)
	return nil
}


func (p *program) run() {
	p.logger.Info("Git Ghost daemon started",
		"backup_repo", p.cfg.BackupRepo,
		"scan_interval", p.cfg.ScanInterval,
		"watch_dirs", len(p.cfg.WatchDirs),
	)

	ticker := time.NewTicker(time.Duration(p.cfg.ScanInterval) * time.Second)
	defer ticker.Stop()

	
	p.backupAll()

	for {
		select {
		case <-p.exit:
			p.logger.Info("Git Ghost daemon stopped")
			return
		case <-ticker.C:
			p.backupAll()
		}
	}
}


func (p *program) backupAll() {
	slog.Info("Scanning repositories")

	repos, err := scanner.ScanRepositories(p.cfg.WatchDirs)
	if err != nil {
		slog.Error("Failed to scan repositories", "error", err)
		return
	}

	if len(repos) == 0 {
		slog.Warn("No repositories found")
		return
	}

	slog.Info("Repositories found", "count", len(repos))

	dbPath := p.cfg.DBpath
	if dbPath == "" {
		
		_, defaultDbPath, err := config.GetDefaultPaths()
		if err != nil {
			slog.Error("Failed to get default DB path", "error", err)
			return
		}
		dbPath = defaultDbPath
	}

	db, err := database.Open(dbPath)
	if err != nil {
		slog.Error("Failed to open database", "error", err)
		return
	}
	defer db.Close()

	successCount := 0
	skippedCount := 0

	for _, repo := range repos {
		
		

		needsBackup, err := db.GetHashRepo(repo.Path, repo.CurrentCommit.Hash)
		if err != nil {
			slog.Error("Failed to check repo state", "path", repo.Path, "error", err)
			db.UpdateRepoState(repo.Path, "error", err.Error())
			continue
		}

		if !needsBackup {
			slog.Info("Skipped (no changes)", "path", repo.Path)
			skippedCount++
			continue
		}

		gitRepo, err := git.PlainOpen(repo.Path)
		if err != nil {
			slog.Error("Failed to open repository", "path", repo.Path, "error", err)
			db.UpdateRepoState(repo.Path, "error", err.Error())
			continue
		}

		repoName := filepath.Base(repo.Path)
		err = backup.BackupRepository(gitRepo, repoName, p.cfg.BackupRepo, p.token)
		if err != nil {
			slog.Error("Failed to backup", "path", repo.Path, "error", err)
			db.UpdateRepoState(repo.Path, "error", err.Error())
			continue
		}

		db.UpdateRepoState(repo.Path, "ok", "", repo.CurrentCommit.Hash)
		slog.Info("Backed up successfully", "path", repo.Path)
		successCount++
	}

	slog.Info("Backup cycle complete",
		"success", successCount,
		"skipped", skippedCount,
		"total", len(repos),
	)
}

var serviceCmd = &cobra.Command{
	Use:   "",
	Short: "Manage git-ghost as a system service (works on Linux, Windows, macOS)",
	Long: `Manage git-ghost as a system service.

On Linux: creates a systemd service
On Windows: creates a Windows Service
On macOS: creates a launchd service

Examples:
  git-ghost service install          Install the service
  git-ghost service start            Start the service
  git-ghost service stop             Stop the service
  git-ghost service status           Check service status
  git-ghost service run              Run in foreground (for debugging)`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		action := args[0]

		configPath, _ := cmd.Flags().GetString("config")
		if configPath == "" {
			defaultConfigPath, _, err := config.GetDefaultPaths()
			if err != nil {
				config.PrintError(fmt.Sprintf("Failed to get config path: %v", err))
				return
			}
			configPath = defaultConfigPath
		}

	
		var cfg config.Config
		needsConfig := action == "run" || action == "install"
		if needsConfig {
			loadedCfg, err := config.LoadConfig(configPath)
			if err != nil {
				config.PrintError(fmt.Sprintf("Config not found at %s. Run 'git-ghost init' first.", configPath))
				return
			}
			cfg = loadedCfg
		}

		
		token := os.Getenv("GIT_GHOST_TOKEN")
		if token == "" && action == "run" {
			config.PrintError("GIT_GHOST_TOKEN not set")
			return
		}

		
		execPath, err := os.Executable()
		if err != nil {
			config.PrintError(fmt.Sprintf("Failed to get executable path: %v", err))
			return
		}

		
		svcConfig := &service.Config{
			Name:        "git-ghost",
			DisplayName: "Git Ghost Backup Daemon",
			Description: "Automatic git backup daemon that backs up your repositories to GitHub",
			Arguments:   []string{"service", "run", "--config", configPath},
			EnvVars: map[string]string{
				"GIT_GHOST_TOKEN":       token,
				"GIT_GHOST_CONFIG_PATH": configPath,
			},
			Executable: execPath,
		}

		prg := &program{
			cfg:   cfg,
			token: token,
		}

		s, err := service.New(prg, svcConfig)
		if err != nil {
			config.PrintError(fmt.Sprintf("Failed to create service: %v", err))
			return
		}

		prg.logger, err = s.Logger(nil)
		if err != nil {
			config.PrintError(fmt.Sprintf("Failed to get service logger: %v", err))
			return
		}

		switch action {
		case "install":
			if err := s.Install(); err != nil {
				config.PrintError(fmt.Sprintf("Install failed: %v", err))
				return
			}
			config.PrintSuccess("Service installed")
			config.PrintInfo("Now run: git-ghost service start")

		case "uninstall":
			if err := s.Uninstall(); err != nil {
				config.PrintError(fmt.Sprintf("Uninstall failed: %v", err))
				return
			}
			config.PrintSuccess("Service removed")

		case "start":
			if err := s.Start(); err != nil {
				config.PrintError(fmt.Sprintf("Start failed: %v", err))
				return
			}
			config.PrintSuccess("Service started")
			config.PrintInfo("Check logs with: git-ghost service status")

		case "stop":
			if err := s.Stop(); err != nil {
				config.PrintError(fmt.Sprintf("Stop failed: %v", err))
				return
			}
			config.PrintSuccess("Service stopped")

		case "restart":
			if err := s.Restart(); err != nil {
				config.PrintError(fmt.Sprintf("Restart failed: %v", err))
				return
			}
			config.PrintSuccess("Service restarted")

		case "status":
			status, err := s.Status()
			if err != nil {
				config.PrintError(fmt.Sprintf("Failed to get status: %v", err))
				return
			}
			switch status {
			case service.StatusRunning:
				config.PrintSuccess("Service is RUNNING")
			case service.StatusStopped:
				config.PrintWarning("Service is STOPPED")
			default:
				config.PrintInfo("Service status: unknown")
			}

		case "run":
			
			config.PrintInfo("Running in foreground mode...")
			if err := s.Run(); err != nil {
				config.PrintError(fmt.Sprintf("Run failed: %v", err))
				return
			}

		default:
			config.PrintError("Unknown action. Use: install|uninstall|start|stop|restart|status|run")
		}
	},
}

func init() {
	serviceCmd.Flags().String("config", "", "Path to config file")
	rootCmd.AddCommand(serviceCmd)
}