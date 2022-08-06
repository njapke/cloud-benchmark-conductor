package actions

import (
	"context"
	"fmt"
	"strings"

	"github.com/christophwitzko/master-thesis/pkg/gcloud"
	"github.com/christophwitzko/master-thesis/pkg/logger"
)

type actionInstallArtillery struct {
	log *logger.Logger
}

func NewActionInstallArtillery(log *logger.Logger) *actionInstallArtillery {
	return &actionInstallArtillery{log: log}
}

func (a *actionInstallArtillery) Name() string {
	return "install-artillery"
}

const artilleryCheckString = "Artillery Core: 2."

func (a *actionInstallArtillery) Run(ctx context.Context, instance gcloud.Instance) error {
	lp := instance.LogPrefix() + "[" + a.Name() + "]"
	a.log.Infof("%s installing node & artillery...", lp)
	stdout, stderr, err := instance.Run(ctx, "node -v || true")
	if err != nil {
		return fmt.Errorf("failed to run nodesource repo: %w\nSTDERR: %s\nSTDOUT: %s", err, stderr, stdout)
	}
	nodeVersion, _, _ := strings.Cut(stdout, "\n")
	installNode := true
	if strings.HasPrefix(nodeVersion, "v16") {
		installNode = false
		a.log.Infof("%s node %s is already installed!", lp, nodeVersion)
	}
	if installNode {
		a.log.Infof("%s installing node...", lp)
		stdout, stderr, err = instance.Run(ctx, "curl -sL https://deb.nodesource.com/setup_16.x | sudo bash -")
		if err != nil {
			return fmt.Errorf("failed to run nodesource repo: %w\nSTDERR: %s\nSTDOUT: %s", err, stderr, stdout)
		}

		stdout, stderr, err = instance.Run(ctx, "sudo apt-get install -y nodejs && node -v")
		if err != nil {
			return fmt.Errorf("failed to run install nodejs: %w\nSTDERR: %s\nSTDOUT: %s", err, stderr, stdout)
		}
	}

	stdout, stderr, err = instance.Run(ctx, "artillery --version || true")
	if err != nil {
		return fmt.Errorf("failed to get artillery version: %w\nSTDERR: %s\nSTDOUT: %s", err, stderr, stdout)
	}
	if strings.Contains(stdout, artilleryCheckString) {
		a.log.Infof("%s artillery is already installed!", lp)
		return nil
	}

	a.log.Infof("%s installing or updating artillery...", lp)
	// TODO: very slow, maybe introduce caching?
	stdout, stderr, err = instance.Run(ctx, "sudo npm install --location=global artillery@latest")
	if err != nil {
		return fmt.Errorf("failed to install artillery: %w\nSTDERR: %s\nSTDOUT: %s", err, stderr, stdout)
	}

	stdout, stderr, err = instance.Run(ctx, "artillery --version")
	if err != nil {
		return fmt.Errorf("failed to get artillery version: %w\nSTDERR: %s\nSTDOUT: %s", err, stderr, stdout)
	}
	if !strings.Contains(stdout, artilleryCheckString) {
		return fmt.Errorf("artillery is not installed correctly: %w\nSTDERR: %s\nSTDOUT: %s", err, stderr, stdout)
	}
	return nil
}
