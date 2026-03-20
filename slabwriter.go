package logf

import (
	"errors"
	"fmt"
	"io"
	"sync"
	"sync/atomic"
	"time"
)

const (
	defaultSlabSize  = 4 * 1024
	defaultSlabCount = 4
)

// SlabWriter is an async buffered writer that keeps your logging goroutines
// fast by decoupling them from slow I/O. It uses a pool of pre-allocated
// linear byte slabs — producers memcpy into a slab, and a background
// goroutine writes full slabs to the destination in big, efficient batches.
//
// Architecture:
//
//	N goroutines (producers)    background I/O goroutine
//	  Write(p)                   ┌──────────┐
//	    ↓ mu.Lock                │   pool   │ ←── recycle after write
//	    ↓ memcpy into slab   ←── └──────────┘
//	    ↓ slab full?
//	    ↓ yes → send slab ──→ full chan ──→ w.Write(slab) → destination
//	    ↓ grab fresh slab ←── pool
//	    ↓ mu.Unlock
//
// Each slab is a contiguous []byte. The I/O goroutine writes it in a
// single Write call — always linear, no wrap-around, no partial writes.
// After Write completes, the slab is returned to the pool for reuse.
//
// Capacity planning:
//
// The two parameters — slabSize and slabCount — control throughput and
// burst tolerance independently.
//
// slabSize determines the batch size per Write call. With a typical log
// message of 256 bytes and slabSize of 16 KB, each Write delivers
// 64 messages. The maximum sustained throughput is:
//
//	throughput = slabSize / (writeLatency × msgSize)
//
// For example, with 1 ms network latency and 256-byte messages:
//
//	16 KB slab:  64,000 msgs/sec
//	64 KB slab: 256,000 msgs/sec
//
// slabCount determines burst tolerance — how many slabs producers
// can fill while the I/O goroutine is blocked on a slow Write. This
// absorbs temporary latency spikes without dropping messages or
// blocking the consumer.
//
// During a latency spike the consumer keeps filling free slabs. The
// number of slabs acts as a time buffer:
//
//	burstTime = slabCount × slabSize / (msgRate × msgSize)
//
// For example, with 4 slabs × 4 KB (default) and 256-byte messages:
//
//	 1,000 msgs/sec:  absorbs a    ~64 ms spike
//	10,000 msgs/sec:  absorbs a    ~6 ms spike
//	50,000 msgs/sec:  absorbs a   ~1.2 ms spike
//
// More configurations (256-byte messages, 50,000 msgs/sec):
//
//	 4 slabs × 16 KB:  absorbs a   ~5 ms spike
//	 8 slabs × 16 KB:  absorbs a  ~10 ms spike
//	16 slabs × 16 KB:  absorbs a  ~20 ms spike
//	 8 slabs × 64 KB:  absorbs a  ~40 ms spike
//	16 slabs × 64 KB:  absorbs a  ~80 ms spike
//
// Memory cost is slabCount × slabSize, plus one extra slabSize when
// FlushInterval is enabled (reusable buffer for idle flush).
// Typical configurations:
//
//	 4 ×  4 KB =  16 KB  (default, lightweight)
//	 8 × 64 KB = 512 KB  (general purpose, good burst tolerance)
//	16 × 64 KB =   1 MB  (high throughput + long spike tolerance)
//
// When all slabs are in flight and the pool is empty, Write blocks
// until a slab is recycled. With DropOnFull, Write never blocks:
// if the I/O goroutine cannot keep up, the current slab's data is
// silently discarded and the slab is reused. Use Dropped to monitor
// the total number of messages lost.
//
// # Message integrity
//
// Write guarantees that each message is either fully delivered or
// fully dropped — never partially written (torn). This holds for
// messages of any size:
//
//   - len(p) <= slabSize: if the message does not fit in the remaining
//     slab space, an early swap is performed before writing. The message
//     always lands in a single slab. On drop the entire slab (including
//     the message) is discarded atomically.
//
//   - len(p) > slabSize: the message is allocated in a dedicated
//     oversized buffer and sent through the I/O goroutine as a single
//     write. The oversized buffer is discarded after write (not returned
//     to the pool). One allocation per oversized message.
//
// # Performance notes
//
// The early swap may leave unused space at the tail of a slab when a
// message does not fit in the remainder. With typical log messages
// (200–500 bytes) and slab sizes (16–64 KB), utilization stays above
// 99 %. Fragmentation becomes noticeable only when message size
// approaches slab size (e.g. msg/slab > 50 %), which is unusual for
// structured logging.
//
// Oversized messages (> slabSize) incur one heap allocation (make +
// copy) per write. This is acceptable because such messages are rare;
// typical log entries are 100–500 bytes while the default slab is 4 KB.
//
// # Concurrency
//
// Write and Flush are safe for concurrent use.
// Write and Flush must not be called after or concurrently with Close.
// Close itself is idempotent.
type SlabWriter struct {
	mu            sync.Mutex
	slabSize      int
	pool          chan []byte   // free slabs
	full          chan []byte   // filled slabs → I/O goroutine
	current       []byte        // active slab
	pos           int           // write position in current slab
	msgCount      int           // messages in current slab (for drop accounting)
	flushBuf      []byte        // reusable buffer for idle flush (ioLoop only)
	w             Writer        // destination
	errW          io.Writer     // destination for I/O error reports (nil = discard)
	errCount      int64         // consecutive write errors (ioLoop only, no mutex)
	flushInterval time.Duration // idle flush interval (0 = no idle flush)
	stop          chan struct{} // closed to signal shutdown
	done          chan struct{} // closed when I/O goroutine exits
	closeErr      error         // Flush/Sync errors from drain; set before done is closed
	closeOnce     sync.Once     // prevents double-close panic
	dropOnFull    bool          // if true, drop data instead of blocking on full pool
	dropped       int64         // total messages dropped (protected by mu)
	written       int64         // total messages accepted by Write (protected by mu)
	writeErrors   atomic.Int64  // total write errors (ioLoop only)
}

