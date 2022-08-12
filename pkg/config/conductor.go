package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/christophwitzko/master-thesis/pkg/cli"
	"github.com/hashicorp/go-multierror"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/crypto/ssh"
)

type ConductorMicrobenchmarkConfig struct {
	Name          string
	Repository    string
	Runs          int
	SuiteRuns     int `yaml:"suiteRuns"`
	V1, V2        string
	ExcludeFilter string `yaml:"excludeFilter"`
	IncludeFilter string `yaml:"includeFilter"`
	Functions     []string
	Outputs       []string `yaml:"outputs"`
}

func (c *ConductorMicrobenchmarkConfig) Validate() error {
	var confErr error
	if c.Repository == "" {
		confErr = multierror.Append(confErr, fmt.Errorf("missing microbenchmark repository"))
	}
	if c.V1 == "" {
		confErr = multierror.Append(confErr, fmt.Errorf("missing microbenchmark v1"))
	}
	if c.V2 == "" {
		confErr = multierror.Append(confErr, fmt.Errorf("missing microbenchmark v2"))
	}
	if len(c.Functions) != 0 && (c.IncludeFilter != "" || c.ExcludeFilter != "") {
		confErr = multierror.Append(confErr, fmt.Errorf("cannot use functions and include/exclude filters"))
	}
	return confErr
}

type ConductorApplicationConfig struct {
	Name       string
	Repository string
	V1, V2     string
	Package    string
	LogFilter  string `yaml:"logFilter"`
	Benchmark  *ConductorApplicationBenchmarkConfig
}

func (c *ConductorApplicationConfig) Validate() error {
	var confErr error
	if c.Repository == "" {
		confErr = multierror.Append(confErr, fmt.Errorf("missing application repository"))
	}
	if c.V1 == "" {
		confErr = multierror.Append(confErr, fmt.Errorf("missing application v1"))
	}
	if c.V2 == "" {
		confErr = multierror.Append(confErr, fmt.Errorf("missing application v2"))
	}
	if err := c.Benchmark.Validate(); err != nil {
		confErr = multierror.Append(confErr, err)
	}
	return confErr
}

type ConductorApplicationBenchmarkConfig struct {
	Config    string
	Reference string
	Output    string
}

func (c *ConductorApplicationBenchmarkConfig) Validate() error {
	var confErr error
	if c.Reference == "" {
		confErr = multierror.Append(confErr, fmt.Errorf("missing application benchmark reference"))
	}
	return confErr
}

type ConductorConfig struct {
	Project        string
	Region         string
	Zone           string
	InstanceType   string     `yaml:"instanceType"`
	SSHPrivateKey  string     `yaml:"sshPrivateKey"`
	SSHSigner      ssh.Signer `yaml:"-"`
	GoVersion      string     `yaml:"goVersion"`
	Timeout        time.Duration
	Microbenchmark *ConductorMicrobenchmarkConfig
	Application    *ConductorApplicationConfig
}

func (c *ConductorConfig) Validate() error {
	var confErr error
	if c.Project == "" {
		confErr = multierror.Append(confErr, fmt.Errorf("missing project"))
	}
	if c.Region == "" {
		confErr = multierror.Append(confErr, fmt.Errorf("missing region"))
	}
	if c.Zone == "" {
		confErr = multierror.Append(confErr, fmt.Errorf("missing zone"))
	}
	if c.InstanceType == "" {
		confErr = multierror.Append(confErr, fmt.Errorf("missing instance type"))
	}
	if c.SSHPrivateKey == "" {
		confErr = multierror.Append(confErr, fmt.Errorf("missing ssh private key"))
	}

	if err := c.Microbenchmark.Validate(); err != nil {
		confErr = multierror.Append(confErr, err)
	}
	if err := c.Application.Validate(); err != nil {
		confErr = multierror.Append(confErr, err)
	}

	return confErr
}

