package config

import (
	"fmt"
	"os"

	"github.com/christophwitzko/master-thesis/pkg/cli"
	"github.com/hashicorp/go-multierror"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type ConductorConfig struct {
	Project      string
	Region       string
	Zone         string
	InstanceType string `yaml:"instanceType"`
	SSHPublicKey string `yaml:"sshPublicKey"`
}

func NewConductorConfig(cmd *cobra.Command) (*ConductorConfig, error) {
	c := &ConductorConfig{
		Project:      viper.GetString("project"),
		Region:       viper.GetString("region"),
		Zone:         viper.GetString("zone"),
		InstanceType: viper.GetString("instanceType"),
		SSHPublicKey: viper.GetString("sshPublicKey"),
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
	if c.SSHPublicKey == "" {
		confErr = multierror.Append(confErr, fmt.Errorf("missing public key"))
	}
	if confErr != nil {
		return nil, confErr
	}

	return c, nil
}

func ConductorSetupFlagsAndViper(cmd *cobra.Command) {
	cmd.PersistentFlags().String("config", "", "config file")
	cmd.PersistentFlags().String("project", os.Getenv("CLOUDSDK_CORE_PROJECT"), "google cloud project")
	cmd.PersistentFlags().String("region", os.Getenv("CLOUDSDK_COMPUTE_REGION"), "compute region")
	cmd.PersistentFlags().String("zone", os.Getenv("CLOUDSDK_COMPUTE_ZONE"), "compute zone")
	cli.Must(viper.BindPFlags(cmd.PersistentFlags()))

	viper.SetDefault("instanceType", "f1-micro")
}

func InitConfig(cmd *cobra.Command, defaultConfigFile string) error {
	configFile := cli.MustGetString(cmd, "config")
	if configFile != "" {
		viper.SetConfigFile(configFile)
	} else {
		viper.AddConfigPath(".")
		viper.SetConfigName(defaultConfigFile)
	}
	viper.SetConfigType("yaml")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return err
		}
	}
	return nil
}
