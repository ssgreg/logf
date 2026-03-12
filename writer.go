package logf

import (
	"context"
	"io"
	"runtime"
	"sync"
)

// NewWriter returns an EntryWriter that encodes entries in the calling
// goroutine. Encoding is fully parallel across goroutines — the Encoder
// handles internal cloning and buffer pooling. The provided io.Writer
// must be safe for concurrent use.
func NewWriter(level Level, w io.Writer, enc Encoder) EntryWriter {
	return &pooledWriter{level: level, w: w, enc: enc}
}

type pooledWriter struct {
	level Level
	w     io.Writer
	enc   Encoder
}

func (w *pooledWriter) WriteEntry(_ context.Context, entry Entry) error {
	buf, err := w.enc.Encode(entry)
	if err != nil {
		return err
	}
	_, err = w.w.Write(buf.Bytes())
	buf.Free()
	return err
}

func (w *pooledWriter) Enabled(_ context.Context, lvl Level) bool {
	return w.level.Enabled(lvl)
}

// NewAsyncWriter returns an EntryWriter that encodes entries in the calling
// goroutine, then sends the encoded bytes through a channel to a single
// consumer goroutine that performs the actual I/O.
// The returned function must be called to flush and close the writer.
func NewAsyncWriter(level Level, w io.Writer, enc Encoder) (EntryWriter, func()) {
	capacity := runtime.NumCPU() * 2
	aw := &asyncWriter{
		level: level,
		w:     w,
		enc:   enc,
		ch:    make(chan *Buffer, capacity),
		done:  make(chan struct{}),
	}
	go aw.consumer()
	return aw, aw.close
}

type asyncWriter struct {
	level  Level
	w      io.Writer
	enc    Encoder
	ch     chan *Buffer
	done   chan struct{}
	closed sync.Once
}

func (w *asyncWriter) WriteEntry(_ context.Context, entry Entry) error {
	buf, err := w.enc.Encode(entry)
	if err != nil {
		return err
	}
	w.ch <- buf
	return nil
}

func (w *asyncWriter) Enabled(_ context.Context, lvl Level) bool {
	return w.level.Enabled(lvl)
}

func (w *asyncWriter) consumer() {
	for buf := range w.ch {
		_, _ = w.w.Write(buf.Bytes())
		buf.Free()
	}
	close(w.done)
}

func (w *asyncWriter) close() {
	w.closed.Do(func() {
		close(w.ch)
		<-w.done
	})
}
