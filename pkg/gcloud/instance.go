package gcloud

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/christophwitzko/master-thesis/pkg/config"
	"github.com/christophwitzko/master-thesis/pkg/logger"
	"golang.org/x/crypto/ssh"
	"google.golang.org/api/compute/v1"
)

type Instance struct {
	config           *config.ConductorConfig
	internalInstance *compute.Instance
	sshClient        *SSHClient
}

func (i *Instance) Name() string {
	return i.internalInstance.Name
}

func (i *Instance) ExternalIP() string {
	return i.internalInstance.NetworkInterfaces[0].AccessConfigs[0].NatIP
}

func (i *Instance) SSHEndpoint() string {
	return i.ExternalIP() + ":22"
}

func (i *Instance) waitForSSHPortReady(ctx context.Context) error {
	publicSSHEndpoint := i.SSHEndpoint()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(time.Second):
			conn, err := net.DialTimeout("tcp4", publicSSHEndpoint, time.Second)
			if err == nil {
				_ = conn.Close()
				return nil
			}
		}
	}
}

func (i *Instance) establishSSHConnection(ctx context.Context) error {
	if i.sshClient != nil {
		return nil
	}
	if err := i.waitForSSHPortReady(ctx); err != nil {
		return err
	}
	sshEndpoint := i.SSHEndpoint()
	var dialer net.Dialer
	tcpConn, err := dialer.DialContext(ctx, "tcp4", sshEndpoint)
	if err != nil {
		return err
	}
	sshConn, chans, reqs, err := ssh.NewClientConn(tcpConn, sshEndpoint, &ssh.ClientConfig{
		User:            "ubuntu",
		Auth:            []ssh.AuthMethod{ssh.PublicKeys(i.config.SSHSigner)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	})
	if err != nil {
		return MaybeMultiError(fmt.Errorf("failed to create ssh client: %w", err), tcpConn.Close())
	}
	i.sshClient = &SSHClient{sshClient: ssh.NewClient(sshConn, chans, reqs)}
	return nil
}

func (i *Instance) Close() error {
	if i.sshClient != nil {
		err := i.sshClient.Close()
		i.sshClient = nil
		return err
	}
	return nil
}

func (i *Instance) Run(ctx context.Context, logger *logger.Logger, cmd string) error {
	if err := i.establishSSHConnection(ctx); err != nil {
		return err
	}
	return i.sshClient.Run(ctx, func(out string, err string) {
		logger.Printf("[%s|%s] %s%s", i.Name(), cmd, out, err)
	}, cmd)
}
