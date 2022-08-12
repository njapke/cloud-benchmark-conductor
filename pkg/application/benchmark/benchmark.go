package benchmark

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/christophwitzko/master-thesis/pkg/gcloud/storage"
	"github.com/christophwitzko/master-thesis/pkg/logger"
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

func RunArtillery(ctx context.Context, log *logger.Logger, config *Config, targetEndpoint string) error {
	configDir := filepath.Dir(config.ConfigFile)
	outputFileName := fmt.Sprintf("%s.json", targetEndpoint)
	outputFile := filepath.Join(configDir, outputFileName)

	args := []string{
		"run",
		fmt.Sprintf("--target=http://%s", targetEndpoint),
		fmt.Sprintf("--output=%s", outputFile),
		config.ConfigFile,
	}
	log.Infof("[%s] running: artillery %s", targetEndpoint, strings.Join(args, " "))
	cmd := exec.CommandContext(ctx, "artillery", args...)
	cmd.Dir = configDir
	logPipeRead, logPipeWrite := io.Pipe()
	cmd.Stdout = logPipeWrite
	cmd.Stderr = logPipeWrite
	defer logPipeWrite.Close()
	go log.PrefixedReader(fmt.Sprintf("[%s]", targetEndpoint), logPipeRead)
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("artillery run failed: %w", err)
	}

	log.Infof("[%s] artillery run finished", targetEndpoint)
	if config.outputURL == nil {
		log.Warnf("[%s] no results output configured, skipping upload", targetEndpoint)
		return nil
	}
	log.Infof("[%s] uploading results...", targetEndpoint)
	outputFileFd, err := os.Open(outputFile)
	if err != nil {
		return fmt.Errorf("failed to open output file: %w", err)
	}
	defer outputFileFd.Close()
	outputFileObjectName := filepath.Join(config.outputURL.Path, outputFileName)
	err = storage.UploadToBucket(ctx, config.outputURL.Host, filepath.Join(config.outputURL.Path, outputFileName), outputFileFd)
	if err != nil {
		return err
	}
	log.Infof("[%s] results uploaded to gs://%s%s", targetEndpoint, config.outputURL.Host, outputFileObjectName)
	return nil
}