func NewConductorConfig(cmd *cobra.Command) (*ConductorConfig, error) {
	c := &ConductorConfig{
		Project:       viper.GetString("project"),
		Region:        viper.GetString("region"),
		Zone:          viper.GetString("zone"),
		InstanceType:  viper.GetString("instanceType"),
		SSHPrivateKey: viper.GetString("sshPrivateKey"),
		GoVersion:     viper.GetString("goVersion"),
		Timeout:       viper.GetDuration("timeout"),
		Microbenchmark: &ConductorMicrobenchmarkConfig{
			Name:          viper.GetString("microbenchmark.name"),
			Repository:    viper.GetString("microbenchmark.repository"),
			Runs:          viper.GetInt("microbenchmark.runs"),
			SuiteRuns:     viper.GetInt("microbenchmark.suiteRuns"),
			V1:            viper.GetString("microbenchmark.v1"),
			V2:            viper.GetString("microbenchmark.v2"),
			ExcludeFilter: viper.GetString("microbenchmark.excludeFilter"),
			IncludeFilter: viper.GetString("microbenchmark.includeFilter"),
			Functions:     viper.GetStringSlice("microbenchmark.functions"),
			Outputs:       viper.GetStringSlice("microbenchmark.outputs"),
		},
		Application: &ConductorApplicationConfig{
			Name:       viper.GetString("application.name"),
			Repository: viper.GetString("application.repository"),
			V1:         viper.GetString("application.v1"),
			V2:         viper.GetString("application.v2"),
			Package:    viper.GetString("application.package"),
			LogFilter:  viper.GetString("application.logFilter"),
			Benchmark: &ConductorApplicationBenchmarkConfig{
				Config:    viper.GetString("application.benchmark.config"),
				Reference: viper.GetString("application.benchmark.reference"),
				Output:    viper.GetString("application.benchmark.output"),
			},
		},
	}

	if err := c.Validate(); err != nil {
		return nil, err
	}

	var privateKeyData []byte
	if strings.HasPrefix(c.SSHPrivateKey, "-----BEGIN OPENSSH PRIVATE KEY-----") {
		// load private key directly from config
		privateKeyData = []byte(c.SSHPrivateKey)
	} else {
		// load private key form file
		pkFileData, err := os.ReadFile(c.SSHPrivateKey)
		if err != nil {
			return nil, err
		}
		privateKeyData = pkFileData
	}

	sshSigner, err := ssh.ParsePrivateKey(privateKeyData)
	if err != nil {
		return nil, err
	}
	c.SSHSigner = sshSigner
	return c, nil
}

