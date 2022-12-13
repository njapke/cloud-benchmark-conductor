package application

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/christophwitzko/masters-thesis/pkg/logger"
	"github.com/christophwitzko/masters-thesis/pkg/merror"
)

type PidCallbackFunc func(pid int) error

func Build(ctx context.Context, log *logger.Logger, buildPath, buildPackage, outputFile string) error {
	if !strings.HasPrefix(buildPackage, "./") {
		return fmt.Errorf("build package must be a relative path")
	}
	args := []string{
		"build",
		"-o", outputFile,
		buildPackage,
	}
	cmd := exec.CommandContext(ctx, "go", args...)
	cmd.Dir = buildPath
	cmd.Env = append(os.Environ(), "CGO_ENABLED=0")
	logPipeRead, logPipeWrite := io.Pipe()
	cmd.Stdout = logPipeWrite
	cmd.Stderr = logPipeWrite
	defer logPipeWrite.Close()
	go log.PrefixedReader(fmt.Sprintf("building %s", buildPath), logPipeRead)
	log.Infof("running in %s: go %s", buildPath, strings.Join(args, " "))
	return cmd.Run()
}

func Run(ctx context.Context, log *logger.Logger, execFile string, env []string, pidCallback PidCallbackFunc) error {
	cmd := exec.Command(execFile)
	cmd.Dir = filepath.Dir(execFile)
	cmd.Env = append(os.Environ(), env...)
	logPipeRead, logPipeWrite := io.Pipe()
	cmd.Stdout = logPipeWrite
	cmd.Stderr = logPipeWrite
	cmd.SysProcAttr = &syscall.SysProcAttr{
		// prevent parent sending signals to child processes
		Setpgid: true,
	}
	defer logPipeWrite.Close()

	go log.PrefixedReader(fmt.Sprintf("|%s|", filepath.Base(execFile)), logPipeRead)
	log.Infof("running %s with env=%v", execFile, env)
	errCh := make(chan error, 1)
	go func() {
		defer close(errCh)
		if err := cmd.Start(); err != nil {
			errCh <- err
			return
		}
		if pidCallback != nil {
			if err := pidCallback(cmd.Process.Pid); err != nil {
				errCh <- fmt.Errorf("failed to call pid callback: %w", err)
				return
			}
		}
		if err := cmd.Wait(); err != nil {
			errCh <- err
			return
		}
	}()

	select {
	case <-ctx.Done():
		log.Warnf("killing %s", execFile)
		killErr := cmd.Process.Signal(syscall.SIGKILL)
		waitErr := <-errCh // should be a signal: killed error
		return merror.MaybeMultiError(ctx.Err(), killErr, waitErr)
	case err := <-errCh:
		return err
	}
}
