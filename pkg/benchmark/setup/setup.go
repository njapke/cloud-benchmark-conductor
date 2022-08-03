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

func checkRefIfExists(refType string, repo *git.Repository, tagName string) (bool, error) {
	var refName plumbing.ReferenceName
	if refType == "tag" {
		refName = plumbing.NewTagReferenceName(tagName)
	} else if refType == "branch" {
		refName = plumbing.NewBranchReferenceName(tagName)
	} else {
		return false, fmt.Errorf("unknown ref type: %s", refType)
	}
	_, err := repo.Reference(refName, false)
	if err != nil {
		if err == plumbing.ErrReferenceNotFound {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func checkoutRef(repo *git.Repository, refName string) error {
	repoTree, err := repo.Worktree()
	if err != nil {
		return err
	}
	checkoutOptions := &git.CheckoutOptions{}
	isTag, err := checkRefIfExists("tag", repo, refName)
	if err != nil {
		return err
	}
	if isTag {
		checkoutOptions.Branch = plumbing.NewTagReferenceName(refName)
	} else {
		isBranch, err := checkRefIfExists("branch", repo, refName)
		if err != nil {
			return err
		}
		if isBranch {
			checkoutOptions.Branch = plumbing.NewBranchReferenceName(refName)
		} else {
			checkoutOptions.Hash = plumbing.NewHash(refName)
		}
	}
	err = repoTree.Checkout(checkoutOptions)
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
	if err := checkoutRef(repoV1, refV1); err != nil {
		return "", "", err
	}
	log.Infof("checking out v2: %s", refV2)
	if err := checkoutRef(repoV2, refV2); err != nil {
		return "", "", err
	}
	return sourcePathV1, sourcePathV2, nil
}
