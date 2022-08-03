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

func applyTemplate(tmplStr string, runIndex int) (string, error) {
	tmpl, err := template.New("tmpl").Parse(tmplStr)
	if err != nil {
		return "", err
	}
	buf := &bytes.Buffer{}
	err = tmpl.Execute(buf, struct{ RunIndex int }{RunIndex: runIndex})
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
		cmd = append(cmd, fmt.Sprintf("--include-filter='%s'", mbConf.ExcludeFilter))
	}
	for _, output := range mbConf.Outputs {
		finalOutput, err := applyTemplate(output, runIndex)
		if err != nil {
			return "", err
		}
		cmd = append(cmd, fmt.Sprintf("--output='%s'", finalOutput))
	}
	return strings.Join(cmd, " "), nil
}

func Microbenchmark(ctx context.Context, log *logger.Logger, service *gcloud.Service, mbConf *config.ConductorMicrobenchmarkConfig, runIndex int) error {
	runnerName := fmt.Sprintf("runner-%d", runIndex)
	log.Infof("creating instance %s...", runnerName)
	instance, err := service.CreateInstance(ctx, runnerName)
	if err != nil {
		return err
	}
	// close open ssh connection
	defer instance.Close()

	err = instance.ExecuteActions(ctx, actions.NewActionInstallGo(log), actions.NewActionInstallMicrobenchmarkRunner(log))
	if err != nil {
		return err
	}
	cmd, err := getMbRunnerCmd(mbConf, runIndex)
	if err != nil {
		return err
	}
	log.Infof("running: %s", cmd)
	return instance.RunWithLog(ctx, log, cmd)
}
