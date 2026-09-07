package backup

import (
	"git-ghost/internal/config"
	"log/slog"
	"os"
	"time"
	"fmt"
	"path/filepath"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	
	"github.com/go-git/go-git/v5/plumbing/transport/http"
)

func GetVersion(cfg *config.Config, repoName string, date time.Time, branch string,token string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home dir: %w", err)
	}
	
	restoreDir := filepath.Join(
		home, 
		"git-ghost-restores", 
		repoName, 
		fmt.Sprintf("%s_%s", date.Format("2006-01-02_15-04"), branch),
	)


	branchname := fmt.Sprintf(
		"backup/%s/%s/%s", 
		repoName, 
		date.Format("2006-01-02_15-04"),  
		branch,
	)

	fmt.Printf("Cloning %s (branch: %s)...\n", repoName, branchname)
	fmt.Printf("Destination: %s\n", restoreDir)

	if _, err := os.Stat(restoreDir); err == nil {
		if err := os.RemoveAll(restoreDir); err != nil {
			return "", fmt.Errorf("failed to remove existing directory: %w", err)
		}
	}

	
	if err := os.MkdirAll(filepath.Dir(restoreDir), 0755); err != nil {
		return "", fmt.Errorf("failed to create directory: %w", err)
	}
	
	
	cloneOpts := &git.CloneOptions{
		URL:           cfg.BackupRepo,
		SingleBranch:  true,
		ReferenceName: plumbing.NewBranchReferenceName(branchname),
	}
	if token != "" {
		cloneOpts.Auth = &http.BasicAuth{
			Username: "git-ghost",
			Password: token,
		}
	}
	slog.Info("try to get version repo", "branchname", branchname)
	_, err = git.PlainClone(restoreDir, false, cloneOpts)

	
	if err != nil {
		slog.Error("Failed to clone repository", "error", err)
		return "", err
	}
	slog.Info("The repository cloned successfully.", "path", restoreDir)

	return restoreDir, nil

}
