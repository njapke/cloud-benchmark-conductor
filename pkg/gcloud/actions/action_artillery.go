package actions

import (
	"context"
	"fmt"
	"strings"

	"github.com/christophwitzko/masters-thesis/pkg/gcloud"
	"github.com/christophwitzko/masters-thesis/pkg/logger"
	"github.com/christophwitzko/masters-thesis/pkg/retry"
)

type actionInstallArtillery struct {
	log *logger.Logger
}

func NewActionInstallArtillery(log *logger.Logger) gcloud.Action {
	return &actionInstallArtillery{log: log}
}

func (a *actionInstallArtillery) Name() string {
	return "install-artillery"
}

const artilleryCheckString = "Artillery Core: 2."

func runInstallNode(ctx context.Context, instance gcloud.Instance, log *logger.Logger, lp string) error {
	err := retry.OnError(ctx, log, lp, func() error {
		log.Infof("%s installing nodesource repo...", lp)
		stdout, stderr, runErr := instance.Run(ctx, "curl -sL https://deb.nodesource.com/setup_16.x | sudo bash -")
		if runErr != nil {
			return fmt.Errorf("failed to run nodesource repo: %w\nSTDERR: %s\nSTDOUT: %s", runErr, stderr, stdout)
		}
		return nil
	})
	if err != nil {
		return err
	}
	err = retry.OnError(ctx, log, lp, func() error {
		log.Infof("%s installing node...", lp)
		stdout, stderr, runErr := instance.Run(ctx, "sudo apt-get install -y nodejs && node -v")
		if runErr != nil {
			return fmt.Errorf("failed to run install nodejs: %w\nSTDERR: %s\nSTDOUT: %s", runErr, stderr, stdout)
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

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
		err = runInstallNode(ctx, instance, a.log, lp)
		if err != nil {
			return err
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

	a.log.Infof("%s installing artillery...", lp)
	// create new cache archive with: sudo npm install --location=global artillery@latest && tar cvzf ~/artillery.tgz ./artillery/
	installArtillery := []string{
		"gsutil cp gs://mt-npm-cache/artillery.tgz ./",
		"sudo tar xzf ./artillery.tgz -C /usr/lib/node_modules/",
		"sudo ln -s /usr/lib/node_modules/artillery/bin/run /usr/bin/artillery",
	}
	stdout, stderr, err = instance.Run(ctx, strings.Join(installArtillery, " && "))
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
