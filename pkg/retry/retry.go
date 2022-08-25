package retry

import (
	"context"
	"time"

	"github.com/christophwitzko/master-thesis/pkg/logger"
)

func OnError(ctx context.Context, log *logger.Logger, prefix string, fn func() error) error {
	var lastErr error
	for i := 1; i <= 3; i++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		err := fn()
		if err == nil {
			return nil
		}
		lastErr = err
		log.Warnf("%s error at attempt %d: %v", prefix, i, err)
		time.Sleep(time.Second)
	}
	return lastErr
}
