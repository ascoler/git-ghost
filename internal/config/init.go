package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
	"log/slog"
)

func InitConfig(configDir string) (*Config, error) {
	reader := bufio.NewReader(os.Stdin)

	
	defaultBackupRepo := ""
	home, _ := os.UserHomeDir()
	defaultWatchDir := filepath.Join(home, "projects")
	defaultScanInterval := 300 

	
	fmt.Printf("Enter backup repository URL: ")
	backupRepo, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("failed to read input: %w", err)
	}
	backupRepo = strings.TrimSpace(backupRepo)

	if backupRepo == "" {
		backupRepo = defaultBackupRepo
	}

	if backupRepo == "" {
		return nil, fmt.Errorf("backup_repo is required")
	}

	
	fmt.Printf("Enter directories to watch (comma-separated, default: %s): ", defaultWatchDir)
	watchDirsInput, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("failed to read input: %w", err)
	}
	watchDirsInput = strings.TrimSpace(watchDirsInput)

	var watchDirs []string
	if watchDirsInput == "" {
		watchDirs = []string{defaultWatchDir}
	} else {
		for _, dir := range strings.Split(watchDirsInput, ",") {
			dir = strings.TrimSpace(dir)
			if dir != "" {
				
				if strings.HasPrefix(dir, "~") {
					dir = filepath.Join(home, dir[1:])
				}
				watchDirs = append(watchDirs, dir)
			}
		}
	}

	if len(watchDirs) == 0 {
		return nil, fmt.Errorf("watch_dirs cannot be empty")
	}

	
	fmt.Printf("Enter scan interval in seconds (default: %d): ", defaultScanInterval)
	scanIntervalInput, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("failed to read input: %w", err)
	}
	scanIntervalInput = strings.TrimSpace(scanIntervalInput)

	scanInterval := defaultScanInterval
	if scanIntervalInput != "" {
		parsed, err := strconv.Atoi(scanIntervalInput)
		if err != nil {
			slog.Warn("Invalid scan interval, using default", "default", defaultScanInterval)
		} else if parsed < 10 {
			slog.Warn("Scan interval too small, using default", "default", defaultScanInterval)
		} else {
			scanInterval = parsed
		}
	}

	
	config := &Config{
		BackupRepo:   backupRepo,
		WatchDirs:    watchDirs,
		ScanInterval: scanInterval,
	}

	
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create config directory: %w", err)
	}

	
	configData, err := yaml.Marshal(config)
	if err != nil {
		slog.Error("Failed to marshal config data", "error", err)
		return nil, fmt.Errorf("failed to marshal config: %w", err)
	}

	
	configFilePath := filepath.Join(configDir, "config.yaml")
	if err := os.WriteFile(configFilePath, configData, 0644); err != nil {
		slog.Error("Failed to write config file", "error", err)
		return nil, fmt.Errorf("failed to write config file: %w", err)
	}

	slog.Info("Config saved", "path", configFilePath)

	return config, nil
}