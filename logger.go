package logf

import "time"

// Level TODO
type Level uint32

// Field TODO
type Field struct {
	Key   string
	Value interface{}
}

// Severity levels.
const (
	// Panic: system is unusable. Calls panic after handling.
	PanicLevel Level = iota
	// Fatal: critical conditions. Calls os.Exit(1) after handling.
	FatalLevel
	// Error: error conditions.
	ErrorLevel
	// Warning: warning conditions.
	WarnLevel
	// Informational: informational messages.
	InfoLevel
	// Debug: debug-level messages.
	DebugLevel
)

// Stringer
func (l Level) String() string {
	switch l {
	case DebugLevel:
		return "debug"
	case InfoLevel:
		return "info"
	case WarnLevel:
		return "warning"
	case ErrorLevel:
		return "error"
	case FatalLevel:
		return "fatal"
	case PanicLevel:
		return "panic"
	default:
		return "unknown"
	}
}

// StdLogger TODO
type StdLogger interface {
	Print(...interface{})
	Printf(string, ...interface{})
	Println(...interface{})

	Fatal(...interface{})
	Fatalf(string, ...interface{})
	Fatalln(...interface{})

	Panic(...interface{})
	Panicf(string, ...interface{})
	Panicln(...interface{})
}

// Logger TODO
type Logger interface {
	// StdLogger

	// Debug(...interface{})
	// Info(string)
	// Warn(...interface{})
	// Error(...interface{})

	// Debugf(string, ...interface{})
	// Infof(string, ...interface{})
	// Warnf(string, ...interface{})
	// Errorf(string, ...interface{})

	Fields() ([]Field, Logger)
	Level() Level
	Close()
}

// FieldLogger TODO
type FieldLogger interface {
	Logger
	// WithField(key string, v interface{}) FieldLogger
	// WithField2(k1 string, v1 interface{}, k2 string, v2 interface{}) FieldLogger
	// WithField10(k1 string, v1 interface{}, k2 string, v2 interface{}, k3 string, v3 interface{}, k4 string, v4 interface{}, k5 string, v5 interface{}, k6 string, v6 interface{}, k7 string, v7 interface{}, k8 string, v8 interface{}, k9 string, v9 interface{}, k10 string, v10 interface{}) FieldLogger

	// WithFields(field ...Field) FieldLogger
	// WithFields(fields Fields) FieldLogger
	// WithError(err error)

	// WithFields1(func() []Field) FieldLogger
	WithInt(key string, v int) FieldLogger
	WithFloat64(key string, v float64) FieldLogger
	WithAny(k string, v interface{}) FieldLogger
	WithStr(k string, v string) FieldLogger
	WithTime(k string, v time.Time) FieldLogger
	WithErr(v error) FieldLogger

	Info() FieldLogger
	Msg(string)
	Msgf(string, ...interface{})
	Logger() FieldLogger
}

// Formatter TODO
type Formatter interface {
	Format(*Buffer, *Entry) error
}

// Appender TODO
type Appender interface {
	Append(*Entry) error
	Close() error
	Flush() error
}
