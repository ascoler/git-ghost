package scanner

import (
	"context"
	

	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"time"
)


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

type gitghoststate struct {
	mu sync.RWMutex
	visited map[string]bool

}
func newGitGhostState() *gitghoststate {
	return &gitghoststate{
		visited: make(map[string]bool),
	}
}
func (s *gitghoststate) isVisited(path string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.visited[path]
}

func (s *gitghoststate) markVisited(path string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.visited[path] = true
}
func (s *gitghoststate) tryMark(path string) bool {
    s.mu.Lock()
    defer s.mu.Unlock()
    
    if s.visited[path] {
        return false 
    }
    
    s.visited[path] = true
    return true 
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

func worker(ctx context.Context, jobs <-chan string, results chan<- *RepositoryInfo, wg *sync.WaitGroup, state *gitghoststate) {
    defer wg.Done()

    for {
        select {
        case <-ctx.Done():
            return

        case task, ok := <-jobs:
            if !ok {
                return
            }

           
            if !state.tryMark(task) {
                slog.Debug("Skipping already processed", "path", task)
                continue
            }

            repoInfo := func() *RepositoryInfo {
                defer func() {
                    if r := recover(); r != nil {
                        slog.Error("Worker panic", "path", task, "error", r)
                    }
                }()

                if !IsGitRepository(task) {
                    return nil
                }

                info, err := InfoRepository(task)
                if err != nil {
                    slog.Warn("Failed to scan repository", "path", task, "error", err)
                    return nil
                }

                return info
            }()

            if repoInfo == nil {
                
                state.mu.Lock()
                delete(state.visited, task)
                state.mu.Unlock()
                continue
            }

            select {
            case results <- repoInfo:
                slog.Debug("Scanned", "path", task)
            case <-ctx.Done():
                return
            }
        }
    }
}
func collectPaths(dirs []string) ([]string,error){
	paths := []string{}
	for j := 0; j < len(dirs);j++{
		err := filepath.WalkDir(dirs[j], func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				slog.Warn("Error accessing path", "path", path, "error", err)
				return nil
			}

			if d.IsDir() && ShouldSkip(d.Name()) {
				return filepath.SkipDir
			}

			if d.IsDir() && d.Name() == ".git" {
				repoPath := filepath.Dir(path)
				paths = append(paths, repoPath)
				
				
				
				return filepath.SkipDir
			}

			return nil
		})

		if err != nil {
			slog.Warn("Failed to walk directory", "error", err)
		}
	}
	return paths,nil
}
func ScanRepositories(watchDirs []string) ([]*RepositoryInfo, error) {
	paths,err := collectPaths(watchDirs)
	if err != nil{
		slog.Error("Error to find repository")
	}
	
	
	ctx,cancel := context.WithTimeout(context.Background(),30 * time.Second)
	defer cancel()
	jobs := make(chan string,len(paths))
	results := make(chan *RepositoryInfo,len(paths))
	var wg sync.WaitGroup
	state := newGitGhostState()
	for i := 0; i < 5;i++{
		wg.Add(1)
		go worker(ctx,jobs,results,&wg,state)
	}
	for _,p := range paths{
		jobs <- p
	}
	close(jobs)
	go func ()  {
		wg.Wait()
		close(results)
	}()


	var repos []*RepositoryInfo
    for r := range results {
        repos = append(repos, r)
    }
    
    return repos, nil
}
