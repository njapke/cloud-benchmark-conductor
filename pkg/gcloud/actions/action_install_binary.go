package actions

import (
	"context"
	"fmt"
	"strings"

	"github.com/christophwitzko/masters-thesis/pkg/assets"
	"github.com/christophwitzko/masters-thesis/pkg/gcloud"
	"github.com/christophwitzko/masters-thesis/pkg/logger"
)

type actionInstallBinary struct {
	log        *logger.Logger
	binaryName string
	binary     assets.Binary
}

func NewActionInstallBinary(log *logger.Logger, name string, binary assets.Binary) gcloud.Action {
	return &actionInstallBinary{
		log:        log,
		binaryName: name,
		binary:     binary,
	}
}

func (a *actionInstallBinary) Name() string {
	return fmt.Sprintf("install-binary-%s", a.binaryName)
}

func (a *actionInstallBinary) getBinaryHash(ctx context.Context, instance gcloud.Instance) (string, error) {
	stdout, stderr, err := instance.Run(ctx, fmt.Sprintf("sha512sum /usr/bin/%s || true", a.binaryName))
	if err != nil {
		return "", fmt.Errorf("failed to run go version: %w\nSTDERR: %s\nSTDOUT: %s", err, stderr, stdout)
	}
	stdout, _, _ = strings.Cut(stdout, " ")
	return stdout, nil
}

func (a *actionInstallBinary) Run(ctx context.Context, instance gcloud.Instance) error {
	lp := instance.LogPrefix() + "[" + a.Name() + "]"
	a.log.Infof("%s installing %s...", lp, a.binaryName)
	foundMbHash, err := a.getBinaryHash(ctx, instance)
	if err != nil {
		return err
	}
	if foundMbHash == a.binary.GetHash() {
		a.log.Infof("%s %s is already installed", lp, a.binaryName)
		return nil
	}

	err = instance.CopyFile(ctx, a.binary.GetReader(), fmt.Sprintf("/tmp/%s", a.binaryName))
	if err != nil {
		return err
	}
	stdout, stderr, err := instance.Run(ctx, fmt.Sprintf("sudo mv /tmp/%s /usr/bin/", a.binaryName))
	if err != nil {
		return fmt.Errorf("failed to move binary: %w\nSTDERR: %s\nSTDOUT: %s", err, stderr, stdout)
	}
	return nil
}