func ConductorSetupFlagsAndViper(cmd *cobra.Command) {
	cmd.PersistentFlags().StringP("config", "c", "", "config file")

	cmd.PersistentFlags().String("project", os.Getenv("CLOUDSDK_CORE_PROJECT"), "google cloud project")
	cmd.PersistentFlags().String("region", os.Getenv("CLOUDSDK_COMPUTE_REGION"), "compute region")
	cmd.PersistentFlags().String("zone", os.Getenv("CLOUDSDK_COMPUTE_ZONE"), "compute zone")
	cmd.PersistentFlags().StringP("ssh-private-key", "i", "", "path to ssh private key")
	cmd.PersistentFlags().String("instance-type", "f1-micro", "instance type")
	cmd.PersistentFlags().String("go-version", "1.19", "go version")
	cmd.PersistentFlags().String("microbenchmark-name", "mb", "name of the microbenchmark")
	cmd.PersistentFlags().String("microbenchmark-repository", "", "repository of the microbenchmark")
	cmd.PersistentFlags().Int("microbenchmark-runs", 3, "number of parallel runs")
	cmd.PersistentFlags().Int("microbenchmark-suite-runs", 3, "number of suite runs")
	cmd.PersistentFlags().String("microbenchmark-v1", "", "v1 of the microbenchmark to run")
	cmd.PersistentFlags().String("microbenchmark-v2", "", "v2 of the microbenchmark to run")
	cmd.PersistentFlags().String("microbenchmark-exclude-filter", "", "exclude filter for the microbenchmark")
	cmd.PersistentFlags().String("microbenchmark-include-filter", "", "include filter for the microbenchmark")
	cmd.PersistentFlags().StringArray("microbenchmark-output", []string{"-"}, "outputs of the microbenchmark")
	cmd.PersistentFlags().Duration("timeout", 30*time.Minute, "timeout for the benchmark execution")
	cmd.PersistentFlags().String("application-name", "app", "name of the application")
	cmd.PersistentFlags().String("application-repository", "", "repository of the application")
	cmd.PersistentFlags().String("application-v1", "", "v1 of the application to run")
	cmd.PersistentFlags().String("application-v2", "", "v2 of the application to run")
	cmd.PersistentFlags().String("application-package", "./", "package that should be build and run")
	cmd.PersistentFlags().String("application-log-filter", "", "filter application logs")
	cmd.PersistentFlags().String("application-benchmark-config", "", "application benchmark config")
	cmd.PersistentFlags().String("application-benchmark-reference", "", "application benchmark reference")
	cmd.PersistentFlags().String("application-benchmark-output", "", "application benchmark output path")
	cmd.PersistentFlags().StringArray("microbenchmark-function", []string{}, "functions to include in the microbenchmark")

	cli.Must(viper.BindPFlag("project", cmd.PersistentFlags().Lookup("project")))
	cli.Must(viper.BindPFlag("region", cmd.PersistentFlags().Lookup("region")))
	cli.Must(viper.BindPFlag("zone", cmd.PersistentFlags().Lookup("zone")))
	cli.Must(viper.BindPFlag("sshPrivateKey", cmd.PersistentFlags().Lookup("ssh-private-key")))
	cli.Must(viper.BindPFlag("instanceType", cmd.PersistentFlags().Lookup("instance-type")))
	cli.Must(viper.BindPFlag("goVersion", cmd.PersistentFlags().Lookup("go-version")))
	cli.Must(viper.BindPFlag("microbenchmark.name", cmd.PersistentFlags().Lookup("microbenchmark-name")))
	cli.Must(viper.BindPFlag("microbenchmark.repository", cmd.PersistentFlags().Lookup("microbenchmark-repository")))
	cli.Must(viper.BindPFlag("microbenchmark.runs", cmd.PersistentFlags().Lookup("microbenchmark-runs")))
	cli.Must(viper.BindPFlag("microbenchmark.suiteRuns", cmd.PersistentFlags().Lookup("microbenchmark-suite-runs")))
	cli.Must(viper.BindPFlag("microbenchmark.v1", cmd.PersistentFlags().Lookup("microbenchmark-v1")))
	cli.Must(viper.BindPFlag("microbenchmark.v2", cmd.PersistentFlags().Lookup("microbenchmark-v2")))
	cli.Must(viper.BindPFlag("microbenchmark.excludeFilter", cmd.PersistentFlags().Lookup("microbenchmark-exclude-filter")))
	cli.Must(viper.BindPFlag("microbenchmark.includeFilter", cmd.PersistentFlags().Lookup("microbenchmark-include-filter")))
	cli.Must(viper.BindPFlag("microbenchmark.outputs", cmd.PersistentFlags().Lookup("microbenchmark-output")))
	cli.Must(viper.BindPFlag("timeout", cmd.PersistentFlags().Lookup("timeout")))
	cli.Must(viper.BindPFlag("application.name", cmd.PersistentFlags().Lookup("application-name")))
	cli.Must(viper.BindPFlag("application.repository", cmd.PersistentFlags().Lookup("application-repository")))
	cli.Must(viper.BindPFlag("application.v1", cmd.PersistentFlags().Lookup("application-v1")))
	cli.Must(viper.BindPFlag("application.v2", cmd.PersistentFlags().Lookup("application-v2")))
	cli.Must(viper.BindPFlag("application.package", cmd.PersistentFlags().Lookup("application-package")))
	cli.Must(viper.BindPFlag("application.logFilter", cmd.PersistentFlags().Lookup("application-log-filter")))
	cli.Must(viper.BindPFlag("application.benchmark.config", cmd.PersistentFlags().Lookup("application-benchmark-config")))
	cli.Must(viper.BindPFlag("application.benchmark.reference", cmd.PersistentFlags().Lookup("application-benchmark-reference")))
	cli.Must(viper.BindPFlag("application.benchmark.output", cmd.PersistentFlags().Lookup("application-benchmark-output")))
	cli.Must(viper.BindPFlag("microbenchmark.functions", cmd.PersistentFlags().Lookup("microbenchmark-function")))
}
