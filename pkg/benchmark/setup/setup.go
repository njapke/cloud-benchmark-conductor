package setup

import (
	"fmt"
	"os"
	"path"

	"github.com/christophwitzko/master-thesis/pkg/logger"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	cp "github.com/otiai10/copy"
)

func checkoutBranch(repo *git.Repository, branchName string) error {
	repoTree, err := repo.Worktree()
	if err != nil {
		return err
	}
	err = repoTree.Checkout(&git.CheckoutOptions{
		Branch: plumbing.NewBranchReferenceName(branchName),
	})
	if err != nil {
		return err
	}
	return nil
}

func SourcePathsFromGitRepository(log *logger.Logger, benchDir, repoUrl, refV1, refV2 string) (string, string, error) {
	if err := os.RemoveAll(benchDir); err != nil {
		return "", "", err
	}
	if err := os.MkdirAll(benchDir, 0755); err != nil {
		return "", "", err
	}

	sourcePathV1 := path.Join(benchDir, "v1")
	sourcePathV2 := path.Join(benchDir, "v2")
	log.Infof("cloning %s to %s", repoUrl, sourcePathV1)
	repoV1, err := git.PlainClone(sourcePathV1, false, &git.CloneOptions{
		URL:  repoUrl,
		Tags: git.AllTags,
	})
	if err != nil {
		return "", "", fmt.Errorf("failed to clone %s to %s: %w", repoUrl, sourcePathV1, err)
	}

	log.Infof("duplicating repository to %s", sourcePathV2)
	if err := cp.Copy(sourcePathV1, sourcePathV2); err != nil {
		return "", "", fmt.Errorf("failed to copy %s to %s: %w", sourcePathV1, sourcePathV2, err)
	}
	repoV2, err := git.PlainOpen(sourcePathV2)
	if err != nil {
		return "", "", fmt.Errorf("failed to open %s: %w", sourcePathV1, err)
	}

	log.Infof("checking out v1: %s", refV1)
	if err := checkoutBranch(repoV1, refV1); err != nil {
		return "", "", err
	}
	log.Infof("checking out v2: %s", refV2)
	if err := checkoutBranch(repoV2, refV2); err != nil {
		return "", "", err
	}
	return sourcePathV1, sourcePathV2, nil
}
