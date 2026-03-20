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
	w := &testHandler{}
	logger := New(w)

	assert.Equal(t, 0, logger.callerSkip)
	assert.Equal(t, true, logger.addCaller)
	assert.Nil(t, logger.bag)
	assert.Empty(t, logger.name)
	assert.Equal(t, w, logger.w)
}

func TestLoggerCallerSpecifiedByDefault(t *testing.T) {
	w := &testHandler{}
	logger := New(w)

	logger.Error(ctx, "")
	assert.NotZero(t, w.Entry.CallerPC)
	file, _ := callerFrame(w.Entry.CallerPC)
	assert.Equal(t, "logf/logger_test.go", fileWithPackage(file))
}

func TestLoggerCallerDisabled(t *testing.T) {
	w := &testHandler{}
	logger := New(w).WithCaller(false)

	logger.Error(ctx, "")
	assert.Zero(t, w.Entry.CallerPC)
}

func TestLoggerCallerSpecifiedWithSkip(t *testing.T) {
	w := &testHandler{}
	logger := New(w).WithCaller(true).WithCallerSkip(1)

	logger.Error(ctx, "")
	assert.NotZero(t, w.Entry.CallerPC)
	file, _ := callerFrame(w.Entry.CallerPC)
	assert.Equal(t, "testing/testing.go", fileWithPackage(file))
}

func TestLoggerNoNameByDefault(t *testing.T) {
	w := &testHandler{}
	logger := New(w)

	logger.Error(ctx, "")
	assert.Empty(t, w.Entry.LoggerName)
}

func TestLoggerEmptyName(t *testing.T) {
	w := &testHandler{}
	logger := New(w).WithName("")

	logger.Error(ctx, "")
	assert.Empty(t, w.Entry.LoggerName)
}

func TestLoggerName(t *testing.T) {
	w := &testHandler{}

	// Set a name for the logger.
	logger := New(w).WithName("1")
	logger.Error(ctx, "")
	assert.Equal(t, "1", w.Entry.LoggerName)

	// Set a nested name for the logger.
	logger = logger.WithName("2")
	logger.Error(ctx, "")
	assert.Equal(t, "1.2", w.Entry.LoggerName)
}

func TestLoggerLevelFiltering(t *testing.T) {
	w := newLeveledTestHandler(LevelError)
	logger := New(w)

	logger.Info(ctx, "filtered")
	assert.Empty(t, w.Entries)

	logger.Error(ctx, "passed")
	assert.Len(t, w.Entries, 1)
	assert.Equal(t, "passed", w.Entries[0].Text)
}

func TestLoggerAtLevel(t *testing.T) {
	w := newLeveledTestHandler(LevelError)
	logger := New(w)

	// Expected the callback should be called with the same severity level.
	called := false
	logger.AtLevel(ctx, LevelError, func(log LogFunc) {
		called = true
		assert.NotEmpty(t, log)
		log(ctx, "at-level")
	})
	assert.Len(t, w.Entries, 1)
	assert.Equal(t, "at-level", w.Entries[0].Text)
	assert.True(t, called)

	// Expected the callback should NOT be called with DEBUG severity level.
	called = false
	logger.AtLevel(ctx, LevelDebug, func(log LogFunc) {
		called = true
	})
	assert.False(t, called)
}

func TestLoggerNoFieldsByDefault(t *testing.T) {
	w := &testHandler{}

	logger := New(w)
	logger.Error(ctx, "")
	assert.Nil(t, w.Entry.LoggerBag)
}

func TestLoggerFields(t *testing.T) {
	w := &testHandler{}

	// Add one Field.
	logger := New(w).With(String("first", "v"), String("second", "v"))
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
	w := &testHandler{}
	logger := New(w)

	logger.Error(ctx, "")
	assert.Nil(t, w.Entry.LoggerBag)

	logger = logger.With(String("k", "v"))
	logger.Error(ctx, "")
	assert.NotNil(t, w.Entry.LoggerBag)
	assert.Equal(t, logger.bag, w.Entry.LoggerBag)
}

