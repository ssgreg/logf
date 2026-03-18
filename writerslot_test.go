package logf

import (
	"bytes"
	"errors"
	"sync"
	"testing"
)

func TestWriterSlotDropBeforeSet(t *testing.T) {
	slot := NewWriterSlot()
	n, err := slot.Write([]byte("dropped"))
	if err != nil {
		t.Fatal(err)
	}
	if n != 7 {
		t.Fatalf("got %d, want 7", n)
	}

	var buf bytes.Buffer
	slot.Set(&buf)

	_, _ = slot.Write([]byte("hello"))
	if buf.String() != "hello" {
		t.Fatalf("got %q, want %q", buf.String(), "hello")
	}
}

func TestWriterSlotBufferBeforeSet(t *testing.T) {
	slot := NewWriterSlot(WithSlotBuffer(1024))
	_, _ = slot.Write([]byte("early1\n"))
	_, _ = slot.Write([]byte("early2\n"))

	var buf bytes.Buffer
	slot.Set(&buf)

	_, _ = slot.Write([]byte("late\n"))

	want := "early1\nearly2\nlate\n"
	if buf.String() != want {
		t.Fatalf("got %q, want %q", buf.String(), want)
	}
}

func TestWriterSlotBufferOverflow(t *testing.T) {
	slot := NewWriterSlot(WithSlotBuffer(10))
	_, _ = slot.Write([]byte("12345"))    // 5 bytes, fits
	_, _ = slot.Write([]byte("67890"))    // 5 bytes, fits exactly
	_, _ = slot.Write([]byte("overflow")) // 8 bytes, doesn't fit — dropped entirely

	var buf bytes.Buffer
	slot.Set(&buf)
	_, _ = slot.Write([]byte("trigger")) // triggers flush of buffered data

	want := "1234567890trigger"
	if buf.String() != want {
		t.Fatalf("got %q, want %q", buf.String(), want)
	}
}

func TestWriterSlotSetSwap(t *testing.T) {
	slot := NewWriterSlot()

	var buf1, buf2 bytes.Buffer
	slot.Set(&buf1)
	_, _ = slot.Write([]byte("first"))

	slot.Set(&buf2)
	_, _ = slot.Write([]byte("second"))

	if buf1.String() != "first" {
		t.Fatalf("buf1: got %q, want %q", buf1.String(), "first")
	}
	if buf2.String() != "second" {
		t.Fatalf("buf2: got %q, want %q", buf2.String(), "second")
	}
}

func TestWriterSlotFlushSync(t *testing.T) {
	slot := NewWriterSlot()

	// Before Set — no-op.
	if err := slot.Flush(); err != nil {
		t.Fatal(err)
	}
	if err := slot.Sync(); err != nil {
		t.Fatal(err)
	}

	// After Set — delegates.
	w := &trackingWriter{}
	slot.Set(w)

	_ = slot.Flush()
	if !w.flushed {
		t.Error("Flush not delegated")
	}
	_ = slot.Sync()
	if !w.synced {
		t.Error("Sync not delegated")
	}
}

func TestWriterSlotWriteError(t *testing.T) {
	slot := NewWriterSlot()
	errBroken := errors.New("broken")
	slot.Set(&failWriter{err: errBroken})

	_, err := slot.Write([]byte("data"))
	if !errors.Is(err, errBroken) {
		t.Fatalf("got %v, want %v", err, errBroken)
	}
}

type safeBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (b *safeBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.Write(p)
}

func (b *safeBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.String()
}

func TestWriterSlotConcurrent(t *testing.T) {
	slot := NewWriterSlot(WithSlotBuffer(4096))
	var wg sync.WaitGroup

	// Writers before Set.
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				_, _ = slot.Write([]byte("x"))
			}
		}()
	}

	wg.Wait()

	buf := &safeBuffer{}
	slot.Set(buf)

	// Writers after Set.
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				_, _ = slot.Write([]byte("y"))
			}
		}()
	}
	wg.Wait()

	data := buf.String()
	yCount := 0
	for _, c := range data {
		if c == 'y' {
			yCount++
		}
	}
	if yCount != 1000 {
		t.Fatalf("got %d y's, want 1000", yCount)
	}
}

func TestWriterSlotConcurrentSetAndWrite(t *testing.T) {
	// Set is called while Write goroutines are active.
	// No panics, no data races. Some writes may be dropped (pre-Set).
	slot := NewWriterSlot(WithSlotBuffer(4096))
	buf := &safeBuffer{}
	var wg sync.WaitGroup

	// Start writers immediately.
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 1000; j++ {
				_, _ = slot.Write([]byte("w"))
			}
		}()
	}

	// Set while writers are running.
	slot.Set(buf)

	wg.Wait()

	// At least some writes should have reached the buffer.
	data := buf.String()
	if len(data) == 0 {
		t.Fatal("no writes reached the writer")
	}
	for _, c := range data {
		if c != 'w' && c != 'x' {
			t.Fatalf("unexpected byte %q in output", c)
		}
	}
}

// --- test helpers ---

type trackingWriter struct {
	flushed bool
	synced  bool
}

func (w *trackingWriter) Write(p []byte) (int, error) { return len(p), nil }
func (w *trackingWriter) Flush() error                { w.flushed = true; return nil }
func (w *trackingWriter) Sync() error                 { w.synced = true; return nil }

type failWriter struct{ err error }

func (w *failWriter) Write([]byte) (int, error) { return 0, w.err }
func (w *failWriter) Flush() error              { return w.err }
func (w *failWriter) Sync() error               { return w.err }
