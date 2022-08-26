package retry

import (
	"context"
	"fmt"
	"testing"

	"github.com/christophwitzko/master-thesis/pkg/logger"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/require"
)

func TestOnError(t *testing.T) {
	logrusLogger, hook := test.NewNullLogger()
	log := &logger.Logger{Logger: logrusLogger}
	err := OnError(context.Background(), log, "[test]", func() error {
		return fmt.Errorf("always error")
	})
	require.Error(t, err)
	require.Contains(t, hook.LastEntry().Message, "always error")
	require.Len(t, hook.Entries, 3)
}

func TestOnErrorFailOnce(t *testing.T) {
	logrusLogger, hook := test.NewNullLogger()
	log := &logger.Logger{Logger: logrusLogger}
	firstRun := true
	err := OnError(context.Background(), log, "[test]", func() error {
		if firstRun {
			firstRun = false
			return fmt.Errorf("always error")
		}
		return nil
	})
	require.NoError(t, err)
	require.Contains(t, hook.LastEntry().Message, "always error")
	require.Len(t, hook.Entries, 1)
}

func TestOnErrorWithHandler(t *testing.T) {
	attempts := 0
	err := OnErrorWithHandler(context.Background(), func(attempt int, err error) {
		attempts++
		require.Error(t, err)
	}, func() error {
		return fmt.Errorf("always error")
	})
	require.Error(t, err)
	require.Equal(t, 3, attempts)
}
