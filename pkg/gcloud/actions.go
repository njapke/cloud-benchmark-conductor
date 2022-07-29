package gcloud

import (
	"context"
	"fmt"

	"github.com/christophwitzko/master-thesis/pkg/logger"
)

type Action interface {
	Run(ctx context.Context, log *logger.Logger, instance *Instance) error
	Name() string
}

type ActionInstallGo struct{}

func (i *ActionInstallGo) Name() string {
	return "install-go"
}

const PATHWithGoBin = `PATH="/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:/usr/games:/usr/local/games:/snap/bin:/usr/local/go/bin"`

func (i *ActionInstallGo) Run(ctx context.Context, log *logger.Logger, instance *Instance) error {
	lp := instance.logPrefix() + "[" + i.Name() + "]"

	log.Infof("%s downloading go%s...", lp, instance.Config.GoVersion)
	goDownloadUrl := fmt.Sprintf("https://go.dev/dl/go%s.linux-amd64.tar.gz", instance.Config.GoVersion)
	stdout, stderr, err := instance.Run(ctx, fmt.Sprintf("curl -SL -o go.tgz %s", goDownloadUrl))
	if err != nil {
		return fmt.Errorf("failed to download go: %w\nSTDERR: %s\nSTDOUT: %s", err, stderr, stdout)
	}

	log.Infof("%s installing go...", lp)
	stdout, stderr, err = instance.Run(ctx, "sudo tar -C /usr/local -xzf go.tgz")
	if err != nil {
		return fmt.Errorf("failed to install go: %w\nSTDERR: %s\nSTDOUT: %s", err, stderr, stdout)
	}
	stderr, stdout, err = instance.Run(ctx, "echo '"+PATHWithGoBin+"' | sudo tee /etc/environment")
	if err != nil {
		return fmt.Errorf("failed to add go to PATH: %w\nSTDERR: %s\nSTDOUT: %s", err, stderr, stdout)
	}

	// TODO: check go version
	//err = instance.RunWithLog(ctx, log, "go version")
	//if err != nil {
	//	return fmt.Errorf("failed to run go version: %w\nSTDERR: %s\nSTDOUT: %s", err, stderr, stdout)
	//}
	//log.Info(stdout)
	//log.Info(stderr)
	//if !strings.Contains(stderr, fmt.Sprintf("go version go%s", instance.Config.GoVersion)) {
	//	return fmt.Errorf("go version did not match: %s", stdout)
	//}
	log.Infof("%s go installed successfully", lp)
	return nil
}
