package profile

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

var (
	// this limits the maximum profile duration to 1 minute
	profileClientTimeout = time.Minute
	profileClient        = &http.Client{
		Transport: &http.Transport{
			ResponseHeaderTimeout: profileClientTimeout,
			ExpectContinueTimeout: profileClientTimeout,
			IdleConnTimeout:       profileClientTimeout,
		},
		Timeout: profileClientTimeout,
	}
)

func Fetch(ctx context.Context, endpoint, outputFile string) error {
	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return err
	}
	out, err := os.OpenFile(outputFile, os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	res, err := profileClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		// read the response body for a more detailed error message
		body, _ := io.ReadAll(res.Body)
		return fmt.Errorf("unexpected %d status code while fetching profile: %s", res.StatusCode, string(body))
	}
	_, err = io.Copy(out, res.Body)
	if err != nil {
		return err
	}
	return nil
}
