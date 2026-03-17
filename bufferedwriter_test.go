package logf

import (
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeClock returns a ticker whose channel is controlled manually.
type fakeClock struct {
	ch chan time.Time
}

func newFakeClock() *fakeClock {
	return &fakeClock{ch: make(chan time.Time, 1)}
}

func (c *fakeClock) NewTicker(time.Duration) *time.Ticker {
	t := time.NewTicker(time.Hour) // real ticker, never fires
	t.Stop()
	t.C = c.ch // replace channel
	return t
}

func (c *fakeClock) tick() {
	c.ch <- time.Now()
}

// recordWriter records write/flush/sync calls.
type recordWriter struct {
	mu       sync.Mutex
	data     []byte
	flushN   int
	syncN    int
	writeN   int
	writeErr error
}

func (w *recordWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.writeErr != nil {
		return 0, w.writeErr
	}
	w.data = append(w.data, p...)
	w.writeN++
	return len(p), nil
}

func (w *recordWriter) Flush() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.flushN++
	return nil
}

func (w *recordWriter) Sync() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.syncN++
	return nil
}

func (w *recordWriter) snapshot() (data []byte, writes, flushes, syncs int) {
	w.mu.Lock()
	defer w.mu.Unlock()
	cp := make([]byte, len(w.data))
	copy(cp, w.data)
	return cp, w.writeN, w.flushN, w.syncN
}

// --- Basic buffering ---

func TestBufferedWriterBuffersUntilThreshold(t *testing.T) {
	rec := &recordWriter{}
	bw, closeFn := NewBufferedWriter(rec,
		WithBufSize(10),
		WithSyncInterval(0), // no timer
	)

	n, err := bw.Write([]byte("hello")) // 5 bytes, below threshold
	require.NoError(t, err)
	assert.Equal(t, 5, n)

	_, writes, _, _ := rec.snapshot()
	assert.Equal(t, 0, writes) // not flushed yet

	// "world!" (6 bytes): 5+6=11 > 10 → pre-flush "hello", then "world!" stays in buffer (6 < 10).
	n, err = bw.Write([]byte("world!"))
	require.NoError(t, err)
	assert.Equal(t, 6, n)

	data, writes, _, _ := rec.snapshot()
	assert.Equal(t, 1, writes) // pre-flush of "hello"
	assert.Equal(t, "hello", string(data))

	// Flush remaining.
	require.NoError(t, bw.Flush())
	data, writes, _, _ = rec.snapshot()
	assert.Equal(t, 2, writes)
	assert.Equal(t, "helloworld!", string(data))

	_ = closeFn()
}

func TestBufferedWriterFlush(t *testing.T) {
	rec := &recordWriter{}
	bw, closeFn := NewBufferedWriter(rec,
		WithBufSize(1024),
		WithSyncInterval(0),
	)
	defer func() { _ = closeFn() }()

	_, _ = bw.Write([]byte("data"))

	_, writes, _, _ := rec.snapshot()
	assert.Equal(t, 0, writes)

	require.NoError(t, bw.Flush())

	data, writes, _, _ := rec.snapshot()
	assert.Equal(t, 1, writes)
	assert.Equal(t, "data", string(data))
}

func TestBufferedWriterFlushEmpty(t *testing.T) {
	rec := &recordWriter{}
	bw, closeFn := NewBufferedWriter(rec,
		WithBufSize(1024),
		WithSyncInterval(0),
	)
	defer func() { _ = closeFn() }()

	require.NoError(t, bw.Flush()) // no-op

	_, writes, _, _ := rec.snapshot()
	assert.Equal(t, 0, writes)
}

func TestBufferedWriterSync(t *testing.T) {
	rec := &recordWriter{}
	bw, closeFn := NewBufferedWriter(rec,
		WithBufSize(1024),
		WithSyncInterval(0),
	)
	defer func() { _ = closeFn() }()

	_, _ = bw.Write([]byte("abc"))
	require.NoError(t, bw.Sync())

	data, writes, _, syncs := rec.snapshot()
	assert.Equal(t, 1, writes)
	assert.Equal(t, 1, syncs)
	assert.Equal(t, "abc", string(data))
}

// --- Partial write / overflow protection ---

func TestBufferedWriterLargeWriteFlushesFirst(t *testing.T) {
	rec := &recordWriter{}
	bw, closeFn := NewBufferedWriter(rec,
		WithBufSize(10),
		WithSyncInterval(0),
	)
	defer func() { _ = closeFn() }()

	// Write 5 bytes (below threshold).
	_, _ = bw.Write([]byte("aaaaa"))
	// Write 8 bytes — doesn't fit (5+8=13 > 10), should flush existing first.
	_, _ = bw.Write([]byte("bbbbbbbb"))

	_, writes, _, _ := rec.snapshot()
	// Two flushes: first for "aaaaa", second for "bbbbbbbb" (8 >= 10? no, 8 < 10).
	// Actually: flush "aaaaa" (existing), then append "bbbbbbbb" (8 < 10, stays in buffer).
	assert.Equal(t, 1, writes) // only the pre-flush
	_ = bw.Flush()

	data, writes, _, _ := rec.snapshot()
	assert.Equal(t, 2, writes)
	assert.Equal(t, "aaaaabbbbbbbb", string(data))
}

