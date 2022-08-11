package netutil

import (
	"context"
	"net"
	"time"
)

func WaitForPortOpen(ctx context.Context, endpoint string) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(time.Second):
			conn, err := net.DialTimeout("tcp4", endpoint, time.Second)
			if err == nil {
				_ = conn.Close()
				return nil
			}
		}
	}
}
