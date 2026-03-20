package logf

import (
	"context"
	"io"
	"time"
)

// Entry is a single log record — the thing that travels through the pipeline
// from Logger to Handler to Encoder. It carries the message, level, timestamp,
// caller info, and all accumulated fields (both from Logger.With and from
// context). You rarely create one yourself; the Logger builds it for you on
// every Debug/Info/Warn/Error call.
type Entry struct {
	// LoggerBag holds logger-scoped fields added via Logger.With. These are
	// typically service-level context like "component" or "version".
	LoggerBag *Bag

	// Bag holds request-scoped fields extracted from context by ContextHandler.
	// Think trace IDs, request metadata — anything you stuff into the context
	// via logf.With(ctx, ...).
	Bag *Bag

	// Fields are the per-call fields passed directly to Debug/Info/Warn/Error.
	Fields []Field

	// Level is the severity of this log record.
	Level Level

	// Time is when this log record was created (usually time.Now()).
	Time time.Time

	// LoggerName is the dot-separated name set via Logger.WithName.
	// Empty string means the logger has no name.
	LoggerName string

	// Text is the human-readable log message.
	Text string

	// CallerPC is the program counter of the call site. Zero means caller
	// reporting is disabled or unavailable.
	CallerPC uintptr
}

// Handler is the core interface that processes log entries. Implement it to
// control where and how logs are written. The built-in handlers — SyncHandler,
// ContextHandler, and Router — cover most use cases, but you can wrap or
// replace them for custom behavior like sampling, rate-limiting, or
// sending logs to an external service.
type Handler interface {
	Handle(context.Context, Entry) error
	Enabled(context.Context, Level) bool
}

// NewSyncHandler returns the simplest possible Handler — it encodes each
// entry right there in the calling goroutine and writes it immediately.
// No routing, no buffering, no background goroutines. Think of it as the
// "just write it" handler.
//
// Encoding is fully parallel across goroutines (the Encoder handles its
// own cloning and buffer pooling), but the provided io.Writer must be
// safe for concurrent use. Great for benchmarks, tests, and simple
// single-destination setups.
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
