package logf

import (
	"sync/atomic"
	"time"
)

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

func NewErrorEntry(text string) Entry {
	return Entry{
		LoggerID: -1,
		Level:    LevelError,
		Time:     time.Now(),
		Text:     text,
	}
}

func NewLogger(level LevelChecker, w ChannelWriter) *Logger {
	return &Logger{
		level: level,
		id:    atomic.AddInt32(&nextID, 1),
		ch:    w,
	}
}

func NewDisabled() *Logger {
	return NewLogger(
		func(Level) bool {
			return false
		}, nil)
}

type CheckedLogger struct {
	logger *Logger
	level  Level
}

type Logger struct {
	level LevelChecker
	id    int32
	ch    ChannelWriter

	fields     []Field
	name       string
	addCaller  bool
	callerSkip int
}

func (l CheckedLogger) Write(text string, fs ...Field) {
	l.logger.write(l.level, text, fs)
}

func (l *Logger) AtLevel(lvl Level, fn func(CheckedLogger)) {
	if !l.level(lvl) {
		return
	}

	fn(CheckedLogger{l, lvl})
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

func (l *Logger) Error(text string, fs ...Field) {
	if !l.level(LevelError) {
		return
	}

	l.write(LevelError, text, fs)
}

func (l *Logger) Warn(text string, fs ...Field) {
	if !l.level(LevelWarn) {
		return
	}

	l.write(LevelWarn, text, fs)
}

func (l *Logger) write(lv Level, text string, fs []Field) {
	for i := range fs {
		snapshotField(&fs[i])
	}

	e := Entry{l.id, l.fields, fs, lv, time.Now(), text, l.name, EntryCaller{}}
	if l.addCaller {
		e.Caller = NewEntryCaller(2 + l.callerSkip)
	}

	l.ch.Write(e)
}

func (l *Logger) clone() *Logger {
	return &Logger{
		level:      l.level,
		id:         atomic.AddInt32(&nextID, 1),
		ch:         l.ch,
		fields:     l.fields,
		name:       l.name,
		addCaller:  l.addCaller,
		callerSkip: l.callerSkip,
	}
}

var nextID int32
