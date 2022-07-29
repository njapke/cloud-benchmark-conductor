package gcloud

import (
	"bufio"
	"bytes"
	"context"
	"fmt"

	"github.com/bramvdbogaerde/go-scp"
	"golang.org/x/crypto/ssh"
)

type LoggerFunction func(stdout string, stderr string)

type SSHClient struct {
	sshClient *ssh.Client
}

func (c *SSHClient) Close() error {
	return c.sshClient.Close()
}

func (c *SSHClient) openSSHSession(loggerFn LoggerFunction) (*ssh.Session, error) {
	session, err := c.sshClient.NewSession()
	if err != nil {
		return nil, fmt.Errorf("failed to create ssh session: %w", err)
	}

	stdout, err := session.StdoutPipe()
	if err != nil {
		return nil, MaybeMultiError(fmt.Errorf("failed to create stdout pipe: %w", err), session.Close())
	}
	stderr, err := session.StderrPipe()
	if err != nil {
		return nil, MaybeMultiError(fmt.Errorf("failed to create stderr pipe: %w", err), session.Close())
	}
	go func() {
		logLineScanner := bufio.NewScanner(stdout)
		for logLineScanner.Scan() {
			loggerFn(logLineScanner.Text(), "")
		}
	}()
	go func() {
		logLineScanner := bufio.NewScanner(stderr)
		for logLineScanner.Scan() {
			loggerFn("", logLineScanner.Text())
		}
	}()
	return session, nil
}

func (c *SSHClient) Run(ctx context.Context, loggerFn LoggerFunction, cmd string) error {
	session, err := c.openSSHSession(loggerFn)
	if err != nil {
		return err
	}
	defer session.Close()

	if err := session.Start(cmd); err != nil {
		return fmt.Errorf("failed to start command %s: %w", cmd, err)
	}

	waitErrCh := make(chan error, 1)
	go func() {
		waitErrCh <- session.Wait()
	}()

	select {
	case <-ctx.Done():
		// send SIGINT to the process
		signalErr := session.Signal(ssh.SIGINT)
		// wait for termination
		waitErr := <-waitErrCh
		return MaybeMultiError(ctx.Err(), signalErr, waitErr)
	case err := <-waitErrCh:
		return err
	}
}

func (c *SSHClient) CopyFile(ctx context.Context, data *bytes.Reader, remotePath string) error {
	scpClient, err := scp.NewClientBySSH(c.sshClient)
	if err != nil {
		return fmt.Errorf("failed to create scp client: %w", err)
	}
	defer scpClient.Close()
	return scpClient.Copy(ctx, data, remotePath, "0755", data.Size())
}
