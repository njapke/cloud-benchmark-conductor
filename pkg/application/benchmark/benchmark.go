package benchmark

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/christophwitzko/master-thesis/pkg/gcloud/storage"
	"github.com/christophwitzko/master-thesis/pkg/logger"
	"github.com/christophwitzko/master-thesis/pkg/retry"
)

type Config struct {
	Tool          string
	ConfigDir     string
	ConfigFile    string
	OutputPath    string
	outputURLHost string
	outputURLPath string
}

func (c *Config) Validate() error {
	if c.ConfigFile == "" {
		return errors.New("config file is required")
	}
	if c.OutputPath == "" {
		return nil
	}
	var err error
	c.outputURLHost, c.outputURLPath, err = storage.ParseURL(c.OutputPath)
	return err
}

func (c *Config) GetOutputObjectName(fileName string) string {
	return fmt.Sprintf("gs://%s%s", c.outputURLHost, filepath.Join(c.outputURLPath, fileName))
}

func (c *Config) UploadToBucketFromReader(ctx context.Context, fileName string, inputFile io.Reader) error {
	return storage.UploadToBucket(ctx, c.outputURLHost, filepath.Join(c.outputURLPath, fileName), inputFile)
}

func (c *Config) UploadToBucketFromFile(ctx context.Context, fileName, inputFile string) error {
	return storage.UploadFileToBucket(ctx, c.outputURLHost, filepath.Join(c.outputURLPath, fileName), inputFile)
}

type TargetInfo struct {
	Name       string
	Endpoint   string
	OutputFile string
}

func (t *TargetInfo) OutputFileName() string {
	return filepath.Base(t.OutputFile)
}

func Run(ctx context.Context, log *logger.Logger, config *Config, targetInfo *TargetInfo) error {
	var args []string
	switch config.Tool {
	case "artillery":
		args = []string{
			"run",
			fmt.Sprintf("--target=http://%s", targetInfo.Endpoint),
			fmt.Sprintf("--output=%s", targetInfo.OutputFile),
			config.ConfigFile,
		}
	case "k6":
		args = []string{
			"run",
			"--env", fmt.Sprintf("target=%s", targetInfo.Endpoint),
			"--tag", fmt.Sprintf("version=%s", targetInfo.Name),
			"--out", fmt.Sprintf("csv=%s", targetInfo.OutputFile),
			config.ConfigFile,
		}
	default:
		return fmt.Errorf("unknown benchmark tool: %s", config.Tool)
	}
	return runWithArgs(ctx, log, config, targetInfo, args)
}

func runWithArgs(ctx context.Context, log *logger.Logger, config *Config, targetInfo *TargetInfo, args []string) error {
	log.Infof("[%s] running: %s %s", targetInfo.Name, config.Tool, strings.Join(args, " "))
	cmd := exec.CommandContext(ctx, config.Tool, args...)
	cmd.Dir = config.ConfigDir
	logPipeRead, logPipeWrite := io.Pipe()
	cmd.Stdout = logPipeWrite
	cmd.Stderr = logPipeWrite
	defer logPipeWrite.Close()
	logPrefix := fmt.Sprintf("[%s]", targetInfo.Name)
	go log.PrefixedReader(logPrefix, logPipeRead)
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("%s run failed: %w", config.Tool, err)
	}

	log.Infof("[%s] %s run finished", targetInfo.Name, config.Tool)
	if config.outputURLHost == "" {
		log.Warnf("[%s] no results output configured, skipping upload", targetInfo.Name)
		return nil
	}
	log.Infof("[%s] uploading results...", targetInfo.Name)
	outputFileName := targetInfo.OutputFileName()
	err = retry.OnError(ctx, log, logPrefix, func() error {
		return config.UploadToBucketFromFile(ctx, outputFileName, targetInfo.OutputFile)
	})
	if err != nil {
		return err
	}
	log.Infof("[%s] results uploaded to %s", targetInfo.Name, config.GetOutputObjectName(outputFileName))
	return nil
}
