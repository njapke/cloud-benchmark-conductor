package run

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"text/template"
	"time"

	"github.com/christophwitzko/masters-thesis/pkg/assets"
	"github.com/christophwitzko/masters-thesis/pkg/config"
	"github.com/christophwitzko/masters-thesis/pkg/gcloud"
	"github.com/christophwitzko/masters-thesis/pkg/gcloud/actions"
	"github.com/christophwitzko/masters-thesis/pkg/logger"
)

type appTmplData struct {
	Timestamp string
	Name      string
	V1, V2    string
}

func applyAppBenchOutputTemplate(appConf *config.ConductorApplicationConfig, tmplStr string) (string, error) {
	tmpl, err := template.New("tmpl").Parse(tmplStr)
	if err != nil {
		return "", err
	}
	buf := &bytes.Buffer{}
	err = tmpl.Execute(buf, appTmplData{
		Timestamp: currentTimestamp,
		Name:      appConf.Name,
		V1:        appConf.V1,
		V2:        appConf.V2,
	})
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

func getAppBenchRunnerCmd(timeout time.Duration, appConf *config.ConductorApplicationConfig, targets []string) (string, error) {
	resultsOutput, err := applyAppBenchOutputTemplate(appConf, appConf.Benchmark.Output)
	if err != nil {
		return "", err
	}
	cmd := []string{
		"application-benchmark-runner",
		fmt.Sprintf("--git-repository='%s' --reference='%s'", appConf.Repository, appConf.Benchmark.Reference),
		fmt.Sprintf("--results-output='%s'", resultsOutput),
		fmt.Sprintf("--timeout=%s", timeout),
	}
	if appConf.Benchmark.Config != "" {
		cmd = append(cmd, fmt.Sprintf("--config='%s'", appConf.Benchmark.Config))
	}
	if appConf.Benchmark.Tool != "" {
		cmd = append(cmd, fmt.Sprintf("--tool=%s", appConf.Benchmark.Tool))
	}
	for _, target := range targets {
		cmd = append(cmd, fmt.Sprintf("--target='%s'", target))
	}
	for _, env := range appConf.Benchmark.Env {
		cmd = append(cmd, fmt.Sprintf("--env='%s'", env))
	}
	return strings.Join(cmd, " "), nil
}

func ApplicationBenchmark(ctx context.Context, log *logger.Logger, service gcloud.Service, targets []string) error {
	conf := service.Config()
	appConf := conf.Application
	runnerName := fmt.Sprintf("%s-application-benchmark", appConf.Name)
	log.Infof("[%s] creating or getting instance...", runnerName)
	instance, err := service.GetOrCreateInstance(ctx, runnerName, appConf.Benchmark.InstanceType)
	if err != nil {
		return err
	}
	// close open ssh connection
	defer instance.Close()

	log.Infof("[%s] external IP: %s", runnerName, instance.ExternalIP())
	log.Infof("[%s] setting up instance...", runnerName)
	var installBenchmarkTool gcloud.Action
	if appConf.Benchmark.Tool == "k6" {
		installBenchmarkTool = actions.NewActionInstallBinary(log, "k6", assets.K6)
	} else {
		installBenchmarkTool = actions.NewActionInstallArtillery(log)
	}
	err = instance.ExecuteActions(ctx,
		installBenchmarkTool,
		actions.NewActionInstallBinary(log, "application-benchmark-runner", assets.ApplicationBenchmarkRunner),
	)
	if err != nil {
		return err
	}
	cmd, err := getAppBenchRunnerCmd(conf.Timeout, appConf, targets)
	if err != nil {
		return err
	}
	log.Infof("[%s] running: %s", runnerName, cmd)
	return instance.RunWithLogger(ctx, func(stdout, stderr string) {
		log.Infof("[%s] %s%s", runnerName, stdout, stderr)
	}, cmd)
}
