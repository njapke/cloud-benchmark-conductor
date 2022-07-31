package gcloud

import (
	"context"
	"fmt"
	"strings"

	"github.com/christophwitzko/master-thesis/pkg/logger"
)

type Action interface {
	Run(ctx context.Context, instance *Instance) error
	Name() string
}

type actionInstallGo struct {
	log *logger.Logger
}

func NewActionInstallGo(log *logger.Logger) *actionInstallGo {
	return &actionInstallGo{log: log}
}

func (i *actionInstallGo) Name() string {
	return "install-go"
}

const PATHWithGoBin = `PATH="/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:/usr/games:/usr/local/games:/snap/bin:/usr/local/go/bin"`

func (i *actionInstallGo) Run(ctx context.Context, instance *Instance) error {
	lp := instance.logPrefix() + "[" + i.Name() + "]"

	i.log.Infof("%s go version: %s...", lp, instance.Config.GoVersion)
	i.log.Infof("%s downloading...", lp)
	goDownloadUrl := fmt.Sprintf("https://go.dev/dl/go%s.linux-amd64.tar.gz", instance.Config.GoVersion)
	stdout, stderr, err := instance.Run(ctx, fmt.Sprintf("curl -SL -o go.tgz %s", goDownloadUrl))
	if err != nil {
		return fmt.Errorf("failed to download go: %w\nSTDERR: %s\nSTDOUT: %s", err, stderr, stdout)
	}

	i.log.Infof("%s installing...", lp)
	stdout, stderr, err = instance.Run(ctx, "sudo tar -C /usr/local -xzf go.tgz")
	if err != nil {
		return fmt.Errorf("failed to install go: %w\nSTDERR: %s\nSTDOUT: %s", err, stderr, stdout)
	}
	stderr, stdout, err = instance.Run(ctx, "echo '"+PATHWithGoBin+"' | sudo tee /etc/environment")
	if err != nil {
		return fmt.Errorf("failed to add go to PATH: %w\nSTDERR: %s\nSTDOUT: %s", err, stderr, stdout)
	}

	if err := instance.Reconnect(ctx); err != nil {
		return fmt.Errorf("failed to reconnect: %w", err)
	}

	i.log.Infof("%s verifying installation...", lp)
	stdout, stderr, err = instance.Run(ctx, "go version")
	if err != nil {
		return fmt.Errorf("failed to run go version: %w\nSTDERR: %s\nSTDOUT: %s", err, stderr, stdout)
	}
	if !strings.Contains(stdout, fmt.Sprintf("go version go%s", instance.Config.GoVersion)) {
		return fmt.Errorf("go version did not match: %s", stdout)
	}
	i.log.Infof("%s go%s installed successfully", lp, instance.Config.GoVersion)
	return nil
}

type actionInstallMicrobenchmarkRunner struct {
	log *logger.Logger
}

func NewActionInstallMicrobenchmarkRunner(log *logger.Logger) *actionInstallMicrobenchmarkRunner {
	return &actionInstallMicrobenchmarkRunner{log: log}
}

func (i *actionInstallMicrobenchmarkRunner) Name() string {
	return "install-microbenchmark-runner"
}
