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

type ConductorConfig struct {
	Project       string
	Region        string
	Zone          string
	InstanceType  string     `yaml:"instanceType"`
	SSHPrivateKey string     `yaml:"sshPrivateKey"`
	SSHSigner     ssh.Signer `yaml:"-"`
	GoVersion     string     `yaml:"goVersion"`
}

func NewConductorConfig(cmd *cobra.Command) (*ConductorConfig, error) {
	c := &ConductorConfig{
		Project:       viper.GetString("project"),
		Region:        viper.GetString("region"),
		Zone:          viper.GetString("zone"),
		InstanceType:  viper.GetString("instanceType"),
		SSHPrivateKey: viper.GetString("sshPrivateKey"),
		GoVersion:     viper.GetString("goVersion"),
	}

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
	if confErr != nil {
		return nil, confErr
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

	cli.Must(viper.BindPFlag("config", cmd.PersistentFlags().Lookup("config")))
	cli.Must(viper.BindPFlag("project", cmd.PersistentFlags().Lookup("project")))
	cli.Must(viper.BindPFlag("region", cmd.PersistentFlags().Lookup("region")))
	cli.Must(viper.BindPFlag("zone", cmd.PersistentFlags().Lookup("zone")))
	cli.Must(viper.BindPFlag("sshPrivateKey", cmd.PersistentFlags().Lookup("ssh-private-key")))
	cli.Must(viper.BindPFlag("instanceType", cmd.PersistentFlags().Lookup("instance-type")))
	cli.Must(viper.BindPFlag("goVersion", cmd.PersistentFlags().Lookup("go-version")))
}
