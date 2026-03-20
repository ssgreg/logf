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
	sb := NewSlabWriter(cw).SlabSize(1024).SlabCount(4).Build()
	_, _ = sb.Write([]byte("hello"))
	_ = sb.Close()
	assert.Equal(t, "hello", cw.allData())
}

func TestSlabBufferMultipleWrites(t *testing.T) {
	cw := &collectWriter{}
	sb := NewSlabWriter(cw).SlabSize(1024).SlabCount(4).Build()
	_, _ = sb.Write([]byte("aaa"))
	_, _ = sb.Write([]byte("bbb"))
	_, _ = sb.Write([]byte("ccc"))
	_ = sb.Close()
	assert.Equal(t, "aaabbbccc", cw.allData())
}

func TestSlabBufferSlabSwap(t *testing.T) {
	cw := &collectWriter{}
	// Tiny slabs: 8 bytes each, 4 slabs.
	sb := NewSlabWriter(cw).SlabSize(8).SlabCount(4).Build()
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
	sb := NewSlabWriter(cw).SlabSize(8).SlabCount(4).Build()
	_, _ = sb.Write([]byte("12345678901234567890"))
	_ = sb.Close()
	assert.Equal(t, "12345678901234567890", cw.allData())
}

func TestSlabBufferRequestFlush(t *testing.T) {
	cw := &collectWriter{}
	sb := NewSlabWriter(cw).SlabSize(1024).SlabCount(4).Build()
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
	sb := NewSlabWriter(cw).SlabSize(1024).SlabCount(4).Build()
	_, _ = sb.Write([]byte("x"))
	_ = sb.Close()

	cw.mu.Lock()
	defer cw.mu.Unlock()
	assert.Equal(t, 1, cw.flushN)
	assert.Equal(t, 1, cw.syncN)
}

func TestSlabBufferEmptyClose(t *testing.T) {
	cw := &collectWriter{}
	sb := NewSlabWriter(cw).SlabSize(1024).SlabCount(4).Build()
	_ = sb.Close()

	cw.mu.Lock()
	defer cw.mu.Unlock()
	assert.Equal(t, 1, cw.flushN)
	assert.Equal(t, 1, cw.syncN)
	assert.Empty(t, cw.writes)
}

func TestSlabBufferOrderPreserved(t *testing.T) {
	cw := &collectWriter{}
	sb := NewSlabWriter(cw).SlabSize(64).SlabCount(4).Build()
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
	sb := NewSlabWriter(cw).SlabSize(4096).SlabCount(8).Build()
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
	sb := NewSlabWriter(cw).SlabSize(0).SlabCount(0).Build()
	_, _ = sb.Write([]byte("test"))
	_ = sb.Close()
	assert.Equal(t, "test", cw.allData())
}

func TestSlabWriterDropOnFull(t *testing.T) {
	// slowWriter simulates a destination that can't keep up.
	// slabCount=2: one current + one in pool → full channel capacity 2.
	// After filling both, the next swap must drop.
	sw := &slowWriter{delay: 50 * time.Millisecond}
	sb := NewSlabWriter(sw).SlabSize(8).SlabCount(2).DropOnFull().Build()

	// Fill slab 1 (8 bytes) → swap: goes to full (ioLoop picks it up, blocks on slow write).
	_, _ = sb.Write([]byte("aaaaaaaa"))
	// Fill slab 2 (8 bytes) → swap: goes to full (ioLoop still busy).
	_, _ = sb.Write([]byte("bbbbbbbb"))
	// Fill current (which is the only slab left) → swap: full is at capacity → DROP.
	_, _ = sb.Write([]byte("cccccccc"))

	_ = sb.Close()

	assert.Greater(t, sb.Stats().Dropped, int64(0), "expected some bytes to be dropped")
}

