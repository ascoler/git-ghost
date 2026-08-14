package scanner

import (
	"log/slog"
	"fmt"
)



type RepositoryInfo struct {
	Path string 
	CurrentCommit CommitInfo
	Branches []BranchInfo
}





func ScanRepositories(watchDirs []string) ([]*RepositoryInfo, error) {
	var repositories []*RepositoryInfo
	var failedCount int
	
	for i := 0; i < len(watchDirs); i++ {
		repo, err := ScanRepository(watchDirs[i])
		if err != nil {
			slog.Warn("Failed to scan repository", "path", watchDirs[i], "error", err)
			failedCount++
			continue 
		}
		
		commit,err := GetCurrentCommit(repo)
		if err != nil {
			slog.Warn("Failed to get current commit", "path", watchDirs[i], "error", err)
			failedCount++
			continue 
		}
		
		branches,err := GetBranches(repo)
		if err != nil {
			slog.Warn("Failed to get branches", "path", watchDirs[i], "error", err)
			failedCount++
			continue 
		}
		
		repositories = append(repositories, &RepositoryInfo{
			Path:          watchDirs[i],
			CurrentCommit: GetCommitInfo(commit),
			Branches:      branches,
		})
	}
	if len(repositories) == 0 && len(watchDirs) > 0 {
        return nil, fmt.Errorf("failed to scan all repositories (%d failed)", failedCount)
    }
    
    return repositories, nil
	
}

