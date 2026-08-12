package config

import (
	"gopkg.in/yaml.v3"
	"log/slog"
	"os"
)

type Config struct {
	BackupRepo string `yaml:"backup_repo"`
	WatchDirs []string `yaml:"watch_dirs"`
	ScanInterval int `yaml:"scan_interval"`
}

func LoadConfig(path string) (Config, error) {
	data,err := os.ReadFile(path)
	if err != nil {
		slog.Error("Failed to read config file", "error", err)
		return Config{}, err
	}
	var config Config
	err = yaml.Unmarshal(data,&config)
	if err != nil {
		slog.Error("Failed to unmarshal config file", "error", err)
		return Config{}, err
	}

	return config,nil



}