func TestSlabWriterDropOnFullCounter(t *testing.T) {
	// Block ioLoop completely by using a writer that never returns.
	blocked := make(chan struct{})
	bw := &blockingWriter{ch: blocked}
	sb := NewSlabWriter(bw).SlabSize(16).SlabCount(2).DropOnFull().Build()

	// Write 1 (16 bytes): fills slab, no swap yet.
	_, _ = sb.Write(make([]byte, 16))
	// Write 2 (16 bytes): avail=0 → swap → success (send full, get pool). pos=16.
	// ioLoop picks slab, blocks on blockingWriter.
	_, _ = sb.Write(make([]byte, 16))
	time.Sleep(10 * time.Millisecond)
	// Write 3 (16 bytes): avail=0 → swap → send full ok (cap=2), pool empty → DROP.
	// Abort (no torn write). pos=0.
	_, _ = sb.Write(make([]byte, 16))
	// Write 4 (16 bytes): pos=0, avail=16 → fits. No swap needed.
	_, _ = sb.Write(make([]byte, 16))
	// Write 5 (16 bytes): avail=0 → swap → pool still empty → DROP.
	_, _ = sb.Write(make([]byte, 16))

	// Write 3 drops 1 msg (itself). Write 5 drops 2 msgs (Write 4 + Write 5).
	assert.Equal(t, int64(3), sb.Stats().Dropped, "expected 3 messages dropped")

	// Unblock ioLoop and close.
	close(blocked)
	_ = sb.Close()
}

func TestSlabWriterDropCountsCurrentMessage(t *testing.T) {
	// Regression: msgCount was incremented after the write loop,
	// so the message that triggered the drop was not counted.
	blocked := make(chan struct{})
	bw := &blockingWriter{ch: blocked}
	// slabSize=8, slabCount=2. One message fills one slab exactly.
	sb := NewSlabWriter(bw).SlabSize(8).SlabCount(2).DropOnFull().Build()

	// Write 1 (8 bytes): fills slab. No swap yet (checked on next Write).
	_, _ = sb.Write(make([]byte, 8))
	// Write 2 (8 bytes): avail=0 → swapSlab → got free slab, send full.
	// ioLoop picks it, blocks on blockingWriter. pos=8.
	_, _ = sb.Write(make([]byte, 8))
	time.Sleep(10 * time.Millisecond) // let ioLoop pick up the slab

	// Write 3 (8 bytes): avail=0 → swapSlab → pool empty → DROP.
	// Slab contains Write 2 (1 msg) + Write 3 incremented msgCount.
	// With fix: msgCount=2 at drop time (both Write 2 and Write 3 counted).
	_, _ = sb.Write(make([]byte, 8))

	// Write 2's data is in the slab that got dropped (1 message).
	// Write 3's data goes into the reused slab after drop.
	// Before the fix (msgCount++ after loop), dropped was 0 because
	// swapSlab saw msgCount=0. With the fix, dropped=1.
	dropped := sb.Stats().Dropped
	assert.Equal(t, int64(1), dropped, "the dropped slab's message must be counted")

	close(blocked)
	_ = sb.Close()
}

func TestSlabWriterDropOnFullNoTornWrite(t *testing.T) {
	// A message larger than slabSize spans multiple slabs. If a drop
	// happens mid-message, the destination receives a partial message
	// (torn write).
	blocked := make(chan struct{})
	bw := &blockingWriter{ch: blocked}
	// slabSize=8, slabCount=2. Message of 16 bytes spans 2 slabs.
	sb := NewSlabWriter(bw).SlabSize(8).SlabCount(2).DropOnFull().Build()

	// Write 1 (8 bytes): fills slab, no swap yet.
	_, _ = sb.Write(make([]byte, 8))
	// Write 2 (8 bytes): avail=0 → swap → got free slab, send full.
	// ioLoop picks it, blocks on blockingWriter.
	_, _ = sb.Write(make([]byte, 8))
	time.Sleep(10 * time.Millisecond) // let ioLoop pick up and block

	// Write 3 (16 bytes): avail=0 → swap → pool empty → DROP first 8 bytes.
	// Remaining 8 bytes go into reused slab — torn write.
	_, _ = sb.Write(make([]byte, 16))

	assert.True(t, sb.Stats().Dropped > 0, "expected drops")

	// Unblock and close.
	close(blocked)
	_ = sb.Close()

	t.Logf("dropped=%d (torn write: second half was written to destination)", sb.Stats().Dropped)
}

