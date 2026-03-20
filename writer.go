package logf

import (
	"io"
	"os"
	"syscall"
)

// Writer extends io.Writer with Flush and Sync — the two operations
// needed for reliable log delivery. Flush pushes buffered data to the
// underlying output, and Sync commits it to stable storage (think
// fsync). The Router calls Flush and Sync during its close sequence
// to make sure nothing is left in flight.
type Writer interface {
	io.Writer
	Flush() error
	Sync() error
}

// WriterFromIO upgrades a plain io.Writer to a Writer. If w already
// implements Writer, it is returned as-is — no wrapping overhead.
// Otherwise, the wrapper discovers Flush and Sync capabilities from
// the underlying type:
//   - Sync calls w.Sync() if available (e.g. *os.File)
//   - Flush calls w.Flush() if available (e.g. *bufio.Writer)
//   - Missing methods become no-ops
func WriterFromIO(w io.Writer) Writer {
	if sw, ok := w.(Writer); ok {
		return sw
	}
	return ioWriter{
		Writer:  w,
		flusher: asFlusher(w),
		syncer:  asSyncer(w),
	}
}

type flusher interface{ Flush() error }
type syncer interface{ Sync() error }

type ioWriter struct {
	io.Writer
	flusher flusher
	syncer  syncer
}

func (w ioWriter) Flush() error {
	if w.flusher != nil {
		return w.flusher.Flush()
	}
	return nil
}

func (w ioWriter) Sync() error {
	if w.syncer != nil {
		return w.syncer.Sync()
	}
	return nil
}

func asFlusher(w io.Writer) flusher {
	if f, ok := w.(flusher); ok {
		return f
	}
	return nil
}

func asSyncer(w io.Writer) syncer {
	s, ok := w.(syncer)
	if !ok {
		return nil
	}
	// Probe: if Sync fails with EINVAL or ENOTSUP the writer is bound
	// to a special file (pipe, socket) that doesn't support fsync.
	// Disable future syncs.
	if err := s.Sync(); err != nil {
		if pathErr, ok := err.(*os.PathError); ok {
			if errno, ok := pathErr.Err.(syscall.Errno); ok {
				if errno == syscall.EINVAL || errno == syscall.ENOTSUP {
					return nil
				}
			}
		}
	}
	return s
}
