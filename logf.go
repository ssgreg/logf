package logf

import (
	"sync/atomic"
	"time"
)

var nextID int32

func NewDisabled() *Logger {
	return NewLogger(Discard, nil)
}

func NewLogger(level Level, w ChannelWriter) *Logger {
	return &Logger{
		level: level,
		ch:    w,
	}
}

type Logger struct {
	level  Level
	id     int32
	fields []Field
	ch     ChannelWriter

	name       string
	addCaller  bool
	callerSkip int
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
	if l.level < Debug {
		return
	}

	l.log(Debug, text, fs)
}

func (l *Logger) Info(text string, fs ...Field) {
	if l.level < Info {
		return
	}

	l.log(Info, text, fs)
}

func (l *Logger) log(lv Level, text string, fs []Field) {
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
		level:  l.level,
		ch:     l.ch,
		fields: nil,
		id:     atomic.AddInt32(&nextID, 1),
		name:   l.name,
	}
}
