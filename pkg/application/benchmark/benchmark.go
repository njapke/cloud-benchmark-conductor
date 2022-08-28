package benchmark

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/christophwitzko/master-thesis/pkg/gcloud/storage"
	"github.com/christophwitzko/master-thesis/pkg/logger"
	"github.com/christophwitzko/master-thesis/pkg/retry"
)

type Config struct {
	ConfigFile string
	OutputPath string
	outputURL  *url.URL
}

func (c *Config) Validate() error {
	if c.ConfigFile == "" {
		return errors.New("config file is required")
	}
	if c.OutputPath == "" {
		return nil
	}
	u, err := url.Parse(c.OutputPath)
	if err != nil {
		return fmt.Errorf("invalid output path: %w", err)
	}
	if u.Scheme != "gs" && u.Scheme != "gcs" {
		return fmt.Errorf("invalid output path scheme: %s", u.Scheme)
	}
	c.outputURL = u
	return nil
}

func (c *Config) GetOutputObjectName(fileName string) string {
	return fmt.Sprintf("gs://%s%s", c.outputURL.Host, filepath.Join(c.outputURL.Path, fileName))
}

func (c *Config) UploadToBucketFromReader(ctx context.Context, fileName string, inputFile io.Reader) error {
	return storage.UploadToBucket(ctx, c.outputURL.Host, filepath.Join(c.outputURL.Path, fileName), inputFile)
}

func (c *Config) UploadToBucketFromFile(ctx context.Context, fileName, inputFile string) error {
	return storage.UploadFileToBucket(ctx, c.outputURL.Host, filepath.Join(c.outputURL.Path, fileName), inputFile)
}

type TargetInfo struct {
	Name       string
	Endpoint   string
	OutputFile string
}

func (t *TargetInfo) OutputFileName() string {
	return filepath.Base(t.OutputFile)
}

func RunArtillery(ctx context.Context, log *logger.Logger, config *Config, targetInfo *TargetInfo) error {
	args := []string{
		"run",
		fmt.Sprintf("--target=http://%s", targetInfo.Endpoint),
		fmt.Sprintf("--output=%s", targetInfo.OutputFile),
		config.ConfigFile,
	}
	log.Infof("[%s] running: artillery %s", targetInfo.Name, strings.Join(args, " "))
	cmd := exec.CommandContext(ctx, "artillery", args...)
	cmd.Dir = filepath.Dir(config.ConfigFile)
	logPipeRead, logPipeWrite := io.Pipe()
	cmd.Stdout = logPipeWrite
	cmd.Stderr = logPipeWrite
	defer logPipeWrite.Close()
	logPrefix := fmt.Sprintf("[%s]", targetInfo.Name)
	go log.PrefixedReader(logPrefix, logPipeRead)
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("artillery run failed: %w", err)
	}

	log.Infof("[%s] artillery run finished", targetInfo.Name)
	if config.outputURL == nil {
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
