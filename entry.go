package logf

import (
	"context"
	"sync"
	"time"
)

// Entry holds a single log message and fields.
type Entry struct {
	// LoggerBag holds logger-scoped fields (from Logger.With).
	LoggerBag *Bag

	// Bag holds request-scoped fields (from context via ContextWriter).
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

// EntryWriter is the interface that should do real logging stuff.
type EntryWriter interface {
	WriteEntry(context.Context, Entry) error
	Enabled(context.Context, Level) bool
}

// NewSyncWriter returns an EntryWriter that encodes and writes entries
// synchronously in the caller's goroutine. It is safe for concurrent use.
func NewSyncWriter(level Level, appender Appender) EntryWriter {
	return &syncWriter{level: level, appender: appender}
}

// NewUnbufferedEntryWriter is a deprecated alias for NewSyncWriter.
//
// Deprecated: Use NewSyncWriter instead.
var NewUnbufferedEntryWriter = NewSyncWriter

type syncWriter struct {
	mu       sync.Mutex
	level    Level
	appender Appender
}

func (w *syncWriter) WriteEntry(_ context.Context, entry Entry) error {
	w.mu.Lock()
	err := w.appender.Append(entry)
	w.mu.Unlock()
	return err
}

func (w *syncWriter) Enabled(_ context.Context, lvl Level) bool {
	return w.level.Enabled(lvl)
}

// nopWriter is an EntryWriter that discards everything.
// Used by NewDisabledLogger.
type nopWriter struct{}

func (nopWriter) WriteEntry(context.Context, Entry) error { return nil }

func (nopWriter) Enabled(context.Context, Level) bool { return false }
