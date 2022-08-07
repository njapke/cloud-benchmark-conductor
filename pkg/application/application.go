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

	"github.com/christophwitzko/master-thesis/pkg/logger"
	"github.com/christophwitzko/master-thesis/pkg/merror"
)

func Build(ctx context.Context, log *logger.Logger, buildPath, outputFile string) error {
	args := []string{
		"build",
		"-o", outputFile,
		"./",
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

func Run(ctx context.Context, log *logger.Logger, execFile, bindAddress string) error {
	cmd := exec.Command(execFile)
	cmd.Dir = filepath.Dir(execFile)
	cmd.Env = append(os.Environ(), "CGO_ENABLED=0", fmt.Sprintf("BIND_ADDRESS=%s", bindAddress))
	logPipeRead, logPipeWrite := io.Pipe()
	cmd.Stdout = logPipeWrite
	cmd.Stderr = logPipeWrite
	cmd.SysProcAttr = &syscall.SysProcAttr{
		// prevent parent sending signals to child processes
		Setpgid: true,
	}
	defer logPipeWrite.Close()

	go log.PrefixedReader(fmt.Sprintf("|%s|", filepath.Base(execFile)), logPipeRead)
	log.Infof("running %s with BIND_ADDRESS=%s", execFile, bindAddress)
	errCh := make(chan error, 1)
	go func() {
		if err := cmd.Run(); err != nil {
			errCh <- err
		}
		close(errCh)
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
