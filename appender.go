package logf

import (
	"io"
	"os"
	"sync"
	"syscall"
)

// Appender is the interface for your own strategies for outputting log
// entries.
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

// NewDiscardAppender returns a new Appender that does nothing.
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

// NewWriteAppender returns a new Appender with the given Writer and Encoder.
func NewWriteAppender(w io.Writer, enc Encoder) Appender {
	s, _ := w.(syncer)

	if s != nil {
		err := s.Sync()
		// Check for EINVAL and ENOTSUP - known errors if Writer is bound to
		// a special File (e.g., a pipe or socket) which does not support
		// synchronization.
		if pathErr, ok := err.(*os.PathError); ok {
			if errno, ok := pathErr.Err.(syscall.Errno); ok {
				if errno == syscall.EINVAL || errno == syscall.ENOTSUP {
					// Disable future syncs.
					s = nil
				}
			}
		}
	}

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

func (a *writeAppender) Sync() error {
	if a.s == nil {
		return nil
	}

	return a.s.Sync()
}

func (a *writeAppender) Flush() error {
	if a.buf.Len() != 0 {
		defer a.buf.Reset()
		_, err := a.w.Write(a.buf.Bytes())

		return err
	}

	return nil
}

type compositeAppender struct {
	appenders []Appender
}

// NewCompositeAppender returns appender which can concurrently perform operations
// for several appenders.
func NewCompositeAppender(appenders ...Appender) Appender {
	return &compositeAppender{appenders}
}

func (ca *compositeAppender) Append(entry Entry) (lastErr error) {
	var wg sync.WaitGroup
	wg.Add(len(ca.appenders))

	for _, a := range ca.appenders {
		go func(a Appender) {
			defer wg.Done()

			err := a.Append(entry)
			if err != nil {
				lastErr = err
			}
		}(a)
	}

	wg.Wait()

	return lastErr
}

func (ca *compositeAppender) Sync() (lastErr error) {
	var wg sync.WaitGroup
	wg.Add(len(ca.appenders))

	for _, a := range ca.appenders {
		go func(a Appender) {
			defer wg.Done()

			err := a.Sync()
			if err != nil {
				lastErr = err
			}
		}(a)
	}

	wg.Wait()

	return lastErr
}

func (ca *compositeAppender) Flush() (lastErr error) {
	var wg sync.WaitGroup
	wg.Add(len(ca.appenders))

	for _, a := range ca.appenders {
		go func(a Appender) {
			defer wg.Done()

			err := a.Flush()
			if err != nil {
				lastErr = err
			}
		}(a)
	}

	wg.Wait()

	return lastErr
}