// SlabStats is a snapshot of SlabWriter runtime statistics. Pull it from
// Stats() and feed it to your metrics system to keep an eye on queue
// depth, drop rates, and write errors.
type SlabStats struct {
	QueuedSlabs int   // slabs waiting for I/O
	FreeSlabs   int   // slabs available in pool
	TotalSlabs  int   // total slab count
	Dropped     int64 // total messages dropped (dropOnFull mode)
	Written     int64 // total messages accepted by Write
	WriteErrors int64 // total write errors
}

// SlabWriterBuilder accumulates configuration for a SlabWriter. Create
// one with NewSlabWriter, set options via chained method calls, and
// finalize with Build.
type SlabWriterBuilder struct {
	w             io.Writer
	slabSize      int
	slabCount     int
	flushInterval time.Duration
	dropOnFull    bool
	errW          io.Writer
}

// NewSlabWriter returns a builder for a SlabWriter that will write to w.
// Call Build to create the SlabWriter. Default slab size is 4 KB and
// default slab count is 4.
func NewSlabWriter(w io.Writer) *SlabWriterBuilder {
	return &SlabWriterBuilder{
		w:         w,
		slabSize:  defaultSlabSize,
		slabCount: defaultSlabCount,
	}
}

// SlabSize sets the size of each slab buffer in bytes. Larger slabs
// mean fewer I/O calls but more memory. Default is 4 KB.
func (b *SlabWriterBuilder) SlabSize(n int) *SlabWriterBuilder {
	b.slabSize = n
	return b
}

// SlabCount sets the number of slab buffers in the pool. More slabs
// give better burst tolerance but use more memory. Default is 4.
func (b *SlabWriterBuilder) SlabCount(n int) *SlabWriterBuilder {
	b.slabCount = n
	return b
}

