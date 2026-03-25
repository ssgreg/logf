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
	logger := logf.DisabledLogger()

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

func TestWithGroup(t *testing.T) {
	w := &mockHandler{}
	logger := logf.New(w)
	ctx := New(context.Background(), logger)

	ctx = WithGroup(ctx, "http")
	Info(ctx, "req", logf.Int("status", 200))

	assert.Equal(t, 1, len(w.entries))
	bag := w.entries[0].LoggerBag
	assert.NotNil(t, bag)
	assert.Equal(t, "http", bag.Group())
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
	logger := logf.New(&mockHandler{})
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
			w := &mockHandler{}
			logger := logf.New(w)
			ctx := New(context.Background(), logger)
			test.Log(ctx, text)
			assert.Equal(t, 1, len(w.entries))
			assert.Equal(t, text, w.entries[0].Text)
			assert.Equal(t, test.Level, w.entries[0].Level)
		})
	}
}

func TestLog(t *testing.T) {
	w := &mockHandler{}
	logger := logf.New(w)
	ctx := New(context.Background(), logger)

	Log(ctx, logf.LevelWarn, "custom level", logf.String("k", "v"))
	assert.Equal(t, 1, len(w.entries))
	assert.Equal(t, "custom level", w.entries[0].Text)
	assert.Equal(t, logf.LevelWarn, w.entries[0].Level)
	assert.Equal(t, 1, len(w.entries[0].Fields))
}

func TestAtLevelCallerPointsToCallSite(t *testing.T) {
	w := &mockHandler{}
	logger := logf.New(w)
	ctx := New(context.Background(), logger)

	_, _, expectedLine, _ := runtime.Caller(0)
	AtLevel(ctx, logf.LevelInfo, func(log logf.LogFunc) {
		log(ctx, "at-level-test") // expectedLine + 2
	})

	assert.Equal(t, 1, len(w.entries))
	pc := w.entries[0].CallerPC
	assert.NotZero(t, pc, "CallerPC should be set")

	frames := runtime.CallersFrames([]uintptr{pc})
	f, _ := frames.Next()

	assert.Equal(t, expectedLine+2, f.Line,
		"caller should point to log() call site, got %s:%d", f.File, f.Line)
	assert.True(t, strings.HasSuffix(f.File, "context_test.go"),
		"expected context_test.go, got %s", f.File)
}

func TestCallerPointsToCallSite(t *testing.T) {
	w := &mockHandler{}
	logger := logf.New(w)
	ctx := New(context.Background(), logger)

	_, expectedFile, expectedLine, _ := runtime.Caller(0)
	Debug(ctx, "caller-test") // expectedLine + 1

	assert.Equal(t, 1, len(w.entries))
	pc := w.entries[0].CallerPC
	assert.NotZero(t, pc, "CallerPC should be set")

	frames := runtime.CallersFrames([]uintptr{pc})
	f, _ := frames.Next()

	assert.Equal(t, expectedLine+1, f.Line)
	assert.True(t, strings.HasSuffix(f.File, expectedFile[strings.LastIndex(expectedFile, "/")+1:]),
		"expected file %s, got %s", expectedFile, f.File)
}

type mockHandler struct {
	entries []logf.Entry
}

func (w *mockHandler) Handle(_ context.Context, e logf.Entry) error {
	w.entries = append(w.entries, e)
	return nil
}

func (w *mockHandler) Enabled(_ context.Context, _ logf.Level) bool {
	return true
}
