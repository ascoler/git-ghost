package config

import (
	"bufio"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	colorReset   = "\033[0m"
	colorRed     = "\033[31m"
	colorGreen   = "\033[32m"
	colorYellow  = "\033[33m"
	colorBlue    = "\033[34m"
	colorMagenta = "\033[35m"
	colorCyan    = "\033[36m"
	colorGray    = "\033[90m"
	colorBold    = "\033[1m"
)

func GetColorCyan() string {
	return colorCyan
}

func GetColorReset() string {
	return colorReset
}
const logo = `
` + colorMagenta + colorBold + ` ██████╗ ██╗████████╗      ██████╗ ██╗  ██╗ ██████╗ ███████╗████████╗
██╔════╝ ██║╚══██╔══╝     ██╔════╝ ██║  ██║██╔═══██╗██╔════╝╚══██╔══╝
██║  ███╗██║   ██║        ██║  ███╗███████║██║   ██║███████╗   ██║   
██║   ██║██║   ██║        ██║   ██║██╔══██║██║   ██║╚════██║   ██║   
╚██████╔╝██║   ██║        ╚██████╔╝██║  ██║╚██████╔╝███████║   ██║   
 ╚═════╝ ╚═╝   ╚═╝         ╚═════╝ ╚═╝  ╚═╝ ╚═════╝ ╚══════╝   ╚═╝` + colorReset + `
`

func PrintHeader(msg string) {
	fmt.Printf("\n%s%s%s%s\n", colorBold, colorCyan, msg, colorReset)
	fmt.Printf("%s%s%s\n", colorCyan, strings.Repeat("─", len(msg)), colorReset)
}
func PrintLogo() {
	fmt.Print(logo)
	fmt.Printf("%s%sAutomatic git backup daemon%s\n\n", colorGray, colorBold, colorReset)
}

func PrintPrompt(label string, hint string) {
	if hint != "" {
		fmt.Printf("%s%s▶ %s%s %s%s%s\n", colorBold, colorYellow, label, colorReset, colorGray, hint, colorReset)
	} else {
		fmt.Printf("%s%s▶ %s%s\n", colorBold, colorYellow, label, colorReset)
	}
	fmt.Printf("  ")
}


func PrintSuccess(msg string) {
	fmt.Printf("%s%s✓ %s%s\n", colorBold, colorGreen, msg, colorReset)
}


func PrintInfo(msg string) {
	fmt.Printf("%s%sℹ %s%s\n", colorBold, colorBlue, msg, colorReset)
}


func PrintWarning(msg string) {
	fmt.Printf("%s%s⚠ %s%s\n", colorBold, colorYellow, msg, colorReset)
}


func PrintError(msg string) {
	fmt.Printf("%s%s✗ %s%s\n", colorBold, colorRed, msg, colorReset)
}


func InitConfig() (*Config, error) {
	reader := bufio.NewReader(os.Stdin)
	configPath,err := GetDefaultConfigPath()
	if err != nil {
		return nil, fmt.Errorf("failed to get config directory: %w", err)
	}
	configDir := filepath.Dir(configPath)
	
	fmt.Print(logo)
	fmt.Printf("%s%sAutomatic git backup daemon%s\n\n", colorGray, colorBold, colorReset)

	PrintHeader("Welcome to Git Ghost Setup")
	PrintInfo("Let's configure your backup daemon")

	
	home, _ := os.UserHomeDir()
	defaultWatchDir := filepath.Join(home, "projects")
	defaultScanInterval := 300 

	
	fmt.Println()
	PrintPrompt("Backup repository URL", "(e.g. https://github.com/user/git-ghost-backups.git)")
	backupRepo, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("failed to read input: %w", err)
	}
	backupRepo = strings.TrimSpace(backupRepo)

	if backupRepo == "" {
		PrintError("Backup repository URL is required")
		return nil, fmt.Errorf("backup_repo is required")
	}
	PrintSuccess("Backup repository set")

	
	fmt.Println()
	PrintPrompt("Directories to watch", fmt.Sprintf("(default: %s)", defaultWatchDir))
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
	PrintSuccess(fmt.Sprintf("Watching %d director%s", len(watchDirs), func() string {
		if len(watchDirs) == 1 {
			return "y"
		}
		return "ies"
	}()))

	
	fmt.Println()
	PrintPrompt("Scan interval in seconds", fmt.Sprintf("(default: %d)", defaultScanInterval))
	scanIntervalInput, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("failed to read input: %w", err)
	}
	scanIntervalInput = strings.TrimSpace(scanIntervalInput)

	scanInterval := defaultScanInterval
	if scanIntervalInput != "" {
		parsed, err := strconv.Atoi(scanIntervalInput)
		if err != nil {
			PrintWarning(fmt.Sprintf("Invalid scan interval, using default: %d", defaultScanInterval))
		} else if parsed < 10 {
			PrintWarning(fmt.Sprintf("Scan interval too small, using default: %d", defaultScanInterval))
		} else {
			scanInterval = parsed
		}
	}
	PrintSuccess(fmt.Sprintf("Scan interval set to %d seconds", scanInterval))

	
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

	
	fmt.Println()
	PrintHeader("Configuration Saved")
	PrintSuccess(fmt.Sprintf("Config saved to %s%s%s", colorBold, configFilePath, colorReset))

	
	fmt.Println()
	PrintHeader("Next Step: GitHub Token")
	fmt.Printf("  %sBefore starting the daemon, you need to set up a GitHub token:%s\n\n", colorGray, colorReset)

	fmt.Printf("  %s1. Create a new token at:%s\n", colorBold, colorReset)
	fmt.Printf("     %s%shttps://github.com/settings/tokens/new%s\n\n", colorCyan, colorBold, colorReset)

	fmt.Printf("  %s2. Give it a name like 'git-ghost'%s\n\n", colorBold, colorReset)

	fmt.Printf("  %s3. Select scope:%s %srepo%s (full control of private repositories)\n\n", colorBold, colorReset, colorYellow, colorReset)

	fmt.Printf("  %s4. Generate the token and copy it%s\n\n", colorBold, colorReset)

	fmt.Printf("  %s5. Set it as environment variable:%s\n", colorBold, colorReset)
	fmt.Printf("     %s%s%s%s\n", colorCyan, colorBold, "export GIT_GHOST_TOKEN=\"ghp_your_token_here\"", colorReset)
	fmt.Printf("     %s%sor on Windows PowerShell:%s\n", colorGray, colorBold, colorReset)
	fmt.Printf("     %s%s$env:GIT_GHOST_TOKEN=\"ghp_your_token_here\"%s\n\n", colorCyan, colorBold, colorReset)

	fmt.Printf("  %s⚠ Important:%s %sNever commit this token to git!%s\n\n", colorYellow, colorReset, colorRed, colorReset)

	
	fmt.Printf("%s%sOnce you've set the token, run:%s\n", colorBold, colorCyan, colorReset)
	fmt.Printf("  %s%sgit-ghost start%s\n\n", colorBold, colorGreen, colorReset)

	return config, nil
}