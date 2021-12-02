package logf

import (
	"sync/atomic"
	"time"
)

// NewLogger returns a new Logger with a given Level and EntryWriter.
func NewLogger(level LevelCheckerGetter, w EntryWriter) *Logger {
	return &Logger{
		level: level.LevelChecker(),
		id:    atomic.AddInt32(&nextID, 1),
		w:     w,
	}
}

// NewDisabledLogger return a new Logger that logs nothing as fast as
// possible.
func NewDisabledLogger() *Logger {
	return NewLogger(
		LevelCheckerGetterFunc(func() LevelChecker {
			return func(Level) bool {
				return false
			}
		}), nil)
}

var defaultDisabledLogger = NewDisabledLogger()

// DisabledLogger returns a default instance of a Logger that logs nothing
// as fast as possible.
func DisabledLogger() *Logger {
	return defaultDisabledLogger
}

// Logger is the fast, asynchronous, structured logger.
//
// The Logger wraps EntryWriter to check logging level and provide a bit of
// syntactic sugar.
type Logger struct {
	level LevelChecker
	id    int32
	w     EntryWriter

	fields     []Field
	name       string
	addCaller  bool
	callerSkip int
}

// LogFunc allows to log a message with a bound level.
type LogFunc func(string, ...Field)

// AtLevel calls the given fn if logging a message at the specified level
// is enabled, passing a LogFunc with the bound level.
func (l *Logger) AtLevel(lvl Level, fn func(LogFunc)) {
	if !l.level(lvl) {
		return
	}

	fn(func(text string, fs ...Field) {
		l.write(lvl, text, fs)
	})
}

// WithLevel returns a new logger with the given additional level checker.
func (l *Logger) WithLevel(levelChecker LevelChecker) *Logger {
	newLevel := levelChecker //levelChecker.LevelChecker()

	cc := l.clone()
	cc.level = func(lvl Level) bool {
		return newLevel(lvl) && l.level(lvl)
	}

	return cc
}

// WithName returns a new Logger adding the given name to the calling one.
// Name separator is a period.
//
// Loggers have no name by default.
func (l *Logger) WithName(n string) *Logger {
	if n == "" {
		return l
	}

	cc := l.clone()
	if cc.name == "" {
		cc.name = n
	} else {
		cc.name += "." + n
	}

	return cc
}

// WithCaller returns a new Logger that adds a special annotation parameters
// to each logging message, such as the filename and line number of a caller.
func (l *Logger) WithCaller() *Logger {
	cc := l.clone()
	cc.addCaller = true

	return cc
}

// WithCallerSkip returns a new Logger with increased number of skipped
// frames. It's usable to build a custom wrapper for the Logger.
func (l *Logger) WithCallerSkip(skip int) *Logger {
	cc := l.clone()
	cc.callerSkip = skip

	return cc
}

// With returns a new Logger with the given additional fields.
func (l *Logger) With(fs ...Field) *Logger {
	// This code attempts to archive optimum performance with minimum
	// allocations count. Do not change it unless the following benchmarks
	// will show a better performance:
	// - BenchmarkAccumulateFields
	// - BenchmarkAccumulateFieldsWithAccumulatedFields

	var cc *Logger
	if len(l.fields) == 0 {
		// The fastest way. Use passed 'fs' as is.
		cc = l.clone()
		cc.fields = fs
	} else {
		// The less efficient path forces us to copy parent's fields.
		c := make([]Field, 0, len(l.fields)+len(fs))
		c = append(c, l.fields...)
		c = append(c, fs...)

		cc = l.clone()
		cc.fields = c
	}

	for i := range cc.fields[len(l.fields):] {
		snapshotField(&cc.fields[i])
	}

	return cc
}

// Debug logs a debug message with the given text, optional fields and
// fields passed to the Logger using With function.
func (l *Logger) Debug(text string, fs ...Field) {
	if !l.level(LevelDebug) {
		return
	}

	l.write(LevelDebug, text, fs)
}

// Info logs an info message with the given text, optional fields and
// fields passed to the Logger using With function.
func (l *Logger) Info(text string, fs ...Field) {
	if !l.level(LevelInfo) {
		return
	}

	l.write(LevelInfo, text, fs)
}

// Warn logs a warning message with the given text, optional fields and
// fields passed to the Logger using With function.
func (l *Logger) Warn(text string, fs ...Field) {
	if !l.level(LevelWarn) {
		return
	}

	l.write(LevelWarn, text, fs)
}

// Error logs an error message with the given text, optional fields and
// fields passed to the Logger using With function.
func (l *Logger) Error(text string, fs ...Field) {
	if !l.level(LevelError) {
		return
	}

	l.write(LevelError, text, fs)
}

func (l *Logger) write(lv Level, text string, fs []Field) {
	// Snapshot non-const fields.
	for i := range fs {
		snapshotField(&fs[i])
	}

	e := Entry{l.id, l.name, l.fields, fs, lv, time.Now(), text, EntryCaller{}}
	if l.addCaller {
		e.Caller = NewEntryCaller(2 + l.callerSkip)
	}

	l.w.WriteEntry(e)
}

func (l *Logger) clone() *Logger {
	// Field names should be omitted in order not to forget the new fields.
	return &Logger{
		l.level,
		atomic.AddInt32(&nextID, 1),
		l.w,
		l.fields,
		l.name,
		l.addCaller,
		l.callerSkip,
	}
}

var nextID int32
