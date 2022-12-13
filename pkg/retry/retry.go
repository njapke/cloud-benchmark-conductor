package retry

import (
	"context"
	"time"

	"github.com/christophwitzko/masters-thesis/pkg/logger"
)

func HandleSilently(attempt int, err error) {}

func OnError(ctx context.Context, log *logger.Logger, prefix string, fn func() error) error {
	return OnErrorWithHandler(ctx, func(attempt int, err error) {
		log.Warnf("%s error at attempt %d: %v", prefix, attempt, err)
	}, fn)
}

func OnErrorWithHandler(ctx context.Context, handler func(attempt int, err error), fn func() error) error {
	var lastErr error
	for i := 1; i <= 3; i++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		if i > 1 {
			time.Sleep(500 * time.Millisecond)
		}
		err := fn()
		if err == nil {
			return nil
		}
		lastErr = err
		handler(i, err)
	}
	return lastErr
}