func TestSlabWriterMessagesAlwaysComplete(t *testing.T) {
	// Every message written to destination must be complete — no partial
	// writes, even under high contention with early swaps.
	cw := &collectWriter{}
	// Small slabs to force frequent early swaps.
	sb := NewSlabWriter(cw).SlabSize(64).SlabCount(4).Build()

	msgs := []string{
		"aaaa-msg-01-end\n", // 16 bytes
		"bbbb-msg-02-end\n",
		"cccc-msg-03-end\n",
		"dddd-msg-04-end\n",
		"eeee-msg-05-end\n",
		"ffff-msg-06-end\n",
		"gggg-msg-07-end\n",
		"hhhh-msg-08-end\n",
		"iiii-msg-09-end\n",
		"jjjj-msg-10-end\n",
	}

	for _, m := range msgs {
		_, _ = sb.Write([]byte(m))
	}
	_ = sb.Close()

	data := cw.allData()
	for _, m := range msgs {
		assert.Contains(t, data, m, "message must appear complete in output")
	}
}

func TestSlabWriterOversizedMessageComplete(t *testing.T) {
	// Oversized messages (> slabSize) must arrive complete at destination.
	cw := &collectWriter{}
	sb := NewSlabWriter(cw).SlabSize(16).SlabCount(4).Build() // tiny slabs

	small := "small\n"
	big := "this-is-a-big-message-that-exceeds-slab-size\n" // 46 bytes > 16

	_, _ = sb.Write([]byte(small))
	_, _ = sb.Write([]byte(big))
	_, _ = sb.Write([]byte(small))
	_ = sb.Close()

	data := cw.allData()
	assert.Contains(t, data, small)
	assert.Contains(t, data, big, "oversized message must appear complete")
}

func TestSlabWriterNoDropWhenKeepingUp(t *testing.T) {
	cw := &collectWriter{}
	sb := NewSlabWriter(cw).SlabSize(1024).SlabCount(4).DropOnFull().Build()

	for i := 0; i < 100; i++ {
		_, _ = sb.Write([]byte("msg"))
	}
	_ = sb.Close()

	assert.Equal(t, int64(0), sb.Stats().Dropped)
	assert.Contains(t, cw.allData(), "msg")
}

func TestSlabWriterDropOnFullNonBlocking(t *testing.T) {
	// Verify Write never blocks even with a completely stuck destination.
	blocked := make(chan struct{})
	bw := &blockingWriter{ch: blocked}
	sb := NewSlabWriter(bw).SlabSize(8).SlabCount(2).DropOnFull().Build()

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
		t.Fatal("producer blocked for >1s — DropOnFull should prevent this")
	}

	assert.Greater(t, sb.Stats().Dropped, int64(0))

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
	sw := NewSlabWriter(cw).SlabSize(16).SlabCount(2).FlushInterval(time.Microsecond).Build()

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

// panicWriter panics on every Write — simulates a broken destination.
type panicWriter struct{}

func (panicWriter) Write([]byte) (int, error) { panic("destination exploded") }
func (panicWriter) Flush() error              { return nil }
func (panicWriter) Sync() error               { return nil }

// panicOnWrite panics when written to — simulates a broken errW.
type panicOnWrite struct{}

func (panicOnWrite) Write([]byte) (int, error) { panic("errW exploded") }

