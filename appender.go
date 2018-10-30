package logf

import (
	"io"
)

// Appender defines the interface for your own strategies for outputting
// log entries.
type Appender interface {
	// Append logs the Entry in Appender specific way.
	Append(Entry) error

	// Flush writes uncommitted changes to the underlying buffer of Writer.
	Flush() error

	// Sync writes uncommitted changes to the stable storage. For files this
	// means flushing the file system's in-memory copy of recently written
	// data to disk.
	Sync() error
}

func NewWriteAppender(w io.Writer, enc Encoder) Appender {
	s, _ := w.(syncer)

	return &writeAppender{w, s, enc, NewBuffer(4096 * 2)}
}

// syncer provides access the the Sync function of a Writer.
type syncer interface {
	Sync() error
}

type writeAppender struct {
	w   io.Writer
	s   syncer
	enc Encoder
	buf *Buffer
}

func (a *writeAppender) Append(entry Entry) error {
	err := a.enc.Encode(a.buf, entry)
	if err != nil {
		return err
	}
	// TODO: fix hardcode
	if len(a.buf.Buf) > 4096 {
		a.Flush()
	}

	return nil
}

func (a *writeAppender) Sync() (err error) {
	defer func() {
		if a.s != nil {
			syncErr := a.s.Sync()
			if syncErr != nil && err == nil {
				err = syncErr
			}
		}
	}()

	return a.Flush()
}

func (a *writeAppender) Flush() error {
	if len(a.buf.Buf) != 0 {
		defer a.buf.Reset()
		_, err := a.w.Write(a.buf.Buf)

		return err
	}

	return nil
}
