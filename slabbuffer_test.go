package logf

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
