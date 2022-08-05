package run

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"text/template"

	"github.com/christophwitzko/master-thesis/pkg/config"
	"github.com/christophwitzko/master-thesis/pkg/gcloud"
	"github.com/christophwitzko/master-thesis/pkg/gcloud/actions"
	"github.com/christophwitzko/master-thesis/pkg/logger"
)

type tmplData struct {
	Name     string
	RunIndex int
}

func applyTemplate(mbConf *config.ConductorMicrobenchmarkConfig, runIndex int, tmplStr string) (string, error) {
	tmpl, err := template.New("tmpl").Parse(tmplStr)
	if err != nil {
		return "", err
	}
	buf := &bytes.Buffer{}
	err = tmpl.Execute(buf, tmplData{
		Name:     mbConf.Name,
		RunIndex: runIndex,
	})
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

func getMbRunnerCmd(mbConf *config.ConductorMicrobenchmarkConfig, runIndex int) (string, error) {
	cmd := []string{
		"microbenchmark-runner",
		fmt.Sprintf("--run %d", runIndex),
		fmt.Sprintf("--git-repository='%s' --v1='%s' --v2='%s'", mbConf.Repository, mbConf.V1, mbConf.V2),
	}
	if mbConf.ExcludeFilter != "" {
		cmd = append(cmd, fmt.Sprintf("--exclude-filter='%s'", mbConf.ExcludeFilter))
	}
	if mbConf.IncludeFilter != "" {
		cmd = append(cmd, fmt.Sprintf("--include-filter='%s'", mbConf.IncludeFilter))
	}
	for _, output := range mbConf.Outputs {
		finalOutput, err := applyTemplate(mbConf, runIndex, output)
		if err != nil {
			return "", err
		}
		cmd = append(cmd, fmt.Sprintf("--output='%s'", finalOutput))
	}
	return strings.Join(cmd, " "), nil
}

func Microbenchmark(ctx context.Context, log *logger.Logger, service gcloud.Service, runIndex int) error {
	mbConf := service.Config().Microbenchmark
	runnerName := fmt.Sprintf("%s-runner-%d", mbConf.Name, runIndex)
	log.Infof("[%s] creating or getting instance...", runnerName)
	instance, err := service.GetOrCreateInstance(ctx, runnerName)
	if err != nil {
		return err
	}
	// close open ssh connection
	defer instance.Close()

	log.Infof("[%s] setting up instance...", runnerName)
	err = instance.ExecuteActions(ctx, actions.NewActionInstallGo(log), actions.NewActionInstallMicrobenchmarkRunner(log))
	if err != nil {
		return err
	}
	cmd, err := getMbRunnerCmd(mbConf, runIndex)
	if err != nil {
		return err
	}
	log.Infof("[%s] running: %s", runnerName, cmd)
	return instance.RunWithLogger(ctx, func(stdout, stderr string) {
		log.Infof("[%s] %s%s", runnerName, stdout, stderr)
	}, cmd)
}
