package logf

import (
	"io"
	"sync"
	"time"
)

const (
	defaultBufferSize   = 2 * PageSize
	defaultSyncInterval = 30 * time.Second
)

// NewBufferedWriter wraps an io.Writer into a Writer that buffers writes
// and flushes automatically when the buffer exceeds a size threshold or
// a periodic interval elapses. The returned close function syncs
// remaining data and stops the timer.
//
// The BufferedWriter is safe for concurrent use: an internal mutex
// protects the buffer from the periodic sync goroutine.
//
// Options:
//   - WithBufSize(n) — buffer capacity in bytes (default 8 KB)
//   - WithSyncInterval(d) — periodic sync interval (default 30s)
func NewBufferedWriter(w io.Writer, opts ...BufferedWriterOption) (Writer, func() error) {
	bw := &bufferedWriter{
		w:            WriterFromIO(w),
		bufSize:      defaultBufferSize,
		syncInterval: defaultSyncInterval,
	}
	for _, opt := range opts {
		opt(bw)
	}
	bw.buf = make([]byte, 0, bw.bufSize)
	if bw.clock == nil {
		bw.clock = realClock{}
	}
	bw.startTimer()
	return bw, bw.close
}

// BufferedWriterOption configures a BufferedWriter.
type BufferedWriterOption func(*bufferedWriter)

// WithBufSize sets the buffer capacity in bytes.
func WithBufSize(n int) BufferedWriterOption {
	return func(bw *bufferedWriter) {
		if n > 0 {
			bw.bufSize = n
		}
	}
}

// WithSyncInterval sets the periodic sync interval (flush + fsync).
// Zero disables periodic syncing.
func WithSyncInterval(d time.Duration) BufferedWriterOption {
	return func(bw *bufferedWriter) {
		bw.syncInterval = d
	}
}

// Clock controls time for BufferedWriter. Useful for testing.
type Clock interface {
	NewTicker(d time.Duration) *time.Ticker
}

type realClock struct{}

func (realClock) NewTicker(d time.Duration) *time.Ticker { return time.NewTicker(d) }

// WithClock sets the clock used for periodic sync. Intended for testing.
func WithClock(c Clock) BufferedWriterOption {
	return func(bw *bufferedWriter) {
		bw.clock = c
	}
}

type bufferedWriter struct {
	mu           sync.Mutex
	w            Writer
	buf          []byte
	bufSize      int
	syncInterval time.Duration
	clock        Clock
	ticker       *time.Ticker
	stopCh       chan struct{}
	done         chan struct{}
	stopped      sync.Once
}

func (bw *bufferedWriter) Write(p []byte) (int, error) {
	bw.mu.Lock()
	defer bw.mu.Unlock()
	// If p doesn't fit and buffer is not empty, flush first to avoid
	// unbounded growth (same logic as bufio.Writer).
	if len(p)+len(bw.buf) > bw.bufSize && len(bw.buf) > 0 {
		if err := bw.flushLocked(); err != nil {
			return 0, err
		}
	}
	bw.buf = append(bw.buf, p...)
	if len(bw.buf) >= bw.bufSize {
		return len(p), bw.flushLocked()
	}
	return len(p), nil
}

func (bw *bufferedWriter) Flush() error {
	bw.mu.Lock()
	defer bw.mu.Unlock()
	return bw.flushLocked()
}

func (bw *bufferedWriter) Sync() error {
	bw.mu.Lock()
	defer bw.mu.Unlock()
	return bw.syncLocked()
}

// flushLocked writes buffered data to the underlying writer.
// Caller must hold bw.mu.
func (bw *bufferedWriter) flushLocked() error {
	if len(bw.buf) == 0 {
		return nil
	}
	_, err := bw.w.Write(bw.buf)
	bw.buf = bw.buf[:0]
	return err
}

// syncLocked flushes buffered data and syncs the underlying writer.
// Caller must hold bw.mu.
func (bw *bufferedWriter) syncLocked() error {
	if err := bw.flushLocked(); err != nil {
		return err
	}
	return bw.w.Sync()
}

func (bw *bufferedWriter) startTimer() {
	if bw.syncInterval <= 0 {
		return
	}
	bw.stopCh = make(chan struct{})
	bw.done = make(chan struct{})
	bw.ticker = bw.clock.NewTicker(bw.syncInterval)
	go bw.timerLoop()
}

func (bw *bufferedWriter) timerLoop() {
	defer close(bw.done)
	for {
		select {
		case <-bw.ticker.C:
			bw.mu.Lock()
			_ = bw.syncLocked()
			bw.mu.Unlock()
		case <-bw.stopCh:
			return
		}
	}
}

func (bw *bufferedWriter) close() error {
	var err error
	bw.stopped.Do(func() {
		if bw.ticker != nil {
			bw.ticker.Stop()
			close(bw.stopCh)
			<-bw.done // wait for timerLoop to exit
		}
		bw.mu.Lock()
		err = bw.syncLocked()
		bw.mu.Unlock()
	})
	return err
}