func TestLoggerLeveledLog(t *testing.T) {
	w := &testHandler{}
	logger := New(w)

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

func TestUnbufferedHandler(t *testing.T) {
	w := newLeveledTestHandler(LevelDebug)
	logger := New(w)

	// Check function not panic.
	logger.Debug(ctx, "debug")
	logger.Info(ctx, "info")
	logger.Warn(ctx, "warn")
	logger.Error(ctx, "error")
	logger.AtLevel(ctx, LevelError, func(log LogFunc) {
		log(ctx, "at-level-error")
	})

	require.Equal(t, 5, len(w.Entries))

	assert.Equal(t, "debug", w.Entries[0].Text)
	assert.Equal(t, "info", w.Entries[1].Text)
	assert.Equal(t, "warn", w.Entries[2].Text)
	assert.Equal(t, "error", w.Entries[3].Text)
	assert.Equal(t, "at-level-error", w.Entries[4].Text)

	assert.Equal(t, LevelDebug, w.Entries[0].Level)
	assert.Equal(t, LevelInfo, w.Entries[1].Level)
	assert.Equal(t, LevelWarn, w.Entries[2].Level)
	assert.Equal(t, LevelError, w.Entries[3].Level)
	assert.Equal(t, LevelError, w.Entries[4].Level)
}

func TestLoggerWithGroup(t *testing.T) {
	w := &testHandler{}
	logger := New(w).WithGroup("http")

	logger.Error(ctx, "done", Int("status", 200))
	require.NotNil(t, w.Entry.LoggerBag)
	assert.Equal(t, "http", w.Entry.LoggerBag.Group())
}

func TestLoggerWithGroupEmpty(t *testing.T) {
	w := &testHandler{}
	logger := New(w)
	same := logger.WithGroup("")

	// Empty group name returns the same logger.
	assert.Equal(t, logger, same)
}

func TestLoggerWithGroupChain(t *testing.T) {
	w := &testHandler{}
	logger := New(w).
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
	sink := &testHandler{}
	logger := New(sink).WithName("fb").WithName("agent")
	slogger := logger.Slog()

	slogger.Info("request")

	require.NotNil(t, sink.Entry)
	assert.Equal(t, "request", sink.Entry.Text)
	assert.Equal(t, LevelInfo, sink.Entry.Level)
	assert.Equal(t, "fb.agent", sink.Entry.LoggerName)
}

func TestLoggerSlogWithFields(t *testing.T) {
	sink := &testHandler{}
	logger := New(sink).With(String("env", "prod"))
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
	sink := &testHandler{}
	logger := New(sink)
	slogger := logger.Slog()

	slogger.Info("hello")

	require.NotNil(t, sink.Entry)
	assert.Equal(t, "", sink.Entry.LoggerName)
}

func TestContextlessMethods(t *testing.T) {
	w := &testHandler{}
	logger := New(w).WithCaller(false)

	logger.Debugx("debug", String("k", "v"))
	logger.Infox("info")
	logger.Warnx("warn")
	logger.Errorx("error")

	require.Equal(t, 4, len(w.Entries))
	assert.Equal(t, "debug", w.Entries[0].Text)
	assert.Equal(t, LevelDebug, w.Entries[0].Level)
	assert.Equal(t, "info", w.Entries[1].Text)
	assert.Equal(t, LevelInfo, w.Entries[1].Level)
	assert.Equal(t, "warn", w.Entries[2].Text)
	assert.Equal(t, LevelWarn, w.Entries[2].Level)
	assert.Equal(t, "error", w.Entries[3].Text)
	assert.Equal(t, LevelError, w.Entries[3].Level)

	// Fields are passed through.
	require.Equal(t, 1, len(w.Entries[0].Fields))
	assert.Equal(t, "k", w.Entries[0].Fields[0].Key)

	// Disabled level: nothing logged.
	w2 := newLeveledTestHandler(LevelError)
	logger2 := New(w2).WithCaller(false)

	logger2.Debugx("nope")
	logger2.Infox("nope")
	logger2.Warnx("nope")
	assert.Equal(t, 0, len(w2.Entries))

	logger2.Errorx("yes")
	require.Equal(t, 1, len(w2.Entries))
	assert.Equal(t, "yes", w2.Entries[0].Text)
}

func TestNilContext(t *testing.T) {
	w := &testHandler{}
	logger := New(w).WithCaller(false)

	// Use a typed nil to avoid SA1012 lint warnings.
	// We intentionally test nil ctx support here.
	var noCtx context.Context

	logger.Debug(noCtx, "debug", String("k", "v"))
	logger.Info(noCtx, "info")
	logger.Warn(noCtx, "warn")
	logger.Error(noCtx, "error")
	logger.Log(noCtx, LevelInfo, "log")
	logger.AtLevel(noCtx, LevelInfo, func(log LogFunc) {
		log(noCtx, "at-level")
	})

	require.Equal(t, 6, len(w.Entries))
	assert.Equal(t, "debug", w.Entries[0].Text)
	assert.Equal(t, "at-level", w.Entries[5].Text)

	// Enabled must not panic with nil context.
	assert.True(t, logger.Enabled(noCtx, LevelInfo))
}

func TestNopHandlerHandle(t *testing.T) {
	h := nopHandler{}
	err := h.Handle(context.Background(), Entry{Text: "should be discarded", Level: LevelError})
	assert.NoError(t, err)
	assert.False(t, h.Enabled(context.Background(), LevelError))
}

func TestLogDepth(t *testing.T) {
	w := &testHandler{}
	logger := New(w)

	LogDepth(logger, ctx, 0, LevelError, "depth-test", String("k", "v"))

	require.NotNil(t, w.Entry)
	assert.Equal(t, "depth-test", w.Entry.Text)
	assert.Equal(t, LevelError, w.Entry.Level)
	require.Equal(t, 1, len(w.Entry.Fields))
	assert.Equal(t, "k", w.Entry.Fields[0].Key)

	// CallerPC should point to this test function.
	assert.NotZero(t, w.Entry.CallerPC)
	file, _ := callerFrame(w.Entry.CallerPC)
	assert.Equal(t, "logf/logger_test.go", fileWithPackage(file))
}

func TestLogDepthFilteredByLevel(t *testing.T) {
	w := newLeveledTestHandler(LevelError)
	logger := New(w)

	LogDepth(logger, ctx, 0, LevelDebug, "should-not-appear")
	assert.Empty(t, w.Entries)
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
