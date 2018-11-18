package logf

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoggerDisabled(t *testing.T) {
	logger := NewDisabledLogger()

	logger.WithCallerSkip(0)
	logger.WithCaller()
	logger.WithName("")
	logger.With(String("", ""))

	logger.Debug("")
	logger.Info("")
	logger.Warn("")
	logger.Error("")

	logger.AtLevel(LevelError, func(LogFunc) { require.FailNow(t, "can't be here") })
}
