package run

import (
	"context"
	"fmt"

	"github.com/christophwitzko/master-thesis/pkg/gcloud"
	"github.com/christophwitzko/master-thesis/pkg/logger"
)

type MicrobenchmarkConfig struct {
	Repository    string
	V1, V2        string
	RunIndex      int
	ExcludeFilter string
}

func (m MicrobenchmarkConfig) Cmd() string {
	return fmt.Sprintf("microbenchmark-runner --run %d --v1 %s --v2 %s --git-repository %s --exclude-filter=\"%s\"", m.RunIndex, m.V1, m.V2, m.Repository, m.ExcludeFilter)
}

func Microbenchmark(ctx context.Context, log *logger.Logger, service *gcloud.Service, mc MicrobenchmarkConfig) error {
	runnerName := fmt.Sprintf("runner-%d", mc.RunIndex)
	log.Infof("creating instance %s...", runnerName)
	instance, err := service.CreateInstance(ctx, runnerName)
	if err != nil {
		return err
	}
	// close open ssh connection
	defer instance.Close()

	err = instance.ExecuteActions(ctx, gcloud.NewActionInstallGo(log), gcloud.NewActionInstallMicrobenchmarkRunner(log))
	if err != nil {
		return err
	}
	return instance.RunWithLog(ctx, log, mc.Cmd())
}
