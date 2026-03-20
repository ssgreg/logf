package logf

import (
	"bytes"
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testEncoder is a minimal Encoder for router tests.
type testEncoder struct {
	prefix string
	err    error
}

func (e *testEncoder) Encode(entry Entry) (*Buffer, error) {
	if e.err != nil {
		return nil, e.err
	}
	buf := GetBuffer()
	buf.AppendString(e.prefix)
	buf.AppendString(entry.Text)
	return buf, nil
}

func (e *testEncoder) Clone() Encoder {
	return &testEncoder{prefix: e.prefix, err: e.err}
}

// spyWriter is an io.Writer that records all writes, flush, and sync calls.
type spyWriter struct {
	mu     sync.Mutex
	writes []string
	flushN int
	syncN  int
}

func (w *spyWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.writes = append(w.writes, string(p))
	return len(p), nil
}

func (w *spyWriter) Flush() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.flushN++
	return nil
}

func (w *spyWriter) Sync() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.syncN++
	return nil
}



func (w *spyWriter) allData() string {
	w.mu.Lock()
	defer w.mu.Unlock()
	var s string
	for _, wr := range w.writes {
		s += wr
	}
	return s
}

func (w *spyWriter) flushCount() int {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.flushN
}

func (w *spyWriter) syncCount() int {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.syncN
}

// --- Build tests ---

func TestRouterBuildNoOutputs(t *testing.T) {
	_, _, err := NewRouter().Build()
	require.Error(t, err)
}

func TestRouterBuildEmptyRoute(t *testing.T) {
	_, _, err := NewRouter().
		Route(&testEncoder{}).
		Build()
	require.Error(t, err)
}

func TestRouterBuildValid(t *testing.T) {
	h, closeFn, err := NewRouter().
		Route(&testEncoder{}, Output(LevelDebug, &bytes.Buffer{})).
		Build()
	require.NoError(t, err)
	require.NotNil(t, h)
	require.NotNil(t, closeFn)
	_ = closeFn()
}

// --- Enabled tests ---

func TestRouterEnabled(t *testing.T) {
	h, closeFn, err := NewRouter().
		Route(&testEncoder{},
			Output(LevelInfo, &bytes.Buffer{}),
			Output(LevelError, &bytes.Buffer{}),
		).
		Build()
	require.NoError(t, err)
	defer func() { _ = closeFn() }()

	ctx := context.Background()
	// LevelInfo output is the widest (Info accepts Error, Warn, Info).
	// minLevel = LevelInfo(2) — highest numeric = widest.
	assert.True(t, h.Enabled(ctx, LevelInfo))
	assert.True(t, h.Enabled(ctx, LevelError))
	// Debug(3) > Info(2), so Info.Enabled(Debug) = 2 >= 3 = false.
	assert.False(t, h.Enabled(ctx, LevelDebug))
}

func TestRouterEnabledMultipleRoutes(t *testing.T) {
	h, closeFn, err := NewRouter().
		Route(&testEncoder{}, Output(LevelError, &bytes.Buffer{})).
		Route(&testEncoder{}, Output(LevelDebug, &bytes.Buffer{})).
		Build()
	require.NoError(t, err)
	defer func() { _ = closeFn() }()

	ctx := context.Background()
	// Debug(3) is the widest output → minLevel=Debug(3).
	assert.True(t, h.Enabled(ctx, LevelDebug))
	assert.True(t, h.Enabled(ctx, LevelError))
}

// --- Handle / fan-out tests ---

func TestRouterSingleOutput(t *testing.T) {
	spy := &spyWriter{}
	h, closeFn, err := NewRouter().
		Route(&testEncoder{prefix: "J:"}, Output(LevelDebug, spy)).
		Build()
	require.NoError(t, err)

	require.NoError(t, h.Handle(context.Background(), Entry{Text: "hello", Level: LevelInfo}))
	_ = closeFn()

	assert.Equal(t, "J:hello", spy.allData())
}

func TestRouterFanOutSameEncoder(t *testing.T) {
	s1 := &spyWriter{}
	s2 := &spyWriter{}
	h, closeFn, err := NewRouter().
		Route(&testEncoder{prefix: "J:"},
			Output(LevelDebug, s1),
			Output(LevelDebug, s2),
		).
		Build()
	require.NoError(t, err)

	require.NoError(t, h.Handle(context.Background(), Entry{Text: "msg", Level: LevelInfo}))
	_ = closeFn()

	assert.Equal(t, "J:msg", s1.allData())
	assert.Equal(t, "J:msg", s2.allData())
}

func TestRouterMultipleEncoders(t *testing.T) {
	sJSON := &spyWriter{}
	sText := &spyWriter{}
	h, closeFn, err := NewRouter().
		Route(&testEncoder{prefix: "J:"}, Output(LevelDebug, sJSON)).
		Route(&testEncoder{prefix: "T:"}, Output(LevelDebug, sText)).
		Build()
	require.NoError(t, err)

	require.NoError(t, h.Handle(context.Background(), Entry{Text: "x", Level: LevelInfo}))
	_ = closeFn()

	assert.Equal(t, "J:x", sJSON.allData())
	assert.Equal(t, "T:x", sText.allData())
}

