package benchmark

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/christophwitzko/master-thesis/pkg/logger"
)

func RunArtillery(ctx context.Context, log *logger.Logger, configFile, targetEndpoint string) (ArtilleryResult, error) {
	normalizedTargetEndpoint := strings.ReplaceAll(targetEndpoint, ":", "_")
	configDir := filepath.Dir(configFile)
	outputFile := filepath.Join(configDir, fmt.Sprintf("%s.json", normalizedTargetEndpoint))
	args := []string{
		"run",
		fmt.Sprintf("--target=http://%s", targetEndpoint),
		fmt.Sprintf("--output=%s", outputFile),
		configFile,
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
		return ArtilleryResult{}, fmt.Errorf("artillery run failed: %w", err)
	}
	outputFileData, err := os.ReadFile(outputFile)
	if err != nil {
		return ArtilleryResult{}, fmt.Errorf("failed to read output file: %w", err)
	}
	var res ArtilleryResult
	err = json.Unmarshal(outputFileData, &res)
	if err != nil {
		return ArtilleryResult{}, fmt.Errorf("failed to unmarshal output file: %w", err)
	}
	return res, nil
}