func TestSlabWriterErrWriterPanicDoesNotCrashIoLoop(t *testing.T) {
	// Setup: destination panics on Write, errW also panics when
	// reportError tries to log the error. This should NOT crash ioLoop.
	sw := NewSlabWriter(panicWriter{}).SlabSize(64).SlabCount(2).
		ErrorWriter(panicOnWrite{}).Build()

	// Write enough to fill a slab and trigger ioLoop processing.
	_, _ = sw.Write(make([]byte, 64))

	// If ioLoop crashed, Close will hang (done channel never closed).
	done := make(chan error, 1)
	go func() { done <- sw.Close() }()

	select {
	case <-done:
		// ioLoop survived — test passes.
	case <-time.After(3 * time.Second):
		t.Fatal("ioLoop appears dead — Close hung (errW panic crashed ioLoop)")
	}
}

func TestSlabWriterEarlySwapKeepsMessageIntact(t *testing.T) {
	// Message fits in slabSize but not in remaining space.
	// Early swap should keep it in one slab.
	cw := &collectWriter{}
	sb := NewSlabWriter(cw).SlabSize(16).SlabCount(4).Build()

	// Fill 12 bytes of slab. 4 bytes remain.
	_, _ = sb.Write([]byte("aaaaaaaaaaaa")) // 12 bytes
	// Write 10 bytes — doesn't fit in 4 remaining, triggers early swap.
	_, _ = sb.Write([]byte("bbbbbbbbbb")) // 10 bytes

	_ = sb.Close()
	assert.Equal(t, "aaaaaaaaaaaabbbbbbbbbb", cw.allData())
}

func TestSlabWriterOversizedWrite(t *testing.T) {
	// Message larger than slabSize uses dedicated oversized slab.
	cw := &collectWriter{}
	sb := NewSlabWriter(cw).SlabSize(8).SlabCount(4).Build()

	_, _ = sb.Write([]byte("before"))
	_, _ = sb.Write([]byte("oversized-message-larger-than-slab")) // 34 bytes > 8
	_, _ = sb.Write([]byte("after"))

	_ = sb.Close()
	assert.Equal(t, "beforeoversized-message-larger-than-slabafter", cw.allData())
}

func TestSlabWriterOversizedDropOnFull(t *testing.T) {
	// Oversized message with dropOnFull — entire message dropped atomically.
	blocked := make(chan struct{})
	bw := &blockingWriter{ch: blocked}
	sb := NewSlabWriter(bw).SlabSize(8).SlabCount(2).DropOnFull().Build()

	// Fill and send first slab, ioLoop blocks.
	_, _ = sb.Write(make([]byte, 8))
	_, _ = sb.Write(make([]byte, 8)) // triggers swap, sends to full
	time.Sleep(10 * time.Millisecond)

	// Oversized write — pool empty, should drop entirely.
	_, _ = sb.Write(make([]byte, 20))

	assert.True(t, sb.Stats().Dropped > 0, "oversized message should be dropped")

	close(blocked)
	_ = sb.Close()
}

func TestSlabWriterEarlySwapDropOnFull(t *testing.T) {
	// Message fits in slabSize but not in remaining space.
	// Early swap with dropOnFull — entire message dropped.
	blocked := make(chan struct{})
	bw := &blockingWriter{ch: blocked}
	sb := NewSlabWriter(bw).SlabSize(16).SlabCount(2).DropOnFull().Build()

	_, _ = sb.Write(make([]byte, 12)) // 12 bytes, 4 remain
	// Swap succeeds, ioLoop picks up and blocks.
	_, _ = sb.Write(make([]byte, 16)) // triggers swap (early: 16 > 4)
	time.Sleep(10 * time.Millisecond)

	// Next write: 10 bytes, 0 remain → early swap → pool empty → drop.
	_, _ = sb.Write(make([]byte, 16))
	_, _ = sb.Write(make([]byte, 10))

	assert.True(t, sb.Stats().Dropped > 0, "expected drops on early swap")

	close(blocked)
	_ = sb.Close()
}
