package logf

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var ctx = context.Background()

func TestLoggerNew(t *testing.T) {
	w := &testEntryWriter{}
	logger := NewLogger(w)

	assert.Equal(t, 0, logger.callerSkip)
	assert.Equal(t, true, logger.addCaller)
	assert.Nil(t, logger.bag)
	assert.Empty(t, logger.name)
	assert.Equal(t, w, logger.w)
}

func TestLoggerCallerSpecifiedByDefault(t *testing.T) {
	w := &testEntryWriter{}
	logger := NewLogger(w)

	logger.Error(ctx, "")
	assert.NotZero(t, w.Entry.CallerPC)
	file, _ := callerFrame(w.Entry.CallerPC)
	assert.Equal(t, "logf/logger_test.go", fileWithPackage(file))
}

func TestLoggerCallerDisabled(t *testing.T) {
	w := &testEntryWriter{}
	logger := NewLogger(w).WithCaller(false)

	logger.Error(ctx, "")
	assert.Zero(t, w.Entry.CallerPC)
}

func TestLoggerCallerSpecifiedWithSkip(t *testing.T) {
	w := &testEntryWriter{}
	logger := NewLogger(w).WithCaller(true).WithCallerSkip(1)

	logger.Error(ctx, "")
	assert.NotZero(t, w.Entry.CallerPC)
	file, _ := callerFrame(w.Entry.CallerPC)
	assert.Equal(t, "testing/testing.go", fileWithPackage(file))
}

func TestLoggerNoNameByDefault(t *testing.T) {
	w := &testEntryWriter{}
	logger := NewLogger(w)

	logger.Error(ctx, "")
	assert.Empty(t, w.Entry.LoggerName)
}

func TestLoggerEmptyName(t *testing.T) {
	w := &testEntryWriter{}
	logger := NewLogger(w).WithName("")

	logger.Error(ctx, "")
	assert.Empty(t, w.Entry.LoggerName)
}

func TestLoggerName(t *testing.T) {
	w := &testEntryWriter{}

	// Set a name for the logger.
	logger := NewLogger(w).WithName("1")
	logger.Error(ctx, "")
	assert.Equal(t, "1", w.Entry.LoggerName)

	// Set a nested name for the logger.
	logger = logger.WithName("2")
	logger.Error(ctx, "")
	assert.Equal(t, "1.2", w.Entry.LoggerName)
}

func TestLoggerLevelFiltering(t *testing.T) {
	a := &testAppender{}
	w := NewSyncWriter(LevelError, a)
	logger := NewLogger(w)

	logger.Info(ctx, "filtered")
	assert.Empty(t, a.Entries)

	logger.Error(ctx, "passed")
	assert.Len(t, a.Entries, 1)
	assert.Equal(t, "passed", a.Entries[0].Text)
}

func TestLoggerAtLevel(t *testing.T) {
	a := &testAppender{}
	w := NewSyncWriter(LevelError, a)
	logger := NewLogger(w)

	// Expected the callback should be called with the same severity level.
	called := false
	logger.AtLevel(ctx, LevelError, func(log LogFunc) {
		called = true
		assert.NotEmpty(t, log)
		log(ctx, "at-level")
	})
	assert.Len(t, a.Entries, 1)
	assert.Equal(t, "at-level", a.Entries[0].Text)
	assert.True(t, called)

	// Expected the callback should NOT be called with DEBUG severity level.
	called = false
	logger.AtLevel(ctx, LevelDebug, func(log LogFunc) {
		called = true
	})
	assert.False(t, called)
}

func TestLoggerNoFieldsByDefault(t *testing.T) {
	w := &testEntryWriter{}

	logger := NewLogger(w)
	logger.Error(ctx, "")
	assert.Nil(t, w.Entry.LoggerBag)
}

