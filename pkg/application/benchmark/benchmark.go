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

	"cloud.google.com/go/storage"
	"github.com/christophwitzko/master-thesis/pkg/logger"
)

func uploadDataToBucket(ctx context.Context, config *Config, outputFile string, data []byte) error {
	client, err := storage.NewClient(ctx)
	if err != nil {
		return err
	}
	defer client.Close()

	// check if bucket exists
	bucket := client.Bucket(config.outputURL.Host)
	_, err = bucket.Attrs(ctx)
	if err != nil {
		if errors.Is(err, storage.ErrBucketNotExist) {
			return fmt.Errorf("bucket %s does not exist", config.outputURL.Host)
		}
		return err
	}

	objectName := strings.TrimPrefix(filepath.Join(config.outputURL.Path, outputFile), "/")
	objectWriter := bucket.Object(objectName).NewWriter(ctx)
	objectWriter.ContentType = "application/json"
	_, err = objectWriter.Write(data)
	if err != nil {
		return err
	}
	return objectWriter.Close()
}

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
	outputFileData, err := os.ReadFile(outputFile)
	if err != nil {
		return fmt.Errorf("failed to read output file: %w", err)
	}
	err = uploadDataToBucket(ctx, config, outputFileName, outputFileData)
	if err != nil {
		return err
	}
	log.Infof("[%s] results uploaded to %s/%s", targetEndpoint, config.OutputPath, outputFileName)
	return nil
}
