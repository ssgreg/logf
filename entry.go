package logf

import (
	"context"
	"time"
)

// Entry holds a single log message and fields.
type Entry struct {
	// LoggerBag holds logger-scoped fields (from Logger.With).
	// Bag.Version() is used as a cache key by encoders.
	LoggerBag *Bag

	// Bag holds request-scoped fields (from context via ContextWriter).
	// Bag.Version() is used as a cache key by encoders.
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
	WriteEntry(context.Context, Entry)
	Enabled(context.Context, Level) bool
}

// NewUnbufferedEntryWriter returns an implementation of EntryWriter which
// puts entries directly to the appender immediately and synchronously.
func NewUnbufferedEntryWriter(level Level, appender Appender) EntryWriter {
	return unbufferedEntryWriter{level: level, appender: appender}
}

type unbufferedEntryWriter struct {
	level    Level
	appender Appender
}

func (w unbufferedEntryWriter) WriteEntry(_ context.Context, entry Entry) {
	_ = w.appender.Append(entry)
}

func (w unbufferedEntryWriter) Enabled(_ context.Context, lvl Level) bool {
	return w.level.Enabled(lvl)
}

// nopWriter is an EntryWriter that discards everything.
// Used by NewDisabledLogger.
type nopWriter struct{}

func (nopWriter) WriteEntry(context.Context, Entry) {}

func (nopWriter) Enabled(context.Context, Level) bool { return false }
