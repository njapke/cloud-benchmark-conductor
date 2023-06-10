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

type mbTmplData struct {
	Timestamp string
	Name      string
	RunIndex  int
	V1, V2    string
}

func applyMbOutputTemplate(mbConf *config.ConductorMicrobenchmarkConfig, runIndex int, tmplStr string) (string, error) {
	tmpl, err := template.New("tmpl").Parse(tmplStr)
	if err != nil {
		return "", err
	}
	buf := &bytes.Buffer{}
	err = tmpl.Execute(buf, mbTmplData{
		Timestamp: currentTimestamp,
		Name:      mbConf.Name,
		RunIndex:  runIndex,
		V1:        mbConf.V1,
		V2:        mbConf.V2,
	})
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

func getMbRunnerCmd(timeout time.Duration, mbConf *config.ConductorMicrobenchmarkConfig, runIndex int) (string, error) {
	cmd := []string{
		"microbenchmark-runner",
		fmt.Sprintf("--run %d", runIndex),
		fmt.Sprintf("--suite-runs %d", mbConf.SuiteRuns),
		fmt.Sprintf("--git-repository='%s' --v1='%s' --v2='%s'", mbConf.Repository, mbConf.V1, mbConf.V2),
		fmt.Sprintf("--timeout=%s", timeout),
	}
	if mbConf.ExcludeFilter != "" {
		cmd = append(cmd, fmt.Sprintf("--exclude-filter='%s'", mbConf.ExcludeFilter))
	}
	if mbConf.IncludeFilter != "" {
		cmd = append(cmd, fmt.Sprintf("--include-filter='%s'", mbConf.IncludeFilter))
	}
	if len(mbConf.Functions) != 0 {
		for _, function := range mbConf.Functions {
			cmd = append(cmd, fmt.Sprintf("--function='%s'", function))
		}
	}
	for _, output := range mbConf.Outputs {
		finalOutput, err := applyMbOutputTemplate(mbConf, runIndex, output)
		if err != nil {
			return "", err
		}
		cmd = append(cmd, fmt.Sprintf("--output='%s'", finalOutput))
	}
	for _, env := range mbConf.Env {
		cmd = append(cmd, fmt.Sprintf("--env='%s'", env))
	}
	return strings.Join(cmd, " "), nil
}

func Microbenchmark(ctx context.Context, log *logger.Logger, service gcloud.Service, runIndex int) error {
	conf := service.Config()
	mbConf := conf.Microbenchmark
	runnerName := fmt.Sprintf("%s-runner-%d", mbConf.Name, runIndex)
	log.Infof("[%s] creating or getting instance...", runnerName)
	instance, err := service.GetOrCreateInstance(ctx, runnerName, mbConf.InstanceType)
	if err != nil {
		return err
	}
	// close open ssh connection
	defer instance.Close()

	log.Infof("[%s] external IP: %s", runnerName, instance.ExternalIP())
	log.Infof("[%s] setting up instance...", runnerName)
	err = instance.ExecuteActions(ctx,
		actions.NewActionInstallGo(log),
		actions.NewActionInstallBinary(log, "microbenchmark-runner", assets.MicrobenchmarkRunner),
	)
	if err != nil {
		return err
	}
	cmd, err := getMbRunnerCmd(conf.Timeout, mbConf, runIndex)
	if err != nil {
		return err
	}
	log.Infof("[%s] running: %s", runnerName, cmd)
	return instance.RunWithLogger(ctx, func(stdout, stderr string) {
		log.Infof("[%s] %s%s", runnerName, stdout, stderr)
	}, cmd)
}
