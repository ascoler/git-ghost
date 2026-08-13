package scanner

import (
	"log/slog"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)
type BranchInfo struct {
	Name string 
	Hash string
}

func ScanRepository(repoPath string) (*git.Repository, error) {
	repo,err := git.PlainOpen(repoPath)
	if err != nil {
		slog.Error("Failed to open repository", "path", repoPath, "error", err)
		return nil,err
	}
	slog.Info("The repository opened successfully.", "path", repoPath)

	return repo,nil
}

func GetRemote(repo *git.Repository) (*git.Remote, error) {
	remote,err := repo.Remote("origin")
	if err != nil {
		slog.Error("Failed to get remote", "error", err)
		return nil,err
	}
	slog.Info("The remote retrieved successfully.", "name", remote.Config().Name)

	return remote,nil
}

func GetHead(repo *git.Repository) (*plumbing.Reference, error) {
	head,err := repo.Head()
	if err != nil {
		slog.Error("Failed to get HEAD", "error", err)
		return nil,err
	}
	slog.Info("The HEAD retrieved successfully.", "name", head.Name().String())

	return head,nil

}

func GetBranches(repo *git.Repository) ([]BranchInfo, error) {
	var branches []BranchInfo

	branchIter, err := repo.Branches()
	if err != nil {
		slog.Error("Failed to get branches", "error", err)
		return nil, err
	}

	err = branchIter.ForEach(func(ref *plumbing.Reference) error {
		branches = append(branches, BranchInfo{
			Name: ref.Name().Short(),
			Hash: ref.Hash().String(),
		})
		return nil
	})
	if err != nil {
		slog.Error("Failed to iterate over branches", "error", err)
		return nil, err
	}

	slog.Info("Branches retrieved successfully.", "count", len(branches))
	return branches, nil
}



