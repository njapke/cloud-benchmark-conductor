package actions

import (
	"bytes"
	"context"
	"fmt"

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

func (a *actionInstallMicrobenchmarkRunner) Run(ctx context.Context, instance *gcloud.Instance) error {
	lp := instance.LogPrefix() + "[" + a.Name() + "]"
	a.log.Infof("%s installing microbenchmark-runner", lp)
	err := instance.CopyFile(ctx, bytes.NewReader(assets.MicrobenchmarkRunnerBinaryLinuxAmd64), "/tmp/microbenchmark-runner")
	if err != nil {
		return err
	}
	stdout, stderr, err := instance.Run(ctx, "sudo mv /tmp/microbenchmark-runner /usr/bin/")
	if err != nil {
		return fmt.Errorf("failed to move microbenchmark-runner: %w\nSTDERR: %s\nSTDOUT: %s", err, stderr, stdout)
	}
	return nil
}
