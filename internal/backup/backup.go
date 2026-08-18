package backup

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"git-ghost/internal/scanner"
)

const remoteName = "git-ghost-backup"

func AddFilesToCommit(workTree *git.Worktree) error {
	err := workTree.AddWithOptions(&git.AddOptions{
		All: true,
	})
	if err != nil {
		slog.Error("Failed to add files to commit", "error", err)
		return err
	}
	return nil
}

func CommitChanges(workTree *git.Worktree) (*plumbing.Hash, error) {
	hash, err := workTree.Commit("Backup changes", &git.CommitOptions{
		Author: &object.Signature{
			Name:  "Git Ghost",
			Email: "git-ghost@localhost",
			When:  time.Now(),
		},
	})
	if err != nil {
		slog.Error("Failed to commit changes", "error", err)
		return nil, err
	}
	return &hash, nil
}

func EnsureRemote(repo *git.Repository, backupURL string) error {
	_, err := repo.Remote(remoteName)
	if err == nil {
		return nil
	}

	_, err = repo.CreateRemote(&config.RemoteConfig{
		Name: remoteName,
		URLs: []string{backupURL},
	})
	if err != nil {
		slog.Error("Failed to create remote", "error", err)
		return fmt.Errorf("failed to create remote: %w", err)
	}

	return nil
}

func PushBackup(repo *git.Repository, repoName string, branches []scanner.BranchInfo, token string) error {
	timestamp := time.Now().Format("2006-01-02_15-04")

	var refSpecs []config.RefSpec
	for _, branch := range branches {
		src := "refs/heads/" + branch.Name
		dst := "refs/heads/backup/" + repoName + "/" + timestamp + "/" + branch.Name
		refSpecs = append(refSpecs, config.RefSpec(src+":"+dst))
	}

	err := repo.Push(&git.PushOptions{
		RemoteName: remoteName,
		RefSpecs:   refSpecs,
		Auth: &http.BasicAuth{
			Username: "git-ghost",
			Password: token,
		},
	})
	if err != nil {
		if err == git.NoErrAlreadyUpToDate {
			slog.Info("Backup is already up to date")
			return nil
		}
		slog.Error("Failed to push backup", "error", err)
		return fmt.Errorf("failed to push backup: %w", err)
	}

	slog.Info("Backup pushed successfully", "repo", repoName, "branches", len(branches))
	return nil
}

func BackupRepository(repo *git.Repository, repoName string, backupURL string, token string) error {
	err := EnsureRemote(repo, backupURL)
	if err != nil {
		return err
	}

	branches, err := scanner.GetBranches(repo)
	if err != nil {
		return fmt.Errorf("failed to get branches: %w", err)
	}

	err = PushBackup(repo, repoName, branches, token)
	if err != nil {
		return err
	}

	return nil
}