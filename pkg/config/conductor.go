package config

import (
	"fmt"
	"os"
	"strings"

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
	V1, V2        string
	ExcludeFilter string   `yaml:"excludeFilter"`
	IncludeFilter string   `yaml:"includeFilter"`
	Outputs       []string `yaml:"outputs"`
}

type ConductorConfig struct {
	Project        string
	Region         string
	Zone           string
	InstanceType   string     `yaml:"instanceType"`
	SSHPrivateKey  string     `yaml:"sshPrivateKey"`
	SSHSigner      ssh.Signer `yaml:"-"`
	GoVersion      string     `yaml:"goVersion"`
	Microbenchmark *ConductorMicrobenchmarkConfig
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

	if c.Microbenchmark.Repository == "" {
		confErr = multierror.Append(confErr, fmt.Errorf("missing microbenchmark repository"))
	}
	if c.Microbenchmark.V1 == "" {
		confErr = multierror.Append(confErr, fmt.Errorf("missing microbenchmark v1"))
	}
	if c.Microbenchmark.V2 == "" {
		confErr = multierror.Append(confErr, fmt.Errorf("missing microbenchmark v2"))
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
		Microbenchmark: &ConductorMicrobenchmarkConfig{
			Name:          viper.GetString("microbenchmark.name"),
			Repository:    viper.GetString("microbenchmark.repository"),
			Runs:          viper.GetInt("microbenchmark.runs"),
			V1:            viper.GetString("microbenchmark.v1"),
			V2:            viper.GetString("microbenchmark.v2"),
			ExcludeFilter: viper.GetString("microbenchmark.excludeFilter"),
			IncludeFilter: viper.GetString("microbenchmark.includeFilter"),
			Outputs:       viper.GetStringSlice("microbenchmark.outputs"),
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
	cmd.PersistentFlags().String("go-version", "1.18.4", "go version")
	cmd.PersistentFlags().String("microbenchmark-name", "mb", "name of the microbenchmark")
	cmd.PersistentFlags().String("microbenchmark-repository", "", "repository of the microbenchmark")
	cmd.PersistentFlags().Int("microbenchmark-runs", 3, "number of parallel runs")
	cmd.PersistentFlags().String("microbenchmark-v1", "", "v1 of the microbenchmark to run")
	cmd.PersistentFlags().String("microbenchmark-v2", "", "v2 of the microbenchmark to run")
	cmd.PersistentFlags().String("microbenchmark-exclude-filter", "", "exclude filter for the microbenchmark")
	cmd.PersistentFlags().String("microbenchmark-include-filter", "", "include filter for the microbenchmark")
	cmd.PersistentFlags().StringArray("microbenchmark-output", []string{"-"}, "outputs of the microbenchmark")

	cli.Must(viper.BindPFlag("config", cmd.PersistentFlags().Lookup("config")))
	cli.Must(viper.BindPFlag("project", cmd.PersistentFlags().Lookup("project")))
	cli.Must(viper.BindPFlag("region", cmd.PersistentFlags().Lookup("region")))
	cli.Must(viper.BindPFlag("zone", cmd.PersistentFlags().Lookup("zone")))
	cli.Must(viper.BindPFlag("sshPrivateKey", cmd.PersistentFlags().Lookup("ssh-private-key")))
	cli.Must(viper.BindPFlag("instanceType", cmd.PersistentFlags().Lookup("instance-type")))
	cli.Must(viper.BindPFlag("goVersion", cmd.PersistentFlags().Lookup("go-version")))
	cli.Must(viper.BindPFlag("microbenchmark.name", cmd.PersistentFlags().Lookup("microbenchmark-name")))
	cli.Must(viper.BindPFlag("microbenchmark.repository", cmd.PersistentFlags().Lookup("microbenchmark-repository")))
	cli.Must(viper.BindPFlag("microbenchmark.runs", cmd.PersistentFlags().Lookup("microbenchmark-runs")))
	cli.Must(viper.BindPFlag("microbenchmark.v1", cmd.PersistentFlags().Lookup("microbenchmark-v1")))
	cli.Must(viper.BindPFlag("microbenchmark.v2", cmd.PersistentFlags().Lookup("microbenchmark-v2")))
	cli.Must(viper.BindPFlag("microbenchmark.excludeFilter", cmd.PersistentFlags().Lookup("microbenchmark-exclude-filter")))
	cli.Must(viper.BindPFlag("microbenchmark.includeFilter", cmd.PersistentFlags().Lookup("microbenchmark-include-filter")))
	cli.Must(viper.BindPFlag("microbenchmark.outputs", cmd.PersistentFlags().Lookup("microbenchmark-output")))
}
