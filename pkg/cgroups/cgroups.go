package cgroups

import (
	"fmt"

	"github.com/containerd/cgroups/v2"
)

var (
	defaultMountPoint = "/sys/fs/cgroup"
	defaultCgroupName = "/app-runner"
)

func Setup() error {
	m, err := v2.LoadManager(defaultMountPoint, defaultCgroupName)
	if err == nil {
		err = m.Delete()
		if err != nil {
			return fmt.Errorf("failed to delete already existing cgroup: %v", err)
		}
	}
	m, err = v2.NewManager(defaultMountPoint, defaultCgroupName, &v2.Resources{CPU: &v2.CPU{}})
	if err != nil {
		return fmt.Errorf("failed to create cgroup manager: %w", err)
	}

	// setup two equally weighted child groups
	weight := uint64(50)
	resources := &v2.Resources{
		CPU: &v2.CPU{
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

func AddProcess(name string, pid uint64) error {
	m, err := v2.LoadManager(defaultMountPoint, defaultCgroupName+"/"+name)
	if err != nil {
		return err
	}
	return m.AddProc(pid)
}
