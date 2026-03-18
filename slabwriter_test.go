package logf

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// blockingWriter blocks on every Write until ch is closed.
type blockingWriter struct {
	ch chan struct{}
}

func (w *blockingWriter) Write(p []byte) (int, error) {
	<-w.ch
	return len(p), nil
}

type collectWriter struct {
	mu     sync.Mutex
	writes []string
	flushN int
	syncN  int
}

func (w *collectWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.writes = append(w.writes, string(p))
	return len(p), nil
}

func (w *collectWriter) Flush() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.flushN++
	return nil
}

func (w *collectWriter) Sync() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.syncN++
	return nil
}

func (w *collectWriter) allData() string {
	w.mu.Lock()
	defer w.mu.Unlock()
	var s string
	for _, wr := range w.writes {
		s += wr
	}
	return s
}

func TestSlabBufferBasicWrite(t *testing.T) {
	cw := &collectWriter{}
	sb := NewSlabWriter(cw, 1024, 4)
	_, _ = sb.Write([]byte("hello"))
	_ = sb.Close()
	assert.Equal(t, "hello", cw.allData())
}

func TestSlabBufferMultipleWrites(t *testing.T) {
	cw := &collectWriter{}
	sb := NewSlabWriter(cw, 1024, 4)
	_, _ = sb.Write([]byte("aaa"))
	_, _ = sb.Write([]byte("bbb"))
	_, _ = sb.Write([]byte("ccc"))
	_ = sb.Close()
	assert.Equal(t, "aaabbbccc", cw.allData())
}

func TestSlabBufferSlabSwap(t *testing.T) {
	cw := &collectWriter{}
	// Tiny slabs: 8 bytes each, 4 slabs.
	sb := NewSlabWriter(cw, 8, 4)
	// Each write is 5 bytes → fits in one slab, next write triggers swap.
	_, _ = sb.Write([]byte("aaaaa"))
	_, _ = sb.Write([]byte("bbbbb"))
	_, _ = sb.Write([]byte("ccccc"))
	_ = sb.Close()
	assert.Equal(t, "aaaaabbbbbccccc", cw.allData())
}

func TestSlabBufferWriteLargerThanSlab(t *testing.T) {
	cw := &collectWriter{}
	// Slab is 8 bytes, message is 20 bytes → spans multiple slabs.
	sb := NewSlabWriter(cw, 8, 4)
	_, _ = sb.Write([]byte("12345678901234567890"))
	_ = sb.Close()
	assert.Equal(t, "12345678901234567890", cw.allData())
}

func TestSlabBufferRequestFlush(t *testing.T) {
	cw := &collectWriter{}
	sb := NewSlabWriter(cw, 1024, 4)
	_, _ = sb.Write([]byte("data"))
	_ = sb.Flush()
	_ = sb.Close()

	assert.Equal(t, "data", cw.allData())
	// requestFlush sends partial slab → at least one write before close.
	cw.mu.Lock()
	assert.GreaterOrEqual(t, len(cw.writes), 1)
	cw.mu.Unlock()
}

func TestSlabBufferCloseFlushesAndSyncs(t *testing.T) {
	cw := &collectWriter{}
	sb := NewSlabWriter(cw, 1024, 4)
	_, _ = sb.Write([]byte("x"))
	_ = sb.Close()

	cw.mu.Lock()
	defer cw.mu.Unlock()
	assert.Equal(t, 1, cw.flushN)
	assert.Equal(t, 1, cw.syncN)
}

func TestSlabBufferEmptyClose(t *testing.T) {
	cw := &collectWriter{}
	sb := NewSlabWriter(cw, 1024, 4)
	_ = sb.Close()

	cw.mu.Lock()
	defer cw.mu.Unlock()
	assert.Equal(t, 1, cw.flushN)
	assert.Equal(t, 1, cw.syncN)
	assert.Empty(t, cw.writes)
}

func TestSlabBufferOrderPreserved(t *testing.T) {
	cw := &collectWriter{}
	sb := NewSlabWriter(cw, 64, 4)
	for i := 0; i < 100; i++ {
		_, _ = sb.Write([]byte{byte('0' + i%10)})
	}
	_ = sb.Close()

	data := cw.allData()
	require.Len(t, data, 100)
	for i := 0; i < 100; i++ {
		assert.Equal(t, byte('0'+i%10), data[i], "mismatch at position %d", i)
	}
}

func TestSlabBufferHighThroughput(t *testing.T) {
	cw := &collectWriter{}
	sb := NewSlabWriter(cw, 4096, 8)
	msg := make([]byte, 200)
	for i := range msg {
		msg[i] = 'A'
	}
	const n = 10000
	for i := 0; i < n; i++ {
		_, _ = sb.Write(msg)
	}
	_ = sb.Close()

	assert.Len(t, cw.allData(), n*200)
}

