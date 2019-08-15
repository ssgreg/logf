package logfc

import (
	"context"
	"testing"

	"github.com/ssgreg/logf"

	"github.com/stretchr/testify/assert"
)

func TestNewAndGet(t *testing.T) {
	ctx := context.Background()
	logger := logf.NewDisabledLogger()

	assert.Equal(t, logf.DisabledLogger(), Get(ctx))
	assert.Equal(t, logger, Get(New(ctx, logger)))
	assert.Equal(t, logger, MustGet(New(ctx, logger)))
	assert.Panics(t, func() {
		_ = MustGet(ctx)
	})
}

func TestWith(t *testing.T) {
	logger := logf.DisabledLogger()
	ctx := New(context.Background(), logger)

	assert.NotEqual(t, logger, Get(With(ctx, logf.Int("int", 42))))
	assert.NotEqual(t, logger, Get(MustWith(ctx, logf.Int("int", 42))))
}

func TestWithName(t *testing.T) {
	logger := logf.DisabledLogger()
	ctx := New(context.Background(), logger)

	assert.NotEqual(t, logger, Get(WithName(ctx, "n")))
	assert.NotEqual(t, logger, Get(MustWithName(ctx, "n")))
}

func TestCaller(t *testing.T) {
	logger := logf.DisabledLogger()
	ctx := New(context.Background(), logger)

	assert.NotEqual(t, logger, Get(WithCaller(ctx)))
	assert.NotEqual(t, logger, Get(MustWithCaller(ctx)))
}

func TestCallerSkip(t *testing.T) {
	logger := logf.DisabledLogger()
	ctx := New(context.Background(), logger)

	assert.NotEqual(t, logger, Get(WithCallerSkip(ctx, 1)))
	assert.NotEqual(t, logger, Get(MustWithCallerSkip(ctx, 1)))
}

func TestAtLevel(t *testing.T) {
	logger := logf.NewLogger(logf.LevelInfo, logf.NewUnbufferedEntryWriter(logf.NewDiscardAppender()))
	ctx := New(context.Background(), logger)

	called := false
	AtLevel(ctx, logf.LevelInfo, func(logf.LogFunc) {
		called = true
	})

	assert.Equal(t, true, called)
}

func TestMustAtLevel(t *testing.T) {
	logger := logf.NewLogger(logf.LevelInfo, logf.NewUnbufferedEntryWriter(logf.NewDiscardAppender()))
	ctx := New(context.Background(), logger)

	called := false
	MustAtLevel(ctx, logf.LevelInfo, func(logf.LogFunc) {
		called = true
	})

	assert.Equal(t, true, called)
}

func TestLevels(t *testing.T) {
	tests := []struct {
		Name        string
		Log         func(ctx context.Context, text string, fs ...logf.Field)
		Level       logf.Level
		ShouldPanic bool
	}{
		{
			Name:        "Debug",
			Log:         Debug,
			Level:       logf.LevelDebug,
			ShouldPanic: false,
		},
		{
			Name:        "MustDebug",
			Log:         MustDebug,
			Level:       logf.LevelDebug,
			ShouldPanic: true,
		},
		{
			Name:        "Info",
			Log:         Info,
			Level:       logf.LevelInfo,
			ShouldPanic: false,
		},
		{
			Name:        "MustInfo",
			Log:         MustInfo,
			Level:       logf.LevelInfo,
			ShouldPanic: true,
		},
		{
			Name:        "Warn",
			Log:         Warn,
			Level:       logf.LevelWarn,
			ShouldPanic: false,
		},
		{
			Name:        "MustWarn",
			Log:         MustWarn,
			Level:       logf.LevelWarn,
			ShouldPanic: true,
		},
		{
			Name:        "Error",
			Log:         Error,
			Level:       logf.LevelError,
			ShouldPanic: false,
		},
		{
			Name:        "MustError",
			Log:         MustError,
			Level:       logf.LevelError,
			ShouldPanic: true,
		},
	}
	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			if test.ShouldPanic {
				assert.Panics(t, func() {
					test.Log(context.Background(), "test")
				})
			} else {
				assert.NotPanics(t, func() {
					test.Log(context.Background(), "test")
				})
			}
			text := "test text"
			appender := mockAppender{}
			logger := logf.NewLogger(logf.LevelDebug, logf.NewUnbufferedEntryWriter(&appender))
			ctx := New(context.Background(), logger)
			test.Log(ctx, text)
			assert.Equal(t, 1, len(appender.entries))
			assert.Equal(t, text, appender.entries[0].Text)
			assert.Equal(t, test.Level, appender.entries[0].Level)
		})
	}
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
