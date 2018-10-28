package logf

import (
	"time"
)

// Severity levels.
const (
	Discard Level = iota
	// Error: error conditions.
	Error
	// Warning: warning conditions.
	Warn
	// Informational: informational messages.
	Info
	// Debug: debug- messages.
	Debug
)

type Level uint32

// Stringer
func (l Level) String() string {
	switch l {
	case Debug:
		return "debug"
	case Info:
		return "info"
	case Warn:
		return "warning"
	case Error:
		return "error"
	default:
		return "unknown"
	}
}

type Entry struct {
	LoggerID      int32
	DerivedFields []Field
	Fields        []Field
	Level         Level
	Time          time.Time
	Text          string
	LoggerName    string
	Caller        EntryCaller
}

// Formatter TODO
type Formatter interface {
	Format(*Buffer, Entry) error
}

// Appender TODO
type Appender interface {
	Append(Entry) error
	Close() error
	Flush() error
}
