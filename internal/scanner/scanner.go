package scanner

import (
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
)
//сделать concurrent сканирование репозиториев
type RepositoryInfo struct {
	Path          string
	CurrentCommit CommitInfo
	Branches      []BranchInfo
}
type BackupCommitInfo struct{
	Hash string 
	Date string 
	Author string
	Message string
}
func IsGitRepository(path string) bool {
	_, err := os.Stat(filepath.Join(path, ".git"))
	return !os.IsNotExist(err)
}

func InfoRepository(path string) (*RepositoryInfo, error) {
	repo, err := ScanRepository(path)
	
	if err != nil {
		return nil, err
	}

	commit, err := GetCurrentCommit(repo)
	if err != nil {
		return nil, err
	}

	branches, err := GetBranches(repo)
	if err != nil {
		return nil, err
	}

	return &RepositoryInfo{
		Path:          path,
		CurrentCommit: GetCommitInfo(commit),
		Branches:      branches,
	}, nil
}

func ShouldSkip(name string) bool {
	skipDirs := map[string]bool{
		"node_modules": true,
		"vendor":       true,
		"__pycache__":  true,
		".cache":       true,
		".Trash":       true,
		"dist":         true,
		"build":        true,
	}
	return skipDirs[name]
}


func ScanRepositories(watchDirs []string) ([]*RepositoryInfo, error) {
	var repositories []*RepositoryInfo
	var failedCount int

	for i := 0; i < len(watchDirs); i++ {
		dir := watchDirs[i]

		if IsGitRepository(dir) {
			repoInfo, err := InfoRepository(dir)
			if err != nil {
				slog.Warn("Failed to scan repository", "path", dir, "error", err)
				failedCount++
				continue
			}
			repositories = append(repositories, repoInfo)
			continue
		}

		err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				slog.Warn("Error accessing path", "path", path, "error", err)
				return nil
			}

			if d.IsDir() && ShouldSkip(d.Name()) {
				return filepath.SkipDir
			}

			if d.IsDir() && d.Name() == ".git" {
				repoPath := filepath.Dir(path)
				repoInfo, err := InfoRepository(repoPath)
				if err != nil {
					slog.Warn("Failed to scan repository", "path", repoPath, "error", err)
					failedCount++
					return filepath.SkipDir
				}
				repositories = append(repositories, repoInfo)
				return filepath.SkipDir
			}

			return nil
		})

		if err != nil {
			slog.Warn("Failed to walk directory", "path", dir, "error", err)
		}
	}

	if len(repositories) == 0 && len(watchDirs) > 0 {
		return nil, fmt.Errorf("failed to scan all repositories (%d failed)", failedCount)
	}

	return repositories, nil
}