func TestLoggerFields(t *testing.T) {
	w := &testEntryWriter{}

	// Add one Field.
	logger := NewLogger(w).With(String("first", "v"), String("second", "v"))
	logger.Error(ctx, "")
	assert.Equal(t, 2, len(w.Entry.LoggerBag.Fields()))

	// Add second Field.
	logger = logger.With(String("third", "v"))
	logger.Error(ctx, "")
	assert.Equal(t, 3, len(w.Entry.LoggerBag.Fields()))

	// Check order. First field should go first.
	assert.Equal(t, "first", w.Entry.LoggerBag.Fields()[0].Key)
	assert.Equal(t, "second", w.Entry.LoggerBag.Fields()[1].Key)
	assert.Equal(t, "third", w.Entry.LoggerBag.Fields()[2].Key)
}


func TestLoggerBag(t *testing.T) {
	w := &testEntryWriter{}
	logger := NewLogger(w)

	logger.Error(ctx, "")
	assert.Nil(t, w.Entry.LoggerBag)

	logger = logger.With(String("k", "v"))
	logger.Error(ctx, "")
	assert.NotNil(t, w.Entry.LoggerBag)
	assert.Equal(t, logger.bag, w.Entry.LoggerBag)
}

func TestLoggerLeveledLog(t *testing.T) {
	w := &testEntryWriter{}
	logger := NewLogger(w)

	cases := []struct {
		Level Level
		Fn    func(context.Context, string, ...Field)
	}{
		{LevelError, logger.Error},
		{LevelWarn, logger.Warn},
		{LevelInfo, logger.Info},
		{LevelDebug, logger.Debug},
	}

	for _, c := range cases {
		t.Run(c.Level.String(), func(t *testing.T) {
			c.Fn(ctx, c.Level.String(), Int("key", 42))

			assert.Equal(t, c.Level.String(), w.Entry.Text)
			assert.Equal(t, 1, len(w.Entry.Fields))
			assert.Equal(t, c.Level, w.Entry.Level)
			assert.WithinDuration(t, time.Now(), w.Entry.Time, time.Second*2)
		})
	}
}

func TestLoggerChecker(t *testing.T) {
	logger := DisabledLogger()

	logger.Error(ctx, "")
	logger.Warn(ctx, "")
	logger.Info(ctx, "")
	logger.Debug(ctx, "")
	// No panic — disabled logger discards everything.
}

func TestLoggerDisabled(t *testing.T) {
	logger := DisabledLogger()
	assert.Equal(t, 0, logger.callerSkip)
	assert.Equal(t, true, logger.addCaller)
	assert.Nil(t, logger.bag)
	assert.Empty(t, logger.name)
	assert.NotNil(t, logger.w)

	loggerWithCallerSkip := logger.WithCallerSkip(1)
	assert.Equal(t, 1, loggerWithCallerSkip.callerSkip)

	loggerWithCaller := logger.WithCaller(true)
	assert.Equal(t, true, loggerWithCaller.addCaller)

	loggerWithName := logger.WithName("name")
	assert.Equal(t, "name", loggerWithName.name)

	loggerWithFields := logger.With(String("k", "v"))
	assert.NotNil(t, loggerWithFields.bag)

	// Check function not panic.
	logger.Debug(ctx, "")
	logger.Info(ctx, "")
	logger.Warn(ctx, "")
	logger.Error(ctx, "")
	logger.AtLevel(ctx, LevelError, func(LogFunc) { require.FailNow(t, "can't be here") })
}

