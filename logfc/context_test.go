package logfc

import (
	"context"
	"runtime"
	"strings"
	"testing"

	"github.com/ssgreg/logf/v2"

	"github.com/stretchr/testify/assert"
)

func TestNewAndGet(t *testing.T) {
	ctx := context.Background()
	logger := logf.NewDisabledLogger()

	assert.Equal(t, logf.DisabledLogger(), Get(ctx))
	assert.Equal(t, logger, Get(New(ctx, logger)))
}

func TestWith(t *testing.T) {
	logger := logf.DisabledLogger()
	ctx := New(context.Background(), logger)

	assert.NotEqual(t, logger, Get(With(ctx, logf.Int("int", 42))))
}

func TestWithName(t *testing.T) {
	logger := logf.DisabledLogger()
	ctx := New(context.Background(), logger)

	assert.NotEqual(t, logger, Get(WithName(ctx, "n")))
}

func TestCaller(t *testing.T) {
	logger := logf.DisabledLogger().WithCaller(false)
	ctx := New(context.Background(), logger)

	assert.NotEqual(t, logger, Get(WithCaller(ctx)))
}

func TestCallerSkip(t *testing.T) {
	logger := logf.DisabledLogger()
	ctx := New(context.Background(), logger)

	assert.NotEqual(t, logger, Get(WithCallerSkip(ctx, 1)))
}

func TestAtLevel(t *testing.T) {
	logger := logf.NewLogger(logf.NewUnbufferedEntryWriter(logf.LevelDebug, logf.NewDiscardAppender()))
	ctx := New(context.Background(), logger)

	called := false
	AtLevel(ctx, logf.LevelInfo, func(logf.LogFunc) {
		called = true
	})

	assert.Equal(t, true, called)
}

func TestLevels(t *testing.T) {
	tests := []struct {
		Name  string
		Log   func(ctx context.Context, text string, fs ...logf.Field)
		Level logf.Level
	}{
		{"Debug", Debug, logf.LevelDebug},
		{"Info", Info, logf.LevelInfo},
		{"Warn", Warn, logf.LevelWarn},
		{"Error", Error, logf.LevelError},
	}
	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			// No panic when no logger in context.
			assert.NotPanics(t, func() {
				test.Log(context.Background(), "test")
			})

			text := "test text"
			appender := mockAppender{}
			logger := logf.NewLogger(logf.NewUnbufferedEntryWriter(logf.LevelDebug, &appender))
			ctx := New(context.Background(), logger)
			test.Log(ctx, text)
			assert.Equal(t, 1, len(appender.entries))
			assert.Equal(t, text, appender.entries[0].Text)
			assert.Equal(t, test.Level, appender.entries[0].Level)
		})
	}
}

func TestAtLevelCallerPointsToCallSite(t *testing.T) {
	appender := mockAppender{}
	logger := logf.NewLogger(logf.NewUnbufferedEntryWriter(logf.LevelDebug, &appender))
	ctx := New(context.Background(), logger)

	_, _, expectedLine, _ := runtime.Caller(0)
	AtLevel(ctx, logf.LevelInfo, func(log logf.LogFunc) {
		log(ctx, "at-level-test") // expectedLine + 2
	})

	assert.Equal(t, 1, len(appender.entries))
	pc := appender.entries[0].CallerPC
	assert.NotZero(t, pc, "CallerPC should be set")

	frames := runtime.CallersFrames([]uintptr{pc})
	f, _ := frames.Next()

	assert.Equal(t, expectedLine+2, f.Line,
		"caller should point to log() call site, got %s:%d", f.File, f.Line)
	assert.True(t, strings.HasSuffix(f.File, "context_test.go"),
		"expected context_test.go, got %s", f.File)
}

func TestCallerPointsToCallSite(t *testing.T) {
	appender := mockAppender{}
	logger := logf.NewLogger(logf.NewUnbufferedEntryWriter(logf.LevelDebug, &appender))
	ctx := New(context.Background(), logger)

	_, expectedFile, expectedLine, _ := runtime.Caller(0)
	Debug(ctx, "caller-test") // expectedLine + 1

	assert.Equal(t, 1, len(appender.entries))
	pc := appender.entries[0].CallerPC
	assert.NotZero(t, pc, "CallerPC should be set")

	frames := runtime.CallersFrames([]uintptr{pc})
	f, _ := frames.Next()

	assert.Equal(t, expectedLine+1, f.Line)
	assert.True(t, strings.HasSuffix(f.File, expectedFile[strings.LastIndex(expectedFile, "/")+1:]),
		"expected file %s, got %s", expectedFile, f.File)
}

type mockAppender struct {
	entries []logf.Entry
}

func (a *mockAppender) Append(entry logf.Entry) error {
	a.entries = append(a.entries, entry)

	return nil
}

func (a *mockAppender) Sync() (err error) {
	return nil
}

func (a *mockAppender) Flush() error {
	return nil
}
