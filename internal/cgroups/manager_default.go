//go:build !linux

package cgroups

import (
	"fmt"

	"github.com/opencontainers/runtime-spec/specs-go"
)

func (c *Manager) MemoryEventFD() (int, uint32, error) {
	return 0, 0, fmt.Errorf("MemoryEventFD: not supported")
}

func (c *Manager) waitForEvents(ec chan<- Event, errCh chan<- error) {
	errCh <- fmt.Errorf("waitForEvents: not supported")
	close(ec)
	close(errCh)
}

func setDevices(path string, devices []specs.LinuxDeviceCgroup) error {
	if len(devices) == 0 {
		return nil
	}
	return fmt.Errorf("setDevices: not supported")
}
