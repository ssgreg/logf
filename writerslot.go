package logf

import (
	"io"
	"sync"
	"sync/atomic"
)

// NewWriterSlot returns a new WriterSlot ready for use. Before Set is
// called, writes are either silently dropped or buffered (if you pass
// WithSlotBuffer).
func NewWriterSlot(opts ...WriterSlotOption) *WriterSlot {
	s := &WriterSlot{}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// WriterSlot is a placeholder Writer you can wire into a Logger now and
// connect to a real destination later via Set. This solves the
// chicken-and-egg problem where you need a Logger at startup but the
// actual output (file, network, etc.) is not ready yet.
//
// Before Set is called, writes are either dropped or buffered (if
// WithSlotBuffer was used). After Set, all writes go straight to the
// real writer with no extra overhead.
//
//	slot := logf.NewWriterSlot()
//	logger := logf.NewLogger().Output(slot).Build()
//	// ... later, when destination is ready:
//	slot.Set(file)
//
// WriterSlot is safe for concurrent Write/Flush/Sync calls.
// Set itself is NOT safe for concurrent calls — call it from a single
// goroutine.
type WriterSlot struct {
	w       atomic.Pointer[writerRef]
	flushed atomic.Bool
	mu      sync.Mutex // protects buf
	buf     []byte
	bufSize int
}

// writerRef wraps Writer to avoid double indirection through atomic.Pointer.
type writerRef struct{ w Writer }

// WriterSlotOption configures a WriterSlot at creation time.
type WriterSlotOption func(*WriterSlot)

// WithSlotBuffer enables buffering of early writes before Set is called,
// keeping up to size bytes in memory so you do not lose startup logs.
// Writes that do not fit entirely are dropped (no partial writes). The
// buffer is flushed to the real writer on the first Write after Set.
func WithSlotBuffer(size int) WriterSlotOption {
	return func(s *WriterSlot) {
		if size > 0 {
			s.bufSize = size
			s.buf = make([]byte, 0, size)
		}
	}
}

// Write writes p to the real writer if Set has been called. Before Set,
// data is buffered (if WithSlotBuffer was used) or silently dropped.
func (s *WriterSlot) Write(p []byte) (int, error) {
	if ref := s.w.Load(); ref != nil {
		if s.bufSize > 0 && !s.flushed.Load() {
			s.flushPending(ref.w)
		}
		return ref.w.Write(p)
	}

	// No writer yet — buffer or drop.
	if s.bufSize > 0 {
		s.mu.Lock()
		if s.buf != nil {
			if len(p) <= s.bufSize-len(s.buf) {
				s.buf = append(s.buf, p...)
			}
			// else: doesn't fit entirely — drop
		}
		s.mu.Unlock()
	}

	return len(p), nil
}

// flushPending writes the pre-Set buffer to w under the mutex.
// All concurrent Write calls block until the buffer is flushed,
// guaranteeing that buffered data appears before any post-Set writes.
// This happens at most once per WriterSlot lifetime.
func (s *WriterSlot) flushPending(w Writer) {
	s.mu.Lock()
	buf := s.buf
	if buf == nil {
		s.mu.Unlock()
		return
	}
	if len(buf) > 0 {
		_, _ = w.Write(buf)
		_ = w.Flush()
	}
	s.buf = nil
	s.flushed.Store(true)
	s.mu.Unlock()
}

// Flush delegates to the real writer's Flush. No-op before Set.
func (s *WriterSlot) Flush() error {
	if ref := s.w.Load(); ref != nil {
		return ref.w.Flush()
	}
	return nil
}

// Sync delegates to the real writer's Sync. No-op before Set.
func (s *WriterSlot) Sync() error {
	if ref := s.w.Load(); ref != nil {
		return ref.w.Sync()
	}
	return nil
}

// Set connects the slot to a real writer. Any buffered data will be
// flushed on the next Write call, preserving temporal ordering without
// blocking Set itself. The writer is automatically wrapped via
// WriterFromIO if needed.
//
// Set is NOT safe for concurrent calls — call it from a single goroutine.
func (s *WriterSlot) Set(w io.Writer) {
	prev := s.w.Swap(&writerRef{WriterFromIO(w)})
	if prev != nil {
		_ = prev.w.Sync()
	}
}
