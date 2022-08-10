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
	"github.com/christophwitzko/master-thesis/pkg/merror"
	"golang.org/x/crypto/ssh"
	computepb "google.golang.org/genproto/googleapis/cloud/compute/v1"
)

type Instance interface {
	Config() *config.ConductorConfig
	Name() string
	ExternalIP() string
	InternalIP() string
	SSHEndpoint() string
	LogPrefix() string
	RunWithLogger(ctx context.Context, logger LoggerFunc, cmd string) error
	Run(ctx context.Context, cmd string) (stdout, stderr string, err error)
	Reconnect(ctx context.Context) error
	CopyFile(ctx context.Context, data *bytes.Reader, file string) error
	ExecuteActions(ctx context.Context, actions ...Action) error
	Close() error
}

type instance struct {
	config           *config.ConductorConfig
	internalInstance *computepb.Instance
	sshPortReady     bool
	sshPortMutex     sync.Mutex
	sshClient        *sshClient
	sshClientMutex   sync.Mutex
}

// Config returns the global config
func (i *instance) Config() *config.ConductorConfig {
	return i.config
}

// Name returns the instance name without the prefix
func (i *instance) Name() string {
	return trimPrefixName(*i.internalInstance.Name)
}

// ExternalIP returns the external IP of the instance
func (i *instance) ExternalIP() string {
	return *i.internalInstance.NetworkInterfaces[0].AccessConfigs[0].NatIP
}

// InternalIP returns the external IP of the instance
func (i *instance) InternalIP() string {
	return *i.internalInstance.NetworkInterfaces[0].NetworkIP
}

// SSHEndpoint returns the public SSH endpoint of the instance
func (i *instance) SSHEndpoint() string {
	return i.ExternalIP() + ":22"
}

// LogPrefix returns the log prefix that contains the instance name
func (i *instance) LogPrefix() string {
	return fmt.Sprintf("[%s]", i.Name())
}

func (i *instance) waitForSSHPortReady(ctx context.Context) error {
	i.sshPortMutex.Lock()
	defer i.sshPortMutex.Unlock()
	if err := ctx.Err(); err != nil {
		return err
	}
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

func (i *instance) newSSHClient(ctx context.Context) (*sshClient, error) {
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
		Auth:            []ssh.AuthMethod{ssh.PublicKeys(i.config.SSHSigner)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	})
	if err != nil {
		return nil, merror.MaybeMultiError(fmt.Errorf("failed to create ssh client: %w", err), tcpConn.Close())
	}
	return &sshClient{sshClient: ssh.NewClient(sshConn, chans, reqs)}, nil
}

func (i *instance) ensureSSHClient(ctx context.Context) error {
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

// Close SSH connection if open
func (i *instance) Close() error {
	i.sshClientMutex.Lock()
	defer i.sshClientMutex.Unlock()
	if i.sshClient != nil {
		err := i.sshClient.Close()
		i.sshClient = nil
		return err
	}
	return nil
}

// Reconnect to instance via SSH
func (i *instance) Reconnect(ctx context.Context) error {
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

// RunWithLogger runs a command on the instance and calls a LoggerFunc for every new line in stdout and stderr
func (i *instance) RunWithLogger(ctx context.Context, logger LoggerFunc, cmd string) error {
	if err := i.ensureSSHClient(ctx); err != nil {
		return err
	}
	return i.sshClient.Run(ctx, logger, cmd)
}

// Run runs a command on the instance and returns stdout and stderr as string
func (i *instance) Run(ctx context.Context, cmd string) (string, string, error) {
	var stdout strings.Builder
	var stderr strings.Builder
	err := i.RunWithLogger(ctx, func(out, err string) {
		if out != "" {
			stdout.WriteString(out + "\n")
		}
		if err != "" {
			stderr.WriteString(err + "\n")
		}
	}, cmd)
	return stdout.String(), stderr.String(), err
}

// CopyFile copies a file from a bytes.Reader to a remote instance
func (i *instance) CopyFile(ctx context.Context, data *bytes.Reader, file string) error {
	if err := i.ensureSSHClient(ctx); err != nil {
		return err
	}
	return i.sshClient.CopyFile(ctx, data, file, "0755")
}

// ExecuteActions executes a list of actions on the instance
func (i *instance) ExecuteActions(ctx context.Context, actions ...Action) error {
	for _, a := range actions {
		if err := a.Run(ctx, i); err != nil {
			return fmt.Errorf("failed to run action %s: %w", a.Name(), err)
		}
	}
	return nil
}
