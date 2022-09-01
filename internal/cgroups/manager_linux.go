package cgroups

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"syscall"

	"github.com/opencontainers/runtime-spec/specs-go"
	"golang.org/x/sys/unix"
)

// MemoryEventFD returns inotify file descriptor and 'memory.events' inotify watch descriptor
func (c *Manager) MemoryEventFD() (int, uint32, error) {
	fpath := filepath.Join(c.path, "memory.events")
	fd, err := syscall.InotifyInit()
	if err != nil {
		return 0, 0, errors.New("failed to create inotify fd")
	}
	wd, err := syscall.InotifyAddWatch(fd, fpath, unix.IN_MODIFY)
	if err != nil {
		syscall.Close(fd)
		return 0, 0, fmt.Errorf("failed to add inotify watch for %q: %w", fpath, err)
	}
	// monitor to detect process exit/cgroup deletion
	evpath := filepath.Join(c.path, "cgroup.events")
	if _, err = syscall.InotifyAddWatch(fd, evpath, unix.IN_MODIFY); err != nil {
		syscall.Close(fd)
		return 0, 0, fmt.Errorf("failed to add inotify watch for %q: %w", evpath, err)
	}

	return fd, uint32(wd), nil
}

func (c *Manager) waitForEvents(ec chan<- Event, errCh chan<- error) {
	defer close(errCh)

	fd, _, err := c.MemoryEventFD()
	if err != nil {
		errCh <- err
		return
	}
	defer syscall.Close(fd)

	for {
		buffer := make([]byte, syscall.SizeofInotifyEvent*10)
		bytesRead, err := syscall.Read(fd, buffer)
		if err != nil {
			errCh <- err
			return
		}
		if bytesRead >= syscall.SizeofInotifyEvent {
			out := make(map[string]interface{})
			if err := readKVStatsFile(c.path, "memory.events", out); err != nil {
				// When cgroup is deleted read may return -ENODEV instead of -ENOENT from open.
				if _, statErr := os.Lstat(filepath.Join(c.path, "memory.events")); !os.IsNotExist(statErr) {
					errCh <- err
				}
				return
			}
			e, err := parseMemoryEvents(out)
			if err != nil {
				errCh <- err
				return
			}
			ec <- e
			if c.isCgroupEmpty() {
				return
			}
		}
	}
}

func setDevices(path string, devices []specs.LinuxDeviceCgroup) error {
	if len(devices) == 0 {
		return nil
	}
	insts, license, err := DeviceFilter(devices)
	if err != nil {
		return err
	}
	dirFD, err := unix.Open(path, unix.O_DIRECTORY|unix.O_RDONLY|unix.O_CLOEXEC, 0o600)
	if err != nil {
		return fmt.Errorf("cannot get dir FD for %s", path)
	}
	defer unix.Close(dirFD)
	if _, err := LoadAttachCgroupDeviceFilter(insts, license, dirFD); err != nil {
		if !canSkipEBPFError(devices) {
			return err
		}
	}
	return nil
}
