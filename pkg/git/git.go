package git

import (
	"errors"
	"fmt"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	cp "github.com/otiai10/copy"
)

type referenceType string

const (
	referenceTypeTag    = referenceType("tag")
	referenceTypeBranch = referenceType("branch")
)

func checkRefIfExists(refType referenceType, repo *git.Repository, tagName string) (bool, error) {
	var refName plumbing.ReferenceName
	switch refType {
	case referenceTypeTag:
		refName = plumbing.NewTagReferenceName(tagName)
	case referenceTypeBranch:
		refName = plumbing.NewBranchReferenceName(tagName)
	default:
		return false, fmt.Errorf("unknown ref type: %s", refType)
	}
	_, err := repo.Reference(refName, false)
	if err != nil {
		if errors.Is(err, plumbing.ErrReferenceNotFound) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func CheckoutReference(repo *git.Repository, refName string) error {
	repoTree, err := repo.Worktree()
	if err != nil {
		return err
	}
	checkoutOptions := &git.CheckoutOptions{}
	isTag, err := checkRefIfExists(referenceTypeTag, repo, refName)
	if err != nil {
		return err
	}
	if isTag {
		checkoutOptions.Branch = plumbing.NewTagReferenceName(refName)
		return repoTree.Checkout(checkoutOptions)
	}
	isBranch, err := checkRefIfExists(referenceTypeBranch, repo, refName)
	if err != nil {
		return err
	}
	if isBranch {
		checkoutOptions.Branch = plumbing.NewBranchReferenceName(refName)
	} else {
		checkoutOptions.Hash = plumbing.NewHash(refName)
	}
	return repoTree.Checkout(checkoutOptions)
}

// Clone clones the given repository to the provided paths.
func Clone(repoURL string, destinationDirs ...string) ([]*git.Repository, error) {
	if len(destinationDirs) == 0 {
		return nil, errors.New("no destination directory specified")
	}

	mainRepoDir := destinationDirs[0]
	repo, err := git.PlainClone(mainRepoDir, false, &git.CloneOptions{
		URL:  repoURL,
		Tags: git.AllTags,
	})
	if err != nil {
		return nil, err
	}
	repos := make([]*git.Repository, len(destinationDirs))
	repos[0] = repo
	for i, destDir := range destinationDirs[1:] {
		err := cp.Copy(mainRepoDir, destDir)
		if err != nil {
			return nil, err
		}
		repo, err := git.PlainOpen(destDir)
		if err != nil {
			return nil, err
		}
		repos[i+1] = repo
	}
	return repos, nil
}

type CheckoutOption struct {
	DestinationDir string
	RefName        string
}

func NewCheckoutOption(destDir, refName string) *CheckoutOption {
	return &CheckoutOption{
		DestinationDir: destDir,
		RefName:        refName,
	}
}

func CloneAndCheckout(repoURL string, checkoutOptions ...*CheckoutOption) error {
	destinationDirs := make([]string, len(checkoutOptions))
	for i, option := range checkoutOptions {
		destinationDirs[i] = option.DestinationDir
	}
	repos, err := Clone(repoURL, destinationDirs...)
	if err != nil {
		return err
	}
	for i, repo := range repos {
		err := CheckoutReference(repo, checkoutOptions[i].RefName)
		if err != nil {
			return err
		}
	}
	return nil
}
