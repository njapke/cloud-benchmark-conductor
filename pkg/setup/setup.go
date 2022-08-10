package setup

import (
	"fmt"
	"os"
	"path"

	"github.com/christophwitzko/master-thesis/pkg/git"
	"github.com/christophwitzko/master-thesis/pkg/logger"
)

func SourcePaths(log *logger.Logger, checkoutDir, gitRepository, sourcePathOrRefV1, sourcePathOrRefV2 string) (string, string, error) {
	if gitRepository != "" {
		return SourcePathsFromGitRepository(log, checkoutDir, gitRepository, sourcePathOrRefV1, sourcePathOrRefV2)
	}
	return sourcePathOrRefV1, sourcePathOrRefV2, nil
}

func CreateDirectory(p string) error {
	if err := os.RemoveAll(p); err != nil {
		return err
	}
	if err := os.MkdirAll(p, 0o755); err != nil {
		return err
	}
	return nil
}

func SourcePathsFromGitRepository(log *logger.Logger, checkoutDir, repoURL, refV1, refV2 string) (string, string, error) {
	if err := CreateDirectory(checkoutDir); err != nil {
		return "", "", err
	}

	sourcePathV1 := path.Join(checkoutDir, "v1")
	sourcePathV2 := path.Join(checkoutDir, "v2")
	log.Infof("cloning %s", repoURL)
	log.Infof("checking out v1: %s (%s)", refV1, sourcePathV1)
	log.Infof("checking out v2: %s (%s)", refV2, sourcePathV2)
	err := git.CloneAndCheckout(repoURL,
		git.NewCheckoutOption(sourcePathV1, refV1),
		git.NewCheckoutOption(sourcePathV2, refV2),
	)
	if err != nil {
		return "", "", fmt.Errorf("failed to clone or checkout %s: %w", repoURL, err)
	}
	return sourcePathV1, sourcePathV2, nil
}

func ApplicationBenchmarkPath(log *logger.Logger, checkoutDir, gitRepository, referenceOrPath string) (string, error) {
	if gitRepository == "" {
		return referenceOrPath, nil
	}
	if err := CreateDirectory(checkoutDir); err != nil {
		return "", err
	}
	log.Infof("cloning %s", gitRepository)
	log.Infof("checking out %s (%s)", referenceOrPath, checkoutDir)
	err := git.CloneAndCheckout(gitRepository,
		git.NewCheckoutOption(checkoutDir, referenceOrPath),
	)
	if err != nil {
		return "", fmt.Errorf("failed to clone or checkout %s: %w", gitRepository, err)
	}
	return checkoutDir, nil
}