// --- Level filtering ---

func TestRouterLevelFiltering(t *testing.T) {
	sAll := &spyWriter{}
	sErr := &spyWriter{}
	h, closeFn, err := NewRouter().
		Route(&testEncoder{prefix: ""},
			Output(LevelDebug, sAll),
			Output(LevelError, sErr),
		).
		Build()
	require.NoError(t, err)

	ctx := context.Background()
	// Info(2): Debug(3).Enabled(Info(2)) = 3>=2 = true, Error(0).Enabled(Info(2)) = 0>=2 = false.
	require.NoError(t, h.Handle(ctx, Entry{Text: "info", Level: LevelInfo}))
	// Error(0): Debug(3).Enabled(Error(0)) = true, Error(0).Enabled(Error(0)) = true.
	require.NoError(t, h.Handle(ctx, Entry{Text: "err", Level: LevelError}))
	_ = closeFn()

	assert.Equal(t, "infoerr", sAll.allData())
	assert.Equal(t, "err", sErr.allData())
}

func TestRouterGroupSkippedByLevel(t *testing.T) {
	spy := &spyWriter{}
	h, closeFn, err := NewRouter().
		Route(&testEncoder{}, Output(LevelError, spy)).
		Build()
	require.NoError(t, err)

	require.NoError(t, h.Handle(context.Background(), Entry{Text: "debug", Level: LevelDebug}))
	_ = closeFn()

	assert.Empty(t, spy.allData())
}

// --- Encode error ---

func TestRouterEncodeError(t *testing.T) {
	encErr := errors.New("encode fail")
	h, closeFn, err := NewRouter().
		Route(&testEncoder{err: encErr}, Output(LevelDebug, &bytes.Buffer{})).
		Build()
	require.NoError(t, err)
	defer func() { _ = closeFn() }()

	err = h.Handle(context.Background(), Entry{Text: "x", Level: LevelInfo})
	assert.ErrorIs(t, err, encErr)
}

// --- Close behavior ---

func TestRouterCloseFlushesAndSyncs(t *testing.T) {
	spy := &spyWriter{}
	h, closeFn, err := NewRouter().
		Route(&testEncoder{prefix: ""}, Output(LevelDebug, spy)).
		Build()
	require.NoError(t, err)

	require.NoError(t, h.Handle(context.Background(), Entry{Text: "x", Level: LevelInfo}))
	_ = closeFn()

	assert.Equal(t, 1, spy.flushCount())
	assert.Equal(t, 1, spy.syncCount())
}

func TestRouterDoubleClose(t *testing.T) {
	_, closeFn, err := NewRouter().
		Route(&testEncoder{}, Output(LevelDebug, &bytes.Buffer{})).
		Build()
	require.NoError(t, err)

	_ = closeFn()
	_ = closeFn() // should not panic
}

// --- All messages delivered ---

func TestRouterAllDelivered(t *testing.T) {
	spy := &spyWriter{}
	h, closeFn, err := NewRouter().
		Route(&testEncoder{prefix: ""},
			Output(LevelDebug, spy),
		).
		Build()
	require.NoError(t, err)

	ctx := context.Background()
	n := 100
	for i := 0; i < n; i++ {
		require.NoError(t, h.Handle(ctx, Entry{Text: "m", Level: LevelInfo}))
	}
	_ = closeFn()

	data := spy.allData()
	assert.Len(t, data, n) // n * 1 byte per "m"
}

// --- Ordering ---

// spyWriteCloser is a Writer + io.Closer for testing OutputCloser.
type spyWriteCloser struct {
	spyWriter
	closeCalled bool
}

func (w *spyWriteCloser) Close() error {
	w.closeCalled = true
	return nil
}

func TestRouterOutputCloser(t *testing.T) {
	spy := &spyWriteCloser{}
	h, closeFn, err := NewRouter().
		Route(&testEncoder{prefix: "J:"}, OutputCloser(LevelDebug, spy)).
		Build()
	require.NoError(t, err)

	require.NoError(t, h.Handle(context.Background(), Entry{Text: "hello", Level: LevelInfo}))
	err = closeFn()
	require.NoError(t, err)

	assert.Equal(t, "J:hello", spy.allData())
	assert.True(t, spy.closeCalled, "Close should be called on the writer")
	assert.Equal(t, 1, spy.flushCount())
	assert.Equal(t, 1, spy.syncCount())
}

func TestRouterOrdering(t *testing.T) {
	spy := &spyWriter{}
	h, closeFn, err := NewRouter().
		Route(&testEncoder{prefix: ""},
			Output(LevelDebug, spy),
		).
		Build()
	require.NoError(t, err)

	ctx := context.Background()
	var expected string
	for i := 0; i < 10; i++ {
		s := string(rune('a' + i))
		expected += s
		require.NoError(t, h.Handle(ctx, Entry{
			Text:  s,
			Level: LevelInfo,
		}))
	}
	_ = closeFn()

	assert.Equal(t, expected, spy.allData())
}