func TestBufferedWriterSingleLargeWrite(t *testing.T) {
	rec := &recordWriter{}
	bw, closeFn := NewBufferedWriter(rec,
		WithBufSize(4),
		WithSyncInterval(0),
	)
	defer func() { _ = closeFn() }()

	// Single write larger than buffer — should write through immediately.
	_, _ = bw.Write([]byte("huge-payload"))

	data, writes, _, _ := rec.snapshot()
	assert.Equal(t, 1, writes)
	assert.Equal(t, "huge-payload", string(data))
}

// --- Timer-based sync ---

func TestBufferedWriterTimerSync(t *testing.T) {
	clk := newFakeClock()
	rec := &recordWriter{}
	bw, closeFn := NewBufferedWriter(rec,
		WithBufSize(1024),
		WithSyncInterval(time.Second),
		WithClock(clk),
	)

	_, _ = bw.Write([]byte("tick"))

	_, writes, _, syncs := rec.snapshot()
	assert.Equal(t, 0, writes)
	assert.Equal(t, 0, syncs)

	// Simulate tick — should flush + sync.
	clk.tick()
	// Give the goroutine a moment to process.
	time.Sleep(50 * time.Millisecond)

	data, writes, _, syncs := rec.snapshot()
	assert.Equal(t, 1, writes)
	assert.Equal(t, 1, syncs)
	assert.Equal(t, "tick", string(data))

	_ = closeFn()
}

func TestBufferedWriterTimerMultipleTicks(t *testing.T) {
	clk := newFakeClock()
	rec := &recordWriter{}
	bw, closeFn := NewBufferedWriter(rec,
		WithBufSize(1024),
		WithSyncInterval(time.Second),
		WithClock(clk),
	)

	_, _ = bw.Write([]byte("a"))
	clk.tick()
	time.Sleep(50 * time.Millisecond)

	_, _ = bw.Write([]byte("b"))
	clk.tick()
	time.Sleep(50 * time.Millisecond)

	_ = closeFn()

	data, writes, _, syncs := rec.snapshot()
	assert.Equal(t, "ab", string(data))
	assert.GreaterOrEqual(t, writes, 2)
	assert.GreaterOrEqual(t, syncs, 2) // 2 ticks + close sync
}

// --- Close ---

func TestBufferedWriterCloseSyncs(t *testing.T) {
	rec := &recordWriter{}
	bw, closeFn := NewBufferedWriter(rec,
		WithBufSize(1024),
		WithSyncInterval(0),
	)

	_, _ = bw.Write([]byte("pending"))
	_ = closeFn()

	data, writes, _, syncs := rec.snapshot()
	assert.Equal(t, "pending", string(data))
	assert.Equal(t, 1, writes)
	assert.Equal(t, 1, syncs)
}

func TestBufferedWriterDoubleClose(t *testing.T) {
	_, closeFn := NewBufferedWriter(&recordWriter{},
		WithSyncInterval(0),
	)
	_ = closeFn()
	_ = closeFn() // should not panic
}

func TestBufferedWriterCloseWaitsForTimer(t *testing.T) {
	clk := newFakeClock()
	rec := &recordWriter{}
	_, closeFn := NewBufferedWriter(rec,
		WithBufSize(1024),
		WithSyncInterval(time.Second),
		WithClock(clk),
	)

	// Close should not deadlock even with fake clock.
	_ = closeFn()
}

// --- Write error propagation ---

func TestBufferedWriterWriteError(t *testing.T) {
	writeErr := errors.New("disk full")
	rec := &recordWriter{writeErr: writeErr}
	bw, closeFn := NewBufferedWriter(rec,
		WithBufSize(4),
		WithSyncInterval(0),
	)
	defer func() { _ = closeFn() }()

	// First write fits in buffer — no error yet.
	_, err := bw.Write([]byte("aaa"))
	require.NoError(t, err)

	// Second write triggers flush — error propagated.
	_, err = bw.Write([]byte("bbb"))
	assert.ErrorIs(t, err, writeErr)
}

func TestBufferedWriterFlushError(t *testing.T) {
	writeErr := errors.New("io error")
	rec := &recordWriter{writeErr: writeErr}
	bw, closeFn := NewBufferedWriter(rec,
		WithBufSize(1024),
		WithSyncInterval(0),
	)
	defer func() { _ = closeFn() }()

	_, _ = bw.Write([]byte("x"))
	err := bw.Flush()
	assert.ErrorIs(t, err, writeErr)
}

// --- WriterFromIO passthrough ---

func TestBufferedWriterWriterPassthrough(t *testing.T) {
	rec := &recordWriter{}
	// rec already implements Writer — should be passed through, not double-wrapped.
	bw, closeFn := NewBufferedWriter(rec,
		WithBufSize(1024),
		WithSyncInterval(0),
	)
	defer func() { _ = closeFn() }()

	_, _ = bw.Write([]byte("data"))
	_ = bw.Sync()

	_, _, _, syncs := rec.snapshot()
	assert.Equal(t, 1, syncs) // Sync reached the real writer
}

// --- Option validation ---

func TestWithBufSizeZeroIgnored(t *testing.T) {
	rec := &recordWriter{}
	bw, closeFn := NewBufferedWriter(rec,
		WithBufSize(0), // should be ignored, keep default
		WithSyncInterval(0),
	)
	defer func() { _ = closeFn() }()

	// Write less than default (8KB) — should not flush.
	_, _ = bw.Write([]byte("small"))
	_, writes, _, _ := rec.snapshot()
	assert.Equal(t, 0, writes)
}
