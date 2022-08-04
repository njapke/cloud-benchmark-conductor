package actions

import (
	"context"
	"fmt"
	"strings"

	"github.com/christophwitzko/master-thesis/pkg/gcloud"
	"github.com/christophwitzko/master-thesis/pkg/logger"
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

func getGoVersion(ctx context.Context, instance *gcloud.Instance) (string, error) {
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

func (a *actionInstallGo) Run(ctx context.Context, instance *gcloud.Instance) error {
	lp := instance.LogPrefix() + "[" + a.Name() + "]"

	a.log.Infof("%s version: %s", lp, instance.Config.GoVersion)
	foundGoVersion, err := getGoVersion(ctx, instance)
	if err != nil {
		return err
	}
	if foundGoVersion == instance.Config.GoVersion {
		a.log.Infof("%s go is already installed", lp)
		return nil
	}

	a.log.Infof("%s downloading...", lp)
	goDownloadURL := fmt.Sprintf("https://go.dev/dl/go%s.linux-amd64.tar.gz", instance.Config.GoVersion)
	stdout, stderr, err := instance.Run(ctx, fmt.Sprintf("curl -SL -o go.tgz %s", goDownloadURL))
	if err != nil {
		return fmt.Errorf("failed to download go: %w\nSTDERR: %s\nSTDOUT: %s", err, stderr, stdout)
	}

	a.log.Infof("%s installing...", lp)
	stdout, stderr, err = instance.Run(ctx, "sudo rm -rf  /usr/local/go && sudo tar -C /usr/local -xzf go.tgz")
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
	if foundGoVersion != instance.Config.GoVersion {
		return fmt.Errorf("go version did not match: %s", foundGoVersion)
	}

	a.log.Infof("%s go%s installed successfully", lp, instance.Config.GoVersion)
	return nil
}
