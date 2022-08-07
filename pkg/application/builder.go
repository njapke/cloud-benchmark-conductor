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
	cmd := exec.CommandContext(ctx, execFile)
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
	return cmd.Run()
}