// FlushInterval sets how long the SlabWriter waits for new data
// before flushing a partial slab. Without this, a quiet period could
// leave recent log entries sitting in the buffer. Default is 0 (no
// idle flush — data only goes out when a slab fills up or Close is
// called).
func (b *SlabWriterBuilder) FlushInterval(d time.Duration) *SlabWriterBuilder {
	b.flushInterval = d
	return b
}

// DropOnFull makes Write non-blocking: if the I/O goroutine cannot
// keep up and all slabs are in flight, the current slab's data is
// silently dropped instead of blocking the caller. Use this when you
// would rather lose log messages than add latency to your hot path.
// Monitor dropped messages via Stats().Dropped.
func (b *SlabWriterBuilder) DropOnFull() *SlabWriterBuilder {
	b.dropOnFull = true
	return b
}

// ErrorWriter sets where I/O errors are reported. When the background
// goroutine fails to write a slab, it formats the error and writes it
// to w. By default errors are silently discarded — pass os.Stderr here
// if you want to know about write failures.
func (b *SlabWriterBuilder) ErrorWriter(w io.Writer) *SlabWriterBuilder {
	b.errW = w
	return b
}

// Build creates and starts the SlabWriter. You must call Close when you
// are done to flush remaining data and stop the background I/O
// goroutine — a defer sw.Close() right after creation is the way to go.
func (b *SlabWriterBuilder) Build() *SlabWriter {
	slabSize := b.slabSize
	slabCount := b.slabCount
	if slabSize <= 0 {
		slabSize = defaultSlabSize
	}
	if slabCount <= 0 {
		slabCount = defaultSlabCount
	}
	sb := &SlabWriter{
		slabSize:      slabSize,
		pool:          make(chan []byte, slabCount),
		full:          make(chan []byte, slabCount+1), // +1 for occasional oversized message
		w:             WriterFromIO(b.w),
		stop:          make(chan struct{}),
		done:          make(chan struct{}),
		flushInterval: b.flushInterval,
		dropOnFull:    b.dropOnFull,
		errW:          b.errW,
	}
	if sb.flushInterval > 0 {
		sb.flushBuf = make([]byte, slabSize)
	}
	for i := 0; i < slabCount; i++ {
		sb.pool <- make([]byte, slabSize)
	}
	sb.current = <-sb.pool
	sb.pos = 0
	go sb.ioLoop()
	return sb
}

// Write copies p into the current slab. Every message is guaranteed to be
// either fully written or fully dropped — never partially torn. If the
// message does not fit in the remaining slab space, an early swap puts it
// in a fresh slab. Messages larger than slabSize get a dedicated buffer.
//
// Write is safe for concurrent use. It must not be called after Close.
func (sb *SlabWriter) Write(p []byte) (int, error) {
	sb.mu.Lock()
	sb.written++
	if sb.dropOnFull {
		sb.msgCount++
	}

	avail := sb.slabSize - sb.pos
	if len(p) > avail && sb.pos > 0 {
		// Message doesn't fit in remaining space. Swap early to keep
		// it in one slab (atomic drop guarantee for msg <= slabSize).
		if !sb.swapSlab() {
			sb.mu.Unlock()
			return len(p), nil // whole message dropped
		}
	}

	// Oversized message: use a dedicated buffer to keep it in one piece.
	if len(p) > sb.slabSize {
		sb.writeOversized(p)
		sb.mu.Unlock()
		return len(p), nil
	}

	// After early swap, message is guaranteed to fit — direct copy.
	copy(sb.current[sb.pos:sb.pos+len(p)], p)
	sb.pos += len(p)
	sb.mu.Unlock()
	return len(p), nil
}

// Flush enqueues the current partial slab for writing by the background
// I/O goroutine. It returns immediately without waiting for the write to
// complete — if you need a durable flush, use Close instead. Must not be
// called after Close.
func (sb *SlabWriter) Flush() error {
	sb.mu.Lock()
	if sb.pos > 0 {
		sb.swapSlab()
	}
	sb.mu.Unlock()
	return nil
}

