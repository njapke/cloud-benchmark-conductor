package run

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/christophwitzko/masters-thesis/pkg/assets"
	"github.com/christophwitzko/masters-thesis/pkg/config"
	"github.com/christophwitzko/masters-thesis/pkg/gcloud"
	"github.com/christophwitzko/masters-thesis/pkg/gcloud/actions"
	"github.com/christophwitzko/masters-thesis/pkg/logger"
)

func getAppRunnerCmd(appConf *config.ConductorApplicationConfig) string {
	cmd := []string{
		"application-runner",
		fmt.Sprintf("--git-repository='%s' --v1='%s' --v2='%s'", appConf.Repository, appConf.V1, appConf.V2),
		fmt.Sprintf("--application-package %s", appConf.Package),
		"--bind 0.0.0.0",
	}
	for _, env := range appConf.Env {
		cmd = append(cmd, fmt.Sprintf("--env='%s'", env))
	}
	if appConf.LimitCPU {
		cmd = append(cmd, "--limit-cpu")
		cmd = append([]string{"sudo"}, cmd...)
	}
	return strings.Join(cmd, " ")
}

func Application(ctx context.Context, log *logger.Logger, service gcloud.Service, internalIP chan<- string) error {
	defer close(internalIP)
	appConf := service.Config().Application

	runnerName := fmt.Sprintf("%s-application", appConf.Name)
	log.Infof("[%s] creating or getting instance...", runnerName)
	instance, err := service.GetOrCreateInstance(ctx, runnerName, appConf.InstanceType)
	if err != nil {
		return err
	}
	// close open ssh connection
	defer instance.Close()

	log.Infof("[%s] external IP: %s", runnerName, instance.ExternalIP())
	var logFilterRe *regexp.Regexp
	if appConf.LogFilter != "" {
		logFilterRe, err = regexp.Compile(appConf.LogFilter)
		if err != nil {
			return fmt.Errorf("failed to compile log filter regexp: %w", err)
		}
	}
	log.Infof("[%s] setting up instance...", runnerName)
	err = instance.ExecuteActions(ctx,
		actions.NewActionInstallGo(log),
		actions.NewActionInstallBinary(log, "application-runner", assets.ApplicationRunner),
	)
	if err != nil {
		return err
	}
	internalIP <- instance.InternalIP()
	cmd := getAppRunnerCmd(appConf)
	log.Infof("[%s] running: %s", runnerName, cmd)
	return instance.RunWithLogger(ctx, func(stdout, stderr string) {
		if logFilterRe != nil && (logFilterRe.MatchString(stderr) || logFilterRe.MatchString(stdout)) {
			return
		}
		log.Infof("[%s] %s%s", runnerName, stdout, stderr)
	}, cmd)
}