func TestSlabBufferDefaultParams(t *testing.T) {
	cw := &collectWriter{}
	// Zero values → defaults applied.
	sb := NewSlabWriter(cw, 0, 0)
	_, _ = sb.Write([]byte("test"))
	_ = sb.Close()
	assert.Equal(t, "test", cw.allData())
}

func TestSlabWriterDropOnFull(t *testing.T) {
	// slowWriter simulates a destination that can't keep up.
	// slabCount=2: one current + one in pool → full channel capacity 2.
	// After filling both, the next swap must drop.
	sw := &slowWriter{delay: 50 * time.Millisecond}
	sb := NewSlabWriter(sw, 8, 2, WithDropOnFull())

	// Fill slab 1 (8 bytes) → swap: goes to full (ioLoop picks it up, blocks on slow write).
	_, _ = sb.Write([]byte("aaaaaaaa"))
	// Fill slab 2 (8 bytes) → swap: goes to full (ioLoop still busy).
	_, _ = sb.Write([]byte("bbbbbbbb"))
	// Fill current (which is the only slab left) → swap: full is at capacity → DROP.
	_, _ = sb.Write([]byte("cccccccc"))

	_ = sb.Close()

	assert.Greater(t, sb.Dropped(), int64(0), "expected some bytes to be dropped")
}

func TestSlabWriterDropOnFullCounter(t *testing.T) {
	// Block ioLoop completely by using a writer that never returns.
	blocked := make(chan struct{})
	bw := &blockingWriter{ch: blocked}
	sb := NewSlabWriter(bw, 16, 2, WithDropOnFull())

	// Write 1: fills slab → swap checks pool (1 slab) → got it, send to full. OK.
	// ioLoop picks from full, blocks on blockingWriter.Write.
	_, _ = sb.Write(make([]byte, 16))
	// Write 2: fills slab → swap checks pool (empty) → DROP (16 bytes).
	_, _ = sb.Write(make([]byte, 16))
	// Write 3: fills slab (no swap yet — swap fires when NEXT write sees avail==0).
	_, _ = sb.Write(make([]byte, 16))
	// Write 4: avail==0 → swap checks pool (empty) → DROP (16 bytes).
	_, _ = sb.Write(make([]byte, 16))

	assert.Equal(t, int64(2), sb.Dropped(), "expected 2 messages dropped")

	// Unblock ioLoop and close.
	close(blocked)
	_ = sb.Close()
}

func TestSlabWriterNoDropWhenKeepingUp(t *testing.T) {
	cw := &collectWriter{}
	sb := NewSlabWriter(cw, 1024, 4, WithDropOnFull())

	for i := 0; i < 100; i++ {
		_, _ = sb.Write([]byte("msg"))
	}
	_ = sb.Close()

	assert.Equal(t, int64(0), sb.Dropped())
	assert.Contains(t, cw.allData(), "msg")
}

func TestSlabWriterDropOnFullNonBlocking(t *testing.T) {
	// Verify Write never blocks even with a completely stuck destination.
	blocked := make(chan struct{})
	bw := &blockingWriter{ch: blocked}
	sb := NewSlabWriter(bw, 8, 2, WithDropOnFull())

	done := make(chan struct{})
	go func() {
		defer close(done)
		for i := 0; i < 1000; i++ {
			_, _ = sb.Write([]byte("12345678"))
		}
	}()

	select {
	case <-done:
		// Producer finished without blocking — correct.
	case <-time.After(1 * time.Second):
		t.Fatal("producer blocked for >1s — WithDropOnFull should prevent this")
	}

	assert.Greater(t, sb.Dropped(), int64(0))

	close(blocked)
	_ = sb.Close()
}

func TestSlabWriterFlushPartialNoDeadlock(t *testing.T) {
	// Regression: flushPartial previously used mu.Lock() which could deadlock
	// when a producer held mu (blocked on an empty slab pool inside swapSlab)
	// while ioLoop's idle timer fired and selected the timer case.
	//
	// With TryLock, flushPartial skips the flush when mu is contended,
	// which is correct: contended mu means data is flowing and idle flush
	// is unnecessary.
	//
	// Setup: slabCount=2 forces frequent pool exhaustion; flushInterval=1µs
	// makes the timer fire on nearly every ioLoop iteration.
	cw := &collectWriter{}
	sw := NewSlabWriter(cw, 16, 2, WithFlushInterval(time.Microsecond))

	done := make(chan struct{})
	go func() {
		defer close(done)
		data := make([]byte, 16)
		for i := 0; i < 10000; i++ {
			_, _ = sw.Write(data)
		}
	}()

	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("deadlock: producer blocked for >3s")
	}
	require.NoError(t, sw.Close())
}