// Sync is a no-op on SlabWriter — the real Sync on the underlying writer
// happens during Close.
func (sb *SlabWriter) Sync() error {
	return nil
}

// Close flushes remaining data, drains the queue, stops the background
// I/O goroutine, and calls Flush + Sync on the underlying Writer. Safe
// to call multiple times — subsequent calls return the same error.
func (sb *SlabWriter) Close() error {
	sb.closeOnce.Do(func() {
		sb.mu.Lock()
		if sb.pos > 0 {
			// Non-blocking: current slab is not counted in full or pool,
			// so at least one slot in full is always free.
			sb.full <- sb.current[:sb.pos]
		}
		sb.mu.Unlock()
		close(sb.stop)
		<-sb.done
	})
	return sb.closeErr
}

// Stats returns a point-in-time snapshot of runtime statistics. Safe to
// call concurrently from a metrics scraper or health check endpoint.
func (sb *SlabWriter) Stats() SlabStats {
	sb.mu.Lock()
	written := sb.written
	dropped := sb.dropped
	sb.mu.Unlock()
	return SlabStats{
		QueuedSlabs: len(sb.full),
		FreeSlabs:   len(sb.pool),
		TotalSlabs:  cap(sb.full),
		Dropped:     dropped,
		Written:     written,
		WriteErrors: sb.writeErrors.Load(),
	}
}

// swapSlab sends the current slab (trimmed to pos) to the I/O goroutine
// and grabs a fresh slab from the pool. Returns true on success.
//
// In dropOnFull mode the pool is checked first (non-blocking). If a
// free slab is available, the full send is guaranteed non-blocking by
// the slab-count invariant (pool + full + processing + current = N).
// If the pool is empty, the data is dropped and the current slab is
// reused. Returns false so the caller can abort a multi-slab write.
func (sb *SlabWriter) swapSlab() bool {
	if sb.dropOnFull {
		select {
		case fresh := <-sb.pool:
			sb.full <- sb.current[:sb.pos]
			sb.current = fresh
		default:
			sb.dropped += int64(sb.msgCount)
			sb.pos = 0
			sb.msgCount = 0
			return false
		}
	} else {
		sb.full <- sb.current[:sb.pos]
		sb.current = <-sb.pool
	}
	sb.pos = 0
	sb.msgCount = 0
	return true
}

// writeOversized handles messages larger than slabSize. It flushes the
// current partial slab first (to preserve ordering), then writes the
// oversized message synchronously to the destination. This avoids
// breaking the slab pool invariant (fixed number of equal-sized slabs).
// Oversized messages are rare (anomaly for a logger), so the synchronous
// write is acceptable.
// Caller must hold mu.
func (sb *SlabWriter) writeOversized(p []byte) {
	// Flush current partial slab first to preserve order.
	if sb.pos > 0 {
		if !sb.swapSlab() {
			return // drop mode: pool empty, both current and oversized dropped
		}
	}

	// Send oversized message through the full channel. The extra
	// capacity (+1 in constructor) accommodates one oversized slab
	// beyond the normal pool. processSlab discards oversized slabs
	// instead of returning them to the pool.
	oversized := make([]byte, len(p))
	copy(oversized, p)
	if sb.dropOnFull {
		select {
		case sb.full <- oversized:
		default:
			sb.dropped += int64(1)
		}
	} else {
		sb.full <- oversized
	}
}

