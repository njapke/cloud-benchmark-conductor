package cgroups

import (
	"fmt"

	"github.com/christophwitzko/master-thesis/internal/cgroups"
)

var (
	defaultMountPoint = "/sys/fs/cgroup"
	defaultCgroupName = "/app-runner"
)

func Setup() error {
	m, err := cgroups.LoadManager(defaultMountPoint, defaultCgroupName)
	// return no error if group does not exist
	if err != nil {
		return err
	}
	err = m.Delete()
	if err != nil {
		return fmt.Errorf("failed to delete already existing cgroup: %w", err)
	}

	m, err = cgroups.NewManager(defaultMountPoint, defaultCgroupName, &cgroups.Resources{CPU: &cgroups.CPU{}})
	if err != nil {
		return fmt.Errorf("failed to create cgroup manager: %w", err)
	}

	// setup two equally weighted child groups
	weight := uint64(50)
	resources := &cgroups.Resources{
		CPU: &cgroups.CPU{
			Weight: &weight,
		},
	}
	_, err = m.NewChild("v1", resources)
	if err != nil {
		return fmt.Errorf("failed to create cgroup child group for v1: %w", err)
	}
	_, err = m.NewChild("v2", resources)
	if err != nil {
		return fmt.Errorf("failed to create cgroup child group for v2: %w", err)
	}
	return nil
}

func AddProcess(name string, pid int) error {
	m, err := cgroups.LoadManager(defaultMountPoint, defaultCgroupName+"/"+name)
	if err != nil {
		return err
	}
	return m.AddProc(uint64(pid))
}
