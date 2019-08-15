package logf

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoggerNew(t *testing.T) {
	w := &testEntryWriter{}
	logger := NewLogger(LevelError, w)

	assert.Equal(t, 0, logger.callerSkip)
	assert.Equal(t, false, logger.addCaller)
	assert.True(t, logger.id > 0)
	assert.Empty(t, logger.fields)
	assert.Empty(t, logger.name)
	assert.Equal(t, w, logger.w)
}

func TestLoggerCallerNotSpecifiedByDefault(t *testing.T) {
	w := &testEntryWriter{}
	logger := NewLogger(LevelError, w)

	logger.Error("")
	assert.False(t, w.Entry.Caller.Specified)
}

func TestLoggerCallerSpecified(t *testing.T) {
	w := &testEntryWriter{}
	logger := NewLogger(LevelError, w).WithCaller()

	logger.Error("")
	assert.True(t, w.Entry.Caller.Specified)
	assert.Equal(t, "logf/logger_test.go", w.Entry.Caller.FileWithPackage())
}

func TestLoggerCallerSpecifiedWithSkip(t *testing.T) {
	w := &testEntryWriter{}
	logger := NewLogger(LevelError, w).WithCaller().WithCallerSkip(1)

	logger.Error("")
	assert.True(t, w.Entry.Caller.Specified)
	assert.Equal(t, "testing/testing.go", w.Entry.Caller.FileWithPackage())
}

func TestLoggerNoNameByDefault(t *testing.T) {
	w := &testEntryWriter{}
	logger := NewLogger(LevelError, w)

	logger.Error("")
	assert.Empty(t, w.Entry.LoggerName)
}

func TestLoggerEmptyName(t *testing.T) {
	w := &testEntryWriter{}
	logger := NewLogger(LevelError, w).WithName("")

	logger.Error("")
	assert.Empty(t, w.Entry.LoggerName)
}

func TestLoggerName(t *testing.T) {
	w := &testEntryWriter{}

	// Set a name for the logger.
	logger := NewLogger(LevelError, w).WithName("1")
	logger.Error("")
	assert.Equal(t, "1", w.Entry.LoggerName)

	// Set a nested name for the logger.
	logger = logger.WithName("2")
	logger.Error("")
	assert.Equal(t, "1.2", w.Entry.LoggerName)
}

func TestLoggerAtLevel(t *testing.T) {
	w := &testEntryWriter{}
	logger := NewLogger(LevelError, w)

	// Expected the callback should be called with the same severity level.
	called := false
	logger.AtLevel(LevelError, func(log LogFunc) {
		called = true
		assert.NotEmpty(t, log)
		log("at-level")
	})
	assert.Equal(t, "at-level", w.Entry.Text)
	assert.True(t, called)

	// Expected the callback should NOT be called with DEBUG severity level.
	called = false
	logger.AtLevel(LevelDebug, func(log LogFunc) {
		called = true
	})
	assert.False(t, called)
}

func TestLoggerNoFieldsByDefault(t *testing.T) {
	w := &testEntryWriter{}

	logger := NewLogger(LevelError, w)
	logger.Error("")
	assert.Empty(t, w.Entry.DerivedFields)
}

func TestLoggerFields(t *testing.T) {
	w := &testEntryWriter{}

	// Add one Field.
	logger := NewLogger(LevelError, w).With(String("first", "v"), String("second", "v"))
	logger.Error("")
	assert.Equal(t, 2, len(w.Entry.DerivedFields))

	// Add second Field.
	logger = logger.With(String("third", "v"))
	logger.Error("")
	assert.Equal(t, 3, len(w.Entry.DerivedFields))

	// Check order. First field should go first.
	assert.Equal(t, "first", w.Entry.DerivedFields[0].Key)
	assert.Equal(t, "second", w.Entry.DerivedFields[1].Key)
	assert.Equal(t, "third", w.Entry.DerivedFields[2].Key)
}

func TestLoggerWithFieldsCallSnapshotter(t *testing.T) {
	w := &testEntryWriter{}
	ts := testSnapshotter{}

	NewLogger(LevelError, w).With(Any("snapshotter", &ts))
	assert.True(t, ts.Called)
}

func TestLoggerID(t *testing.T) {
	w := &testEntryWriter{}
	logger := NewLogger(LevelError, w)

	logger.Error("")
	assert.Equal(t, logger.id, w.Entry.LoggerID)
}

func TestLoggerLeveledLog(t *testing.T) {
	w := &testEntryWriter{}
	logger := NewLogger(LevelDebug, w)

	cases := []struct {
		Level Level
		Fn    func(string, ...Field)
	}{
		{LevelError, logger.Error},
		{LevelWarn, logger.Warn},
		{LevelInfo, logger.Info},
		{LevelDebug, logger.Debug},
	}

	for _, c := range cases {
		t.Run(c.Level.String(), func(t *testing.T) {
			ts := testSnapshotter{}
			c.Fn(c.Level.String(), Any("snapshotter", &ts))

			assert.Equal(t, c.Level.String(), w.Entry.Text)
			assert.Equal(t, 1, len(w.Entry.Fields))
			assert.Equal(t, c.Level, w.Entry.Level)
			assert.WithinDuration(t, time.Now(), w.Entry.Time, time.Second*2)
			assert.True(t, ts.Called)
		})
	}
}

func TestLoggerChecker(t *testing.T) {
	w := &testEntryWriter{}
	logger := NewLogger(testLevelCheckerReturningFalse{}, w)

	logger.Error("")
	assert.Nil(t, w.Entry)

	logger.Warn("")
	assert.Nil(t, w.Entry)

	logger.Info("")
	assert.Nil(t, w.Entry)

	logger.Debug("")
	assert.Nil(t, w.Entry)
}

func TestLoggerDisabled(t *testing.T) {
	logger := DisabledLogger()
	assert.Equal(t, 0, logger.callerSkip)
	assert.Equal(t, false, logger.addCaller)
	assert.True(t, logger.id > 0)
	assert.Empty(t, logger.fields)
	assert.Empty(t, logger.name)
	assert.Empty(t, logger.w)

	loggerWithCallerSkip := logger.WithCallerSkip(1)
	assert.Equal(t, 1, loggerWithCallerSkip.callerSkip)

	loggerWithCaller := logger.WithCaller()
	assert.Equal(t, true, loggerWithCaller.addCaller)

	loggerWithName := logger.WithName("name")
	assert.Equal(t, "name", loggerWithName.name)

	loggerWithFields := logger.With(String("k", "v"))
	assert.NotEmpty(t, loggerWithFields.fields)

	// Check function not panic.
	logger.Debug("")
	logger.Info("")
	logger.Warn("")
	logger.Error("")
	logger.AtLevel(LevelError, func(LogFunc) { require.FailNow(t, "can't be here") })
}

func TestUnbufferedWriter(t *testing.T) {
	a := &testAppender{}
	w := NewUnbufferedEntryWriter(a)
	logger := NewLogger(LevelDebug, w)

	// Check function not panic.
	logger.Debug("debug")
	logger.Info("info")
	logger.Warn("warn")
	logger.Error("error")
	logger.AtLevel(LevelError, func(log LogFunc) {
		log("at-level-error")
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