func (sb *SlabWriter) ioLoop() {
	defer close(sb.done)

	var timer *time.Timer
	var timerC <-chan time.Time
	if sb.flushInterval > 0 {
		timer = time.NewTimer(sb.flushInterval)
		timer.Stop()
		timerC = timer.C
	}

	for {
		// Fast path: non-blocking read.
		select {
		case slab := <-sb.full:
			sb.processSlab(slab)
			continue
		case <-sb.stop:
			sb.drain()
			return
		default:
		}

		// Idle — start flush timer if configured.
		if timer != nil {
			timer.Reset(sb.flushInterval)
		}

		select {
		case slab := <-sb.full:
			stopTimer(timer, timerC)
			sb.processSlab(slab)
		case <-timerC:
			sb.flushPartial()
		case <-sb.stop:
			stopTimer(timer, timerC)
			sb.drain()
			return
		}
	}
}

// stopTimer stops the timer and drains its channel if it already fired.
func stopTimer(t *time.Timer, ch <-chan time.Time) {
	if t != nil && !t.Stop() {
		select {
		case <-ch:
		default:
		}
	}
}

// processSlab writes the slab to the destination and recycles it.
// Oversized slabs (from writeOversized) are discarded — they are
// extra and must not be returned to the fixed-size pool.
func (sb *SlabWriter) processSlab(slab []byte) {
	sb.safeWrite(slab)
	if cap(slab) != sb.slabSize {
		return // oversized slab — discard, don't return to pool
	}
	sb.pool <- slab[:cap(slab)]
}

// safeWrite writes p to the destination, recovering from panics.
func (sb *SlabWriter) safeWrite(p []byte) {
	err := sb.tryWrite(p)
	if err != nil {
		sb.reportError(err)
	} else {
		sb.reportOK()
	}
}

func (sb *SlabWriter) tryWrite(p []byte) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic: %v", r)
		}
	}()
	_, err = sb.w.Write(p)
	return err
}

// flushPartial copies the current partial slab data and writes it.
// Called from ioLoop when the idle timer fires. Uses a dedicated
// flushBuf to avoid holding mu during I/O and to avoid channel ops
// on the pool (which would add latency to the next Write).
func (sb *SlabWriter) flushPartial() {
	if !sb.mu.TryLock() {
		return // producer active, data is flowing — no idle flush needed
	}
	if sb.pos == 0 {
		sb.mu.Unlock()
		return
	}
	n := copy(sb.flushBuf[:sb.pos], sb.current[:sb.pos])
	sb.pos = 0
	sb.msgCount = 0
	sb.mu.Unlock()
	// SAFETY: after Unlock, producers may immediately write into current
	// at pos 0, but flushBuf already holds a snapshot. flushBuf is owned
	// exclusively by ioLoop, so no concurrent access is possible.
	sb.safeWrite(sb.flushBuf[:n])
}

// reportError tracks consecutive write errors and reports transitions
// to the error writer: the first error in an episode and recovery.
// Called only from ioLoop (single goroutine, no mutex needed for errCount).
func (sb *SlabWriter) reportError(err error) {
	sb.errCount++
	sb.writeErrors.Add(1)
	if sb.errCount == 1 && sb.errW != nil {
		sb.safeReport("logf: SlabWriter: %v\n", err)
	}
}

// reportOK resets the error counter and reports recovery if there were errors.
func (sb *SlabWriter) reportOK() {
	if sb.errCount == 0 {
		return
	}
	if sb.errW != nil {
		if sb.errCount > 1 {
			sb.safeReport("logf: SlabWriter: recovered after %d errors\n", sb.errCount)
		} else {
			sb.safeReport("logf: SlabWriter: recovered\n")
		}
	}
	sb.errCount = 0
}

// safeReport writes to errW, recovering from panics to avoid crashing ioLoop.
func (sb *SlabWriter) safeReport(format string, args ...any) {
	defer func() { _ = recover() }()
	fmt.Fprintf(sb.errW, format, args...)
}

func (sb *SlabWriter) drain() {
	for {
		select {
		case slab := <-sb.full:
			sb.processSlab(slab)
		default:
			sb.closeErr = errors.Join(sb.closeErr, sb.w.Flush())
			sb.closeErr = errors.Join(sb.closeErr, sb.w.Sync())
			return
		}
	}
}
