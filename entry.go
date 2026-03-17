package logf

import (
	"context"
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

// Handler is the interface that should do real logging stuff.
type Handler interface {
	Handle(context.Context, Entry) error
	Enabled(context.Context, Level) bool
}

// nopHandler is an Handler that discards everything.
// Used by DisabledLogger.
type nopHandler struct{}

func (nopHandler) Handle(context.Context, Entry) error { return nil }

func (nopHandler) Enabled(context.Context, Level) bool { return false }
