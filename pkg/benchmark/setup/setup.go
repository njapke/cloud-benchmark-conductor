package setup

import (
	"fmt"
	"os"
	"path"

	"github.com/christophwitzko/master-thesis/pkg/git"
	"github.com/christophwitzko/master-thesis/pkg/logger"
)

func SourcePathsFromGitRepository(log *logger.Logger, benchDir, repoURL, refV1, refV2 string) (string, string, error) {
	if err := os.RemoveAll(benchDir); err != nil {
		return "", "", err
	}
	if err := os.MkdirAll(benchDir, 0o755); err != nil {
		return "", "", err
	}

	sourcePathV1 := path.Join(benchDir, "v1")
	sourcePathV2 := path.Join(benchDir, "v2")
	log.Infof("cloning %s to %s and %s", repoURL, sourcePathV1, sourcePathV2)
	repos, err := git.Clone(repoURL, sourcePathV1, sourcePathV2)
	if err != nil {
		return "", "", fmt.Errorf("failed to clone %s: %w", repoURL, err)
	}

	log.Infof("checking out v1: %s", refV1)
	if err := git.CheckoutReference(repos[0], refV1); err != nil {
		return "", "", err
	}
	log.Infof("checking out v2: %s", refV2)
	if err := git.CheckoutReference(repos[1], refV2); err != nil {
		return "", "", err
	}
	return sourcePathV1, sourcePathV2, nil
}
