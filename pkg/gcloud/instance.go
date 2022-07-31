package gcloud

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/christophwitzko/master-thesis/pkg/config"
	"github.com/christophwitzko/master-thesis/pkg/logger"
	"golang.org/x/crypto/ssh"
	computepb "google.golang.org/genproto/googleapis/cloud/compute/v1"
)

type Instance struct {
	Config           *config.ConductorConfig
	internalInstance *computepb.Instance
	sshPortReady     bool
	sshPortMutex     sync.Mutex
	sshClient        *SSHClient
	sshClientMutex   sync.Mutex
}

func (i *Instance) Name() string {
	return *i.internalInstance.Name
}

func (i *Instance) ExternalIP() string {
	return *i.internalInstance.NetworkInterfaces[0].AccessConfigs[0].NatIP
}

func (i *Instance) SSHEndpoint() string {
	return i.ExternalIP() + ":22"
}

func (i *Instance) logPrefix() string {
	return fmt.Sprintf("[%s]", i.Name())
}

func (i *Instance) waitForSSHPortReady(ctx context.Context) error {
	i.sshPortMutex.Lock()
	defer i.sshPortMutex.Unlock()
	if i.sshPortReady {
		return nil
	}
	publicSSHEndpoint := i.SSHEndpoint()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(time.Second):
			conn, err := net.DialTimeout("tcp4", publicSSHEndpoint, time.Second)
			if err == nil {
				_ = conn.Close()
				i.sshPortReady = true
				return nil
			}
		}
	}
}

func (i *Instance) ensureSSHClient(ctx context.Context) error {
	i.sshClientMutex.Lock()
	defer i.sshClientMutex.Unlock()
	if i.sshClient != nil {
		return nil
	}
	sshClient, err := i.newSSHClient(ctx)
	if err != nil {
		return err
	}
	i.sshClient = sshClient
	return nil
}

func (i *Instance) newSSHClient(ctx context.Context) (*SSHClient, error) {
	if err := i.waitForSSHPortReady(ctx); err != nil {
		return nil, err
	}
	sshEndpoint := i.SSHEndpoint()
	var dialer net.Dialer
	tcpConn, err := dialer.DialContext(ctx, "tcp4", sshEndpoint)
	if err != nil {
		return nil, err
	}
	sshConn, chans, reqs, err := ssh.NewClientConn(tcpConn, sshEndpoint, &ssh.ClientConfig{
		User:            "ubuntu",
		Auth:            []ssh.AuthMethod{ssh.PublicKeys(i.Config.SSHSigner)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	})
	if err != nil {
		return nil, MaybeMultiError(fmt.Errorf("failed to create ssh client: %w", err), tcpConn.Close())
	}
	return &SSHClient{sshClient: ssh.NewClient(sshConn, chans, reqs)}, nil
}

func (i *Instance) Close() error {
	i.sshClientMutex.Lock()
	defer i.sshClientMutex.Unlock()
	if i.sshClient != nil {
		err := i.sshClient.Close()
		i.sshClient = nil
		return err
	}
	return nil
}

func (i *Instance) Reconnect(ctx context.Context) error {
	i.sshClientMutex.Lock()
	defer i.sshClientMutex.Unlock()
	if i.sshClient != nil {
		// acquire session lock to prevent interrupting commands
		i.sshClient.sshSessionMutex.Lock()
		defer i.sshClient.sshSessionMutex.Unlock()
		if err := i.sshClient.Close(); err != nil {
			return err
		}
	}
	var err error
	i.sshClient, err = i.newSSHClient(ctx)
	return err
}

func (i *Instance) RunWithLog(ctx context.Context, logger *logger.Logger, cmd string) error {
	if err := i.ensureSSHClient(ctx); err != nil {
		return err
	}
	lp := i.logPrefix()
	return i.sshClient.Run(ctx, func(out string, err string) {
		ioType := "OUT"
		ioVal := out
		if out == "" {
			ioType = "ERR"
			ioVal = err
		}
		logger.Printf("%s |%s|%s| %s", lp, cmd, ioType, ioVal)
	}, cmd)
}

func (i *Instance) Run(ctx context.Context, cmd string) (string, string, error) {
	if err := i.ensureSSHClient(ctx); err != nil {
		return "", "", err
	}
	var stdout strings.Builder
	var stderr strings.Builder
	err := i.sshClient.Run(ctx, func(out string, err string) {
		if out != "" {
			stdout.WriteString(out + "\n")
		}
		if err != "" {
			stderr.WriteString(err + "\n")
		}
	}, cmd)
	return stdout.String(), stderr.String(), err
}

func (i *Instance) CopyFile(ctx context.Context, data *bytes.Reader, file string) error {
	if err := i.ensureSSHClient(ctx); err != nil {
		return err
	}
	return i.sshClient.CopyFile(ctx, data, file, "0755")
}

func (i *Instance) ExecuteActions(ctx context.Context, actions ...Action) error {
	for _, action := range actions {
		if err := action.Run(ctx, i); err != nil {
			return fmt.Errorf("failed to run action %s: %w", action.Name(), err)
		}
	}
	return nil
}
