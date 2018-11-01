package logf

import (
	"io"
)

// Appender defines the interface for your own strategies for outputting
// log entries.
type Appender interface {
	io.Closer

	// Append logs the Entry in Appender specific way.
	Append(Entry) error

	// Flush writes uncommitted changes to the underlying buffer of Writer.
	Flush() error

	// Sync writes uncommitted changes to the stable storage. For files this
	// means flushing the file system's in-memory copy of recently written
	// data to disk.
	Sync() error
}

// NewDiscardAppender creates the new instance of an appender that does
// nothing.
func NewDiscardAppender() Appender {
	return &discardAppender{}
}

type discardAppender struct {
}

func (a *discardAppender) Append(Entry) error {
	return nil
}

func (a *discardAppender) Sync() (err error) {
	return nil
}

func (a *discardAppender) Flush() error {
	return nil
}

func (a *discardAppender) Close() error {
	return nil
}

func NewWriteAppender(w io.Writer, enc Encoder) Appender {
	s, _ := w.(syncer)

	return &writeAppender{
		w:   w,
		s:   s,
		enc: enc,
		buf: NewBufferWithCapacity(PageSize * 2),
	}
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
	if a.buf.Len() > PageSize {
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
	if a.buf.Len() != 0 {
		defer a.buf.Reset()
		_, err := a.w.Write(a.buf.Bytes())

		return err
	}

	return nil
}

func (a *writeAppender) Close() error {
	return nil
}
