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

// type StdLogger interface {
// 	Print(...interface{})
// 	Printf(string, ...interface{})
// 	Println(...interface{})

// 	Fatal(...interface{})
// 	Fatalf(string, ...interface{})
// 	Fatalln(...interface{})

// 	Panic(...interface{})
// 	Panicf(string, ...interface{})
// 	Panicln(...interface{})
// }

// type FieldGetter interface {
// 	Fields() ([]Field, FieldGetter)
// }

// type Logger interface {
// 	FieldGetter

// 	On(Level) EntryLogger

// 	New() Context
// }

// type Entry interface {
// 	FieldGetter

// 	Timestamp() time.Time
// 	Level() Level
// 	Format() string
// 	Args() []interface{}
// }

// type Context interface {
// 	Field(k string, v interface{}) Context
// 	Error(v error) Context

// 	Logger() Logger
// }

// type EntryLogger interface {
// 	Field(k string, v interface{}) EntryLogger
// 	Error(v error) EntryLogger

// 	Log(string, ...interface{})
// }

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
