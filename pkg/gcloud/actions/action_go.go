package actions

import (
	"context"
	"fmt"
	"strings"

	"github.com/christophwitzko/master-thesis/pkg/gcloud"
	"github.com/christophwitzko/master-thesis/pkg/logger"
	"github.com/christophwitzko/master-thesis/pkg/retry"
)

type actionInstallGo struct {
	log *logger.Logger
}

func NewActionInstallGo(log *logger.Logger) *actionInstallGo {
	return &actionInstallGo{log: log}
}

func (a *actionInstallGo) Name() string {
	return "install-go"
}

const PATHWithGoBin = `PATH="/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:/usr/games:/usr/local/games:/snap/bin:/usr/local/go/bin"`

func getGoVersion(ctx context.Context, instance gcloud.Instance) (string, error) {
	stdout, stderr, err := instance.Run(ctx, "go version || true")
	if err != nil {
		return "", fmt.Errorf("failed to run go version: %w\nSTDERR: %s\nSTDOUT: %s", err, stderr, stdout)
	}
	goVersion, _, found := strings.Cut(strings.TrimPrefix(stdout, "go version go"), " ")
	if !found {
		return stdout, nil
	}
	return goVersion, nil
}

func (a *actionInstallGo) Run(ctx context.Context, instance gcloud.Instance) error {
	lp := instance.LogPrefix() + "[" + a.Name() + "]"
	desiredGoVersion := instance.Config().GoVersion
	a.log.Infof("%s version: %s", lp, desiredGoVersion)
	foundGoVersion, err := getGoVersion(ctx, instance)
	if err != nil {
		return err
	}
	if foundGoVersion == desiredGoVersion {
		a.log.Infof("%s go is already installed", lp)
		return nil
	}

	goDownloadURL := fmt.Sprintf("https://go.dev/dl/go%s.linux-amd64.tar.gz", desiredGoVersion)
	err = retry.OnError(ctx, a.log, lp, func() error {
		a.log.Infof("%s downloading...", lp)
		stdout, stderr, runErr := instance.Run(ctx, fmt.Sprintf("curl -SL -o go.tgz %s", goDownloadURL))
		if runErr != nil {
			return fmt.Errorf("failed to download go: %w\nSTDERR: %s\nSTDOUT: %s", runErr, stderr, stdout)
		}
		return nil
	})
	if err != nil {
		return err
	}

	a.log.Infof("%s installing...", lp)
	stdout, stderr, err := instance.Run(ctx, "sudo rm -rf  /usr/local/go && sudo tar -C /usr/local -xzf go.tgz")
	if err != nil {
		return fmt.Errorf("failed to install go: %w\nSTDERR: %s\nSTDOUT: %s", err, stderr, stdout)
	}
	stderr, stdout, err = instance.Run(ctx, "echo '"+PATHWithGoBin+"' | sudo tee /etc/environment")
	if err != nil {
		return fmt.Errorf("failed to add go to PATH: %w\nSTDERR: %s\nSTDOUT: %s", err, stderr, stdout)
	}

	err = instance.Reconnect(ctx)
	if err != nil {
		return fmt.Errorf("failed to reconnect: %w", err)
	}

	a.log.Infof("%s verifying installation...", lp)

	foundGoVersion, err = getGoVersion(ctx, instance)
	if err != nil {
		return err
	}
	if foundGoVersion != desiredGoVersion {
		return fmt.Errorf("go version did not match: %s", foundGoVersion)
	}

	a.log.Infof("%s go%s installed successfully", lp, desiredGoVersion)
	return nil
}