func TestUnbufferedWriter(t *testing.T) {
	a := &testAppender{}
	w := NewSyncWriter(LevelDebug, a)
	logger := NewLogger(w)

	// Check function not panic.
	logger.Debug(ctx, "debug")
	logger.Info(ctx, "info")
	logger.Warn(ctx, "warn")
	logger.Error(ctx, "error")
	logger.AtLevel(ctx, LevelError, func(log LogFunc) {
		log(ctx, "at-level-error")
	})

	require.Equal(t, 5, len(a.Entries))

	assert.Equal(t, "debug", a.Entries[0].Text)
	assert.Equal(t, "info", a.Entries[1].Text)
	assert.Equal(t, "warn", a.Entries[2].Text)
	assert.Equal(t, "error", a.Entries[3].Text)
	assert.Equal(t, "at-level-error", a.Entries[4].Text)

	assert.Equal(t, LevelDebug, a.Entries[0].Level)
	assert.Equal(t, LevelInfo, a.Entries[1].Level)
	assert.Equal(t, LevelWarn, a.Entries[2].Level)
	assert.Equal(t, LevelError, a.Entries[3].Level)
	assert.Equal(t, LevelError, a.Entries[4].Level)
}

func TestLoggerWithGroup(t *testing.T) {
	w := &testEntryWriter{}
	logger := NewLogger(w).WithGroup("http")

	logger.Error(ctx, "done", Int("status", 200))
	require.NotNil(t, w.Entry.LoggerBag)
	assert.Equal(t, "http", w.Entry.LoggerBag.Group())
}

func TestLoggerWithGroupEmpty(t *testing.T) {
	w := &testEntryWriter{}
	logger := NewLogger(w)
	same := logger.WithGroup("")

	// Empty group name returns the same logger.
	assert.Equal(t, logger, same)
}

func TestLoggerWithGroupChain(t *testing.T) {
	w := &testEntryWriter{}
	logger := NewLogger(w).
		With(String("service", "api")).
		WithGroup("http").
		With(String("method", "GET"))

	logger.Error(ctx, "done")
	bag := w.Entry.LoggerBag

	// Child node has method field.
	assert.Equal(t, 1, len(bag.OwnFields()))
	assert.Equal(t, "method", bag.OwnFields()[0].Key)

	// Parent is group node.
	assert.Equal(t, "http", bag.Parent().Group())

	// Grandparent has service field.
	assert.Equal(t, 1, len(bag.Parent().Parent().OwnFields()))
	assert.Equal(t, "service", bag.Parent().Parent().OwnFields()[0].Key)
}

func TestLoggerSlog(t *testing.T) {
	sink := &testEntryWriter{}
	logger := NewLogger(sink).WithName("fb").WithName("agent")
	slogger := logger.Slog()

	slogger.Info("request")

	require.NotNil(t, sink.Entry)
	assert.Equal(t, "request", sink.Entry.Text)
	assert.Equal(t, LevelInfo, sink.Entry.Level)
	assert.Equal(t, "fb.agent", sink.Entry.LoggerName)
}

func TestLoggerSlogWithFields(t *testing.T) {
	sink := &testEntryWriter{}
	logger := NewLogger(sink).With(String("env", "prod"))
	slogger := logger.Slog()

	slogger.Info("hello", "key", "value")

	require.NotNil(t, sink.Entry)
	assert.Equal(t, "hello", sink.Entry.Text)
	// Logger's bag is transferred.
	assert.NotNil(t, sink.Entry.LoggerBag)
	assert.Equal(t, "env", sink.Entry.LoggerBag.OwnFields()[0].Key)
	// Per-call fields from slog.
	require.Equal(t, 1, len(sink.Entry.Fields))
	assert.Equal(t, "key", sink.Entry.Fields[0].Key)
}

func TestLoggerSlogNoName(t *testing.T) {
	sink := &testEntryWriter{}
	logger := NewLogger(sink)
	slogger := logger.Slog()

	slogger.Info("hello")

	require.NotNil(t, sink.Entry)
	assert.Equal(t, "", sink.Entry.LoggerName)
}

func TestContext(t *testing.T) {
	// Check if no logger is associated with the Context — returns DisabledLogger.
	assert.Equal(t, DisabledLogger(), FromContext(context.Background()))

	logger := DisabledLogger()
	ctx := NewContext(context.Background(), logger)
	// First try.
	assert.Equal(t, logger, FromContext(ctx))
	// Second try.
	assert.Equal(t, logger, FromContext(ctx))
}
