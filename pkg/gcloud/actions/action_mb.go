package actions

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/christophwitzko/master-thesis/pkg/assets"
	"github.com/christophwitzko/master-thesis/pkg/gcloud"
	"github.com/christophwitzko/master-thesis/pkg/logger"
)

type actionInstallMicrobenchmarkRunner struct {
	log *logger.Logger
}

func NewActionInstallMicrobenchmarkRunner(log *logger.Logger) *actionInstallMicrobenchmarkRunner {
	return &actionInstallMicrobenchmarkRunner{log: log}
}

func (a *actionInstallMicrobenchmarkRunner) Name() string {
	return "install-microbenchmark-runner"
}

func getMbHash(ctx context.Context, instance *gcloud.Instance) (string, error) {
	stdout, stderr, err := instance.Run(ctx, "sha512sum /usr/bin/microbenchmark-runner || true\n")
	if err != nil {
		return "", fmt.Errorf("failed to run go version: %w\nSTDERR: %s\nSTDOUT: %s", err, stderr, stdout)
	}
	stdout, _, _ = strings.Cut(stdout, " ")
	return stdout, nil
}

func (a *actionInstallMicrobenchmarkRunner) Run(ctx context.Context, instance *gcloud.Instance) error {
	lp := instance.LogPrefix() + "[" + a.Name() + "]"
	a.log.Infof("%s installing microbenchmark-runner...", lp)
	foundMbHash, err := getMbHash(ctx, instance)
	if err != nil {
		return err
	}
	if foundMbHash == assets.GetMicrobenchmarkRunnerBinaryLinuxAmd64Hash() {
		a.log.Infof("%s microbenchmark-runner is already installed", lp)
		return nil
	}

	err = instance.CopyFile(ctx, bytes.NewReader(assets.MicrobenchmarkRunnerBinaryLinuxAmd64), "/tmp/microbenchmark-runner")
	if err != nil {
		return err
	}
	stdout, stderr, err := instance.Run(ctx, "sudo mv /tmp/microbenchmark-runner /usr/bin/")
	if err != nil {
		return fmt.Errorf("failed to move microbenchmark-runner: %w\nSTDERR: %s\nSTDOUT: %s", err, stderr, stdout)
	}
	return nil
}
