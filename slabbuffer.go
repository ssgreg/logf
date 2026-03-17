package logf

import (
	"errors"
	"fmt"
	"io"
	"sync"
	"sync/atomic"
	"time"
)

// ErrClosed is returned by Write and Flush after Close has been called.
var ErrClosed = errors.New("logf: SlabWriter is closed")

const (
	defaultSlabSize  = 4 * 1024
	defaultSlabCount = 4
)

// SlabWriter is an async buffered writer that decouples the caller
// from I/O by using a pool of pre-allocated linear byte slabs.
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
// WithFlushInterval is enabled (reusable buffer for idle flush).
// Typical configurations:
//
//	 4 ×  4 KB =  16 KB  (default, lightweight)
//	 8 × 64 KB = 512 KB  (general purpose, good burst tolerance)
//	16 × 64 KB =   1 MB  (high throughput + long spike tolerance)
//
// When all slabs are in flight and the pool is empty, Write blocks
// until a slab is recycled.
//
// Concurrency: Write and Flush are safe for concurrent use.
// Close may be called concurrently with Write; after Close returns,
// Write and Flush return ErrClosed. Close itself is idempotent.
type SlabWriter struct {
	mu            sync.Mutex
	slabSize      int
	pool          chan []byte   // free slabs
	full          chan []byte   // filled slabs → I/O goroutine
	current       []byte        // active slab
	pos           int           // write position in current slab
	flushBuf      []byte        // reusable buffer for idle flush (ioLoop only)
	w             Writer        // destination
	errW          io.Writer     // destination for I/O error reports (nil = discard)
	errCount      int64         // consecutive write errors (ioLoop only, no mutex)
	flushInterval time.Duration // idle flush interval (0 = no idle flush)
	stop          chan struct{} // closed to signal shutdown
	done          chan struct{} // closed when I/O goroutine exits
	closeErr      error         // Flush/Sync errors from drain; set before done is closed
	closed        atomic.Bool   // set by Close, checked by Write/Flush
	closeOnce     sync.Once     // prevents double-close panic
}

// SlabOption configures a SlabWriter.
type SlabOption func(*SlabWriter)

// WithFlushInterval sets the idle flush interval. When no new data
// arrives for this duration, the partial slab is flushed to the
// destination. Default is 0 (no idle flush).
func WithFlushInterval(d time.Duration) SlabOption {
	return func(sw *SlabWriter) {
		sw.flushInterval = d
	}
}

// WithErrorWriter sets a writer for I/O error reports. When the
// background goroutine fails to write a slab to the destination, it
// formats the error and writes it to w. By default errors are
// silently discarded. Typical usage: WithErrorWriter(os.Stderr).
func WithErrorWriter(w io.Writer) SlabOption {
	return func(sw *SlabWriter) {
		sw.errW = w
	}
}

// NewSlabWriter creates a SlabWriter that buffers writes into pre-allocated
// slabs and flushes them to w via a background I/O goroutine. Close must
// be called to flush remaining data and stop the goroutine.
func NewSlabWriter(w io.Writer, slabSize, slabCount int, opts ...SlabOption) *SlabWriter {
	if slabSize <= 0 {
		slabSize = defaultSlabSize
	}
	if slabCount <= 0 {
		slabCount = defaultSlabCount
	}
	sb := &SlabWriter{
		slabSize: slabSize,
		pool:     make(chan []byte, slabCount),
		full:     make(chan []byte, slabCount),
		w:        WriterFromIO(w),
		stop:     make(chan struct{}),
		done:     make(chan struct{}),
	}
	for _, opt := range opts {
		opt(sb)
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

// Write copies p into the current slab. When a slab fills up it is
// sent to the I/O goroutine and a fresh slab is grabbed from the pool.
// Write is safe for concurrent use; it returns ErrClosed after Close.
func (sb *SlabWriter) Write(p []byte) (int, error) {
	if sb.closed.Load() {
		return 0, ErrClosed
	}
	sb.mu.Lock()
	written := 0
	for written < len(p) {
		avail := sb.slabSize - sb.pos
		if avail == 0 {
			sb.swapSlab()
			avail = sb.slabSize
		}
		n := len(p) - written
		if n > avail {
			n = avail
		}
		copy(sb.current[sb.pos:sb.pos+n], p[written:written+n])
		sb.pos += n
		written += n
	}
	sb.mu.Unlock()
	return len(p), nil
}

// Flush sends the current partial slab to the I/O goroutine.
// It returns ErrClosed after Close.
func (sb *SlabWriter) Flush() error {
	if sb.closed.Load() {
		return ErrClosed
	}
	sb.mu.Lock()
	if sb.pos > 0 {
		sb.swapSlab()
	}
	sb.mu.Unlock()
	return nil
}

// Sync is a no-op. The underlying writer's Sync is called on Close.
func (sb *SlabWriter) Sync() error {
	return nil
}

// Close flushes remaining data, stops the I/O goroutine, and calls
// Flush and Sync on the underlying Writer. It is safe to call
// multiple times; subsequent calls return the same error.
func (sb *SlabWriter) Close() error {
	sb.closeOnce.Do(func() {
		sb.closed.Store(true)
		sb.mu.Lock()
		if sb.pos > 0 {
			sb.full <- sb.current[:sb.pos]
		}
		sb.mu.Unlock()
		close(sb.stop)
		<-sb.done
	})
	return sb.closeErr
}

// swapSlab sends the current slab (trimmed to pos) to the I/O goroutine
// and grabs a fresh slab from the pool.
func (sb *SlabWriter) swapSlab() {
	sb.full <- sb.current[:sb.pos]
	sb.current = <-sb.pool
	sb.pos = 0
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
			if timer != nil {
				timer.Stop()
			}
			sb.processSlab(slab)
		case <-timerC:
			sb.flushPartial()
		case <-sb.stop:
			if timer != nil {
				timer.Stop()
			}
			sb.drain()
			return
		}
	}
}

// processSlab writes the slab to the destination and recycles it.
func (sb *SlabWriter) processSlab(slab []byte) {
	if _, err := sb.w.Write(slab); err != nil {
		sb.reportError(err)
	} else {
		sb.reportOK()
	}
	sb.pool <- slab[:cap(slab)]
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
	sb.mu.Unlock()
	if _, err := sb.w.Write(sb.flushBuf[:n]); err != nil {
		sb.reportError(err)
	} else {
		sb.reportOK()
	}
}

// reportError tracks consecutive write errors and reports transitions
// to the error writer: the first error in an episode and recovery.
// Called only from ioLoop (single goroutine, no mutex needed for errCount).
func (sb *SlabWriter) reportError(err error) {
	sb.errCount++
	if sb.errCount == 1 && sb.errW != nil {
		fmt.Fprintf(sb.errW, "logf: SlabWriter: %v\n", err)
	}
}

// reportOK resets the error counter and reports recovery if there were errors.
func (sb *SlabWriter) reportOK() {
	if sb.errCount == 0 {
		return
	}
	if sb.errW != nil {
		if sb.errCount > 1 {
			fmt.Fprintf(sb.errW, "logf: SlabWriter: recovered after %d errors\n", sb.errCount)
		} else {
			fmt.Fprintf(sb.errW, "logf: SlabWriter: recovered\n")
		}
	}
	sb.errCount = 0
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
