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
