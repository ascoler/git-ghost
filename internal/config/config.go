package config

import (
	"log/slog"
	"os"
	"path/filepath"
	"runtime"

	"gopkg.in/yaml.v3"
)

type Config struct {
	BackupRepo   string   `yaml:"backup_repo"`
	WatchDirs    []string `yaml:"watch_dirs"`
	ScanInterval int      `yaml:"scan_interval"`
	DBpath string `yaml:"dbpath"`
}

func LoadConfig(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		slog.Error("Failed to read config file", "error", err)
		return Config{}, err
	}
	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		slog.Error("Failed to unmarshal config file", "error", err)
		return Config{}, err
	}

	return config, nil
}


func GetDefaultConfigPath() (configPath string, dbPath string, err error) {
	var configDir string

	switch runtime.GOOS {
	case "windows":
		appData := os.Getenv("APPDATA")
		if appData == "" {
			home, err := os.UserHomeDir()
			if err != nil {
				return "", "", err
			}
			appData = filepath.Join(home, "AppData", "Roaming")
		}
		configDir = filepath.Join(appData, "git-ghost")

	default:
		home, err := os.UserHomeDir()
		if err != nil {
			return "", "", err
		}
		configDir = filepath.Join(home, ".config", "git-ghost")
	}

	configPath = filepath.Join(configDir, "config.yaml")
	dbPath = filepath.Join(configDir, "git-ghost.db")

	return configPath, dbPath, nil
}