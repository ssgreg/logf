package logf

import (
	"context"
	"io"
	"time"
)

// Entry holds a single log message and fields.
type Entry struct {
	// LoggerBag holds logger-scoped fields (from Logger.With).
	LoggerBag *Bag

	// Bag holds request-scoped fields (from context via ContextHandler).
	Bag *Bag

	// Fields specifies data fields of a log message.
	Fields []Field

	// Level specifies a severity level of a log message.
	Level Level

	// Time specifies a timestamp of a log message.
	Time time.Time

	// LoggerName specifies a non-unique name of a logger.
	// Can be empty.
	LoggerName string

	// Text specifies a text message of a log message.
	Text string

	// CallerPC is the program counter of the caller.
	// Zero means caller info is not available.
	CallerPC uintptr
}

// Handler is the interface that should do real logging stuff.
type Handler interface {
	Handle(context.Context, Entry) error
	Enabled(context.Context, Level) bool
}

// NewSyncHandler returns a Handler that encodes entries in the calling
// goroutine. Encoding is fully parallel across goroutines — the Encoder
// handles internal cloning and buffer pooling. The provided io.Writer
// must be safe for concurrent use.
//
// This is the thinnest possible Handler — no routing, no buffering,
// no background goroutines. Useful for benchmarks where every
// nanosecond matters and for simple single-destination setups.
func NewSyncHandler(level Level, w io.Writer, enc Encoder) Handler {
	return &syncHandler{level: level, w: w, enc: enc}
}

type syncHandler struct {
	level Level
	w     io.Writer
	enc   Encoder
}

func (h *syncHandler) Handle(_ context.Context, entry Entry) error {
	buf, err := h.enc.Encode(entry)
	if err != nil {
		return err
	}
	_, err = h.w.Write(buf.Bytes())
	buf.Free()
	return err
}

func (h *syncHandler) Enabled(_ context.Context, lvl Level) bool {
	return h.level.Enabled(lvl)
}

// nopHandler is an Handler that discards everything.
// Used by DisabledLogger.
type nopHandler struct{}

func (nopHandler) Handle(context.Context, Entry) error {
	return nil
}

func (nopHandler) Enabled(context.Context, Level) bool {
	return false
}
