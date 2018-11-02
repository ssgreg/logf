package logf

import (
	"sync/atomic"
	"time"
)

type EntryWriter interface {
	WriteEntry(Entry)
}

func NewLogger(level LevelCheckerGetter, w EntryWriter) *Logger {
	return &Logger{
		level: level.LevelChecker(),
		id:    atomic.AddInt32(&nextID, 1),
		w:     w,
	}
}

func NewDisabledLogger() *Logger {
	return NewLogger(
		LevelCheckerGetterFunc(func() LevelChecker {
			return func(Level) bool {
				return false
			}
		}), nil)
}

type LogFunc func(string, ...Field)

type Logger struct {
	level LevelChecker
	id    int32
	w     EntryWriter

	fields     []Field
	name       string
	addCaller  bool
	callerSkip int
}

func (l *Logger) AtLevel(lvl Level, fn func(LogFunc)) {
	if !l.level(lvl) {
		return
	}

	fn(func(text string, fs ...Field) {
		l.write(lvl, text, fs)
	})
}

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

func (l *Logger) WithCaller() *Logger {
	cc := l.clone()
	cc.addCaller = true

	return cc
}

func (l *Logger) WithCallerSkip(skip int) *Logger {
	cc := l.clone()
	cc.callerSkip = skip

	return cc
}

func (l *Logger) With(fs ...Field) *Logger {
	// This code attempts to archive optimum performance with minimum
	// allocations count. Do not change it unless the folowing benchmarks
	// will show a better performance:
	// - BenchmarkCreateContextLogger
	// - BenchmarkCreateContextWithAccumulatedContextLogger

	var cc *Logger
	if len(l.fields) == 0 {
		// The fastest way. Use passed 'fs' as is.
		cc = l.clone()
		cc.fields = fs
	} else {
		// The less efficient path forces us to copy parent's fields.
		c := make([]Field, 0, len(l.fields)+len(fs))
		// TODO: change order
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

func (l *Logger) Debug(text string, fs ...Field) {
	if !l.level(LevelDebug) {
		return
	}

	l.write(LevelDebug, text, fs)
}

func (l *Logger) Info(text string, fs ...Field) {
	if !l.level(LevelInfo) {
		return
	}

	l.write(LevelInfo, text, fs)
}

func (l *Logger) Warn(text string, fs ...Field) {
	if !l.level(LevelWarn) {
		return
	}

	l.write(LevelWarn, text, fs)
}

func (l *Logger) Error(text string, fs ...Field) {
	if !l.level(LevelError) {
		return
	}

	l.write(LevelError, text, fs)
}

func (l *Logger) write(lv Level, text string, fs []Field) {
	for i := range fs {
		snapshotField(&fs[i])
	}

	e := Entry{l.id, l.fields, fs, lv, time.Now(), text, l.name, EntryCaller{}}
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
